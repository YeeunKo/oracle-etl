package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository/memory"
)

// TestJobService_CreateJob는 Job 생성을 테스트합니다
func TestJobService_CreateJob(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 먼저 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	// Job 생성
	job, err := svc.CreateJob(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "TRPID-12345678", job.TransportID)
	assert.Equal(t, 1, job.Version)
	assert.Equal(t, domain.JobStatusPending, job.Status)
}

// TestJobService_CreateJobVersionIncrement는 버전 자동 증가를 테스트합니다
func TestJobService_CreateJobVersionIncrement(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	// Job 3개 생성
	for i := 1; i <= 3; i++ {
		job, err := svc.CreateJob(ctx, "TRPID-12345678")
		require.NoError(t, err)
		assert.Equal(t, i, job.Version)
	}
}

// TestJobService_CreateJobTransportNotFound는 존재하지 않는 Transport에 대한 Job 생성 시 에러를 테스트합니다
func TestJobService_CreateJobTransportNotFound(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	_, err := svc.CreateJob(ctx, "non-existent")
	assert.Error(t, err)
}

// TestJobService_GetByID는 Job ID로 조회를 테스트합니다
func TestJobService_GetByID(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 및 Job 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	job, err := svc.CreateJob(ctx, "TRPID-12345678")
	require.NoError(t, err)

	// 조회
	found, err := svc.GetByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, found.ID)
}

// TestJobService_List는 Job 목록 조회를 테스트합니다
func TestJobService_List(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 및 Job 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	for i := 0; i < 5; i++ {
		_, err := svc.CreateJob(ctx, "TRPID-12345678")
		require.NoError(t, err)
	}

	// 전체 목록 조회
	filter := domain.DefaultJobListFilter()
	resp, err := svc.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, resp.Jobs, 5)
	assert.Equal(t, 5, resp.Total)
}

// TestJobService_ListByTransportID는 Transport ID로 필터링을 테스트합니다
func TestJobService_ListByTransportID(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// 두 개의 Transport 생성
	transport1 := domain.NewTransport("TRPID-AAA", "Test1", "", []string{"T1"})
	transport2 := domain.NewTransport("TRPID-BBB", "Test2", "", []string{"T2"})
	require.NoError(t, transportRepo.Create(ctx, transport1))
	require.NoError(t, transportRepo.Create(ctx, transport2))

	// 각 Transport에 Job 생성
	for i := 0; i < 3; i++ {
		_, err := svc.CreateJob(ctx, "TRPID-AAA")
		require.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		_, err := svc.CreateJob(ctx, "TRPID-BBB")
		require.NoError(t, err)
	}

	// Transport ID로 필터링
	filter := domain.DefaultJobListFilter()
	filter.TransportID = "TRPID-AAA"
	resp, err := svc.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, resp.Jobs, 3)
	assert.Equal(t, 3, resp.Total)
}

// TestJobService_StartJob는 Job 시작을 테스트합니다
func TestJobService_StartJob(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 및 Job 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	job, err := svc.CreateJob(ctx, "TRPID-12345678")
	require.NoError(t, err)

	// 시작
	err = svc.StartJob(ctx, job.ID)
	require.NoError(t, err)

	// 상태 확인
	found, err := svc.GetByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.JobStatusRunning, found.Status)
	assert.NotNil(t, found.StartedAt)
}

// TestJobService_CompleteJob는 Job 완료를 테스트합니다
func TestJobService_CompleteJob(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 및 Job 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	job, err := svc.CreateJob(ctx, "TRPID-12345678")
	require.NoError(t, err)

	// 시작 후 완료
	require.NoError(t, svc.StartJob(ctx, job.ID))
	require.NoError(t, svc.CompleteJob(ctx, job.ID))

	// 상태 확인
	found, err := svc.GetByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.JobStatusCompleted, found.Status)
	assert.NotNil(t, found.CompletedAt)
}

// TestJobService_FailJob는 Job 실패를 테스트합니다
func TestJobService_FailJob(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	transportRepo := memory.NewTransportRepository()
	svc := NewJobService(jobRepo, transportRepo)
	ctx := context.Background()

	// Transport 및 Job 생성
	transport := domain.NewTransport("TRPID-12345678", "Test", "Desc", []string{"TABLE1"})
	require.NoError(t, transportRepo.Create(ctx, transport))

	job, err := svc.CreateJob(ctx, "TRPID-12345678")
	require.NoError(t, err)

	// 시작 후 실패
	require.NoError(t, svc.StartJob(ctx, job.ID))
	require.NoError(t, svc.FailJob(ctx, job.ID, "테스트 에러"))

	// 상태 확인
	found, err := svc.GetByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.JobStatusFailed, found.Status)
	assert.NotNil(t, found.Error)
	assert.Equal(t, "테스트 에러", *found.Error)
}
