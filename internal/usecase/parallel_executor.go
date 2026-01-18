// Package usecase는 비즈니스 로직을 구현하는 서비스 레이어입니다.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"oracle-etl/internal/adapter/gcs"
	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/adapter/sse"
	"oracle-etl/internal/domain"
	"oracle-etl/pkg/buffer"
	"oracle-etl/pkg/pool"
)

// ExecutionPlan은 병렬 추출 실행 계획을 정의합니다
type ExecutionPlan struct {
	TransportID  string         // Transport ID
	JobID        string         // Job ID
	JobVersion   string         // Job 버전 (v001, v002, ...)
	Tables       []string       // 추출할 테이블 목록
	Concurrency  int            // 동시 실행 수 (0이면 기본값)
	Owner        string         // 스키마 소유자
	BufferConfig *buffer.Config // 버퍼 설정 (nil이면 기본값)
}

// Validate는 ExecutionPlan의 유효성을 검사합니다
func (p *ExecutionPlan) Validate() error {
	if p.TransportID == "" {
		return errors.New("TransportID는 필수입니다")
	}
	if len(p.Tables) == 0 {
		return errors.New("최소 1개 이상의 테이블이 필요합니다")
	}
	return nil
}

// EffectiveConcurrency는 실제 사용할 동시 실행 수를 반환합니다
func (p *ExecutionPlan) EffectiveConcurrency() int {
	if p.Concurrency <= 0 {
		return buffer.DefaultParallelism
	}
	if p.Concurrency > buffer.MaxParallelism {
		return buffer.MaxParallelism
	}
	return p.Concurrency
}

// EffectiveBufferConfig는 실제 사용할 버퍼 설정을 반환합니다
func (p *ExecutionPlan) EffectiveBufferConfig() buffer.Config {
	if p.BufferConfig != nil {
		return *p.BufferConfig
	}
	return buffer.DefaultConfig()
}

// TableResult는 단일 테이블 추출 결과입니다
type TableResult struct {
	TableName   string        // 테이블 이름
	RowCount    int64         // 처리된 row 수
	ByteCount   int64         // 전송된 바이트 수
	StartTime   time.Time     // 시작 시간
	EndTime     time.Time     // 종료 시간
	Duration    time.Duration // 소요 시간
	GCSPath     string        // GCS 경로
	Error       error         // 에러 (있는 경우)
}

// Success는 테이블 추출이 성공했는지 반환합니다
func (r TableResult) Success() bool {
	return r.Error == nil
}

// RowsPerSecond는 초당 처리 row 수를 반환합니다
func (r TableResult) RowsPerSecond() float64 {
	if r.Duration.Seconds() <= 0 {
		return 0
	}
	return float64(r.RowCount) / r.Duration.Seconds()
}

// ExecutionResult는 전체 실행 결과입니다
type ExecutionResult struct {
	TransportID      string        // Transport ID
	JobID            string        // Job ID
	JobVersion       string        // Job 버전
	TableResults     []TableResult // 테이블별 결과
	TotalRows        int64         // 총 처리 row 수
	TotalBytes       int64         // 총 전송 바이트 수
	SuccessfulTables int           // 성공한 테이블 수
	FailedTables     int           // 실패한 테이블 수
	StartTime        time.Time     // 시작 시간
	EndTime          time.Time     // 종료 시간
}

// Duration은 전체 소요 시간을 반환합니다
func (r *ExecutionResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// RowsPerSecond는 초당 처리 row 수를 반환합니다
func (r *ExecutionResult) RowsPerSecond() float64 {
	duration := r.Duration().Seconds()
	if duration <= 0 {
		return 0
	}
	return float64(r.TotalRows) / duration
}

// Success는 전체 실행이 성공했는지 반환합니다 (부분 실패 없음)
func (r *ExecutionResult) Success() bool {
	return r.FailedTables == 0
}

// ParallelExecutor는 다중 테이블 병렬 추출을 관리합니다
type ParallelExecutor struct {
	oracle      oracle.Repository
	gcs         gcs.Client
	sse         *sse.Broadcaster
	maxWorkers  int
}

// NewParallelExecutor는 새로운 ParallelExecutor를 생성합니다
func NewParallelExecutor(oracleRepo oracle.Repository, gcsClient gcs.Client, sseBroadcaster *sse.Broadcaster, maxWorkers int) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = buffer.DefaultParallelism
	}
	return &ParallelExecutor{
		oracle:     oracleRepo,
		gcs:        gcsClient,
		sse:        sseBroadcaster,
		maxWorkers: maxWorkers,
	}
}

// MaxWorkers는 최대 워커 수를 반환합니다
func (e *ParallelExecutor) MaxWorkers() int {
	return e.maxWorkers
}

