package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/domain"
)

func TestTableHandler_GetTables(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*oracle.MockRepository)
		expectedStatus int
		checkResponse  func(*testing.T, *domain.TableListResponse)
	}{
		{
			name:        "기본 owner로 테이블 목록 조회",
			queryParams: "",
			setupMock: func(m *oracle.MockRepository) {
				// 기본 mock 설정 사용
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *domain.TableListResponse) {
				assert.Equal(t, 2, resp.Total)
				require.Len(t, resp.Tables, 2)
				assert.Equal(t, "VBRP", resp.Tables[0].Name)
				assert.Equal(t, int64(250000), resp.Tables[0].RowCount)
			},
		},
		{
			name:        "특정 owner로 테이블 목록 조회",
			queryParams: "?owner=TESTOWNER",
			setupMock: func(m *oracle.MockRepository) {
				m.MockTables = []domain.TableInfo{
					{Name: "TEST_TABLE", Owner: "TESTOWNER", RowCount: 1000, ColumnCount: 10},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *domain.TableListResponse) {
				assert.Equal(t, 1, resp.Total)
			},
		},
		{
			name:        "조회 실패 시 에러 반환",
			queryParams: "",
			setupMock: func(m *oracle.MockRepository) {
				m.ShouldError = true
				m.ErrorMessage = "DB 연결 실패"
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock 저장소 생성
			mockRepo := oracle.NewMockRepository()
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			// 핸들러 생성
			handler := NewTableHandler(mockRepo, "SAPSR3")

			// Fiber 앱 생성
			app := fiber.New()
			app.Get("/api/tables", handler.GetTables)

			// 요청 생성
			req := httptest.NewRequest(http.MethodGet, "/api/tables"+tt.queryParams, nil)

			// 요청 실행
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// 상태 코드 확인
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// 응답 확인
			if tt.checkResponse != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var response domain.TableListResponse
				err = json.Unmarshal(body, &response)
				require.NoError(t, err)

				tt.checkResponse(t, &response)
			}

			// Mock 메서드 호출 확인
			assert.True(t, mockRepo.GetTablesCalled)
		})
	}
}

func TestTableHandler_GetSampleData(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		queryParams    string
		setupMock      func(*oracle.MockRepository)
		expectedStatus int
		checkResponse  func(*testing.T, *domain.SampleData)
	}{
		{
			name:        "기본 샘플 데이터 조회",
			tableName:   "VBRP",
			queryParams: "",
			setupMock: func(m *oracle.MockRepository) {
				// 기본 mock 설정 사용
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, sample *domain.SampleData) {
				assert.Equal(t, "VBRP", sample.TableName)
				assert.Contains(t, sample.Columns, "MANDT")
				assert.Equal(t, 2, sample.Count)
			},
		},
		{
			name:        "limit 파라미터로 조회",
			tableName:   "VBRP",
			queryParams: "?limit=50",
			setupMock: func(m *oracle.MockRepository) {
				// 기본 mock 설정 사용
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, sample *domain.SampleData) {
				assert.Equal(t, "VBRP", sample.TableName)
			},
		},
		{
			name:        "조회 실패 시 에러 반환",
			tableName:   "VBRP",
			queryParams: "",
			setupMock: func(m *oracle.MockRepository) {
				m.ShouldError = true
				m.ErrorMessage = "테이블 없음"
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock 저장소 생성
			mockRepo := oracle.NewMockRepository()
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			// 핸들러 생성
			handler := NewTableHandler(mockRepo, "SAPSR3")

			// Fiber 앱 생성
			app := fiber.New()
			app.Get("/api/tables/:name/sample", handler.GetSampleData)

			// 요청 생성
			req := httptest.NewRequest(http.MethodGet, "/api/tables/"+tt.tableName+"/sample"+tt.queryParams, nil)

			// 요청 실행
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// 상태 코드 확인
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// 응답 확인
			if tt.checkResponse != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var sample domain.SampleData
				err = json.Unmarshal(body, &sample)
				require.NoError(t, err)

				tt.checkResponse(t, &sample)
			}

			// Mock 메서드 호출 확인
			assert.True(t, mockRepo.GetSampleCalled)
		})
	}
}

func TestTableHandler_GetTables_NoOwner(t *testing.T) {
	// Mock 저장소 생성 (기본 owner 없음)
	mockRepo := oracle.NewMockRepository()
	handler := NewTableHandler(mockRepo, "") // 기본 owner 없음

	// Fiber 앱 생성
	app := fiber.New()
	app.Get("/api/tables", handler.GetTables)

	// 요청 생성 (owner 파라미터 없음)
	req := httptest.NewRequest(http.MethodGet, "/api/tables", nil)

	// 요청 실행
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 상태 코드 확인 (400 Bad Request)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
