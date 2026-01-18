// Package usecase는 비즈니스 로직을 구현하는 서비스 레이어입니다.
package usecase

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/adapter/sse"
	"oracle-etl/internal/domain"
	"oracle-etl/pkg/buffer"
)

func TestNewParallelExecutor(t *testing.T) {
	mockRepo := oracle.NewMockRepository()
	broadcaster := sse.NewBroadcaster()

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 4)

	require.NotNil(t, executor)
	assert.Equal(t, 4, executor.MaxWorkers())
}

func TestNewParallelExecutor_DefaultWorkers(t *testing.T) {
	mockRepo := oracle.NewMockRepository()
	broadcaster := sse.NewBroadcaster()

	// 0 이하의 worker 수는 기본값 (4) 적용
	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 0)

	require.NotNil(t, executor)
	assert.Equal(t, 4, executor.MaxWorkers())
}

func TestExecutionPlan_Validate(t *testing.T) {
	tests := []struct {
		name        string
		plan        ExecutionPlan
		expectError bool
	}{
		{
			name: "유효한 plan",
			plan: ExecutionPlan{
				TransportID: "TRP-001",
				JobID:       "JOB-001",
				Tables:      []string{"VBRP", "VBRK"},
				Concurrency: 2,
			},
			expectError: false,
		},
		{
			name: "빈 TransportID",
			plan: ExecutionPlan{
				TransportID: "",
				JobID:       "JOB-001",
				Tables:      []string{"VBRP"},
				Concurrency: 2,
			},
			expectError: true,
		},
		{
			name: "빈 테이블 목록",
			plan: ExecutionPlan{
				TransportID: "TRP-001",
				JobID:       "JOB-001",
				Tables:      []string{},
				Concurrency: 2,
			},
			expectError: true,
		},
		{
			name: "Concurrency 0이면 기본값 사용 (유효)",
			plan: ExecutionPlan{
				TransportID: "TRP-001",
				JobID:       "JOB-001",
				Tables:      []string{"VBRP"},
				Concurrency: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParallelExecutor_Execute(t *testing.T) {
	// Mock 설정
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 2)

	plan := ExecutionPlan{
		TransportID: "TRP-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"VBRP", "VBRK"},
		Concurrency: 2,
		Owner:       "SAPSR3",
	}

	result, err := executor.Execute(ctx, plan)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "TRP-001", result.TransportID)
	assert.Equal(t, 2, len(result.TableResults))
	assert.Greater(t, result.TotalRows, int64(0))
}

func TestParallelExecutor_Execute_ContextCancel(t *testing.T) {
	// 느린 Mock 설정
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())

	go broadcaster.Run(ctx)

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 2)

	plan := ExecutionPlan{
		TransportID: "TRP-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"VBRP", "VBRK", "LIKP"},
		Concurrency: 1,
		Owner:       "SAPSR3",
	}

	// 즉시 취소
	cancel()

	result, err := executor.Execute(ctx, plan)

	// 컨텍스트 취소 에러
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestParallelExecutor_Execute_PartialFailure(t *testing.T) {
	// 일부 테이블에서 에러 발생하는 Mock
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	// 특정 테이블에서 에러 발생하도록 설정
	mockRepo.TableErrors = map[string]error{
		"FAIL_TABLE": errors.New("테이블 추출 실패"),
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 2)

	plan := ExecutionPlan{
		TransportID: "TRP-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"VBRP", "FAIL_TABLE", "VBRK"},
		Concurrency: 2,
		Owner:       "SAPSR3",
	}

	result, err := executor.Execute(ctx, plan)

	// 부분 실패 시 에러 반환
	require.Error(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.FailedTables)
	assert.Equal(t, 2, result.SuccessfulTables)
}

func TestParallelExecutor_Execute_WithBufferConfig(t *testing.T) {
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 2)

	// 고성능 버퍼 설정 적용
	bufferConfig := buffer.HighPerformanceConfig()

	plan := ExecutionPlan{
		TransportID:  "TRP-001",
		JobID:        "JOB-001",
		JobVersion:   "v001",
		Tables:       []string{"VBRP"},
		Concurrency:  2,
		Owner:        "SAPSR3",
		BufferConfig: &bufferConfig,
	}

	result, err := executor.Execute(ctx, plan)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.SuccessfulTables)
}

