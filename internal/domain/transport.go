// Package domain은 ETL 파이프라인의 핵심 도메인 모델을 정의합니다.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// TransportStatus는 Transport의 상태를 나타냅니다
type TransportStatus string

const (
	// TransportStatusIdle은 대기 상태입니다
	TransportStatusIdle TransportStatus = "idle"
	// TransportStatusRunning은 실행 중 상태입니다
	TransportStatusRunning TransportStatus = "running"
	// TransportStatusFailed는 실패 상태입니다
	TransportStatusFailed TransportStatus = "failed"
)

// CronSchedule은 스케줄 설정을 나타냅니다
type CronSchedule struct {
	Expression string `json:"expression"` // cron 표현식
	Timezone   string `json:"timezone"`   // 시간대 (예: "Asia/Seoul")
}

// Transport는 ETL 전송 구성을 나타냅니다
type Transport struct {
	ID          string          `json:"id"`                    // TRPID-xxx 형식
	Name        string          `json:"name"`                  // Transport 이름
	Description string          `json:"description,omitempty"` // 설명
	Tables      []string        `json:"tables"`                // 대상 테이블 목록
	Enabled     bool            `json:"enabled"`               // 활성화 여부
	Schedule    *CronSchedule   `json:"schedule,omitempty"`    // 선택적 cron 스케줄
	Status      TransportStatus `json:"status"`                // 현재 상태
	CreatedAt   time.Time       `json:"created_at"`            // 생성 시간
	UpdatedAt   time.Time       `json:"updated_at"`            // 수정 시간
}

// GenerateTransportID는 새로운 Transport ID를 생성합니다
// 형식: TRPID-{uuid의 처음 8자}
func GenerateTransportID(uuidStr string) string {
	// uuid가 하이픈을 포함하는 경우 제거
	cleanUUID := strings.ReplaceAll(uuidStr, "-", "")
	if len(cleanUUID) >= 8 {
		return fmt.Sprintf("TRPID-%s", cleanUUID[:8])
	}
	return fmt.Sprintf("TRPID-%s", cleanUUID)
}

// NewTransport는 새로운 Transport를 생성합니다
func NewTransport(id, name, description string, tables []string) *Transport {
	now := time.Now().UTC()
	return &Transport{
		ID:          id,
		Name:        name,
		Description: description,
		Tables:      tables,
		Enabled:     true,
		Status:      TransportStatusIdle,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate는 Transport의 유효성을 검사합니다
func (t *Transport) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transport ID가 비어있습니다")
	}
	if t.Name == "" {
		return fmt.Errorf("transport 이름이 비어있습니다")
	}
	if len(t.Tables) == 0 {
		return fmt.Errorf("테이블 목록이 비어있습니다")
	}
	return nil
}

// CanExecute는 Transport가 실행 가능한지 확인합니다
func (t *Transport) CanExecute() bool {
	return t.Enabled && t.Status != TransportStatusRunning
}

// CreateTransportRequest는 Transport 생성 요청 DTO입니다
type CreateTransportRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tables      []string `json:"tables"`
}

// Validate는 요청의 유효성을 검사합니다
func (r *CreateTransportRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name은 필수입니다")
	}
	if len(r.Tables) == 0 {
		return fmt.Errorf("tables는 최소 1개 이상이어야 합니다")
	}
	return nil
}

// TransportListResponse는 Transport 목록 응답입니다
type TransportListResponse struct {
	Transports []Transport `json:"transports"`
	Total      int         `json:"total"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
}
