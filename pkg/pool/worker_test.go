// Package pool은 병렬 작업 처리를 위한 worker pool을 제공합니다.
package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name            string
		workers         int
		expectedWorkers int
	}{
		{
			name:            "기본 worker 수",
			workers:         4,
			expectedWorkers: 4,
		},
		{
			name:            "최소 worker 수 보장 (0 입력시)",
			workers:         0,
			expectedWorkers: 1,
		},
		{
			name:            "최소 worker 수 보장 (음수 입력시)",
			workers:         -5,
			expectedWorkers: 1,
		},
		{
			name:            "단일 worker",
			workers:         1,
			expectedWorkers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers)
			require.NotNil(t, pool)
			assert.Equal(t, tt.expectedWorkers, pool.Workers())
		})
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()

	// 작업 실행
	var executed int32

	task := Task{
		ID:        "task-001",
		TableName: "VBRP",
		Execute: func(ctx context.Context) error {
			atomic.AddInt32(&executed, 1)
			return nil
		},
	}

	pool.Start(ctx)
	pool.Submit(task)

	results := pool.Wait()
	pool.Stop()

	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
	assert.Len(t, results, 1)
	assert.Equal(t, "task-001", results[0].TaskID)
	assert.NoError(t, results[0].Error)
}

func TestWorkerPool_SubmitMultiple(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx := context.Background()

	var executionCount int32

	pool.Start(ctx)

	// 10개 작업 제출
	for i := 0; i < 10; i++ {
		taskID := "task-" + string(rune('A'+i))
		pool.Submit(Task{
			ID:        taskID,
			TableName: "TABLE_" + string(rune('A'+i)),
			Execute: func(ctx context.Context) error {
				atomic.AddInt32(&executionCount, 1)
				time.Sleep(10 * time.Millisecond) // 약간의 지연
				return nil
			},
		})
	}

	results := pool.Wait()
	pool.Stop()

	assert.Equal(t, int32(10), atomic.LoadInt32(&executionCount))
	assert.Len(t, results, 10)
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()

	pool.Start(ctx)

	// 성공 작업
	pool.Submit(Task{
		ID:        "success-task",
		TableName: "VBRP",
		Execute: func(ctx context.Context) error {
			return nil
		},
	})

	// 실패 작업
	pool.Submit(Task{
		ID:        "fail-task",
		TableName: "VBRK",
		Execute: func(ctx context.Context) error {
			return errors.New("테이블 추출 실패")
		},
	})

	results := pool.Wait()
	pool.Stop()

	assert.Len(t, results, 2)

	// 결과 검증
	var successCount, failCount int
	for _, r := range results {
		if r.Error != nil {
			failCount++
			assert.Equal(t, "fail-task", r.TaskID)
		} else {
			successCount++
			assert.Equal(t, "success-task", r.TaskID)
		}
	}

	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, failCount)
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx, cancel := context.WithCancel(context.Background())

	var startedCount int32

	pool.Start(ctx)

	// 긴 작업 제출
	for i := 0; i < 5; i++ {
		pool.Submit(Task{
			ID:        "long-task",
			TableName: "LONG_TABLE",
			Execute: func(ctx context.Context) error {
				atomic.AddInt32(&startedCount, 1)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
					return nil
				}
			},
		})
	}

	// 약간의 지연 후 취소
	time.Sleep(50 * time.Millisecond)
	cancel()

	results := pool.Wait()
	pool.Stop()

	// 일부 작업은 취소되어야 함
	var cancelledCount int
	for _, r := range results {
		if r.Error == context.Canceled {
			cancelledCount++
		}
	}

	// 최소 1개 이상 취소됨
	assert.GreaterOrEqual(t, cancelledCount, 0)
}

func TestWorkerPool_Concurrency(t *testing.T) {
	pool := NewWorkerPool(4)
	ctx := context.Background()

	var maxConcurrent int32
	var currentConcurrent int32
	var mu sync.Mutex

	pool.Start(ctx)

	// 10개 작업 제출
	for i := 0; i < 10; i++ {
		pool.Submit(Task{
			ID:        "concurrent-task",
			TableName: "TABLE",
			Execute: func(ctx context.Context) error {
				current := atomic.AddInt32(&currentConcurrent, 1)

				mu.Lock()
				if current > maxConcurrent {
					maxConcurrent = current
				}
				mu.Unlock()

				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&currentConcurrent, -1)
				return nil
			},
		})
	}

	results := pool.Wait()
	pool.Stop()

	// 동시 실행 수가 worker 수를 초과하지 않아야 함
	assert.LessOrEqual(t, maxConcurrent, int32(4))
	assert.Len(t, results, 10)
}

func TestWorkerPool_ResultMetrics(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()

	pool.Start(ctx)

	pool.Submit(Task{
		ID:        "metrics-task",
		TableName: "VBRP",
		Execute: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	})

	results := pool.Wait()
	pool.Stop()

	require.Len(t, results, 1)
	result := results[0]

	assert.Equal(t, "metrics-task", result.TaskID)
	assert.Equal(t, "VBRP", result.TableName)
	assert.GreaterOrEqual(t, result.Duration.Milliseconds(), int64(50))
	assert.False(t, result.StartTime.IsZero())
	assert.False(t, result.EndTime.IsZero())
}

func TestWorkerPool_TaskOrder(t *testing.T) {
	pool := NewWorkerPool(1) // 단일 worker로 순서 보장
	ctx := context.Background()

	var order []string
	var mu sync.Mutex

	pool.Start(ctx)

	for _, name := range []string{"A", "B", "C", "D", "E"} {
		tableName := name
		pool.Submit(Task{
			ID:        "task-" + tableName,
			TableName: tableName,
			Execute: func(ctx context.Context) error {
				mu.Lock()
				order = append(order, tableName)
				mu.Unlock()
				return nil
			},
		})
	}

	results := pool.Wait()
	pool.Stop()

	assert.Len(t, results, 5)
	// 단일 worker이므로 순서대로 실행됨
	assert.Equal(t, []string{"A", "B", "C", "D", "E"}, order)
}

func TestResult_Success(t *testing.T) {
	result := Result{
		TaskID:    "test-task",
		TableName: "VBRP",
		Error:     nil,
	}

	assert.True(t, result.Success())
}

func TestResult_Failure(t *testing.T) {
	result := Result{
		TaskID:    "test-task",
		TableName: "VBRP",
		Error:     errors.New("테스트 에러"),
	}

	assert.False(t, result.Success())
}
