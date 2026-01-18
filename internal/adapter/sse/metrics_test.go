// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsTracker(t *testing.T) {
	// MetricsTracker 생성 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")

	require.NotNil(t, tracker)
	assert.Equal(t, "TRPID-12345678", tracker.TransportID)
	assert.Equal(t, "JOB-001", tracker.JobID)
	assert.Equal(t, int64(0), tracker.TotalRows)
	assert.Equal(t, int64(0), tracker.TotalBytes)
}

func TestMetricsTracker_Start(t *testing.T) {
	// 추적 시작 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()

	assert.False(t, tracker.startTime.IsZero(), "시작 시간이 설정되어야 합니다")
}

func TestMetricsTracker_AddTable(t *testing.T) {
	// 테이블 추가 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()

	tracker.AddTable("VBRP", 40000)
	tracker.AddTable("VBAK", 10000)

	assert.Equal(t, 2, len(tracker.tables))
	assert.Equal(t, int64(40000), tracker.tables["VBRP"].TotalRows)
	assert.Equal(t, int64(10000), tracker.tables["VBAK"].TotalRows)
}

func TestMetricsTracker_UpdateProgress(t *testing.T) {
	// 진행률 업데이트 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 40000)

	// 진행률 업데이트
	tracker.UpdateProgress("VBRP", 10000, 512000)

	// 테이블 메트릭 확인
	tableMetric := tracker.tables["VBRP"]
	assert.Equal(t, int64(10000), tableMetric.ProcessedRows)
	assert.Equal(t, int64(512000), tableMetric.BytesWritten)

	// 전체 메트릭 확인
	assert.Equal(t, int64(10000), tracker.TotalRows)
	assert.Equal(t, int64(512000), tracker.TotalBytes)
}

func TestMetricsTracker_GetProgressEvent(t *testing.T) {
	// ProgressEvent 생성 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 40000)
	tracker.UpdateProgress("VBRP", 10000, 512000)

	event := tracker.GetProgressEvent("VBRP")

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, "VBRP", event.Table)
	assert.Equal(t, int64(10000), event.RowsProcessed)
	assert.Equal(t, int64(40000), event.RowsTotal)
	assert.Equal(t, int64(512000), event.BytesWritten)
	assert.Equal(t, 25.0, event.ProgressPercent)
}

func TestMetricsTracker_CalculateRowsPerSecond(t *testing.T) {
	// 초당 처리 row 수 계산 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 100000)

	// 시뮬레이션: 2초 동안 50000 rows 처리
	time.Sleep(100 * time.Millisecond)
	tracker.UpdateProgress("VBRP", 50000, 2560000)

	event := tracker.GetProgressEvent("VBRP")

	// RowsPerSecond가 0보다 커야 함
	assert.Greater(t, event.RowsPerSecond, 0.0)
}

func TestMetricsTracker_GetCompleteEvent(t *testing.T) {
	// CompleteEvent 생성 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 40000)
	tracker.AddTable("VBAK", 10000)
	tracker.UpdateProgress("VBRP", 40000, 2048000)
	tracker.UpdateProgress("VBAK", 10000, 512000)

	// 약간의 시간 경과 시뮬레이션
	time.Sleep(50 * time.Millisecond)

	event := tracker.GetCompleteEvent()

	assert.Equal(t, "TRPID-12345678", event.TransportID)
	assert.Equal(t, "JOB-001", event.JobID)
	assert.Equal(t, int64(50000), event.TotalRows)
	assert.Equal(t, int64(2560000), event.TotalBytes)
	assert.Equal(t, 2, event.TablesCount)
	assert.Greater(t, event.DurationMs, int64(0))
	assert.Greater(t, event.RowsPerSecond, 0.0)
}

func TestMetricsTracker_GetDuration(t *testing.T) {
	// 실행 시간 계산 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()

	time.Sleep(50 * time.Millisecond)

	duration := tracker.GetDuration()
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(50))
}

func TestMetricsTracker_Reset(t *testing.T) {
	// 리셋 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 40000)
	tracker.UpdateProgress("VBRP", 10000, 512000)

	tracker.Reset()

	assert.Equal(t, int64(0), tracker.TotalRows)
	assert.Equal(t, int64(0), tracker.TotalBytes)
	assert.Empty(t, tracker.tables)
	assert.True(t, tracker.startTime.IsZero())
}

func TestMetricsTracker_ConcurrentUpdates(t *testing.T) {
	// 동시 업데이트 안전성 테스트
	tracker := NewMetricsTracker("TRPID-12345678", "JOB-001")
	tracker.Start()
	tracker.AddTable("VBRP", 100000)

	done := make(chan bool)

	// 동시에 여러 고루틴에서 업데이트
	for i := 0; i < 100; i++ {
		go func(idx int) {
			tracker.UpdateProgress("VBRP", int64((idx+1)*1000), int64((idx+1)*51200))
			done <- true
		}(i)
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < 100; i++ {
		<-done
	}

	// 패닉 없이 완료되면 성공
	assert.True(t, tracker.TotalRows > 0)
}
