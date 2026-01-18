// Package usecase는 비즈니스 로직을 구현하는 서비스 레이어입니다.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository"
)

// JobService는 Job 비즈니스 로직을 처리합니다
type JobService struct {
	jobRepo       repository.JobRepository
	transportRepo repository.TransportRepository
}

// NewJobService는 새로운 JobService를 생성합니다
func NewJobService(jobRepo repository.JobRepository, transportRepo repository.TransportRepository) *JobService {
	return &JobService{
		jobRepo:       jobRepo,
		transportRepo: transportRepo,
	}
}

// CreateJob은 새로운 Job을 생성합니다
func (s *JobService) CreateJob(ctx context.Context, transportID string) (*domain.Job, error) {
	// Transport 존재 여부 확인
	_, err := s.transportRepo.GetByID(ctx, transportID)
	if err != nil {
		return nil, fmt.Errorf("transport를 찾을 수 없습니다: %w", err)
	}

	// 최신 버전 조회
	latestVersion, err := s.jobRepo.GetLatestVersionByTransportID(ctx, transportID)
	if err != nil {
		return nil, fmt.Errorf("최신 버전 조회 실패: %w", err)
	}

	// 새 버전
	newVersion := latestVersion + 1

	// Job ID 생성
	now := time.Now().UTC()
	randomPart := uuid.New().String()[:8]
	jobID := domain.GenerateJobID(now, randomPart)

	// Job 엔티티 생성
	job := domain.NewJob(jobID, transportID, newVersion)

	// 저장
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("job 생성 실패: %w", err)
	}

	return job, nil
}

// GetByID는 ID로 Job을 조회합니다
func (s *JobService) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	return s.jobRepo.GetByID(ctx, id)
}

// List는 필터에 따라 Job 목록을 조회합니다
func (s *JobService) List(ctx context.Context, filter domain.JobListFilter) (*domain.JobListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	jobs, total, err := s.jobRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.JobListResponse{
		Jobs:   jobs,
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}, nil
}

// StartJob은 Job을 시작 상태로 변경합니다
func (s *JobService) StartJob(ctx context.Context, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Start()

	return s.jobRepo.Update(ctx, job)
}

// CompleteJob은 Job을 완료 상태로 변경합니다
func (s *JobService) CompleteJob(ctx context.Context, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Complete()

	return s.jobRepo.Update(ctx, job)
}

// FailJob은 Job을 실패 상태로 변경합니다
func (s *JobService) FailJob(ctx context.Context, jobID string, errMsg string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Fail(errors.New(errMsg))

	return s.jobRepo.Update(ctx, job)
}

// CancelJob은 Job을 취소 상태로 변경합니다
func (s *JobService) CancelJob(ctx context.Context, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Cancel()

	return s.jobRepo.Update(ctx, job)
}

// UpdateJob은 Job을 업데이트합니다
func (s *JobService) UpdateJob(ctx context.Context, job *domain.Job) error {
	return s.jobRepo.Update(ctx, job)
}

// GetJobsByTransportID는 특정 Transport의 모든 Job을 조회합니다
func (s *JobService) GetJobsByTransportID(ctx context.Context, transportID string) ([]domain.Job, error) {
	return s.jobRepo.GetByTransportID(ctx, transportID)
}
