package oracle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oracle-etl/internal/domain"
)

func TestDefaultPoolConfig(t *testing.T) {
	// 기본 풀 설정 테스트
	cfg := DefaultPoolConfig()

	assert.Equal(t, 2, cfg.PoolMinConns, "기본 최소 커넥션 수는 2")
	assert.Equal(t, 10, cfg.PoolMaxConns, "기본 최대 커넥션 수는 10")
	assert.Equal(t, 1000, cfg.FetchArraySize, "기본 FetchArraySize는 1000")
	assert.Equal(t, 1000, cfg.PrefetchCount, "기본 PrefetchCount는 1000")
	assert.Equal(t, 30*time.Second, cfg.ConnectTimeout, "기본 연결 타임아웃은 30초")
}

func TestMockRepository_GetStatus(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()

	// 정상 상태 조회 테스트
	ctx := context.Background()
	status, err := mock.GetStatus(ctx)

	require.NoError(t, err)
	assert.True(t, mock.GetStatusCalled)
	assert.True(t, status.Connected)
	assert.Equal(t, "19.0.0.0.0", status.DatabaseVersion)
	assert.Equal(t, "ATP_HIGH", status.InstanceName)
	assert.Equal(t, 2, status.PoolStats.ActiveConnections)
	assert.Equal(t, 8, status.PoolStats.IdleConnections)
	assert.Equal(t, 10, status.PoolStats.MaxConnections)
}

func TestMockRepository_GetStatus_Error(t *testing.T) {
	// Mock 저장소 생성 및 에러 설정
	mock := NewMockRepository()
	mock.ShouldError = true
	mock.ErrorMessage = "연결 실패"

	// 에러 상태 테스트
	ctx := context.Background()
	status, err := mock.GetStatus(ctx)

	require.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "연결 실패")
}

func TestMockRepository_GetTables(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()

	// 테이블 목록 조회 테스트
	ctx := context.Background()
	tables, err := mock.GetTables(ctx, "SAPSR3")

	require.NoError(t, err)
	assert.True(t, mock.GetTablesCalled)
	require.Len(t, tables, 2)

	// 첫 번째 테이블 검증
	assert.Equal(t, "VBRP", tables[0].Name)
	assert.Equal(t, "SAPSR3", tables[0].Owner)
	assert.Equal(t, int64(250000), tables[0].RowCount)
	assert.Equal(t, 274, tables[0].ColumnCount)
}

func TestMockRepository_GetSampleData(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()

	// 샘플 데이터 조회 테스트
	ctx := context.Background()
	sample, err := mock.GetSampleData(ctx, "SAPSR3", "VBRP", 100)

	require.NoError(t, err)
	assert.True(t, mock.GetSampleCalled)
	assert.Equal(t, "VBRP", sample.TableName)
	assert.Contains(t, sample.Columns, "MANDT")
	assert.Contains(t, sample.Columns, "VBELN")
	require.Len(t, sample.Rows, 2)
	assert.Equal(t, "800", sample.Rows[0]["MANDT"])
}

func TestMockRepository_StreamTableData(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()
	mock.MockChunks = []*domain.ChunkResult{
		{
			TableName:     "VBRP",
			ChunkNumber:   1,
			RowCount:      10000,
			IsLastChunk:   false,
			TotalRowsSent: 10000,
		},
		{
			TableName:     "VBRP",
			ChunkNumber:   2,
			RowCount:      5000,
			IsLastChunk:   true,
			TotalRowsSent: 15000,
		},
	}

	// 스트리밍 테스트
	ctx := context.Background()
	opts := domain.DefaultExtractionOptions()

	var receivedChunks []*domain.ChunkResult
	err := mock.StreamTableData(ctx, "SAPSR3", "VBRP", opts, func(chunk *domain.ChunkResult) error {
		receivedChunks = append(receivedChunks, chunk)
		return nil
	})

	require.NoError(t, err)
	assert.True(t, mock.StreamCalled)
	require.Len(t, receivedChunks, 2)
	assert.Equal(t, 1, receivedChunks[0].ChunkNumber)
	assert.False(t, receivedChunks[0].IsLastChunk)
	assert.True(t, receivedChunks[1].IsLastChunk)
}

func TestMockRepository_Ping(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()

	// Ping 테스트
	ctx := context.Background()
	err := mock.Ping(ctx)

	require.NoError(t, err)
	assert.True(t, mock.PingCalled)
}

func TestMockRepository_Close(t *testing.T) {
	// Mock 저장소 생성
	mock := NewMockRepository()

	// Close 테스트
	err := mock.Close()

	require.NoError(t, err)
	assert.True(t, mock.CloseCalled)
}

func TestDefaultExtractionOptions(t *testing.T) {
	// 기본 추출 옵션 테스트
	opts := domain.DefaultExtractionOptions()

	assert.Equal(t, 10000, opts.ChunkSize, "기본 청크 크기는 10000")
	assert.Equal(t, 1000, opts.FetchArraySize, "기본 FetchArraySize는 1000")
	assert.True(t, opts.IncludeColumns, "기본적으로 컬럼 정보 포함")
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		banner   string
		expected string
	}{
		{
			name:     "Oracle 19c 버전",
			banner:   "Oracle Database 19c Enterprise Edition Release 19.0.0.0.0",
			expected: "19.0.0.0.0",
		},
		{
			name:     "Oracle 21c 버전",
			banner:   "Oracle Database 21c Standard Edition Release 21.3.0.0.0",
			expected: "21.3.0.0.0",
		},
		{
			name:     "버전 정보 없음",
			banner:   "Oracle Database",
			expected: "Oracle Database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.banner)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "nil 값",
			input:    nil,
			expected: nil,
		},
		{
			name:     "문자열",
			input:    "test",
			expected: "test",
		},
		{
			name:     "바이트 슬라이스",
			input:    []byte("hello"),
			expected: "hello",
		},
		{
			name:     "정수",
			input:    42,
			expected: 42,
		},
		{
			name:     "시간",
			input:    time.Date(2026, 1, 18, 10, 30, 0, 0, time.UTC),
			expected: "2026-01-18T10:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
