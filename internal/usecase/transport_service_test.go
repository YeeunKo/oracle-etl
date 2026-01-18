package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository/memory"
)

// TestTransportService_Create는 Transport 생성을 테스트합니다
func TestTransportService_Create(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	req := domain.CreateTransportRequest{
		Name:        "Test Transport",
		Description: "Test Description",
		Tables:      []string{"TABLE1", "TABLE2"},
	}

	transport, err := svc.Create(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, transport.ID)
	assert.True(t, len(transport.ID) > 6) // TRPID-xxxxxxxx
	assert.Equal(t, "Test Transport", transport.Name)
	assert.Equal(t, []string{"TABLE1", "TABLE2"}, transport.Tables)
	assert.True(t, transport.Enabled)
	assert.Equal(t, domain.TransportStatusIdle, transport.Status)
}

// TestTransportService_CreateValidation은 유효성 검사를 테스트합니다
func TestTransportService_CreateValidation(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	// 빈 이름
	_, err := svc.Create(ctx, domain.CreateTransportRequest{
		Name:   "",
		Tables: []string{"TABLE1"},
	})
	assert.Error(t, err)

	// 빈 테이블
	_, err = svc.Create(ctx, domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{},
	})
	assert.Error(t, err)
}

// TestTransportService_GetByID는 ID로 Transport 조회를 테스트합니다
func TestTransportService_GetByID(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	// 존재하지 않는 ID
	_, err := svc.GetByID(ctx, "non-existent")
	assert.Error(t, err)

	// Transport 생성 후 조회
	transport, err := svc.Create(ctx, domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	})
	require.NoError(t, err)

	found, err := svc.GetByID(ctx, transport.ID)
	require.NoError(t, err)
	assert.Equal(t, transport.ID, found.ID)
}

// TestTransportService_List는 Transport 목록 조회를 테스트합니다
func TestTransportService_List(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	// 빈 목록
	resp, err := svc.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, resp.Transports)
	assert.Equal(t, 0, resp.Total)

	// 여러 Transport 생성
	for i := 0; i < 5; i++ {
		_, err := svc.Create(ctx, domain.CreateTransportRequest{
			Name:   "Transport",
			Tables: []string{"TABLE"},
		})
		require.NoError(t, err)
	}

	// 페이지네이션 테스트
	resp, err = svc.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Len(t, resp.Transports, 3)
	assert.Equal(t, 5, resp.Total)
	assert.Equal(t, 0, resp.Offset)
	assert.Equal(t, 3, resp.Limit)
}

// TestTransportService_Delete는 Transport 삭제를 테스트합니다
func TestTransportService_Delete(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	transport, err := svc.Create(ctx, domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, transport.ID)
	require.NoError(t, err)

	// 삭제된 Transport 조회
	_, err = svc.GetByID(ctx, transport.ID)
	assert.Error(t, err)
}

// TestTransportService_CanExecute는 실행 가능 여부 확인을 테스트합니다
func TestTransportService_CanExecute(t *testing.T) {
	repo := memory.NewTransportRepository()
	svc := NewTransportService(repo)
	ctx := context.Background()

	transport, err := svc.Create(ctx, domain.CreateTransportRequest{
		Name:   "Test",
		Tables: []string{"TABLE1"},
	})
	require.NoError(t, err)

	// 실행 가능
	assert.True(t, transport.CanExecute())

	// 상태가 Running이면 실행 불가
	err = repo.UpdateStatus(ctx, transport.ID, domain.TransportStatusRunning)
	require.NoError(t, err)

	found, err := svc.GetByID(ctx, transport.ID)
	require.NoError(t, err)
	assert.False(t, found.CanExecute())
}
