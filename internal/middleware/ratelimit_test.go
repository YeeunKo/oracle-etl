package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/config"
)

// TestRateLimit_AllowsWithinLimit는 제한 내 요청 허용을 테스트
func TestRateLimit_AllowsWithinLimit(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		BurstSize:         5,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 첫 번째 요청 - 허용
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestRateLimit_RejectsOverLimit는 제한 초과 시 거부를 테스트
func TestRateLimit_RejectsOverLimit(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2, // 매우 낮은 제한
		BurstSize:         1,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 연속 요청으로 제한 초과
	var lastResp *httptest.ResponseRecorder
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)

		if resp.StatusCode == 429 {
			lastResp = httptest.NewRecorder()
			lastResp.Code = resp.StatusCode
			break
		}
	}

	// 429 응답을 받아야 함
	require.NotNil(t, lastResp)
	assert.Equal(t, 429, lastResp.Code)
}

// TestRateLimit_ReturnsRateLimitHeaders는 Rate Limit 헤더 반환을 테스트
func TestRateLimit_ReturnsRateLimitHeaders(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 100,
		BurstSize:         10,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.2")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Rate Limit 헤더 확인
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"))
}

// TestRateLimit_Disabled는 비활성화 시 모든 요청 허용을 테스트
func TestRateLimit_Disabled(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled: false, // 비활성화
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 많은 요청을 보내도 모두 허용
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.3")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

// TestRateLimit_PerIPTracking는 IP별 추적을 테스트
func TestRateLimit_PerIPTracking(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2,
		BurstSize:         1,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// IP1 - 첫 요청 허용
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	resp1, err := app.Test(req1, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// IP2 - 첫 요청 허용 (다른 IP이므로)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.2")
	resp2, err := app.Test(req2, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}

// TestRateLimit_ReturnsErrorJSON는 제한 초과 시 JSON 에러 응답을 테스트
func TestRateLimit_ReturnsErrorJSON(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		BurstSize:         1,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 제한 초과 요청
	var resp429Found bool
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.50")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)

		if resp.StatusCode == 429 {
			resp429Found = true

			// JSON 응답 확인
			var errResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.Equal(t, "RATE_LIMIT_EXCEEDED", errResp["code"])
			break
		}
	}

	assert.True(t, resp429Found, "429 응답을 받아야 함")
}

// TestRateLimit_UsesAPIKeyForTracking는 API Key 기반 추적을 테스트
func TestRateLimit_UsesAPIKeyForTracking(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 5,
		BurstSize:         2,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 같은 IP, 다른 API Key
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.10")
	req1.Header.Set("X-API-Key", "key-1")
	resp1, err := app.Test(req1, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.10")
	req2.Header.Set("X-API-Key", "key-2")
	resp2, err := app.Test(req2, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}

// TestRateLimit_RetryAfterHeader는 Retry-After 헤더 반환을 테스트
func TestRateLimit_RetryAfterHeader(t *testing.T) {
	app := fiber.New()
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		BurstSize:         1,
	}

	app.Use(NewRateLimitMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 제한 초과까지 요청
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.60")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)

		if resp.StatusCode == 429 {
			// Retry-After 헤더 확인
			retryAfter := resp.Header.Get("Retry-After")
			assert.NotEmpty(t, retryAfter)
			break
		}
	}
}
