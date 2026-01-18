// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler는 헬스 체크 엔드포인트 핸들러입니다
type HealthHandler struct {
	version string
}

// HealthResponse는 /api/health 응답 구조체입니다
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// NewHealthHandler는 새로운 HealthHandler를 생성합니다
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

// Check는 시스템 상태를 확인하고 JSON 응답을 반환합니다
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   h.version,
	}

	return c.JSON(response)
}
