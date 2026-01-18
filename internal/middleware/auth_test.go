package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/config"
)

// TestAPIKeyAuth_ValidKey는 유효한 API Key 인증을 테스트
func TestAPIKeyAuth_ValidKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled: true,
		APIKeys: []string{"valid-api-key-1", "valid-api-key-2"},
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "valid-api-key-1")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestAPIKeyAuth_InvalidKey는 유효하지 않은 API Key 인증 거부를 테스트
func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled: true,
		APIKeys: []string{"valid-api-key"},
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var errResp ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "AUTHENTICATION_ERROR", errResp.Code)
}

// TestAPIKeyAuth_MissingKey는 API Key 누락 시 인증 거부를 테스트
func TestAPIKeyAuth_MissingKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled: true,
		APIKeys: []string{"valid-api-key"},
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// X-API-Key 헤더 없음

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBearerAuth_ValidToken는 유효한 Bearer 토큰 인증을 테스트
func TestBearerAuth_ValidToken(t *testing.T) {
	secret := "test-secret-key"
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		BearerSecret: secret,
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 유효한 JWT 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBearerAuth_ExpiredToken는 만료된 토큰 거부를 테스트
func TestBearerAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		BearerSecret: secret,
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 만료된 JWT 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(-time.Hour).Unix(), // 1시간 전 만료
	})
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBearerAuth_InvalidSignature는 잘못된 서명 토큰 거부를 테스트
func TestBearerAuth_InvalidSignature(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		BearerSecret: "correct-secret",
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// 다른 비밀키로 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBearerAuth_MalformedToken는 형식이 잘못된 토큰 거부를 테스트
func TestBearerAuth_MalformedToken(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		BearerSecret: "test-secret",
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.format")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestAuthMiddleware_Disabled는 인증 비활성화 시 모든 요청 허용을 테스트
func TestAuthMiddleware_Disabled(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled: false, // 인증 비활성화
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// 인증 헤더 없음

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestAuthMiddleware_BothAPIKeyAndBearer는 API Key와 Bearer 둘 다 지원을 테스트
func TestAuthMiddleware_BothAPIKeyAndBearer(t *testing.T) {
	secret := "test-secret"
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		APIKeys:      []string{"valid-api-key"},
		BearerSecret: secret,
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// API Key로 인증
	t.Run("API Key auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "valid-api-key")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// Bearer 토큰으로 인증
	t.Run("Bearer auth", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

// TestAuthMiddleware_SkipsHealthEndpoint는 /api/health 엔드포인트 인증 제외를 테스트
func TestAuthMiddleware_SkipsHealthEndpoint(t *testing.T) {
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled: true,
		APIKeys: []string{"valid-api-key"},
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	// 인증 헤더 없음

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestAuthMiddleware_SetsUserContext는 인증 성공 시 사용자 정보 컨텍스트 설정을 테스트
func TestAuthMiddleware_SetsUserContext(t *testing.T) {
	secret := "test-secret"
	app := fiber.New()
	cfg := &config.AuthConfig{
		Enabled:      true,
		BearerSecret: secret,
	}

	app.Use(NewAuthMiddleware(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(500).SendString("user_id not set")
		}
		return c.SendString(userID.(string))
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
