// Package benchmarks는 성능 벤치마크 테스트를 제공합니다.
package benchmarks

import (
	"context"
	"sync/atomic"
	"testing"

	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/adapter/sse"
	"oracle-etl/internal/domain"
	"oracle-etl/internal/usecase"
	"oracle-etl/pkg/buffer"
)

// BenchmarkParallelExecutor_SingleTable은 단일 테이블 추출 성능을 측정합니다
func BenchmarkParallelExecutor_SingleTable(b *testing.B) {
	mockRepo := createMockRepoWithRows(100000)
	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go broadcaster.Run(ctx)

	executor := usecase.NewParallelExecutor(mockRepo, nil, broadcaster, 4)

	plan := usecase.ExecutionPlan{
		TransportID: "BENCH-001",
		JobID:       "JOB-001",
		JobVersion:  "v001",
		Tables:      []string{"LARGE_TABLE"},
		Concurrency: 4,
		Owner:       "BENCH",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := executor.Execute(ctx, plan)
		if result != nil {
			b.ReportMetric(float64(result.TotalRows), "rows")
			b.ReportMetric(result.RowsPerSecond(), "rows/sec")
		}
	}
}

// BenchmarkParallelExecutor_MultipleTables은 다중 테이블 병렬 추출 성능을 측정합니다
func BenchmarkParallelExecutor_MultipleTables(b *testing.B) {
	tableCounts := []int{1, 2, 4, 8}

	for _, count := range tableCounts {
		b.Run(
			"tables="+itoa(count),
			func(b *testing.B) {
				mockRepo := createMockRepoWithRows(10000)
				broadcaster := sse.NewBroadcaster()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go broadcaster.Run(ctx)

				executor := usecase.NewParallelExecutor(mockRepo, nil, broadcaster, 4)

				tables := make([]string, count)
				for i := 0; i < count; i++ {
					tables[i] = "TABLE_" + itoa(i)
				}

				plan := usecase.ExecutionPlan{
					TransportID: "BENCH-001",
					JobID:       "JOB-001",
					JobVersion:  "v001",
					Tables:      tables,
					Concurrency: 4,
					Owner:       "BENCH",
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _ := executor.Execute(ctx, plan)
					if result != nil {
						b.ReportMetric(float64(result.TotalRows), "total_rows")
					}
				}
			},
		)
	}
}

// BenchmarkParallelExecutor_ConcurrencyScaling은 동시성 레벨에 따른 성능 확장성을 측정합니다
func BenchmarkParallelExecutor_ConcurrencyScaling(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8}
	tableCount := 8

	for _, conc := range concurrencyLevels {
		b.Run(
			"concurrency="+itoa(conc),
			func(b *testing.B) {
				mockRepo := createMockRepoWithRows(5000)
				broadcaster := sse.NewBroadcaster()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go broadcaster.Run(ctx)

				executor := usecase.NewParallelExecutor(mockRepo, nil, broadcaster, conc)

				tables := make([]string, tableCount)
				for i := 0; i < tableCount; i++ {
					tables[i] = "TABLE_" + itoa(i)
				}

				plan := usecase.ExecutionPlan{
					TransportID: "BENCH-001",
					JobID:       "JOB-001",
					JobVersion:  "v001",
					Tables:      tables,
					Concurrency: conc,
					Owner:       "BENCH",
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _ := executor.Execute(ctx, plan)
					if result != nil {
						b.ReportMetric(result.RowsPerSecond(), "rows/sec")
					}
				}
			},
		)
	}
}

// BenchmarkParallelExecutor_BufferConfigs는 다양한 버퍼 설정에서 성능을 비교합니다
func BenchmarkParallelExecutor_BufferConfigs(b *testing.B) {
	configs := map[string]buffer.Config{
		"Default":         buffer.DefaultConfig(),
		"HighPerformance": buffer.HighPerformanceConfig(),
		"LowMemory":       buffer.LowMemoryConfig(),
	}

	for name, cfg := range configs {
		b.Run(name, func(b *testing.B) {
			mockRepo := createMockRepoWithRows(10000)
			broadcaster := sse.NewBroadcaster()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go broadcaster.Run(ctx)

			executor := usecase.NewParallelExecutor(mockRepo, nil, broadcaster, 4)

			plan := usecase.ExecutionPlan{
				TransportID:  "BENCH-001",
				JobID:        "JOB-001",
				JobVersion:   "v001",
				Tables:       []string{"TABLE_1", "TABLE_2", "TABLE_3", "TABLE_4"},
				Concurrency:  4,
				Owner:        "BENCH",
				BufferConfig: &cfg,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, _ := executor.Execute(ctx, plan)
				if result != nil {
					b.ReportMetric(result.RowsPerSecond(), "rows/sec")
				}
			}
		})
	}
}

// createMockRepoWithRows는 지정된 row 수를 반환하는 Mock 저장소를 생성합니다
func createMockRepoWithRows(rowsPerTable int) *oracle.MockRepository {
	mockRepo := oracle.NewMockRepository()

	// 청크 수 계산 (각 청크당 1000 rows)
	chunkSize := 1000
	chunkCount := (rowsPerTable + chunkSize - 1) / chunkSize

	mockRepo.MockChunks = make([]*domain.ChunkResult, chunkCount)
	for i := 0; i < chunkCount; i++ {
		rows := chunkSize
		if i == chunkCount-1 {
			rows = rowsPerTable - (i * chunkSize)
			if rows <= 0 {
				rows = chunkSize
			}
		}
		mockRepo.MockChunks[i] = &domain.ChunkResult{
			ChunkNumber:   i + 1,
			RowCount:      rows,
			IsLastChunk:   i == chunkCount-1,
			TotalRowsSent: int64((i + 1) * chunkSize),
		}
	}

	return mockRepo
}

// BenchmarkTargetThroughput은 목표 처리량(100,000 rows/sec)을 검증합니다
func BenchmarkTargetThroughput(b *testing.B) {
	// 1백만 row 테스트
	mockRepo := createMockRepoWithRows(1000000)
	
	// 고성능 설정
	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go broadcaster.Run(ctx)

	executor := usecase.NewParallelExecutor(mockRepo, nil, broadcaster, 8)

	highPerfConfig := buffer.HighPerformanceConfig()
	plan := usecase.ExecutionPlan{
		TransportID:  "BENCH-TARGET",
		JobID:        "JOB-TARGET",
		JobVersion:   "v001",
		Tables:       []string{"MILLION_ROW_TABLE"},
		Concurrency:  8,
		Owner:        "BENCH",
		BufferConfig: &highPerfConfig,
	}

	var totalRows int64
	var totalDuration float64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ := executor.Execute(ctx, plan)
		if result != nil {
			atomic.AddInt64(&totalRows, result.TotalRows)
			totalDuration += result.Duration().Seconds()
		}
	}

	// 평균 처리량 계산
	avgRowsPerSec := float64(totalRows) / totalDuration
	b.ReportMetric(avgRowsPerSec, "avg_rows/sec")
	
	// 목표: 100,000 rows/sec
	if avgRowsPerSec >= 100000 {
		b.Logf("TARGET ACHIEVED: %.0f rows/sec >= 100,000 rows/sec", avgRowsPerSec)
	} else {
		b.Logf("TARGET NOT MET: %.0f rows/sec < 100,000 rows/sec", avgRowsPerSec)
	}
}
