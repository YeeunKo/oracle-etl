// Package domain은 ETL 파이프라인의 핵심 도메인 모델을 정의합니다.
package domain

import (
	"fmt"
	"time"
)

// JobStatus는 Job의 상태를 나타냅니다
type JobStatus string

const (
	// JobStatusPending은 대기 상태입니다
	JobStatusPending JobStatus = "pending"
	// JobStatusRunning은 실행 중 상태입니다
	JobStatusRunning JobStatus = "running"
	// JobStatusCompleted는 완료 상태입니다
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed는 실패 상태입니다
	JobStatusFailed JobStatus = "failed"
	// JobStatusCancelled은 취소 상태입니다
	JobStatusCancelled JobStatus = "cancelled"
)

// IsTerminal은 Job이 종료 상태인지 확인합니다
func (s JobStatus) IsTerminal() bool {
	return s == JobStatusCompleted || s == JobStatusFailed || s == JobStatusCancelled
}

// JobMetrics는 Job 실행 메트릭을 나타냅니다
type JobMetrics struct {
	TotalRows     int64         `json:"total_rows"`      // 총 처리 row 수
	TotalBytes    int64         `json:"total_bytes"`     // 총 바이트 수
	Duration      time.Duration `json:"duration"`        // 실행 시간
	RowsPerSecond float64       `json:"rows_per_second"` // 초당 처리 row 수
}

// CalculateRowsPerSecond는 초당 처리 row 수를 계산합니다
func (m *JobMetrics) CalculateRowsPerSecond() {
	if m.Duration > 0 {
		m.RowsPerSecond = float64(m.TotalRows) / m.Duration.Seconds()
	}
}

// Job은 Transport의 단일 실행을 나타냅니다
type Job struct {
	ID          string       `json:"id"`                     // JOB-{timestamp}-{random}
	TransportID string       `json:"transport_id"`           // 연결된 Transport ID
	Version     int          `json:"version"`                // JOBVER: Transport별 증가
	Status      JobStatus    `json:"status"`                 // 현재 상태
	StartedAt   *time.Time   `json:"started_at,omitempty"`   // 시작 시간
	CompletedAt *time.Time   `json:"completed_at,omitempty"` // 완료 시간
	Extractions []Extraction `json:"extractions,omitempty"`  // 테이블별 추출 결과
	Error       *string      `json:"error,omitempty"`        // 에러 메시지
	Metrics     JobMetrics   `json:"metrics"`                // 실행 메트릭
	CreatedAt   time.Time    `json:"created_at"`             // 생성 시간
}

// GenerateJobID는 새로운 Job ID를 생성합니다
// 형식: JOB-{YYYYMMDD-HHMMSS}-{random}
func GenerateJobID(timestamp time.Time, randomPart string) string {
	return fmt.Sprintf("JOB-%s-%s", timestamp.Format("20060102-150405"), randomPart)
}

// VersionString은 버전을 문자열로 반환합니다 (v001, v002, ...)
func (j *Job) VersionString() string {
	return fmt.Sprintf("v%03d", j.Version)
}

// NewJob은 새로운 Job을 생성합니다
func NewJob(id, transportID string, version int) *Job {
	now := time.Now().UTC()
	return &Job{
		ID:          id,
		TransportID: transportID,
		Version:     version,
		Status:      JobStatusPending,
		Extractions: make([]Extraction, 0),
		Metrics:     JobMetrics{},
		CreatedAt:   now,
	}
}

// Start는 Job을 시작 상태로 변경합니다
func (j *Job) Start() {
	now := time.Now().UTC()
	j.Status = JobStatusRunning
	j.StartedAt = &now
}

// Complete는 Job을 완료 상태로 변경합니다
func (j *Job) Complete() {
	now := time.Now().UTC()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
	if j.StartedAt != nil {
		j.Metrics.Duration = now.Sub(*j.StartedAt)
		j.Metrics.CalculateRowsPerSecond()
	}
}

// Fail은 Job을 실패 상태로 변경합니다
func (j *Job) Fail(err error) {
	now := time.Now().UTC()
	j.Status = JobStatusFailed
	j.CompletedAt = &now
	if err != nil {
		errStr := err.Error()
		j.Error = &errStr
	}
	if j.StartedAt != nil {
		j.Metrics.Duration = now.Sub(*j.StartedAt)
		j.Metrics.CalculateRowsPerSecond()
	}
}

// Cancel은 Job을 취소 상태로 변경합니다
func (j *Job) Cancel() {
	now := time.Now().UTC()
	j.Status = JobStatusCancelled
	j.CompletedAt = &now
	if j.StartedAt != nil {
		j.Metrics.Duration = now.Sub(*j.StartedAt)
	}
}

// AddExtraction은 Job에 Extraction을 추가합니다
func (j *Job) AddExtraction(ext Extraction) {
	j.Extractions = append(j.Extractions, ext)
}

// UpdateMetrics는 Extractions 기반으로 메트릭을 업데이트합니다
func (j *Job) UpdateMetrics() {
	var totalRows, totalBytes int64
	for _, ext := range j.Extractions {
		totalRows += ext.RowCount
		totalBytes += ext.ByteCount
	}
	j.Metrics.TotalRows = totalRows
	j.Metrics.TotalBytes = totalBytes
	j.Metrics.CalculateRowsPerSecond()
}

// ExecuteJobResponse는 Job 실행 응답입니다
type ExecuteJobResponse struct {
	JobID       string    `json:"job_id"`
	TransportID string    `json:"transport_id"`
	Version     int       `json:"version"`
	Status      JobStatus `json:"status"`
}

// JobListResponse는 Job 목록 응답입니다
type JobListResponse struct {
	Jobs   []Job `json:"jobs"`
	Total  int   `json:"total"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

// JobListFilter는 Job 목록 조회 필터입니다
type JobListFilter struct {
	TransportID string    `json:"transport_id,omitempty"`
	Status      JobStatus `json:"status,omitempty"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

// DefaultJobListFilter는 기본 필터를 반환합니다
func DefaultJobListFilter() JobListFilter {
	return JobListFilter{
		Limit:  20,
		Offset: 0,
	}
}
