// Package middleware는 HTTP 요청 처리를 위한 미들웨어를 제공합니다
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// NewLoggingMiddleware는 구조화된 JSON 로깅 미들웨어를 생성합니다
func NewLoggingMiddleware(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// 요청 ID 생성
		requestID := uuid.New().String()
		c.Locals("request_id", requestID)

		// 다음 핸들러 실행
		err := c.Next()

		// 응답 후 로깅
		latency := time.Since(start)
		status := c.Response().StatusCode()

		// 로그 레벨 결정
		var event *zerolog.Event
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		} else {
			event = logger.Info()
		}

		event.
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Str("latency", latency.String()).
			Msg("request completed")

		return err
	}
}
