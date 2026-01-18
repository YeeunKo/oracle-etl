// Package middleware는 HTTP 미들웨어를 제공합니다.
// M7-05: Enhanced Panic Recovery
package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apperrors "oracle-etl/internal/errors"
)

// NewRecoveryMiddleware는 개선된 패닉 복구 미들웨어를 생성합니다
// - 구조화된 에러 응답 (ErrorResponse)
// - TraceID 지원
// - 요청 컨텍스트 로깅
// - error 타입 패닉 처리
func NewRecoveryMiddleware(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// TraceID 추출 또는 생성
				traceID := c.Get("X-Request-ID")
				if traceID == "" {
					traceID = uuid.New().String()
				}

				// 스택 트레이스 캡처
				stack := debug.Stack()

				// 패닉 값에서 에러 정보 추출
				errResp := extractError(r, traceID)

				// 요청 컨텍스트와 함께 패닉 로깅
				logger.Error().
					Str("panic", fmt.Sprintf("%v", r)).
					Str("stack", string(stack)).
					Str("trace_id", traceID).
					Str("method", c.Method()).
					Str("path", c.Path()).
					Str("ip", c.IP()).
					Str("user_agent", c.Get("User-Agent")).
					Msg("panic recovered")

				// JSON 에러 응답 반환
				c.Status(errResp.HTTPStatus())
				_ = c.JSON(errResp)
			}
		}()

		return c.Next()
	}
}

// extractError는 패닉 값에서 ErrorResponse를 추출합니다
func extractError(r any, traceID string) *apperrors.ErrorResponse {
	// ErrorResponse 타입인 경우
	if errResp, ok := r.(*apperrors.ErrorResponse); ok {
		return errResp.WithTraceID(traceID)
	}

	// 일반 error 타입인 경우
	if err, ok := r.(error); ok {
		return apperrors.NewError(apperrors.ErrCodeInternal, err.Error()).
			WithTraceID(traceID)
	}

	// 문자열 또는 기타 타입인 경우
	return apperrors.NewError(apperrors.ErrCodeInternal, "Internal Server Error").
		WithTraceID(traceID).
		WithDetails(map[string]any{
			"panic_value": fmt.Sprintf("%v", r),
		})
}

// NewRecoveryMiddlewareWithConfig는 설정 가능한 패닉 복구 미들웨어를 생성합니다
func NewRecoveryMiddlewareWithConfig(logger zerolog.Logger, config RecoveryConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// TraceID 추출 또는 생성
				traceID := c.Get(config.TraceIDHeader)
				if traceID == "" {
					traceID = uuid.New().String()
				}

				// 스택 트레이스 캡처
				stack := debug.Stack()

				// 패닉 값에서 에러 정보 추출
				errResp := extractError(r, traceID)

				// 로깅
				logEvent := logger.Error().
					Str("panic", fmt.Sprintf("%v", r)).
					Str("trace_id", traceID).
					Str("method", c.Method()).
					Str("path", c.Path())

				if config.EnableStackTrace {
					logEvent = logEvent.Str("stack", string(stack))
				}

				logEvent.Msg("panic recovered")

				// 콜백 호출 (있는 경우)
				if config.OnPanic != nil {
					config.OnPanic(c, r, stack)
				}

				// JSON 에러 응답 반환
				c.Status(errResp.HTTPStatus())
				_ = c.JSON(errResp)
			}
		}()

		return c.Next()
	}
}

// RecoveryConfig는 패닉 복구 미들웨어 설정입니다
type RecoveryConfig struct {
	// TraceIDHeader는 TraceID를 추출할 헤더 이름입니다 (기본: X-Request-ID)
	TraceIDHeader string
	// EnableStackTrace는 스택 트레이스 로깅 활성화 여부입니다
	EnableStackTrace bool
	// OnPanic은 패닉 발생 시 호출되는 콜백입니다
	OnPanic func(c *fiber.Ctx, panicValue any, stack []byte)
}

// DefaultRecoveryConfig는 기본 설정을 반환합니다
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		TraceIDHeader:    "X-Request-ID",
		EnableStackTrace: true,
		OnPanic:          nil,
	}
}
