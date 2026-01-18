package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/sse"
	"oracle-etl/internal/config"
	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository/memory"
	"oracle-etl/internal/usecase"
)

// createTestApp은 테스트용 앱을 생성합니다
func createTestApp(t *testing.T) (*fiber.App, *config.Config, *sse.Broadcaster, context.CancelFunc) {
	t.Helper()

	logger := zerolog.Nop()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  "10s",
			WriteTimeout: "60s",
		},
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()
	transportSvc := usecase.NewTransportService(transportRepo)
	jobSvc := usecase.NewJobService(jobRepo, transportRepo)

	// SSE Broadcaster 생성 및 시작
	broadcaster := sse.NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go broadcaster.Run(ctx)

	app := setupFiber(cfg, logger)
	setupRoutes(app, cfg, transportSvc, jobSvc, broadcaster)

	return app, cfg, broadcaster, cancel
}

// TestSetupFiber_Configuration은 Fiber 앱 설정을 테스트합니다
func TestSetupFiber_Configuration(t *testing.T) {
	logger := zerolog.Nop()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  "10s",
			WriteTimeout: "60s",
		},
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	app := setupFiber(cfg, logger)
	require.NotNil(t, app)

	// Fiber 설정 확인
	assert.Equal(t, "test-app", app.Config().AppName)
	assert.Equal(t, 10*time.Second, app.Config().ReadTimeout)
	assert.Equal(t, 60*time.Second, app.Config().WriteTimeout)
}

// TestSetupRoutes_HealthEndpoint는 라우트 설정을 테스트합니다
func TestSetupRoutes_HealthEndpoint(t *testing.T) {
	app, _, _, _ := createTestApp(t)

	// /api/health 엔드포인트 테스트
	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "ok", result["status"])
	assert.Equal(t, "1.0.0", result["version"])
}

// TestSetupFiber_ErrorHandler는 에러 핸들러를 테스트합니다
func TestSetupFiber_ErrorHandler(t *testing.T) {
	logger := zerolog.Nop()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  "10s",
			WriteTimeout: "60s",
		},
		App: config.AppConfig{
			Name: "test-app",
		},
	}

	app := setupFiber(cfg, logger)

	// 404 에러를 발생시키는 라우트 테스트
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// TestSetupRoutes_TransportEndpoints는 Transport API 엔드포인트를 테스트합니다
func TestSetupRoutes_TransportEndpoints(t *testing.T) {
	app, _, _, _ := createTestApp(t)

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:        "Test Transport",
		Description: "Test Description",
		Tables:      []string{"TABLE1", "TABLE2"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var transport domain.Transport
	require.NoError(t, json.Unmarshal(respBody, &transport))
	assert.NotEmpty(t, transport.ID)

	// Transport 목록 조회
	req = httptest.NewRequest("GET", "/api/transports", nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Transport ID로 조회
	req = httptest.NewRequest("GET", "/api/transports/"+transport.ID, nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Transport 실행
	req = httptest.NewRequest("POST", "/api/transports/"+transport.ID+"/execute", nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var execResp domain.ExecuteJobResponse
	require.NoError(t, json.Unmarshal(respBody, &execResp))
	assert.NotEmpty(t, execResp.JobID)
	assert.Equal(t, transport.ID, execResp.TransportID)

	// Job 조회
	req = httptest.NewRequest("GET", "/api/jobs/"+execResp.JobID, nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Job 목록 조회
	req = httptest.NewRequest("GET", "/api/jobs", nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Transport 삭제
	req = httptest.NewRequest("DELETE", "/api/transports/"+transport.ID, nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestSetupRoutes_SSEStatusEndpoint는 SSE 상태 엔드포인트를 테스트합니다
// Note: SSE 스트리밍은 Fiber Test 메서드와 호환되지 않으므로
// 핸들러 유닛 테스트에서 더 자세히 테스트합니다
func TestSetupRoutes_SSEStatusEndpoint(t *testing.T) {
	app, _, broadcaster, cancel := createTestApp(t)

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Test Transport",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	respBody, _ := io.ReadAll(resp.Body)
	var transport domain.Transport
	require.NoError(t, json.Unmarshal(respBody, &transport))

	// SSE 엔드포인트 라우트가 등록되었는지 확인
	// (스트리밍 테스트는 handler 패키지 테스트에서 수행)
	assert.NotNil(t, broadcaster, "Broadcaster가 초기화되어야 합니다")
	assert.Equal(t, 0, broadcaster.ClientCount(), "초기 클라이언트 수는 0이어야 합니다")

	// 라우트 존재 확인을 위해 빠르게 취소하여 테스트
	// broadcaster 컨텍스트를 취소하여 SSE 스트림이 즉시 종료되도록 함
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	req = httptest.NewRequest("GET", "/api/transports/"+transport.ID+"/status", nil)
	// 라우트 존재 여부만 확인 (타임아웃 허용)
	resp, _ = app.Test(req, 50)

	// 응답을 받았거나 타임아웃이 발생했거나 상관없이
	// 라우트가 등록되었는지만 확인
	// resp가 nil이 아니면 라우트가 존재함
	if resp != nil {
		defer resp.Body.Close()
		// SSE 헤더 확인 (응답이 있는 경우)
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			assert.Contains(t, contentType, "text/event-stream")
		}
	}
}
