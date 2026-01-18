package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/config"
)

// TestCORS_AllowedOrigin는 허용된 Origin 요청 처리를 테스트
func TestCORS_AllowedOrigin(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com", "https://app.example.com"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// CORS 헤더 확인
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCORS_DisallowedOrigin는 허용되지 않은 Origin 처리를 테스트
func TestCORS_DisallowedOrigin(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://malicious.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// 요청은 처리되지만 CORS 헤더가 없음
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCORS_PreflightRequest는 Preflight OPTIONS 요청 처리를 테스트
func TestCORS_PreflightRequest(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-API-Key"},
		MaxAge:       86400,
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	// Preflight 응답 헤더 확인
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", resp.Header.Get("Access-Control-Max-Age"))
}

// TestCORS_WildcardOrigin는 와일드카드 Origin 처리를 테스트
func TestCORS_WildcardOrigin(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"*"}, // 모든 Origin 허용
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCORS_AllowCredentials는 자격 증명 허용 설정을 테스트
func TestCORS_AllowCredentials(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
}

// TestCORS_Disabled는 CORS 비활성화 시 동작을 테스트
func TestCORS_Disabled(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled: false, // CORS 비활성화
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	// CORS 헤더가 없어야 함
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCORS_AllowedMethods는 허용된 메서드 설정을 테스트
func TestCORS_AllowedMethods(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		AllowHeaders: []string{"Content-Type"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	assert.Contains(t, methods, "GET")
	assert.Contains(t, methods, "POST")
	assert.Contains(t, methods, "PUT")
}

// TestCORS_AllowedHeaders는 허용된 헤더 설정을 테스트
func TestCORS_AllowedHeaders(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-API-Key", "X-Custom-Header"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-API-Key")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	headers := resp.Header.Get("Access-Control-Allow-Headers")
	assert.Contains(t, headers, "Content-Type")
	assert.Contains(t, headers, "Authorization")
	assert.Contains(t, headers, "X-API-Key")
}

// TestCORS_MaxAge는 Max-Age 설정을 테스트
func TestCORS_MaxAge(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Content-Type"},
		MaxAge:       3600, // 1시간
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, "3600", resp.Header.Get("Access-Control-Max-Age"))
}

// TestCORS_NoOriginHeader는 Origin 헤더 없는 요청 처리를 테스트
func TestCORS_NoOriginHeader(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:      true,
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Content-Type"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Origin 헤더 없음

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	// CORS 헤더가 없어야 함 (same-origin 요청)
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestCORS_ExposeHeaders는 노출 헤더 설정을 테스트
func TestCORS_ExposeHeaders(t *testing.T) {
	app := fiber.New()
	cfg := &config.CORSConfig{
		Enabled:       true,
		AllowOrigins:  []string{"https://example.com"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Content-Type"},
		ExposeHeaders: []string{"X-Custom-Response", "X-Request-ID"},
	}

	app.Use(NewCORSMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Set("X-Custom-Response", "value")
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	expose := resp.Header.Get("Access-Control-Expose-Headers")
	assert.Contains(t, expose, "X-Custom-Response")
	assert.Contains(t, expose, "X-Request-ID")
}
