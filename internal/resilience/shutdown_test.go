// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-04: Graceful Shutdown 테스트
package resilience

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGracefulShutdown_Basic은 기본 shutdown 동작을 테스트합니다
func TestGracefulShutdown_Basic(t *testing.T) {
	gs := NewGracefulShutdown()

	// 초기 상태
	assert.False(t, gs.IsShuttingDown())
}

// TestGracefulShutdown_Trigger는 shutdown 트리거를 테스트합니다
func TestGracefulShutdown_Trigger(t *testing.T) {
	gs := NewGracefulShutdown()

	gs.Trigger()

	assert.True(t, gs.IsShuttingDown())
}

// TestGracefulShutdown_RegisterCleanup은 정리 함수 등록을 테스트합니다
func TestGracefulShutdown_RegisterCleanup(t *testing.T) {
	gs := NewGracefulShutdown()

	cleaned := false
	gs.RegisterCleanup("test", func(ctx context.Context) error {
		cleaned = true
		return nil
	})

	ctx := context.Background()
	gs.Trigger()
	_ = gs.Wait(ctx)

	assert.True(t, cleaned)
}

// TestGracefulShutdown_CleanupOrder는 역순 정리를 테스트합니다
func TestGracefulShutdown_CleanupOrder(t *testing.T) {
	gs := NewGracefulShutdown()

	order := make([]string, 0)

	gs.RegisterCleanup("first", func(ctx context.Context) error {
		order = append(order, "first")
		return nil
	})
	gs.RegisterCleanup("second", func(ctx context.Context) error {
		order = append(order, "second")
		return nil
	})
	gs.RegisterCleanup("third", func(ctx context.Context) error {
		order = append(order, "third")
		return nil
	})

	ctx := context.Background()
	gs.Trigger()
	_ = gs.Wait(ctx)

	// LIFO 순서로 정리 (역순)
	assert.Equal(t, []string{"third", "second", "first"}, order)
}

// TestGracefulShutdown_Timeout은 타임아웃 처리를 테스트합니다
func TestGracefulShutdown_Timeout(t *testing.T) {
	gs := NewGracefulShutdown()

	gs.RegisterCleanup("slow", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	gs.Trigger()
	err := gs.Wait(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestGracefulShutdown_WaitChannel은 채널 대기를 테스트합니다
func TestGracefulShutdown_WaitChannel(t *testing.T) {
	gs := NewGracefulShutdown()

	done := gs.Done()

	// 아직 shutdown 전
	select {
	case <-done:
		t.Fatal("should not be done yet")
	default:
	}

	gs.Trigger()

	// shutdown 후 채널이 닫혀야 함
	select {
	case <-done:
		// 예상대로 동작
	case <-time.After(100 * time.Millisecond):
		t.Fatal("done channel should be closed")
	}
}

// TestGracefulShutdown_InFlightRequests는 진행 중 요청 추적을 테스트합니다
func TestGracefulShutdown_InFlightRequests(t *testing.T) {
	gs := NewGracefulShutdown()

	// 요청 시작
	gs.AddInFlight()
	gs.AddInFlight()

	assert.Equal(t, int32(2), gs.InFlightCount())

	// 요청 완료
	gs.DoneInFlight()

	assert.Equal(t, int32(1), gs.InFlightCount())
}

// TestGracefulShutdown_WaitForInFlight는 진행 중 요청 완료 대기를 테스트합니다
func TestGracefulShutdown_WaitForInFlight(t *testing.T) {
	gs := NewGracefulShutdown()

	var completed atomic.Bool

	gs.AddInFlight()

	// 백그라운드에서 요청 완료
	go func() {
		time.Sleep(50 * time.Millisecond)
		gs.DoneInFlight()
	}()

	ctx := context.Background()
	gs.Trigger()

	go func() {
		_ = gs.Wait(ctx)
		completed.Store(true)
	}()

	// 잠시 대기 후 완료 확인
	time.Sleep(100 * time.Millisecond)
	assert.True(t, completed.Load())
}

// TestGracefulShutdown_MultipleTriggersAreSafe는 중복 트리거 안전성을 테스트합니다
func TestGracefulShutdown_MultipleTriggersAreSafe(t *testing.T) {
	gs := NewGracefulShutdown()

	callCount := 0
	gs.RegisterCleanup("counter", func(ctx context.Context) error {
		callCount++
		return nil
	})

	// 여러 번 트리거
	gs.Trigger()
	gs.Trigger()
	gs.Trigger()

	ctx := context.Background()
	_ = gs.Wait(ctx)
	_ = gs.Wait(ctx) // 두 번째 Wait는 즉시 반환

	// 정리 함수는 한 번만 호출되어야 함
	assert.Equal(t, 1, callCount)
}
