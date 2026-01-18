package middleware

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "oracle-etl/internal/errors"
)

// ErrorResponseEnhanced는 개선된 에러 응답 구조체
type ErrorResponseEnhanced struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
	Details any    `json:"details,omitempty"`
}

// ErrorResponse는 에러 응답 구조체 (기존 호환)
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TestRecoveryMiddleware_RecoversPanic은 패닉 복구를 테스트
func TestRecoveryMiddleware_RecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// 500 응답 확인
	assert.Equal(t, 500, resp.StatusCode)
}

// TestRecoveryMiddleware_ReturnsErrorJSON은 패닉 시 JSON 에러 응답을 테스트
func TestRecoveryMiddleware_ReturnsErrorJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// Content-Type 확인
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	// 응답 본문 확인
	var errResp ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "INTERNAL_ERROR", errResp.Code)
	assert.NotEmpty(t, errResp.Message)
}

// TestRecoveryMiddleware_LogsPanic은 패닉 로깅을 테스트
func TestRecoveryMiddleware_LogsPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic message")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	_, err := app.Test(req, -1)
	require.NoError(t, err)

	// 로그에 패닉 정보 포함 확인
	logOutput := buf.String()
	assert.Contains(t, logOutput, "panic")
	assert.Contains(t, logOutput, "error")
}

// TestRecoveryMiddleware_NoAffectOnNormalRequest는 정상 요청에 영향이 없는지 테스트
func TestRecoveryMiddleware_NoAffectOnNormalRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/normal", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}

// TestRecoveryMiddleware_ReturnsTraceID는 TraceID 포함 응답을 테스트
func TestRecoveryMiddleware_ReturnsTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic with trace")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "test-trace-123")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	var errResp ErrorResponseEnhanced
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	// TraceID가 응답에 포함되어야 함
	assert.Contains(t, errResp.TraceID, "test-trace-123")
}

// TestRecoveryMiddleware_LogsRequestContext는 요청 컨텍스트 로깅을 테스트
func TestRecoveryMiddleware_LogsRequestContext(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Post("/api/test", func(c *fiber.Ctx) error {
		panic("context panic")
	})

	req := httptest.NewRequest("POST", "/api/test", nil)
	_, err := app.Test(req, -1)
	require.NoError(t, err)

	// 로그에 요청 정보 포함 확인
	logOutput := buf.String()
	assert.Contains(t, logOutput, "/api/test")
	assert.Contains(t, logOutput, "POST")
}

// TestRecoveryMiddleware_UsesStructuredErrors는 구조화된 에러 사용을 테스트
func TestRecoveryMiddleware_UsesStructuredErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("structured error test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	var errResp ErrorResponseEnhanced
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	// 구조화된 에러 코드 확인
	assert.Equal(t, apperrors.ErrCodeInternal, errResp.Code)
}

// TestRecoveryMiddleware_HandlesPanicWithError는 error 타입 패닉을 테스트
func TestRecoveryMiddleware_HandlesPanicWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewRecoveryMiddleware(logger))
	app.Get("/panic-error", func(c *fiber.Ctx) error {
		panic(apperrors.NewError(apperrors.ErrCodeValidation, "validation failed"))
	})

	req := httptest.NewRequest("GET", "/panic-error", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	var errResp ErrorResponseEnhanced
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)

	// 원본 에러 코드가 유지되어야 함
	assert.Equal(t, apperrors.ErrCodeValidation, errResp.Code)
}
