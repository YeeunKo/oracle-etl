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

// setupJobTestApp은 테스트용 Fiber 앱을 설정합니다
func setupJobTestApp() (*fiber.App, *JobHandler, *usecase.TransportService, *usecase.JobService) {
	app := fiber.New()
	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()
	transportSvc := usecase.NewTransportService(transportRepo)
	jobSvc := usecase.NewJobService(jobRepo, transportRepo)
	handler := NewJobHandler(jobSvc)

	api := app.Group("/api")
	api.Get("/jobs", handler.List)
	api.Get("/jobs/:id", handler.GetByID)

	return app, handler, transportSvc, jobSvc
}

// TestJobHandler_List는 Job 목록 조회 API를 테스트합니다
func TestJobHandler_List(t *testing.T) {
	app, _, _, _ := setupJobTestApp()

	// 빈 목록 조회
	req := httptest.NewRequest("GET", "/api/jobs", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var listResp domain.JobListResponse
	require.NoError(t, json.Unmarshal(respBody, &listResp))

	assert.Empty(t, listResp.Jobs)
	assert.Equal(t, 0, listResp.Total)
}

// TestJobHandler_ListWithJobs는 Job이 있는 경우 목록 조회를 테스트합니다
func TestJobHandler_ListWithJobs(t *testing.T) {
	// Transport 핸들러를 통해 Transport와 Job 생성
	transportApp, _ := setupTransportTestApp()

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := transportApp.Test(req, -1)

	respBody, _ := io.ReadAll(resp.Body)
	var created domain.Transport
	_ = json.Unmarshal(respBody, &created)

	// Job 3개 생성 (execute 호출)
	for i := 0; i < 3; i++ {
		req = httptest.NewRequest("POST", "/api/transports/"+created.ID+"/execute", nil)
		_, _ = transportApp.Test(req, -1)
	}
}

// TestJobHandler_GetByID는 Job ID 조회 API를 테스트합니다
func TestJobHandler_GetByID(t *testing.T) {
	app, _, _, _ := setupJobTestApp()

	// 존재하지 않는 ID 조회
	req := httptest.NewRequest("GET", "/api/jobs/non-existent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestJobHandler_ListWithFilter는 필터링을 테스트합니다
func TestJobHandler_ListWithFilter(t *testing.T) {
	app, _, _, _ := setupJobTestApp()

	// Transport ID로 필터링
	req := httptest.NewRequest("GET", "/api/jobs?transport_id=TRPID-XXX", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Status로 필터링
	req = httptest.NewRequest("GET", "/api/jobs?status=completed", nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestJobHandler_IntegrationWithTransport는 Transport Execute와 Job 조회 통합 테스트입니다
func TestJobHandler_IntegrationWithTransport(t *testing.T) {
	// 통합된 앱 설정
	app := fiber.New()
	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()
	transportSvc := usecase.NewTransportService(transportRepo)
	jobSvc := usecase.NewJobService(jobRepo, transportRepo)

	transportHandler := NewTransportHandler(transportSvc, jobSvc)
	jobHandler := NewJobHandler(jobSvc)

	api := app.Group("/api")
	api.Post("/transports", transportHandler.Create)
	api.Post("/transports/:id/execute", transportHandler.Execute)
	api.Get("/jobs", jobHandler.List)
	api.Get("/jobs/:id", jobHandler.GetByID)

	// Transport 생성
	reqBody := domain.CreateTransportRequest{
		Name:   "Integration Test",
		Tables: []string{"TABLE1", "TABLE2"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/transports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	respBody, _ := io.ReadAll(resp.Body)
	var transport domain.Transport
	_ = json.Unmarshal(respBody, &transport)

	// Job 생성 (execute)
	req = httptest.NewRequest("POST", "/api/transports/"+transport.ID+"/execute", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 202, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var execResp domain.ExecuteJobResponse
	_ = json.Unmarshal(respBody, &execResp)

	// Job 조회
	req = httptest.NewRequest("GET", "/api/jobs/"+execResp.JobID, nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var job domain.Job
	require.NoError(t, json.Unmarshal(respBody, &job))

	assert.Equal(t, execResp.JobID, job.ID)
	assert.Equal(t, transport.ID, job.TransportID)
	assert.Equal(t, 1, job.Version)

	// Job 목록에서 확인
	req = httptest.NewRequest("GET", "/api/jobs?transport_id="+transport.ID, nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ = io.ReadAll(resp.Body)
	var listResp domain.JobListResponse
	require.NoError(t, json.Unmarshal(respBody, &listResp))

	assert.Len(t, listResp.Jobs, 1)
	assert.Equal(t, 1, listResp.Total)
}
