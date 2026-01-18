// Package pool은 병렬 작업 처리를 위한 worker pool을 제공합니다.
// ETL 파이프라인에서 다중 테이블 동시 추출을 위해 사용됩니다.
package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Task는 worker pool에서 실행할 작업을 나타냅니다
type Task struct {
	ID        string                              // 작업 고유 ID
	TableName string                              // 처리할 테이블 이름
	Execute   func(ctx context.Context) error    // 실행할 함수
}

// Result는 작업 실행 결과를 나타냅니다
type Result struct {
	TaskID    string        // 작업 ID
	TableName string        // 테이블 이름
	Error     error         // 에러 (있는 경우)
	StartTime time.Time     // 시작 시간
	EndTime   time.Time     // 종료 시간
	Duration  time.Duration // 소요 시간
}

// Success는 작업이 성공했는지 반환합니다
func (r Result) Success() bool {
	return r.Error == nil
}

// WorkerPool은 병렬 작업을 관리하는 워커 풀입니다
type WorkerPool struct {
	workers   int
	taskQueue chan Task
	results   chan Result
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	started   int32
	taskCount int32
}

// NewWorkerPool은 새로운 WorkerPool을 생성합니다
// workers: 동시에 실행할 worker 수 (최소 1)
func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, 1000),  // 버퍼링된 작업 큐
		results:   make(chan Result, 1000), // 버퍼링된 결과 채널
	}
}

// Workers는 워커 수를 반환합니다
func (p *WorkerPool) Workers() int {
	return p.workers
}

// Results는 결과 채널을 반환합니다 (비동기 결과 소비용)
func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

// Start는 워커 풀을 시작합니다
func (p *WorkerPool) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.started, 0, 1) {
		return // 이미 시작됨
	}

	p.ctx, p.cancel = context.WithCancel(ctx)

	// 워커 시작
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker는 단일 워커 goroutine입니다
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// 컨텍스트 취소 - 남은 작업 처리
			for {
				select {
				case task, ok := <-p.taskQueue:
					if !ok {
						return
					}
					p.executeTask(task)
				default:
					return
				}
			}
		case task, ok := <-p.taskQueue:
			if !ok {
				return // 채널 닫힘
			}
			p.executeTask(task)
		}
	}
}

// executeTask는 단일 작업을 실행합니다
func (p *WorkerPool) executeTask(task Task) {
	startTime := time.Now()

	// 작업 실행
	var err error
	if task.Execute != nil {
		err = task.Execute(p.ctx)
	}

	endTime := time.Now()

	// 결과 전송
	result := Result{
		TaskID:    task.ID,
		TableName: task.TableName,
		Error:     err,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}

	// 컨텍스트 취소를 체크하며 결과 전송 (교착 상태 방지)
	select {
	case p.results <- result:
		// 정상 전송
	case <-p.ctx.Done():
		// 컨텍스트 취소됨 - 결과 드롭
		return
	}

	atomic.AddInt32(&p.taskCount, -1)
}

// Submit은 새로운 작업을 제출합니다
func (p *WorkerPool) Submit(task Task) {
	atomic.AddInt32(&p.taskCount, 1)
	p.taskQueue <- task
}

// Wait는 모든 작업이 완료될 때까지 기다리고 결과를 반환합니다
func (p *WorkerPool) Wait() []Result {
	// 작업 큐 닫기 (더 이상 새 작업 없음)
	close(p.taskQueue)

	// 모든 워커가 완료될 때까지 대기
	p.wg.Wait()

	// 결과 채널 닫기
	close(p.results)

	// 모든 결과 수집
	var results []Result
	for result := range p.results {
		results = append(results, result)
	}

	return results
}

// Stop은 워커 풀을 중지합니다
func (p *WorkerPool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	atomic.StoreInt32(&p.started, 0)
}

// PendingTasks는 대기 중인 작업 수를 반환합니다
func (p *WorkerPool) PendingTasks() int {
	return int(atomic.LoadInt32(&p.taskCount))
}
