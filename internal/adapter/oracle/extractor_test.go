package oracle

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oracle-etl/internal/domain"
)

func TestDataExtractor_ExtractTable(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockRepository)
		opts          domain.ExtractionOptions
		expectedError bool
		expectedChunks int
	}{
		{
			name: "단일 테이블 추출 성공",
			setupMock: func(m *MockRepository) {
				m.MockChunks = []*domain.ChunkResult{
					{TableName: "VBRP", ChunkNumber: 1, RowCount: 10000, IsLastChunk: false, TotalRowsSent: 10000},
					{TableName: "VBRP", ChunkNumber: 2, RowCount: 5000, IsLastChunk: true, TotalRowsSent: 15000},
				}
			},
			opts:           domain.DefaultExtractionOptions(),
			expectedError:  false,
			expectedChunks: 2,
		},
		{
			name: "추출 중 에러 발생",
			setupMock: func(m *MockRepository) {
				m.ShouldError = true
				m.ErrorMessage = "데이터베이스 연결 끊김"
			},
			opts:           domain.DefaultExtractionOptions(),
			expectedError:  true,
			expectedChunks: 0,
		},
		{
			name: "기본 옵션 적용",
			setupMock: func(m *MockRepository) {
				m.MockChunks = []*domain.ChunkResult{
					{TableName: "VBRP", ChunkNumber: 1, RowCount: 100, IsLastChunk: true, TotalRowsSent: 100},
				}
			},
			opts: domain.ExtractionOptions{
				ChunkSize:      0, // 기본값 적용
				FetchArraySize: 0, // 기본값 적용
			},
			expectedError:  false,
			expectedChunks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock 저장소 생성
			mockRepo := NewMockRepository()
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			// Extractor 생성
			extractor := NewDataExtractor(mockRepo)

			// 청크 수집
			var receivedChunks []*domain.ChunkResult
			ctx := context.Background()

			err := extractor.ExtractTable(ctx, "SAPSR3", "VBRP", tt.opts, func(chunk *domain.ChunkResult) error {
				receivedChunks = append(receivedChunks, chunk)
				return nil
			})

			// 결과 검증
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, receivedChunks, tt.expectedChunks)
			}

			// Mock 호출 확인
			assert.True(t, mockRepo.StreamCalled)
		})
	}
}

func TestDataExtractor_ExtractTable_HandlerError(t *testing.T) {
	// Mock 저장소 생성
	mockRepo := NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "VBRP", ChunkNumber: 1, RowCount: 10000, IsLastChunk: false, TotalRowsSent: 10000},
	}

	// Extractor 생성
	extractor := NewDataExtractor(mockRepo)

	// 핸들러에서 에러 반환
	ctx := context.Background()
	err := extractor.ExtractTable(ctx, "SAPSR3", "VBRP", domain.DefaultExtractionOptions(), func(chunk *domain.ChunkResult) error {
		return errors.New("핸들러 에러")
	})

	require.Error(t, err)
}

func TestDataExtractor_ExtractTables(t *testing.T) {
	// Mock 저장소 생성
	mockRepo := NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	// Extractor 생성
	extractor := NewDataExtractor(mockRepo)

	// 청크 카운터 (atomic)
	var chunkCount int32

	ctx := context.Background()
	tables := []string{"VBRP", "VBRK", "LIKP"}

	err := extractor.ExtractTables(ctx, "SAPSR3", tables, domain.DefaultExtractionOptions(), 2, func(chunk *domain.ChunkResult) error {
		atomic.AddInt32(&chunkCount, 1)
		return nil
	})

	require.NoError(t, err)
	// 3개 테이블 x 1개 청크 = 3개 청크
	assert.Equal(t, int32(3), chunkCount)
}

func TestDataExtractor_ExtractTables_ContextCancel(t *testing.T) {
	// Mock 저장소 생성 (느린 응답 시뮬레이션)
	mockRepo := NewMockRepository()
	mockRepo.MockChunks = []*domain.ChunkResult{
		{TableName: "", ChunkNumber: 1, RowCount: 1000, IsLastChunk: true, TotalRowsSent: 1000},
	}

	// Extractor 생성
	extractor := NewDataExtractor(mockRepo)

	// 취소 가능한 컨텍스트
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	tables := []string{"VBRP", "VBRK"}

	err := extractor.ExtractTables(ctx, "SAPSR3", tables, domain.DefaultExtractionOptions(), 1, func(chunk *domain.ChunkResult) error {
		return nil
	})

	// 컨텍스트 취소 에러 확인
	require.Error(t, err)
}
