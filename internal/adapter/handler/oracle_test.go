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

func TestOracleHandler_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*oracle.MockRepository)
		expectedStatus int
		checkResponse  func(*testing.T, *domain.OracleStatus)
	}{
		{
			name: "정상 상태 반환",
			setupMock: func(m *oracle.MockRepository) {
				// 기본 mock 설정 사용
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, status *domain.OracleStatus) {
				assert.True(t, status.Connected)
				assert.Equal(t, "19.0.0.0.0", status.DatabaseVersion)
				assert.Equal(t, "ATP_HIGH", status.InstanceName)
				assert.Equal(t, 2, status.PoolStats.ActiveConnections)
				assert.Equal(t, 8, status.PoolStats.IdleConnections)
				assert.Equal(t, 10, status.PoolStats.MaxConnections)
			},
		},
		{
			name: "연결 실패 시 에러 반환",
			setupMock: func(m *oracle.MockRepository) {
				m.ShouldError = true
				m.ErrorMessage = "Oracle 연결 실패"
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
			handler := NewOracleHandler(mockRepo)

			// Fiber 앱 생성
			app := fiber.New()
			app.Get("/api/oracle/status", handler.GetStatus)

			// 요청 생성
			req := httptest.NewRequest(http.MethodGet, "/api/oracle/status", nil)

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

				var status domain.OracleStatus
				err = json.Unmarshal(body, &status)
				require.NoError(t, err)

				tt.checkResponse(t, &status)
			}

			// Mock 메서드 호출 확인
			assert.True(t, mockRepo.GetStatusCalled)
		})
	}
}
