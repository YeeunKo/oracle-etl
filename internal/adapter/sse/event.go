// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"encoding/json"
	"fmt"
	"time"
)

// 이벤트 타입 상수
const (
	// EventTypeProgress는 진행률 이벤트 타입입니다
	EventTypeProgress = "progress"
	// EventTypeStatus는 상태 변경 이벤트 타입입니다
	EventTypeStatus = "status"
	// EventTypeError는 에러 이벤트 타입입니다
	EventTypeError = "error"
	// EventTypeComplete는 완료 이벤트 타입입니다
	EventTypeComplete = "complete"
)

// 상태 상수
const (
	// StatusRunning은 실행 중 상태입니다
	StatusRunning = "running"
	// StatusCompleted는 완료 상태입니다
	StatusCompleted = "completed"
	// StatusFailed는 실패 상태입니다
	StatusFailed = "failed"
)

// SSEEvent는 SSE 이벤트의 기본 구조입니다
type SSEEvent struct {
	Event string      `json:"-"`    // event type: progress, status, error, complete
	Data  interface{} `json:"data"` // 이벤트 데이터
}

// Format은 SSE 이벤트를 SSE 형식 문자열로 변환합니다
// 형식: event: {type}\ndata: {json}\n\n
func (e *SSEEvent) Format() (string, error) {
	data, err := json.Marshal(e.Data)
	if err != nil {
		return "", fmt.Errorf("이벤트 데이터 직렬화 실패: %w", err)
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", e.Event, string(data)), nil
}

// ProgressEvent는 진행률 이벤트입니다
type ProgressEvent struct {
	TransportID     string  `json:"transport_id"`     // Transport ID
	JobID           string  `json:"job_id"`           // Job ID
	Table           string  `json:"table"`            // 현재 처리 중인 테이블
	RowsProcessed   int64   `json:"rows_processed"`   // 처리된 row 수
	RowsTotal       int64   `json:"rows_total"`       // 총 row 수 (-1 if unknown)
	RowsPerSecond   float64 `json:"rows_per_second"`  // 초당 처리 row 수
	BytesWritten    int64   `json:"bytes_written"`    // 작성된 바이트 수
	ProgressPercent float64 `json:"progress_percent"` // 진행률 (0-100)
}

// CalculateProgressPercent는 진행률을 계산합니다
func (e *ProgressEvent) CalculateProgressPercent() {
	if e.RowsTotal <= 0 {
		e.ProgressPercent = 0.0
		return
	}
	e.ProgressPercent = float64(e.RowsProcessed) / float64(e.RowsTotal) * 100.0
}

// NewProgressEvent는 새로운 ProgressEvent를 생성합니다
func NewProgressEvent(transportID, jobID, table string, rowsProcessed, rowsTotal, bytesWritten int64) *ProgressEvent {
	event := &ProgressEvent{
		TransportID:   transportID,
		JobID:         jobID,
		Table:         table,
		RowsProcessed: rowsProcessed,
		RowsTotal:     rowsTotal,
		BytesWritten:  bytesWritten,
	}
	event.CalculateProgressPercent()
	return event
}

// StatusEvent는 상태 변경 이벤트입니다
type StatusEvent struct {
	TransportID string `json:"transport_id"` // Transport ID
	JobID       string `json:"job_id"`       // Job ID
	Status      string `json:"status"`       // 상태: running, completed, failed
	Message     string `json:"message"`      // 상태 메시지
}

// NewStatusEvent는 새로운 StatusEvent를 생성합니다
func NewStatusEvent(transportID, jobID, status, message string) *StatusEvent {
	return &StatusEvent{
		TransportID: transportID,
		JobID:       jobID,
		Status:      status,
		Message:     message,
	}
}

// ErrorEvent는 에러 이벤트입니다
type ErrorEvent struct {
	TransportID string `json:"transport_id"` // Transport ID
	JobID       string `json:"job_id"`       // Job ID
	Table       string `json:"table"`        // 에러 발생 테이블
	Code        string `json:"code"`         // 에러 코드
	Message     string `json:"message"`      // 에러 메시지
	Timestamp   string `json:"timestamp"`    // 에러 발생 시간 (ISO8601)
}

// NewErrorEvent는 새로운 ErrorEvent를 생성합니다
func NewErrorEvent(transportID, jobID, table, code, message string) *ErrorEvent {
	return &ErrorEvent{
		TransportID: transportID,
		JobID:       jobID,
		Table:       table,
		Code:        code,
		Message:     message,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
}

// CompleteEvent는 완료 이벤트입니다
type CompleteEvent struct {
	TransportID   string  `json:"transport_id"`   // Transport ID
	JobID         string  `json:"job_id"`         // Job ID
	TotalRows     int64   `json:"total_rows"`     // 총 처리 row 수
	TotalBytes    int64   `json:"total_bytes"`    // 총 바이트 수
	DurationMs    int64   `json:"duration_ms"`    // 실행 시간 (밀리초)
	TablesCount   int     `json:"tables_count"`   // 처리된 테이블 수
	RowsPerSecond float64 `json:"rows_per_second"` // 평균 초당 처리 row 수
}

// NewCompleteEvent는 새로운 CompleteEvent를 생성합니다
func NewCompleteEvent(transportID, jobID string, totalRows, totalBytes, durationMs int64, tablesCount int) *CompleteEvent {
	var rowsPerSecond float64
	if durationMs > 0 {
		// ms를 초로 변환하여 계산
		rowsPerSecond = float64(totalRows) / (float64(durationMs) / 1000.0)
	}
	return &CompleteEvent{
		TransportID:   transportID,
		JobID:         jobID,
		TotalRows:     totalRows,
		TotalBytes:    totalBytes,
		DurationMs:    durationMs,
		TablesCount:   tablesCount,
		RowsPerSecond: rowsPerSecond,
	}
}