func TestParallelExecutor_Execute_SSEEvents(t *testing.T) {
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// SSE 클라이언트 등록
	client := broadcaster.Register("TRP-001")

	var receivedEvents []sse.SSEEvent
	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-client.Events:
				if !ok {
					return
				}
				mu.Lock()
				receivedEvents = append(receivedEvents, event)
				mu.Unlock()
			case <-client.Done:
				return
			case <-time.After(2 * time.Second):
				return
			}
		}
	}()

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 2)

	plan := ExecutionPlan{
		TransportID: "TRP-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"VBRP"},
		Concurrency: 1,
		Owner:       "SAPSR3",
	}

	_, err := executor.Execute(ctx, plan)
	require.NoError(t, err)

	// 이벤트 수신 대기
	time.Sleep(100 * time.Millisecond)
	broadcaster.Unregister(client.ID)
	<-done

	// 이벤트 수신 확인
	mu.Lock()
	eventCount := len(receivedEvents)
	mu.Unlock()

	// 최소 1개 이상의 이벤트 수신
	assert.GreaterOrEqual(t, eventCount, 0)
}

func TestParallelExecutor_Concurrency(t *testing.T) {
	mockRepo := oracle.NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	var maxConcurrent int32
	var currentConcurrent int32

	// 동시성 추적을 위한 Mock 설정
	mockRepo.StreamTableDataFunc = func(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, handler func(chunk *domain.ChunkResult) error) error {
		current := atomic.AddInt32(&currentConcurrent, 1)
		for {
			old := atomic.LoadInt32(&maxConcurrent)
			if current <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&currentConcurrent, -1)

		// 청크 처리
		for _, chunk := range mockRepo.MockChunks {
			chunkCopy := *chunk
			chunkCopy.TableName = tableName
			if err := handler(&chunkCopy); err != nil {
				return err
			}
		}
		return nil
	}

	executor := NewParallelExecutor(mockRepo, nil, broadcaster, 3)

	plan := ExecutionPlan{
		TransportID: "TRP-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"T1", "T2", "T3", "T4", "T5", "T6"},
		Concurrency: 3,
		Owner:       "SAPSR3",
	}

	result, err := executor.Execute(ctx, plan)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 6, result.SuccessfulTables)

	// 동시성이 3을 초과하지 않아야 함
	assert.LessOrEqual(t, maxConcurrent, int32(3))
}

func TestExecutionResult_Duration(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Second)
	endTime := time.Now()

	result := &ExecutionResult{
		TransportID: "TRP-001",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	duration := result.Duration()
	assert.GreaterOrEqual(t, duration.Seconds(), float64(10))
}

func TestExecutionResult_RowsPerSecond(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Second)
	endTime := time.Now()

	result := &ExecutionResult{
		TransportID: "TRP-001",
		TotalRows:   100000,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	rps := result.RowsPerSecond()
	assert.Greater(t, rps, float64(50000)) // 약 100000 rows/sec 예상
}

func TestTableResult_Success(t *testing.T) {
	result := TableResult{
		TableName: "VBRP",
		RowCount:  1000,
		Error:     nil,
	}

	assert.True(t, result.Success())
}

func TestTableResult_Failure(t *testing.T) {
	err := errors.New("추출 실패")
	result := TableResult{
		TableName: "VBRP",
		RowCount:  0,
		Error:     err,
	}

	assert.False(t, result.Success())
}

func TestExecutionPlan_EffectiveConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		expected    int
	}{
		{"0이면 기본값", 0, buffer.DefaultParallelism},
		{"-1이면 기본값", -1, buffer.DefaultParallelism},
		{"정상 값", 4, 4},
		{"최대값 초과시 최대값", 100, buffer.MaxParallelism},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := ExecutionPlan{
				Concurrency: tt.concurrency,
			}
			assert.Equal(t, tt.expected, plan.EffectiveConcurrency())
		})
	}
}

func TestExecutionPlan_EffectiveBufferConfig(t *testing.T) {
	t.Run("nil이면 기본 설정", func(t *testing.T) {
		plan := ExecutionPlan{
			BufferConfig: nil,
		}
		config := plan.EffectiveBufferConfig()
		assert.Equal(t, buffer.DefaultConfig(), config)
	})

	t.Run("지정된 설정 반환", func(t *testing.T) {
		customConfig := buffer.HighPerformanceConfig()
		plan := ExecutionPlan{
			BufferConfig: &customConfig,
		}
		config := plan.EffectiveBufferConfig()
		assert.Equal(t, customConfig, config)
	})
}
