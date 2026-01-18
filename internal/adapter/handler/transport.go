// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/usecase"
)

// TransportHandler는 Transport 관련 HTTP 핸들러입니다
type TransportHandler struct {
	transportSvc *usecase.TransportService
	jobSvc       *usecase.JobService
}

// NewTransportHandler는 새로운 TransportHandler를 생성합니다
func NewTransportHandler(transportSvc *usecase.TransportService, jobSvc *usecase.JobService) *TransportHandler {
	return &TransportHandler{
		transportSvc: transportSvc,
		jobSvc:       jobSvc,
	}
}

// Create는 새로운 Transport를 생성합니다
// POST /api/transports
func (h *TransportHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateTransportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "요청 본문을 파싱할 수 없습니다: " + err.Error(),
		})
	}

	transport, err := h.transportSvc.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(transport)
}

// List는 Transport 목록을 조회합니다
// GET /api/transports
func (h *TransportHandler) List(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	resp, err := h.transportSvc.List(c.Context(), offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
	}

	return c.JSON(resp)
}

// GetByID는 ID로 Transport를 조회합니다
// GET /api/transports/:id
func (h *TransportHandler) GetByID(c *fiber.Ctx) error {
	// fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	id := strings.Clone(c.Params("id"))

	transport, err := h.transportSvc.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "TRANSPORT_NOT_FOUND",
			"message": err.Error(),
		})
	}

	return c.JSON(transport)
}

// Delete는 Transport를 삭제합니다
// DELETE /api/transports/:id
func (h *TransportHandler) Delete(c *fiber.Ctx) error {
	// fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	id := strings.Clone(c.Params("id"))

	if err := h.transportSvc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "TRANSPORT_NOT_FOUND",
			"message": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Execute는 Transport를 실행하고 새 Job을 생성합니다
// POST /api/transports/:id/execute
func (h *TransportHandler) Execute(c *fiber.Ctx) error {
	// fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	transportID := strings.Clone(c.Params("id"))

	// Transport 존재 여부 및 실행 가능 여부 확인
	transport, err := h.transportSvc.GetByID(c.Context(), transportID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "TRANSPORT_NOT_FOUND",
			"message": err.Error(),
		})
	}

	if !transport.CanExecute() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "TRANSPORT_NOT_EXECUTABLE",
			"message": "Transport가 이미 실행 중이거나 비활성화 상태입니다",
		})
	}

	// 새 Job 생성
	job, err := h.jobSvc.CreateJob(c.Context(), transportID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "JOB_CREATION_FAILED",
			"message": err.Error(),
		})
	}

	// 응답 생성
	resp := domain.ExecuteJobResponse{
		JobID:       job.ID,
		TransportID: job.TransportID,
		Version:     job.Version,
		Status:      job.Status,
	}

	// Note: 실제 ETL 실행은 비동기로 처리됩니다
	// TODO: M5에서 goroutine으로 Oracle 추출 + GCS 업로드 실행

	return c.Status(fiber.StatusAccepted).JSON(resp)
}
