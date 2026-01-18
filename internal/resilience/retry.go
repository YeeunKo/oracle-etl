// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-02: Retry with Exponential Backoff
package resilience

import (
	"context"
	"time"
)

// RetryConfig는 재시도 설정을 정의합니다
type RetryConfig struct {
	// MaxRetries는 최대 시도 횟수입니다 (1 = 재시도 없음)
	MaxRetries int
	// InitialDelay는 첫 재시도 전 대기 시간입니다
	InitialDelay time.Duration
	// MaxDelay는 최대 대기 시간입니다
	MaxDelay time.Duration
	// Multiplier는 지수 백오프 배수입니다
	Multiplier float64
	// RetryableFunc은 에러가 재시도 가능한지 판단하는 함수입니다
	// nil이면 모든 에러가 재시도 가능으로 간주됩니다
	RetryableFunc func(error) bool
}

// DefaultRetryConfig는 기본 재시도 설정을 반환합니다
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		Multiplier:    2.0,
		RetryableFunc: nil, // 모든 에러 재시도
	}
}

// Retry는 주어진 함수를 설정에 따라 재시도합니다
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return lastErr
			}
			return ctx.Err()
		default:
		}

		// 함수 실행
		lastErr = fn()
		if lastErr == nil {
			return nil // 성공
		}

		// 재시도 가능 여부 확인
		if cfg.RetryableFunc != nil && !cfg.RetryableFunc(lastErr) {
			return lastErr // 재시도 불가능한 에러
		}

		// 마지막 시도였으면 대기 없이 반환
		if attempt == cfg.MaxRetries-1 {
			break
		}

		// 지수 백오프 대기
		select {
		case <-ctx.Done():
			return lastErr
		case <-time.After(delay):
		}

		// 다음 대기 시간 계산
		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return lastErr
}

// RetryWithResult는 결과를 반환하는 함수를 재시도합니다
func RetryWithResult[T any](ctx context.Context, cfg RetryConfig, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return result, lastErr
			}
			return result, ctx.Err()
		default:
		}

		// 함수 실행
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil // 성공
		}

		// 재시도 가능 여부 확인
		if cfg.RetryableFunc != nil && !cfg.RetryableFunc(lastErr) {
			return result, lastErr // 재시도 불가능한 에러
		}

		// 마지막 시도였으면 대기 없이 반환
		if attempt == cfg.MaxRetries-1 {
			break
		}

		// 지수 백오프 대기
		select {
		case <-ctx.Done():
			return result, lastErr
		case <-time.After(delay):
		}

		// 다음 대기 시간 계산
		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return result, lastErr
}
