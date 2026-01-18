// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressEvent_JSON(t *testing.T) {
	// 진행률 이벤트가 올바르게 JSON으로 직렬화되는지 테스트
	event := ProgressEvent{
		TransportID:     "TRPID-12345678",
		JobID:           "JOB-20260118-120000-abc123",
		Table:           "VBRP",
		RowsProcessed:   10000,
		RowsTotal:       40000,
		RowsPerSecond:   5000.5,
		BytesWritten:    1048576,
		ProgressPercent: 25.0,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "TRPID-12345678", decoded["transport_id"])
	assert.Equal(t, "JOB-20260118-120000-abc123", decoded["job_id"])
	assert.Equal(t, "VBRP", decoded["table"])
	assert.Equal(t, float64(10000), decoded["rows_processed"])
	assert.Equal(t, float64(40000), decoded["rows_total"])
	assert.Equal(t, 5000.5, decoded["rows_per_second"])
	assert.Equal(t, float64(1048576), decoded["bytes_written"])
	assert.Equal(t, 25.0, decoded["progress_percent"])
}

func TestProgressEvent_CalculateProgressPercent(t *testing.T) {
	tests := []struct {
		name          string
		rowsProcessed int64
		rowsTotal     int64
		expected      float64
	}{
		{
			name:          "정상 진행률 계산",
			rowsProcessed: 10000,
			rowsTotal:     40000,
			expected:      25.0,
		},
		{
			name:          "총 row 수가 0인 경우",
			rowsProcessed: 10000,
			rowsTotal:     0,
			expected:      0.0,
		},
		{
			name:          "총 row 수가 -1인 경우 (unknown)",
			rowsProcessed: 10000,
			rowsTotal:     -1,
			expected:      0.0,
		},
		{
			name:          "완료 (100%)",
			rowsProcessed: 40000,
			rowsTotal:     40000,
			expected:      100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &ProgressEvent{
				RowsProcessed: tt.rowsProcessed,
				RowsTotal:     tt.rowsTotal,
			}
			event.CalculateProgressPercent()
			assert.Equal(t, tt.expected, event.ProgressPercent)
		})
	}
}

func TestStatusEvent_JSON(t *testing.T) {
	// 상태 변경 이벤트가 올바르게 JSON으로 직렬화되는지 테스트
	event := StatusEvent{
		TransportID: "TRPID-12345678",
		JobID:       "JOB-20260118-120000-abc123",
		Status:      "running",
		Message:     "추출 시작됨",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "TRPID-12345678", decoded["transport_id"])
	assert.Equal(t, "running", decoded["status"])
	assert.Equal(t, "추출 시작됨", decoded["message"])
}

func TestErrorEvent_JSON(t *testing.T) {
	// 에러 이벤트가 올바르게 JSON으로 직렬화되는지 테스트
	now := time.Now().UTC()
	event := ErrorEvent{
		TransportID: "TRPID-12345678",
		JobID:       "JOB-20260118-120000-abc123",
		Table:       "VBRP",
		Code:        "ORACLE_CONNECTION_ERROR",
		Message:     "Oracle 연결 실패",
		Timestamp:   now.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "TRPID-12345678", decoded["transport_id"])
	assert.Equal(t, "VBRP", decoded["table"])
	assert.Equal(t, "ORACLE_CONNECTION_ERROR", decoded["code"])
	assert.Equal(t, "Oracle 연결 실패", decoded["message"])
	assert.NotEmpty(t, decoded["timestamp"])
}

func TestCompleteEvent_JSON(t *testing.T) {
	// 완료 이벤트가 올바르게 JSON으로 직렬화되는지 테스트
	event := CompleteEvent{
		TransportID:   "TRPID-12345678",
		JobID:         "JOB-20260118-120000-abc123",
		TotalRows:     100000,
		TotalBytes:    10485760,
		DurationMs:    8000,
		TablesCount:   5,
		RowsPerSecond: 12500.0,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "TRPID-12345678", decoded["transport_id"])
	assert.Equal(t, float64(100000), decoded["total_rows"])
	assert.Equal(t, float64(10485760), decoded["total_bytes"])
	assert.Equal(t, float64(8000), decoded["duration_ms"])
	assert.Equal(t, float64(5), decoded["tables_count"])
	assert.Equal(t, 12500.0, decoded["rows_per_second"])
}

func TestSSEEvent_Format(t *testing.T) {
	// SSE 이벤트가 올바른 형식으로 포맷되는지 테스트
	progressData := ProgressEvent{
		TransportID:     "TRPID-12345678",
		JobID:           "JOB-20260118-120000-abc123",
		Table:           "VBRP",
		RowsProcessed:   10000,
		RowsTotal:       40000,
		RowsPerSecond:   5000.5,
		BytesWritten:    1048576,
		ProgressPercent: 25.0,
	}

	event := SSEEvent{
		Event: EventTypeProgress,
		Data:  progressData,
	}

	formatted, err := event.Format()
	require.NoError(t, err)

	assert.Contains(t, formatted, "event: progress\n")
	assert.Contains(t, formatted, "data: ")
	assert.Contains(t, formatted, "TRPID-12345678")
	assert.Contains(t, formatted, "\n\n")
}

func TestNewProgressEvent(t *testing.T) {
	// NewProgressEvent 헬퍼 함수 테스트
	event := NewProgressEvent("TRPID-12345678", "JOB-001", "VBRP", 10000, 40000, 1048576)

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, "VBRP", event.Table)
	assert.Equal(t, int64(10000), event.RowsProcessed)
	assert.Equal(t, int64(40000), event.RowsTotal)
	assert.Equal(t, int64(1048576), event.BytesWritten)
	// ProgressPercent는 계산되어야 함
	assert.Equal(t, 25.0, event.ProgressPercent)
}

func TestNewStatusEvent(t *testing.T) {
	// NewStatusEvent 헬퍼 함수 테스트
	event := NewStatusEvent("TRPID-12345678", "JOB-001", StatusRunning, "추출 시작")

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, StatusRunning, event.Status)
	assert.Equal(t, "추출 시작", event.Message)
}

func TestNewErrorEvent(t *testing.T) {
	// NewErrorEvent 헬퍼 함수 테스트
	event := NewErrorEvent("TRPID-12345678", "JOB-001", "VBRP", "ORACLE_ERROR", "연결 실패")

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, "VBRP", event.Table)
	assert.Equal(t, "ORACLE_ERROR", event.Code)
	assert.Equal(t, "연결 실패", event.Message)
	assert.NotEmpty(t, event.Timestamp)
}

func TestNewCompleteEvent(t *testing.T) {
	// NewCompleteEvent 헬퍼 함수 테스트
	event := NewCompleteEvent("TRPID-12345678", "JOB-001", 100000, 10485760, 8000, 5)

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, int64(100000), event.TotalRows)
	assert.Equal(t, int64(10485760), event.TotalBytes)
	assert.Equal(t, int64(8000), event.DurationMs)
	assert.Equal(t, 5, event.TablesCount)
	// RowsPerSecond는 계산되어야 함: 100000 rows / 8 seconds = 12500
	assert.Equal(t, 12500.0, event.RowsPerSecond)
}
