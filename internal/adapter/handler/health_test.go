package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthHandler_Success는 정상적인 health check 응답을 테스트
func TestHealthHandler_Success(t *testing.T) {
	// Fiber 앱 설정
	app := fiber.New()
	handler := NewHealthHandler("1.0.0")
	app.Get("/api/health", handler.Check)

	// 요청 생성 및 실행
	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// 상태 코드 확인
	assert.Equal(t, 200, resp.StatusCode)

	// 응답 본문 확인
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResp HealthResponse
	err = json.Unmarshal(body, &healthResp)
	require.NoError(t, err)

	assert.Equal(t, "ok", healthResp.Status)
	assert.Equal(t, "1.0.0", healthResp.Version)
	assert.False(t, healthResp.Timestamp.IsZero())
}

// TestHealthHandler_ContentType은 Content-Type 헤더를 검증
func TestHealthHandler_ContentType(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("1.0.0")
	app.Get("/api/health", handler.Check)

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

// TestHealthHandler_TimestampFormat은 타임스탬프가 ISO8601 형식인지 확인
func TestHealthHandler_TimestampFormat(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("1.0.0")
	app.Get("/api/health", handler.Check)

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResp HealthResponse
	err = json.Unmarshal(body, &healthResp)
	require.NoError(t, err)

	// 타임스탬프가 현재 시간과 가까운지 확인 (5초 이내)
	now := time.Now()
	diff := now.Sub(healthResp.Timestamp)
	assert.Less(t, diff.Abs(), 5*time.Second)
}
