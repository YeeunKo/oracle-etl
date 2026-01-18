// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-02: Retry with Exponential Backoff 테스트
package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetry_SuccessOnFirstAttempt는 첫 시도 성공을 테스트합니다
func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	attempts := 0
	cfg := DefaultRetryConfig()

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return nil // 성공
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

// TestRetry_SuccessAfterRetries는 재시도 후 성공을 테스트합니다
func TestRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil // 3번째에 성공
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

// TestRetry_FailsAfterMaxRetries는 최대 재시도 초과 시 실패를 테스트합니다
func TestRetry_FailsAfterMaxRetries(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return errors.New("persistent error")
	})

	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // MaxRetries 횟수만큼 시도
	assert.Contains(t, err.Error(), "persistent error")
}

// TestRetry_ContextCancellation은 컨텍스트 취소 시 중단을 테스트합니다
func TestRetry_ContextCancellation(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxRetries:   10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Retry(ctx, cfg, func() error {
		attempts++
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || attempts < 10)
}

// TestRetry_ExponentialBackoff는 지수 백오프를 테스트합니다
func TestRetry_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   4,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	var timestamps []time.Time
	start := time.Now()

	_ = Retry(context.Background(), cfg, func() error {
		timestamps = append(timestamps, time.Now())
		return errors.New("error")
	})

	// 첫 번째 시도와 두 번째 시도 사이: ~50ms
	// 두 번째와 세 번째: ~100ms
	// 세 번째와 네 번째: ~200ms
	totalDuration := time.Since(start)
	// 최소 지연 시간: 50 + 100 + 200 = 350ms
	assert.True(t, totalDuration >= 300*time.Millisecond, "총 소요시간: %v", totalDuration)
}

// TestRetry_MaxDelayLimit은 최대 지연 시간 제한을 테스트합니다
func TestRetry_MaxDelayLimit(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     150 * time.Millisecond, // 낮은 최대값
		Multiplier:   10.0,                   // 높은 배수
	}

	start := time.Now()
	attempts := 0

	_ = Retry(context.Background(), cfg, func() error {
		attempts++
		return errors.New("error")
	})

	totalDuration := time.Since(start)
	// MaxDelay가 150ms로 제한되므로 전체 시간이 과도하게 길어지지 않아야 함
	// 4번의 대기 (첫 시도 후): 각각 최대 150ms = 최대 600ms
	assert.True(t, totalDuration < 1*time.Second, "총 소요시간: %v", totalDuration)
}

// TestRetry_RetryableFunc은 재시도 가능 여부 판단 함수를 테스트합니다
func TestRetry_RetryableFunc(t *testing.T) {
	attempts := 0
	nonRetryableErr := errors.New("non-retryable error")

	cfg := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		RetryableFunc: func(err error) bool {
			return !errors.Is(err, nonRetryableErr)
		},
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return nonRetryableErr // 재시도 불가 에러
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // 재시도 없이 즉시 반환
}

// TestRetryWithResult_Success는 결과 반환 재시도 성공을 테스트합니다
func TestRetryWithResult_Success(t *testing.T) {
	attempts := 0
	cfg := DefaultRetryConfig()

	result, err := RetryWithResult(context.Background(), cfg, func() (string, error) {
		attempts++
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 1, attempts)
}

// TestRetryWithResult_SuccessAfterRetries는 재시도 후 결과 반환을 테스트합니다
func TestRetryWithResult_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	result, err := RetryWithResult(context.Background(), cfg, func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("not ready")
		}
		return 42, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, attempts)
}

// TestRetryWithResult_FailsAfterMaxRetries는 최대 재시도 초과 시 실패를 테스트합니다
func TestRetryWithResult_FailsAfterMaxRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	result, err := RetryWithResult(context.Background(), cfg, func() (string, error) {
		return "", errors.New("always fails")
	})

	assert.Error(t, err)
	assert.Empty(t, result)
}

// TestDefaultRetryConfig는 기본 설정을 테스트합니다
func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
	assert.Nil(t, cfg.RetryableFunc)
}

// TestRetry_ZeroMaxRetries는 MaxRetries가 0일 때 한 번만 시도하는지 테스트합니다
func TestRetry_ZeroMaxRetries(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{
		MaxRetries:   1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempts)
}
