// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"sync"
	"time"
)

// TableMetrics는 테이블별 메트릭을 저장합니다
type TableMetrics struct {
	TotalRows     int64 // 총 row 수
	ProcessedRows int64 // 처리된 row 수
	BytesWritten  int64 // 작성된 바이트 수
}

// MetricsTracker는 ETL 작업의 메트릭을 추적합니다
type MetricsTracker struct {
	TransportID string // Transport ID
	JobID       string // Job ID
	TotalRows   int64  // 총 처리된 row 수
	TotalBytes  int64  // 총 바이트 수

	startTime time.Time              // 시작 시간
	tables    map[string]*TableMetrics // 테이블별 메트릭
	mu        sync.RWMutex           // 동기화를 위한 뮤텍스
}

// NewMetricsTracker는 새로운 MetricsTracker를 생성합니다
func NewMetricsTracker(transportID, jobID string) *MetricsTracker {
	return &MetricsTracker{
		TransportID: transportID,
		JobID:       jobID,
		tables:      make(map[string]*TableMetrics),
	}
}

// Start는 메트릭 추적을 시작합니다
func (m *MetricsTracker) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startTime = time.Now()
}

// AddTable은 추적할 테이블을 추가합니다
func (m *MetricsTracker) AddTable(tableName string, totalRows int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tables[tableName] = &TableMetrics{
		TotalRows:     totalRows,
		ProcessedRows: 0,
		BytesWritten:  0,
	}
}

// UpdateProgress는 테이블의 진행률을 업데이트합니다
func (m *MetricsTracker) UpdateProgress(tableName string, processedRows, bytesWritten int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tableMetric, exists := m.tables[tableName]; exists {
		// 이전 값과의 차이를 계산하여 전체 합계 업데이트
		rowsDiff := processedRows - tableMetric.ProcessedRows
		bytesDiff := bytesWritten - tableMetric.BytesWritten

		tableMetric.ProcessedRows = processedRows
		tableMetric.BytesWritten = bytesWritten

		m.TotalRows += rowsDiff
		m.TotalBytes += bytesDiff
	}
}

// GetProgressEvent는 특정 테이블의 ProgressEvent를 생성합니다
func (m *MetricsTracker) GetProgressEvent(tableName string) *ProgressEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tableMetric, exists := m.tables[tableName]
	if !exists {
		return nil
	}

	event := &ProgressEvent{
		TransportID:   m.TransportID,
		JobID:         m.JobID,
		Table:         tableName,
		RowsProcessed: tableMetric.ProcessedRows,
		RowsTotal:     tableMetric.TotalRows,
		BytesWritten:  tableMetric.BytesWritten,
	}

	// 진행률 계산
	event.CalculateProgressPercent()

	// 초당 처리 row 수 계산
	elapsed := time.Since(m.startTime)
	if elapsed > 0 {
		event.RowsPerSecond = float64(tableMetric.ProcessedRows) / elapsed.Seconds()
	}

	return event
}

// GetCompleteEvent는 완료 이벤트를 생성합니다
func (m *MetricsTracker) GetCompleteEvent() *CompleteEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elapsed := time.Since(m.startTime)
	durationMs := elapsed.Milliseconds()

	var rowsPerSecond float64
	if elapsed > 0 {
		rowsPerSecond = float64(m.TotalRows) / elapsed.Seconds()
	}

	return &CompleteEvent{
		TransportID:   m.TransportID,
		JobID:         m.JobID,
		TotalRows:     m.TotalRows,
		TotalBytes:    m.TotalBytes,
		DurationMs:    durationMs,
		TablesCount:   len(m.tables),
		RowsPerSecond: rowsPerSecond,
	}
}

// GetDuration은 현재까지의 실행 시간을 반환합니다
func (m *MetricsTracker) GetDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.startTime.IsZero() {
		return 0
	}
	return time.Since(m.startTime)
}

// Reset은 모든 메트릭을 초기화합니다
func (m *MetricsTracker) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalRows = 0
	m.TotalBytes = 0
	m.startTime = time.Time{}
	m.tables = make(map[string]*TableMetrics)
}
