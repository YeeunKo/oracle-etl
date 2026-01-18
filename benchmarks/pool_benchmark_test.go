// Package benchmarks는 성능 벤치마크 테스트를 제공합니다.
package benchmarks

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"oracle-etl/pkg/pool"
)

// BenchmarkWorkerPool_Submit은 워커 풀의 작업 제출 성능을 측정합니다
func BenchmarkWorkerPool_Submit(b *testing.B) {
	ctx := context.Background()
	wp := pool.NewWorkerPool(4)
	wp.Start(ctx)
	defer wp.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.Submit(pool.Task{
			ID:        "task",
			TableName: "TEST",
			Execute: func(ctx context.Context) error {
				return nil
			},
		})
	}
	wp.Wait()
}

// BenchmarkWorkerPool_ConcurrencyLevels는 다양한 동시성 레벨에서 성능을 측정합니다
func BenchmarkWorkerPool_ConcurrencyLevels(b *testing.B) {
	levels := []int{1, 2, 4, 8, 16}
	taskCount := 100

	for _, workers := range levels {
		b.Run(
			poolName(workers),
			func(b *testing.B) {
				ctx := context.Background()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					wp := pool.NewWorkerPool(workers)
					wp.Start(ctx)

					for j := 0; j < taskCount; j++ {
						wp.Submit(pool.Task{
							ID:        "task",
							TableName: "TEST",
							Execute: func(ctx context.Context) error {
								// 시뮬레이션: 약간의 작업
								time.Sleep(time.Microsecond * 100)
								return nil
							},
						})
					}
					wp.Wait()
				}
			},
		)
	}
}

// BenchmarkWorkerPool_Throughput은 워커 풀의 처리량을 측정합니다
func BenchmarkWorkerPool_Throughput(b *testing.B) {
	ctx := context.Background()
	wp := pool.NewWorkerPool(8)
	wp.Start(ctx)
	defer wp.Stop()

	var processed int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wp.Submit(pool.Task{
				ID:        "throughput-task",
				TableName: "THROUGHPUT_TEST",
				Execute: func(ctx context.Context) error {
					atomic.AddInt64(&processed, 1)
					return nil
				},
			})
		}
	})
	wp.Wait()

	b.ReportMetric(float64(processed)/b.Elapsed().Seconds(), "tasks/sec")
}

func poolName(workers int) string {
	return "workers=" + itoa(workers)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string('0'+byte(i%10)) + result
		i /= 10
	}
	return result
}