// Execute는 실행 계획에 따라 병렬 추출을 수행합니다
func (e *ParallelExecutor) Execute(ctx context.Context, plan ExecutionPlan) (*ExecutionResult, error) {
	// 컨텍스트 취소 확인
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 유효성 검사
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("실행 계획 유효성 검사 실패: %w", err)
	}

	result := &ExecutionResult{
		TransportID:  plan.TransportID,
		JobID:        plan.JobID,
		JobVersion:   plan.JobVersion,
		TableResults: make([]TableResult, 0, len(plan.Tables)),
		StartTime:    time.Now(),
	}

	// 워커 풀 생성
	concurrency := plan.EffectiveConcurrency()
	if concurrency > e.maxWorkers {
		concurrency = e.maxWorkers
	}

	workerPool := pool.NewWorkerPool(concurrency)
	workerPool.Start(ctx)

	// 버퍼 설정
	bufferConfig := plan.EffectiveBufferConfig()

	// 결과 수집용 채널
	resultCh := make(chan TableResult, len(plan.Tables))
	var wg sync.WaitGroup

	// 작업 제출
	for _, tableName := range plan.Tables {
		table := tableName // 클로저용 복사
		wg.Add(1)

		workerPool.Submit(pool.Task{
			ID:        fmt.Sprintf("%s-%s", plan.JobID, table),
			TableName: table,
			Execute: func(taskCtx context.Context) error {
				defer wg.Done()
				
				tableResult := e.extractTable(taskCtx, plan, table, bufferConfig)
				resultCh <- tableResult
				
				// SSE 이벤트 발송
				if e.sse != nil {
					e.sendTableEvent(plan.TransportID, plan.JobID, tableResult)
				}
				
				return tableResult.Error
			},
		})
	}

	// 워커 풀 결과 수집 (별도 goroutine에서)
	go func() {
		workerPool.Wait()
		close(resultCh)
	}()

	// 결과 수집
	var totalRows, totalBytes int64
	for tableResult := range resultCh {
		result.TableResults = append(result.TableResults, tableResult)
		if tableResult.Success() {
			result.SuccessfulTables++
			atomic.AddInt64(&totalRows, tableResult.RowCount)
			atomic.AddInt64(&totalBytes, tableResult.ByteCount)
		} else {
			result.FailedTables++
		}
	}

	result.TotalRows = totalRows
	result.TotalBytes = totalBytes
	result.EndTime = time.Now()

	// 완료 이벤트 발송
	if e.sse != nil {
		e.sendCompleteEvent(result)
	}

	// 부분 실패 시 에러 반환
	if result.FailedTables > 0 {
		return result, fmt.Errorf("%d개 테이블 추출 실패", result.FailedTables)
	}

	return result, nil
}

// extractTable은 단일 테이블을 추출합니다
func (e *ParallelExecutor) extractTable(ctx context.Context, plan ExecutionPlan, tableName string, bufferConfig buffer.Config) TableResult {
	result := TableResult{
		TableName: tableName,
		StartTime: time.Now(),
	}

	// 추출 옵션 설정
	opts := domain.ExtractionOptions{
		ChunkSize:      bufferConfig.ChunkSize,
		FetchArraySize: bufferConfig.FetchArraySize,
	}

	var rowCount int64

	// 데이터 추출
	err := e.oracle.StreamTableData(ctx, plan.Owner, tableName, opts, func(chunk *domain.ChunkResult) error {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		atomic.AddInt64(&rowCount, int64(chunk.RowCount))

		// 진행률 이벤트 발송
		if e.sse != nil {
			e.sendProgressEvent(plan.TransportID, plan.JobID, tableName, atomic.LoadInt64(&rowCount))
		}

		return nil
	})

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.RowCount = rowCount

	if err != nil {
		result.Error = err
	}

	return result
}

// sendProgressEvent는 진행률 이벤트를 발송합니다
func (e *ParallelExecutor) sendProgressEvent(transportID, jobID, tableName string, rowsProcessed int64) {
	if e.sse == nil {
		return
	}

	e.sse.BroadcastProgress(sse.ProgressEvent{
		TransportID:   transportID,
		JobID:         jobID,
		Table:         tableName,
		RowsProcessed: rowsProcessed,
		RowsTotal:     -1, // 총 row 수 알 수 없음
	})
}

// sendTableEvent는 테이블 완료 이벤트를 발송합니다
func (e *ParallelExecutor) sendTableEvent(transportID, jobID string, result TableResult) {
	if e.sse == nil {
		return
	}

	if !result.Success() {
		// 에러 이벤트 발송
		e.sse.BroadcastError(sse.ErrorEvent{
			TransportID: transportID,
			JobID:       jobID,
			Table:       result.TableName,
			Code:        "EXTRACTION_ERROR",
			Message:     result.Error.Error(),
		})
		return
	}

	// 테이블 완료 이벤트 (상태 이벤트로 발송)
	e.sse.BroadcastStatus(sse.StatusEvent{
		TransportID: transportID,
		JobID:       jobID,
		Status:      sse.StatusCompleted,
		Message:     fmt.Sprintf("테이블 %s 추출 완료: %d rows, %dms", result.TableName, result.RowCount, result.Duration.Milliseconds()),
	})
}

// sendCompleteEvent는 완료 이벤트를 발송합니다
func (e *ParallelExecutor) sendCompleteEvent(result *ExecutionResult) {
	if e.sse == nil {
		return
	}

	e.sse.BroadcastComplete(sse.CompleteEvent{
		TransportID:   result.TransportID,
		JobID:         result.JobID,
		TotalRows:     result.TotalRows,
		TotalBytes:    result.TotalBytes,
		DurationMs:    result.Duration().Milliseconds(),
		TablesCount:   result.SuccessfulTables + result.FailedTables,
	})
}
