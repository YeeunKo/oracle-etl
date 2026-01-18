package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository/memory"
	"oracle-etl/internal/usecase"
)

// setupTransportTestApp은 테스트용 Fiber 앱을 설정합니다
func setupTransportTestApp() (*fiber.App, *TransportHandler) {
	app := fiber.New()
	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()
	transportSvc := usecase.NewTransportService(transportRepo)
	jobSvc := usecase.NewJobService(jobRepo, transportRepo)
	handler := NewTransportHandler(transportSvc, jobSvc)

	api := app.Group("/api")
	api.Post("/transports", handler.Create)
	api.Get("/transports", handler.List)
	api.Get("/transports/:id", handler.GetByID)
	api.Delete("/transports/:id", handler.Delete)
	api.Post("/transports/:id/execute", handler.Execute)

	return app, handler
}

// TestTransportHandler_Create는 Transport 생성 API를 테스트합니다
func TestTransportHandler_Create(t *testing.T) {
	app, _ := setupTransportTestApp()

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
	assert.Equal(t, 201, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var transport domain.Transport
	require.NoError(t, json.Unmarshal(respBody, &transport))

	assert.NotEmpty(t, transport.ID)
	assert.Equal(t, "Test Transport", transport.Name)
	assert.Equal(t, []string{"TABLE1", "TABLE2"}, transport.Tables)
}

// TestTransportHandler_CreateValidationError는 유효성 검사 실패를 테스트합니다
func TestTransportHandler_CreateValidationError(t *testing.T) {
	app, _ := setupTransportTestApp()

	// 빈 이름
	reqBody := domain.CreateTransportRequest{
		Name:   "",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestTransportHandler_List는 Transport 목록 조회 API를 테스트합니다
func TestTransportHandler_List(t *testing.T) {
	app, _ := setupTransportTestApp()

	// Transport 3개 생성
	for i := 0; i < 3; i++ {
		reqBody := domain.CreateTransportRequest{
			Name:   "Transport",
			Tables: []string{"TABLE"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		_, _ = app.Test(req, -1)
	}

	// 목록 조회
	req := httptest.NewRequest("GET", "/api/transports?limit=10&offset=0", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var listResp domain.TransportListResponse
	require.NoError(t, json.Unmarshal(respBody, &listResp))

	assert.Len(t, listResp.Transports, 3)
	assert.Equal(t, 3, listResp.Total)
}

// TestTransportHandler_GetByID는 Transport ID 조회 API를 테스트합니다
func TestTransportHandler_GetByID(t *testing.T) {
	app, _ := setupTransportTestApp()

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	respBody, _ := io.ReadAll(resp.Body)
	var created domain.Transport
	_ = json.Unmarshal(respBody, &created)

	// ID로 조회
	req = httptest.NewRequest("GET", "/api/transports/"+created.ID, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var found domain.Transport
	require.NoError(t, json.Unmarshal(respBody, &found))
	assert.Equal(t, created.ID, found.ID)
}

// TestTransportHandler_GetByIDNotFound는 존재하지 않는 ID 조회를 테스트합니다
func TestTransportHandler_GetByIDNotFound(t *testing.T) {
	app, _ := setupTransportTestApp()

	req := httptest.NewRequest("GET", "/api/transports/non-existent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestTransportHandler_Delete는 Transport 삭제 API를 테스트합니다
func TestTransportHandler_Delete(t *testing.T) {
	app, _ := setupTransportTestApp()

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	respBody, _ := io.ReadAll(resp.Body)
	var created domain.Transport
	_ = json.Unmarshal(respBody, &created)

	// 삭제
	req = httptest.NewRequest("DELETE", "/api/transports/"+created.ID, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	// 삭제된 Transport 조회
	req = httptest.NewRequest("GET", "/api/transports/"+created.ID, nil)
	resp, _ = app.Test(req, -1)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestTransportHandler_Execute는 Transport 실행 API를 테스트합니다
func TestTransportHandler_Execute(t *testing.T) {
	app, _ := setupTransportTestApp()

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	respBody, _ := io.ReadAll(resp.Body)
	var created domain.Transport
	_ = json.Unmarshal(respBody, &created)

	// 실행
	req = httptest.NewRequest("POST", "/api/transports/"+created.ID+"/execute", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 202, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var execResp domain.ExecuteJobResponse
	require.NoError(t, json.Unmarshal(respBody, &execResp))

	assert.NotEmpty(t, execResp.JobID)
	assert.Equal(t, created.ID, execResp.TransportID)
	assert.Equal(t, 1, execResp.Version)
	assert.Equal(t, domain.JobStatusPending, execResp.Status)
}
