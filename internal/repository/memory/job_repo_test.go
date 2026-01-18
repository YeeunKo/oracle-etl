package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/domain"
)

// TestJobRepo_Create는 Job 생성을 테스트합니다
func TestJobRepo_Create(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	job := domain.NewJob("JOB-20260118-120000-abc", "TRPID-12345678", 1)

	err := repo.Create(ctx, job)
	require.NoError(t, err)

	// 중복 생성 시도
	err = repo.Create(ctx, job)
	assert.Error(t, err)
}

// TestJobRepo_GetByID는 ID로 Job 조회를 테스트합니다
func TestJobRepo_GetByID(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	// 존재하지 않는 ID 조회
	_, err := repo.GetByID(ctx, "non-existent")
	assert.Error(t, err)

	// Job 생성 후 조회
	job := domain.NewJob("JOB-20260118-120000-abc", "TRPID-12345678", 1)
	require.NoError(t, repo.Create(ctx, job))

	found, err := repo.GetByID(ctx, "JOB-20260118-120000-abc")
	require.NoError(t, err)
	assert.Equal(t, "TRPID-12345678", found.TransportID)
	assert.Equal(t, 1, found.Version)
}

// TestJobRepo_List는 Job 목록 조회를 테스트합니다
func TestJobRepo_List(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	// 빈 목록
	filter := domain.DefaultJobListFilter()
	list, total, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, 0, total)

	// 여러 Job 생성
	now := time.Now()
	for i := 0; i < 5; i++ {
		job := domain.NewJob(
			domain.GenerateJobID(now.Add(time.Duration(i)*time.Second), "abc"),
			"TRPID-12345678",
			i+1,
		)
		require.NoError(t, repo.Create(ctx, job))
	}

	// 페이지네이션 테스트
	filter.Limit = 3
	filter.Offset = 0
	list, total, err = repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, 5, total)
}

// TestJobRepo_ListByTransportID는 Transport ID로 필터링을 테스트합니다
func TestJobRepo_ListByTransportID(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	// 서로 다른 Transport의 Job 생성
	now := time.Now()
	for i := 0; i < 3; i++ {
		job := domain.NewJob(
			domain.GenerateJobID(now.Add(time.Duration(i)*time.Second), "aaa"),
			"TRPID-AAA",
			i+1,
		)
		require.NoError(t, repo.Create(ctx, job))
	}
	for i := 0; i < 2; i++ {
		job := domain.NewJob(
			domain.GenerateJobID(now.Add(time.Duration(i+10)*time.Second), "bbb"),
			"TRPID-BBB",
			i+1,
		)
		require.NoError(t, repo.Create(ctx, job))
	}

	// Transport ID로 필터링
	filter := domain.DefaultJobListFilter()
	filter.TransportID = "TRPID-AAA"
	list, total, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, 3, total)
}

// TestJobRepo_ListByStatus는 상태로 필터링을 테스트합니다
func TestJobRepo_ListByStatus(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	now := time.Now()
	// 다양한 상태의 Job 생성
	job1 := domain.NewJob(domain.GenerateJobID(now, "a"), "TRPID-AAA", 1)
	job1.Status = domain.JobStatusCompleted
	require.NoError(t, repo.Create(ctx, job1))

	job2 := domain.NewJob(domain.GenerateJobID(now.Add(time.Second), "b"), "TRPID-AAA", 2)
	job2.Status = domain.JobStatusFailed
	require.NoError(t, repo.Create(ctx, job2))

	job3 := domain.NewJob(domain.GenerateJobID(now.Add(2*time.Second), "c"), "TRPID-AAA", 3)
	job3.Status = domain.JobStatusCompleted
	require.NoError(t, repo.Create(ctx, job3))

	// 상태로 필터링
	filter := domain.DefaultJobListFilter()
	filter.Status = domain.JobStatusCompleted
	list, total, err := repo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, 2, total)
}

// TestJobRepo_GetLatestVersionByTransportID는 최신 버전 조회를 테스트합니다
func TestJobRepo_GetLatestVersionByTransportID(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	// Job이 없는 경우
	version, err := repo.GetLatestVersionByTransportID(ctx, "TRPID-NEW")
	require.NoError(t, err)
	assert.Equal(t, 0, version)

	// Job 생성
	now := time.Now()
	for i := 1; i <= 3; i++ {
		job := domain.NewJob(
			domain.GenerateJobID(now.Add(time.Duration(i)*time.Second), "abc"),
			"TRPID-12345678",
			i,
		)
		require.NoError(t, repo.Create(ctx, job))
	}

	// 최신 버전 확인
	version, err = repo.GetLatestVersionByTransportID(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.Equal(t, 3, version)
}

// TestJobRepo_Update는 Job 수정을 테스트합니다
func TestJobRepo_Update(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	job := domain.NewJob("JOB-20260118-120000-abc", "TRPID-12345678", 1)
	require.NoError(t, repo.Create(ctx, job))

	// 상태 변경 및 업데이트
	job.Start()
	err := repo.Update(ctx, job)
	require.NoError(t, err)

	// 확인
	found, err := repo.GetByID(ctx, "JOB-20260118-120000-abc")
	require.NoError(t, err)
	assert.Equal(t, domain.JobStatusRunning, found.Status)
	assert.NotNil(t, found.StartedAt)
}

// TestJobRepo_GetByTransportID는 Transport ID로 Job 조회를 테스트합니다
func TestJobRepo_GetByTransportID(t *testing.T) {
	repo := NewJobRepository()
	ctx := context.Background()

	now := time.Now()
	for i := 1; i <= 3; i++ {
		job := domain.NewJob(
			domain.GenerateJobID(now.Add(time.Duration(i)*time.Second), "abc"),
			"TRPID-12345678",
			i,
		)
		require.NoError(t, repo.Create(ctx, job))
	}

	jobs, err := repo.GetByTransportID(ctx, "TRPID-12345678")
	require.NoError(t, err)
	assert.Len(t, jobs, 3)
}
