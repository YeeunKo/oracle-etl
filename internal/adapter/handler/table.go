// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"github.com/gofiber/fiber/v2"
	"oracle-etl/internal/adapter/oracle"
	"oracle-etl/internal/domain"
)

// TableHandler는 테이블 관련 엔드포인트 핸들러입니다
type TableHandler struct {
	repo         oracle.Repository
	defaultOwner string
}

// NewTableHandler는 새로운 TableHandler를 생성합니다
func NewTableHandler(repo oracle.Repository, defaultOwner string) *TableHandler {
	return &TableHandler{
		repo:         repo,
		defaultOwner: defaultOwner,
	}
}

// GetTables는 테이블 목록을 반환합니다 (GET /api/tables)
// 응답 예시:
//
//	{
//	  "tables": [
//	    {
//	      "name": "VBRP",
//	      "owner": "SAPSR3",
//	      "row_count": 250000,
//	      "column_count": 274
//	    }
//	  ],
//	  "total": 40
//	}
func (h *TableHandler) GetTables(c *fiber.Ctx) error {
	ctx := c.Context()

	// 쿼리 파라미터에서 owner 추출 (없으면 기본값 사용)
	owner := c.Query("owner", h.defaultOwner)
	if owner == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "owner 파라미터가 필요합니다",
		})
	}

	tables, err := h.repo.GetTables(ctx, owner)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "TABLE_LIST_ERROR",
			"message": "테이블 목록 조회 실패",
			"error":   err.Error(),
		})
	}

	response := domain.TableListResponse{
		Tables: tables,
		Total:  len(tables),
	}

	return c.JSON(response)
}

// GetSampleData는 테이블의 샘플 데이터를 반환합니다 (GET /api/tables/:name/sample)
// 응답 예시:
//
//	{
//	  "table": "VBRP",
//	  "columns": ["MANDT", "VBELN", "POSNR", ...],
//	  "rows": [
//	    {"MANDT": "800", "VBELN": "0090000001", ...},
//	    ...
//	  ],
//	  "count": 100
//	}
func (h *TableHandler) GetSampleData(c *fiber.Ctx) error {
	ctx := c.Context()

	// 경로 파라미터에서 테이블 이름 추출
	tableName := c.Params("name")
	if tableName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "테이블 이름이 필요합니다",
		})
	}

	// 쿼리 파라미터에서 owner와 limit 추출
	owner := c.Query("owner", h.defaultOwner)
	if owner == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "owner 파라미터가 필요합니다",
		})
	}

	limit := c.QueryInt("limit", 100)
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // 최대 1000개로 제한
	}

	sample, err := h.repo.GetSampleData(ctx, owner, tableName, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "SAMPLE_DATA_ERROR",
			"message": "샘플 데이터 조회 실패",
			"error":   err.Error(),
		})
	}

	return c.JSON(sample)
}

// GetTableColumns는 테이블의 컬럼 정보를 반환합니다 (GET /api/tables/:name/columns)
func (h *TableHandler) GetTableColumns(c *fiber.Ctx) error {
	ctx := c.Context()

	// 경로 파라미터에서 테이블 이름 추출
	tableName := c.Params("name")
	if tableName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "테이블 이름이 필요합니다",
		})
	}

	// 쿼리 파라미터에서 owner 추출
	owner := c.Query("owner", h.defaultOwner)
	if owner == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "owner 파라미터가 필요합니다",
		})
	}

	columns, err := h.repo.GetTableColumns(ctx, owner, tableName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "COLUMN_INFO_ERROR",
			"message": "컬럼 정보 조회 실패",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"table":   tableName,
		"columns": columns,
		"count":   len(columns),
	})
}
