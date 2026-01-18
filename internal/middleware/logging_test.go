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
)

// LogEntry는 JSON 로그 엔트리 구조체
type LogEntry struct {
	Level     string `json:"level"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Latency   string `json:"latency"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

// TestLoggingMiddleware_LogsRequest는 요청 로깅을 테스트
func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	// 버퍼로 로그 캡처
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewLoggingMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// 로그 검증
	var logEntry LogEntry
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "info", logEntry.Level)
	assert.Equal(t, "GET", logEntry.Method)
	assert.Equal(t, "/test", logEntry.Path)
	assert.Equal(t, 200, logEntry.Status)
	assert.NotEmpty(t, logEntry.Latency)
}

// TestLoggingMiddleware_ErrorLogging은 에러 응답 로깅을 테스트
func TestLoggingMiddleware_ErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewLoggingMiddleware(logger))
	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var logEntry LogEntry
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "error", logEntry.Level)
	assert.Equal(t, 500, logEntry.Status)
}

// TestLoggingMiddleware_RequestID는 요청 ID 생성을 테스트
func TestLoggingMiddleware_RequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	app := fiber.New()
	app.Use(NewLoggingMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req, -1)
	require.NoError(t, err)

	var logEntry LogEntry
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.NotEmpty(t, logEntry.RequestID)
}
