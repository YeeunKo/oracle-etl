// Package oracle은 Oracle 데이터베이스 연결 및 데이터 추출 기능을 제공합니다.
package oracle

import (
	"context"
	"errors"
	"time"

	"sync"

	"oracle-etl/internal/domain"
)

// MockRepository는 테스트용 Oracle 저장소 Mock입니다
type MockRepository struct {
	mu sync.Mutex // 동시성 보호용 뮤텍스

	// 테스트 설정 필드
	ShouldError       bool
	ErrorMessage      string
	MockStatus        *domain.OracleStatus
	MockTables        []domain.TableInfo
	MockColumns       []domain.ColumnInfo
	MockSampleData    *domain.SampleData
	MockChunks        []*domain.ChunkResult
	PingCalled        bool
	CloseCalled       bool
	GetStatusCalled   bool
	GetTablesCalled   bool
	GetSampleCalled   bool
	StreamCalled      bool

	// 테이블별 에러 설정 (부분 실패 테스트용)
	TableErrors map[string]error

	// 커스텀 StreamTableData 함수 (동시성 테스트용)
	StreamTableDataFunc func(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, handler func(chunk *domain.ChunkResult) error) error
}

// NewMockRepository는 새로운 MockRepository를 생성합니다
func NewMockRepository() *MockRepository {
	return &MockRepository{
		MockStatus: &domain.OracleStatus{
			Connected:       true,
			DatabaseVersion: "19.0.0.0.0",
			InstanceName:    "ATP_HIGH",
			PoolStats: domain.PoolStats{
				ActiveConnections: 2,
				IdleConnections:   8,
				MaxConnections:    10,
			},
			CheckedAt: time.Now().UTC(),
		},
		MockTables: []domain.TableInfo{
			{Name: "VBRP", Owner: "SAPSR3", RowCount: 250000, ColumnCount: 274},
			{Name: "VBRK", Owner: "SAPSR3", RowCount: 100000, ColumnCount: 180},
		},
		MockColumns: []domain.ColumnInfo{
			{Name: "MANDT", DataType: "VARCHAR2", Nullable: false, Position: 1},
			{Name: "VBELN", DataType: "VARCHAR2", Nullable: false, Position: 2},
			{Name: "POSNR", DataType: "NUMBER", Nullable: false, Position: 3},
		},
		MockSampleData: &domain.SampleData{
			TableName: "VBRP",
			Columns:   []string{"MANDT", "VBELN", "POSNR"},
			Rows: []map[string]interface{}{
				{"MANDT": "800", "VBELN": "0090000001", "POSNR": float64(1)},
				{"MANDT": "800", "VBELN": "0090000002", "POSNR": float64(2)},
			},
			Count: 2,
		},
		TableErrors: make(map[string]error),
	}
}

// GetStatus는 Oracle 연결 상태 및 풀 통계를 반환합니다
func (m *MockRepository) GetStatus(ctx context.Context) (*domain.OracleStatus, error) {
	m.GetStatusCalled = true
	if m.ShouldError {
		return nil, errors.New(m.ErrorMessage)
	}
	return m.MockStatus, nil
}

// GetTables는 접근 가능한 테이블 목록과 row count를 반환합니다
func (m *MockRepository) GetTables(ctx context.Context, owner string) ([]domain.TableInfo, error) {
	m.GetTablesCalled = true
	if m.ShouldError {
		return nil, errors.New(m.ErrorMessage)
	}
	return m.MockTables, nil
}

// GetTableColumns는 테이블의 컬럼 정보를 반환합니다
func (m *MockRepository) GetTableColumns(ctx context.Context, owner, tableName string) ([]domain.ColumnInfo, error) {
	if m.ShouldError {
		return nil, errors.New(m.ErrorMessage)
	}
	return m.MockColumns, nil
}

// GetSampleData는 테이블의 샘플 데이터를 반환합니다
func (m *MockRepository) GetSampleData(ctx context.Context, owner, tableName string, limit int) (*domain.SampleData, error) {
	m.GetSampleCalled = true
	if m.ShouldError {
		return nil, errors.New(m.ErrorMessage)
	}
	return m.MockSampleData, nil
}

// StreamTableData는 테이블 데이터를 청크 단위로 스트리밍합니다
func (m *MockRepository) StreamTableData(ctx context.Context, owner, tableName string, opts domain.ExtractionOptions, chunkHandler func(chunk *domain.ChunkResult) error) error {
	m.mu.Lock()
	m.StreamCalled = true
	m.mu.Unlock()

	// 커스텀 함수가 있으면 사용
	if m.StreamTableDataFunc != nil {
		return m.StreamTableDataFunc(ctx, owner, tableName, opts, chunkHandler)
	}

	// 테이블별 에러 확인
	if err, ok := m.TableErrors[tableName]; ok {
		return err
	}

	if m.ShouldError {
		return errors.New(m.ErrorMessage)
	}

	// Mock 청크 데이터 스트리밍
	for _, chunk := range m.MockChunks {
		// 테이블 이름 설정
		chunkCopy := *chunk
		if chunkCopy.TableName == "" {
			chunkCopy.TableName = tableName
		}
		if err := chunkHandler(&chunkCopy); err != nil {
			return err
		}
	}
	return nil
}

// Ping은 Oracle 연결을 테스트합니다
func (m *MockRepository) Ping(ctx context.Context) error {
	m.PingCalled = true
	if m.ShouldError {
		return errors.New(m.ErrorMessage)
	}
	return nil
}

// Close는 커넥션 풀을 종료합니다
func (m *MockRepository) Close() error {
	m.CloseCalled = true
	return nil
}

// 인터페이스 구현 확인
var _ Repository = (*MockRepository)(nil)
