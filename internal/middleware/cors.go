// Package middleware는 HTTP 미들웨어를 제공합니다.
// M6-05: CORS 설정
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"oracle-etl/internal/config"
)

// NewCORSMiddleware는 CORS 미들웨어를 생성합니다
func NewCORSMiddleware(cfg *config.CORSConfig) fiber.Handler {
	if cfg == nil || !cfg.Enabled {
		// CORS 비활성화 - 아무 작업도 하지 않음
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	// 오리진 문자열 결합
	origins := strings.Join(cfg.AllowOrigins, ", ")

	// 메서드 문자열 결합
	methods := strings.Join(cfg.AllowMethods, ", ")

	// 헤더 문자열 결합
	headers := strings.Join(cfg.AllowHeaders, ", ")

	// 노출 헤더 문자열 결합
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    exposeHeaders,
		MaxAge:           cfg.MaxAge,
	})
}
