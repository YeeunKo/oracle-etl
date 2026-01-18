// Package memory는 개발 및 테스트용 인메모리 저장소 구현을 제공합니다.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository"
)

// JobRepository는 인메모리 Job 저장소 구현입니다
type JobRepository struct {
	mu   sync.RWMutex
	jobs map[string]*domain.Job
}

// NewJobRepository는 새로운 인메모리 Job 저장소를 생성합니다
func NewJobRepository() repository.JobRepository {
	return &JobRepository{
		jobs: make(map[string]*domain.Job),
	}
}

// Create는 새로운 Job을 생성합니다
func (r *JobRepository) Create(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; exists {
		return fmt.Errorf("job ID '%s'가 이미 존재합니다", job.ID)
	}

	// 복사본 저장
	copied := *job
	copied.Extractions = make([]domain.Extraction, len(job.Extractions))
	copy(copied.Extractions, job.Extractions)
	r.jobs[job.ID] = &copied

	return nil
}

// GetByID는 ID로 Job을 조회합니다
func (r *JobRepository) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, exists := r.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job ID '%s'를 찾을 수 없습니다", id)
	}

	// 복사본 반환
	copied := *job
	copied.Extractions = make([]domain.Extraction, len(job.Extractions))
	copy(copied.Extractions, job.Extractions)
	return &copied, nil
}

// List는 필터에 따라 Job 목록을 조회합니다
func (r *JobRepository) List(ctx context.Context, filter domain.JobListFilter) ([]domain.Job, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 필터 적용
	list := make([]domain.Job, 0)
	for _, j := range r.jobs {
		// TransportID 필터
		if filter.TransportID != "" && j.TransportID != filter.TransportID {
			continue
		}
		// Status 필터
		if filter.Status != "" && j.Status != filter.Status {
			continue
		}
		list = append(list, *j)
	}

	// 생성 시간 기준 정렬 (최신순)
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)

	// offset/limit 적용
	if filter.Offset >= len(list) {
		return []domain.Job{}, total, nil
	}

	end := filter.Offset + filter.Limit
	if end > len(list) {
		end = len(list)
	}

	return list[filter.Offset:end], total, nil
}

// Update는 Job을 수정합니다
func (r *JobRepository) Update(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[job.ID]; !exists {
		return fmt.Errorf("job ID '%s'를 찾을 수 없습니다", job.ID)
	}

	// 복사본 저장
	copied := *job
	copied.Extractions = make([]domain.Extraction, len(job.Extractions))
	copy(copied.Extractions, job.Extractions)
	r.jobs[job.ID] = &copied

	return nil
}

// GetLatestVersionByTransportID는 특정 Transport의 최신 Job 버전을 반환합니다
func (r *JobRepository) GetLatestVersionByTransportID(ctx context.Context, transportID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	maxVersion := 0
	for _, j := range r.jobs {
		if j.TransportID == transportID && j.Version > maxVersion {
			maxVersion = j.Version
		}
	}

	return maxVersion, nil
}

// GetByTransportID는 특정 Transport의 모든 Job을 조회합니다
func (r *JobRepository) GetByTransportID(ctx context.Context, transportID string) ([]domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Job, 0)
	for _, j := range r.jobs {
		if j.TransportID == transportID {
			list = append(list, *j)
		}
	}

	// 버전 기준 정렬 (최신순)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Version > list[j].Version
	})

	return list, nil
}
