// Package repository는 데이터 저장소 인터페이스를 정의합니다.
package repository

import (
	"context"

	"oracle-etl/internal/domain"
)

// JobRepository는 Job 저장소 인터페이스입니다
type JobRepository interface {
	// Create는 새로운 Job을 생성합니다
	Create(ctx context.Context, job *domain.Job) error

	// GetByID는 ID로 Job을 조회합니다
	GetByID(ctx context.Context, id string) (*domain.Job, error)

	// List는 필터에 따라 Job 목록을 조회합니다
	List(ctx context.Context, filter domain.JobListFilter) ([]domain.Job, int, error)

	// Update는 Job을 수정합니다
	Update(ctx context.Context, job *domain.Job) error

	// GetLatestVersionByTransportID는 특정 Transport의 최신 Job 버전을 반환합니다
	GetLatestVersionByTransportID(ctx context.Context, transportID string) (int, error)

	// GetByTransportID는 특정 Transport의 모든 Job을 조회합니다
	GetByTransportID(ctx context.Context, transportID string) ([]domain.Job, error)
}
