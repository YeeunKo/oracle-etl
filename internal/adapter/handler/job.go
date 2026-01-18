// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/usecase"
)

// JobHandler는 Job 관련 HTTP 핸들러입니다
type JobHandler struct {
	jobSvc *usecase.JobService
}

// NewJobHandler는 새로운 JobHandler를 생성합니다
func NewJobHandler(jobSvc *usecase.JobService) *JobHandler {
	return &JobHandler{
		jobSvc: jobSvc,
	}
}

// List는 Job 목록을 조회합니다
// GET /api/jobs
func (h *JobHandler) List(c *fiber.Ctx) error {
	filter := domain.DefaultJobListFilter()

	// Query 파라미터 파싱 - fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	if transportID := c.Query("transport_id"); transportID != "" {
		filter.TransportID = strings.Clone(transportID)
	}
	if status := c.Query("status"); status != "" {
		filter.Status = domain.JobStatus(strings.Clone(status))
	}
	if offset, err := strconv.Atoi(c.Query("offset", "0")); err == nil {
		filter.Offset = offset
	}
	if limit, err := strconv.Atoi(c.Query("limit", "20")); err == nil {
		filter.Limit = limit
	}

	resp, err := h.jobSvc.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		})
	}

	return c.JSON(resp)
}

// GetByID는 ID로 Job을 조회합니다
// GET /api/jobs/:id
func (h *JobHandler) GetByID(c *fiber.Ctx) error {
	// fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	id := strings.Clone(c.Params("id"))

	job, err := h.jobSvc.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "JOB_NOT_FOUND",
			"message": err.Error(),
		})
	}

	return c.JSON(job)
}
