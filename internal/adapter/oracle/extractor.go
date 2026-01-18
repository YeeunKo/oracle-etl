// Package oracle은 Oracle 데이터베이스 연결 및 데이터 추출 기능을 제공합니다.
package oracle

import (
	"context"
	"fmt"
	"sync"

	"oracle-etl/internal/domain"
)

// Extractor는 Oracle에서 데이터를 추출하는 인터페이스입니다
type Extractor interface {
	// ExtractTable은 단일 테이블의 데이터를 추출합니다
	ExtractTable(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, handler func(chunk *domain.ChunkResult) error) error

	// ExtractTables는 여러 테이블의 데이터를 병렬로 추출합니다
	ExtractTables(ctx context.Context, owner string, tableNames []string, opts domain.ExtractionOptions, parallelism int, handler func(chunk *domain.ChunkResult) error) error
}

// DataExtractor는 Extractor 인터페이스의 구현체입니다
type DataExtractor struct {
	repo Repository
}

// NewDataExtractor는 새로운 DataExtractor를 생성합니다
func NewDataExtractor(repo Repository) *DataExtractor {
	return &DataExtractor{
		repo: repo,
	}
}

// ExtractTable은 단일 테이블의 데이터를 추출합니다
func (e *DataExtractor) ExtractTable(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, handler func(chunk *domain.ChunkResult) error) error {
	// 기본값 적용
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = 10000
	}
	if opts.FetchArraySize <= 0 {
		opts.FetchArraySize = 1000
	}

	return e.repo.StreamTableData(ctx, owner, tableName, opts, handler)
}

// ExtractTables는 여러 테이블의 데이터를 병렬로 추출합니다
func (e *DataExtractor) ExtractTables(ctx context.Context, owner string, tableNames []string, opts domain.ExtractionOptions, parallelism int, handler func(chunk *domain.ChunkResult) error) error {
	if parallelism <= 0 {
		parallelism = 4
	}

	// 세마포어 역할을 하는 채널
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup

	// 에러 수집
	errCh := make(chan error, len(tableNames))

	// mutex for handler (thread-safe handler call)
	var mu sync.Mutex

	for _, tableName := range tableNames {
		wg.Add(1)
		go func(table string) {
			defer wg.Done()

			// 세마포어 획득
			sem <- struct{}{}
			defer func() { <-sem }()

			// 컨텍스트 취소 확인
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}

			// 테이블 추출
			err := e.repo.StreamTableData(ctx, owner, table, opts, func(chunk *domain.ChunkResult) error {
				mu.Lock()
				defer mu.Unlock()
				return handler(chunk)
			})
			if err != nil {
				errCh <- fmt.Errorf("테이블 %s 추출 실패: %w", table, err)
			}
		}(tableName)
	}

	// 모든 goroutine 완료 대기
	wg.Wait()
	close(errCh)

	// 에러 수집 및 반환
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d개 테이블 추출 실패: %v", len(errors), errors[0])
	}

	return nil
}

// ExtractionProgress는 추출 진행 상황을 나타냅니다
type ExtractionProgress struct {
	TableName       string  `json:"table_name"`
	RowsProcessed   int64   `json:"rows_processed"`
	RowsPerSecond   float64 `json:"rows_per_second"`
	ProgressPercent float64 `json:"progress_percent"`
	IsComplete      bool    `json:"is_complete"`
}

// ExtractionResult는 테이블 추출 완료 결과를 나타냅니다
type ExtractionResult struct {
	TableName        string `json:"table_name"`
	TotalRows        int64  `json:"total_rows"`
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationSeconds  float64 `json:"duration_seconds"`
	ChunksProcessed  int    `json:"chunks_processed"`
	Error            string `json:"error,omitempty"`
}
