// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-03: Circuit Breaker 테스트
package resilience

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCircuitBreaker_InitialState는 초기 상태를 테스트합니다
func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitConfig())

	assert.Equal(t, StateClosed, cb.State())
}

// TestCircuitBreaker_ClosedState_Success는 Closed 상태에서 성공을 테스트합니다
func TestCircuitBreaker_ClosedState_Success(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitConfig())

	err := cb.Execute(func() error {
		return nil // 성공
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

// TestCircuitBreaker_OpensAfterFailures는 실패 임계값 초과 시 Open 상태 전환을 테스트합니다
func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
	}
	cb := NewCircuitBreaker(cfg)

	// 3번 연속 실패
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errors.New("error")
		})
	}

	assert.Equal(t, StateOpen, cb.State())
}

// TestCircuitBreaker_OpenState_RejectsRequests는 Open 상태에서 요청 거부를 테스트합니다
func TestCircuitBreaker_OpenState_RejectsRequests(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	}
	cb := NewCircuitBreaker(cfg)

	// Circuit 열기
	_ = cb.Execute(func() error {
		return errors.New("error")
	})

	// Open 상태에서 요청
	err := cb.Execute(func() error {
		return nil
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCircuitOpen))
}

// TestCircuitBreaker_TransitionsToHalfOpen은 타임아웃 후 Half-Open 전환을 테스트합니다
func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(cfg)

	// Circuit 열기
	_ = cb.Execute(func() error {
		return errors.New("error")
	})
	assert.Equal(t, StateOpen, cb.State())

	// 타임아웃 대기
	time.Sleep(150 * time.Millisecond)

	// 다음 요청 시 Half-Open으로 전환되어야 함
	_ = cb.Execute(func() error {
		return nil
	})

	// 성공했으므로 Closed로 전환
	assert.Equal(t, StateClosed, cb.State())
}

// TestCircuitBreaker_HalfOpen_Success는 Half-Open에서 성공 시 Closed 전환을 테스트합니다
func TestCircuitBreaker_HalfOpen_Success(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(cfg)

	// Circuit 열기
	_ = cb.Execute(func() error {
		return errors.New("error")
	})

	// 타임아웃 대기
	time.Sleep(60 * time.Millisecond)

	// 2번 성공하면 Closed로 전환
	_ = cb.Execute(func() error { return nil })
	assert.Equal(t, StateHalfOpen, cb.State())

	_ = cb.Execute(func() error { return nil })
	assert.Equal(t, StateClosed, cb.State())
}

// TestCircuitBreaker_HalfOpen_Failure는 Half-Open에서 실패 시 Open 전환을 테스트합니다
func TestCircuitBreaker_HalfOpen_Failure(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(cfg)

	// Circuit 열기
	_ = cb.Execute(func() error {
		return errors.New("error")
	})

	// 타임아웃 대기
	time.Sleep(60 * time.Millisecond)

	// Half-Open에서 실패하면 다시 Open
	_ = cb.Execute(func() error {
		return errors.New("still failing")
	})

	assert.Equal(t, StateOpen, cb.State())
}

// TestCircuitBreaker_ResetCounts는 성공 시 실패 카운트 리셋을 테스트합니다
func TestCircuitBreaker_ResetCounts(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 3,
		SuccessThreshold: 1,
		Timeout:          1 * time.Second,
	}
	cb := NewCircuitBreaker(cfg)

	// 2번 실패
	_ = cb.Execute(func() error { return errors.New("error") })
	_ = cb.Execute(func() error { return errors.New("error") })

	// 1번 성공 (실패 카운트 리셋)
	_ = cb.Execute(func() error { return nil })

	// 다시 2번 실패
	_ = cb.Execute(func() error { return errors.New("error") })
	_ = cb.Execute(func() error { return errors.New("error") })

	// 아직 Closed 상태 (3번 연속이 아님)
	assert.Equal(t, StateClosed, cb.State())
}

// TestCircuitBreaker_ConcurrentAccess는 동시 접근을 테스트합니다
func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 100,
		SuccessThreshold: 10,
		Timeout:          1 * time.Second,
	}
	cb := NewCircuitBreaker(cfg)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = cb.Execute(func() error {
				if n%2 == 0 {
					return errors.New("error")
				}
				return nil
			})
		}(i)
	}

	wg.Wait()
	// 동시 접근에도 패닉 없이 완료되어야 함
	assert.True(t, cb.State() == StateClosed || cb.State() == StateOpen || cb.State() == StateHalfOpen)
}

// TestDefaultCircuitConfig는 기본 설정을 테스트합니다
func TestDefaultCircuitConfig(t *testing.T) {
	cfg := DefaultCircuitConfig()

	assert.Equal(t, 5, cfg.FailureThreshold)
	assert.Equal(t, 3, cfg.SuccessThreshold)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

// TestCircuitBreaker_StateString은 상태 문자열 변환을 테스트합니다
func TestCircuitBreaker_StateString(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
}

// TestCircuitBreaker_Metrics는 메트릭 조회를 테스트합니다
func TestCircuitBreaker_Metrics(t *testing.T) {
	cfg := CircuitConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
	}
	cb := NewCircuitBreaker(cfg)

	// 몇 번 실행
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return errors.New("error") })
	_ = cb.Execute(func() error { return nil })

	metrics := cb.Metrics()

	require.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.TotalRequests)
	assert.Equal(t, int64(2), metrics.Successes)
	assert.Equal(t, int64(1), metrics.Failures)
}

// TestErrCircuitOpen은 Circuit Open 에러를 테스트합니다
func TestErrCircuitOpen(t *testing.T) {
	assert.Equal(t, "circuit breaker is open", ErrCircuitOpen.Error())
}
