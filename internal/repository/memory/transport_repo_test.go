package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/domain"
)

// TestTransportRepo_Create는 Transport 생성을 테스트합니다
func TestTransportRepo_Create(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	transport := domain.NewTransport("TRPID-12345678", "Test Transport", "Description", []string{"TABLE1", "TABLE2"})

	err := repo.Create(ctx, transport)
	require.NoError(t, err)

	// 중복 생성 시도
	err = repo.Create(ctx, transport)
	assert.Error(t, err)
}

// TestTransportRepo_GetByID는 ID로 Transport 조회를 테스트합니다
func TestTransportRepo_GetByID(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	// 존재하지 않는 ID 조회
	_, err := repo.GetByID(ctx, "non-existent")
	assert.Error(t, err)

	// Transport 생성 후 조회
	transport := domain.NewTransport("TRPID-12345678", "Test Transport", "Description", []string{"TABLE1"})
	require.NoError(t, repo.Create(ctx, transport))

	found, err := repo.GetByID(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.Equal(t, "Test Transport", found.Name)
	assert.Equal(t, []string{"TABLE1"}, found.Tables)
}

// TestTransportRepo_List는 Transport 목록 조회를 테스트합니다
func TestTransportRepo_List(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	// 빈 목록
	list, total, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, 0, total)

	// 여러 Transport 생성
	for i := 0; i < 5; i++ {
		transport := domain.NewTransport(
			domain.GenerateTransportID("uuid"+string(rune('0'+i))),
			"Transport "+string(rune('A'+i)),
			"",
			[]string{"TABLE"},
		)
		require.NoError(t, repo.Create(ctx, transport))
	}

	// 페이지네이션 테스트
	list, total, err = repo.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, 5, total)

	list, total, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, 5, total)
}

// TestTransportRepo_Update는 Transport 수정을 테스트합니다
func TestTransportRepo_Update(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	transport := domain.NewTransport("TRPID-12345678", "Original", "Desc", []string{"TABLE1"})
	require.NoError(t, repo.Create(ctx, transport))

	// 수정
	transport.Name = "Updated"
	transport.Tables = []string{"TABLE1", "TABLE2"}
	err := repo.Update(ctx, transport)
	require.NoError(t, err)

	// 확인
	found, err := repo.GetByID(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.Name)
	assert.Len(t, found.Tables, 2)
}

// TestTransportRepo_Delete는 Transport 삭제를 테스트합니다
func TestTransportRepo_Delete(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, repo.Create(ctx, transport))

	err := repo.Delete(ctx, "TRPID-12345678")
	require.NoError(t, err)

	// 삭제된 Transport 조회
	_, err = repo.GetByID(ctx, "TRPID-12345678")
	assert.Error(t, err)

	// 존재하지 않는 ID 삭제
	err = repo.Delete(ctx, "non-existent")
	assert.Error(t, err)
}

// TestTransportRepo_UpdateStatus는 Transport 상태 변경을 테스트합니다
func TestTransportRepo_UpdateStatus(t *testing.T) {
	repo := NewTransportRepository()
	ctx := context.Background()

	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, repo.Create(ctx, transport))

	err := repo.UpdateStatus(ctx, "TRPID-12345678", domain.TransportStatusRunning)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.Equal(t, domain.TransportStatusRunning, found.Status)
}
