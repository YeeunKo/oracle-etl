// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"github.com/gofiber/fiber/v2"
	"oracle-etl/internal/adapter/oracle"
)

// OracleHandler는 Oracle 관련 엔드포인트 핸들러입니다
type OracleHandler struct {
	repo oracle.Repository
}

// NewOracleHandler는 새로운 OracleHandler를 생성합니다
func NewOracleHandler(repo oracle.Repository) *OracleHandler {
	return &OracleHandler{
		repo: repo,
	}
}

// GetStatus는 Oracle 연결 상태를 반환합니다 (GET /api/oracle/status)
// 응답 예시:
//
//	{
//	  "connected": true,
//	  "database_version": "19.0.0.0.0",
//	  "instance_name": "ATP_HIGH",
//	  "pool_stats": {
//	    "active_connections": 2,
//	    "idle_connections": 8,
//	    "max_connections": 10
//	  }
//	}
func (h *OracleHandler) GetStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	status, err := h.repo.GetStatus(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "ORACLE_STATUS_ERROR",
			"message": "Oracle 상태 조회 실패",
			"error":   err.Error(),
		})
	}

	return c.JSON(status)
}
