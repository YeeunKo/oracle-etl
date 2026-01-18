// Package domain은 ETL 파이프라인의 핵심 도메인 모델을 정의합니다.
package domain

import "time"

// TableInfo는 Oracle 테이블 메타데이터를 나타냅니다
type TableInfo struct {
	Name        string `json:"name"`         // 테이블 이름
	Owner       string `json:"owner"`        // 스키마 소유자
	RowCount    int64  `json:"row_count"`    // 테이블 row 수
	ColumnCount int    `json:"column_count"` // 컬럼 수
}

// ColumnInfo는 테이블 컬럼 메타데이터를 나타냅니다
type ColumnInfo struct {
	Name     string `json:"name"`      // 컬럼 이름
	DataType string `json:"data_type"` // 데이터 타입 (VARCHAR2, NUMBER 등)
	Nullable bool   `json:"nullable"`  // NULL 허용 여부
	Position int    `json:"position"`  // 컬럼 위치
}

// SampleData는 테이블 샘플 데이터 조회 결과를 나타냅니다
type SampleData struct {
	TableName string                   `json:"table"`   // 테이블 이름
	Columns   []string                 `json:"columns"` // 컬럼 이름 목록
	Rows      []map[string]interface{} `json:"rows"`    // 데이터 행
	Count     int                      `json:"count"`   // 반환된 row 수
}

// OracleStatus는 Oracle 연결 상태 정보를 나타냅니다
type OracleStatus struct {
	Connected       bool      `json:"connected"`       // 연결 상태
	DatabaseVersion string    `json:"database_version"` // 데이터베이스 버전
	InstanceName    string    `json:"instance_name"`   // 인스턴스 이름
	PoolStats       PoolStats `json:"pool_stats"`      // 커넥션 풀 통계
	CheckedAt       time.Time `json:"checked_at"`      // 상태 확인 시간
	Error           string    `json:"error,omitempty"` // 에러 메시지 (있는 경우)
}

// PoolStats는 커넥션 풀 통계를 나타냅니다
type PoolStats struct {
	ActiveConnections int `json:"active_connections"` // 사용 중인 연결 수
	IdleConnections   int `json:"idle_connections"`   // 유휴 연결 수
	MaxConnections    int `json:"max_connections"`    // 최대 연결 수
}

// TableListResponse는 GET /api/tables 응답 구조체입니다
type TableListResponse struct {
	Tables []TableInfo `json:"tables"` // 테이블 목록
	Total  int         `json:"total"`  // 총 테이블 수
}

// ChunkResult는 청크 단위 데이터 추출 결과를 나타냅니다
type ChunkResult struct {
	TableName     string                   `json:"table_name"`      // 테이블 이름
	ChunkNumber   int                      `json:"chunk_number"`    // 청크 번호
	Rows          []map[string]interface{} `json:"rows"`            // 데이터 행
	RowCount      int                      `json:"row_count"`       // 이 청크의 row 수
	IsLastChunk   bool                     `json:"is_last_chunk"`   // 마지막 청크 여부
	TotalRowsSent int64                    `json:"total_rows_sent"` // 지금까지 전송된 총 row 수
}

// ExtractionOptions는 데이터 추출 옵션을 나타냅니다
type ExtractionOptions struct {
	ChunkSize      int  `json:"chunk_size"`       // 청크당 row 수 (기본값: 10000)
	FetchArraySize int  `json:"fetch_array_size"` // 배치 페치 크기 (기본값: 1000)
	IncludeColumns bool `json:"include_columns"`  // 컬럼 정보 포함 여부
}

// DefaultExtractionOptions는 기본 추출 옵션을 반환합니다
func DefaultExtractionOptions() ExtractionOptions {
	return ExtractionOptions{
		ChunkSize:      10000,
		FetchArraySize: 1000,
		IncludeColumns: true,
	}
}

// ExtractionStatus는 Extraction의 상태를 나타냅니다
type ExtractionStatus string

const (
	// ExtractionStatusPending은 대기 상태입니다
	ExtractionStatusPending ExtractionStatus = "pending"
	// ExtractionStatusRunning은 실행 중 상태입니다
	ExtractionStatusRunning ExtractionStatus = "running"
	// ExtractionStatusCompleted는 완료 상태입니다
	ExtractionStatusCompleted ExtractionStatus = "completed"
	// ExtractionStatusFailed는 실패 상태입니다
	ExtractionStatusFailed ExtractionStatus = "failed"
)

// Extraction은 단일 테이블 추출 결과를 나타냅니다
type Extraction struct {
	ID          string           `json:"id"`                     // 추출 ID
	JobID       string           `json:"job_id"`                 // 연결된 Job ID
	TableName   string           `json:"table_name"`             // 테이블 이름
	Status      ExtractionStatus `json:"status"`                 // 상태
	RowCount    int64            `json:"row_count"`              // 처리된 row 수
	ByteCount   int64            `json:"byte_count"`             // 전송된 바이트 수
	GCSPath     string           `json:"gcs_path,omitempty"`     // GCS 객체 경로
	StartedAt   *time.Time       `json:"started_at,omitempty"`   // 시작 시간
	CompletedAt *time.Time       `json:"completed_at,omitempty"` // 완료 시간
	Error       *string          `json:"error,omitempty"`        // 에러 메시지
}

// NewExtraction은 새로운 Extraction을 생성합니다
func NewExtraction(id, jobID, tableName string) *Extraction {
	return &Extraction{
		ID:        id,
		JobID:     jobID,
		TableName: tableName,
		Status:    ExtractionStatusPending,
	}
}

// Start는 Extraction을 시작 상태로 변경합니다
func (e *Extraction) Start() {
	now := time.Now().UTC()
	e.Status = ExtractionStatusRunning
	e.StartedAt = &now
}

// Complete는 Extraction을 완료 상태로 변경합니다
func (e *Extraction) Complete(rowCount, byteCount int64, gcsPath string) {
	now := time.Now().UTC()
	e.Status = ExtractionStatusCompleted
	e.CompletedAt = &now
	e.RowCount = rowCount
	e.ByteCount = byteCount
	e.GCSPath = gcsPath
}

// Fail은 Extraction을 실패 상태로 변경합니다
func (e *Extraction) Fail(err error) {
	now := time.Now().UTC()
	e.Status = ExtractionStatusFailed
	e.CompletedAt = &now
	if err != nil {
		errStr := err.Error()
		e.Error = &errStr
	}
}
