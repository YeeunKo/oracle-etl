// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-04: Graceful Shutdown
package resilience

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// CleanupFunc는 정리 함수 타입입니다
type CleanupFunc func(ctx context.Context) error

// cleanupEntry는 정리 함수 항목입니다
type cleanupEntry struct {
	name string
	fn   CleanupFunc
}

// GracefulShutdown은 우아한 종료를 관리합니다
type GracefulShutdown struct {
	mu sync.Mutex

	shuttingDown atomic.Bool
	done         chan struct{}
	cleanups     []cleanupEntry
	inFlight     atomic.Int32

	// 한 번만 정리 실행을 보장
	cleanupOnce sync.Once
	cleanupDone chan struct{}
}

// NewGracefulShutdown은 새로운 GracefulShutdown을 생성합니다
func NewGracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		done:        make(chan struct{}),
		cleanups:    make([]cleanupEntry, 0),
		cleanupDone: make(chan struct{}),
	}
}

// RegisterCleanup은 정리 함수를 등록합니다
// 등록된 순서의 역순으로 실행됩니다 (LIFO)
func (gs *GracefulShutdown) RegisterCleanup(name string, fn CleanupFunc) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.cleanups = append(gs.cleanups, cleanupEntry{
		name: name,
		fn:   fn,
	})
}

// Trigger는 종료를 시작합니다
func (gs *GracefulShutdown) Trigger() {
	if gs.shuttingDown.CompareAndSwap(false, true) {
		close(gs.done)
	}
}

// IsShuttingDown은 종료 중인지 확인합니다
func (gs *GracefulShutdown) IsShuttingDown() bool {
	return gs.shuttingDown.Load()
}

// Done은 종료 신호를 받을 채널을 반환합니다
func (gs *GracefulShutdown) Done() <-chan struct{} {
	return gs.done
}

// AddInFlight는 진행 중인 요청을 추가합니다
func (gs *GracefulShutdown) AddInFlight() {
	gs.inFlight.Add(1)
}

// DoneInFlight는 진행 중인 요청 완료를 표시합니다
func (gs *GracefulShutdown) DoneInFlight() {
	gs.inFlight.Add(-1)
}

// InFlightCount는 진행 중인 요청 수를 반환합니다
func (gs *GracefulShutdown) InFlightCount() int32 {
	return gs.inFlight.Load()
}

// Wait는 모든 정리 작업이 완료될 때까지 대기합니다
func (gs *GracefulShutdown) Wait(ctx context.Context) error {
	// 이미 정리 완료된 경우
	select {
	case <-gs.cleanupDone:
		return nil
	default:
	}

	var cleanupErr error

	gs.cleanupOnce.Do(func() {
		defer close(gs.cleanupDone)

		// 진행 중인 요청 완료 대기
		cleanupErr = gs.waitForInFlight(ctx)
		if cleanupErr != nil {
			return
		}

		// 정리 함수 역순 실행
		cleanupErr = gs.runCleanups(ctx)
	})

	// 다른 고루틴이 정리 중인 경우 완료 대기
	select {
	case <-gs.cleanupDone:
		return cleanupErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// waitForInFlight는 진행 중인 요청이 완료될 때까지 대기합니다
func (gs *GracefulShutdown) waitForInFlight(ctx context.Context) error {
	for gs.inFlight.Load() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-awaitZero(&gs.inFlight):
			return nil
		}
	}
	return nil
}

// awaitZero는 카운터가 0이 될 때까지 대기하는 채널을 반환합니다
func awaitZero(counter *atomic.Int32) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for counter.Load() > 0 {
			// 짧은 폴링 (실제 구현에서는 조건 변수 사용 권장)
		}
		close(done)
	}()
	return done
}

// runCleanups는 등록된 정리 함수를 역순으로 실행합니다
func (gs *GracefulShutdown) runCleanups(ctx context.Context) error {
	gs.mu.Lock()
	cleanups := make([]cleanupEntry, len(gs.cleanups))
	copy(cleanups, gs.cleanups)
	gs.mu.Unlock()

	var errs []error

	// 역순 실행 (LIFO)
	for i := len(cleanups) - 1; i >= 0; i-- {
		entry := cleanups[i]

		select {
		case <-ctx.Done():
			return fmt.Errorf("cleanup interrupted: %w", ctx.Err())
		default:
		}

		if err := entry.fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", entry.name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}
