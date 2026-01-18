// Package repository는 데이터 저장소 인터페이스를 정의합니다.
package repository

import (
	"context"

	"oracle-etl/internal/domain"
)

// TransportRepository는 Transport 저장소 인터페이스입니다
type TransportRepository interface {
	// Create는 새로운 Transport를 생성합니다
	Create(ctx context.Context, transport *domain.Transport) error

	// GetByID는 ID로 Transport를 조회합니다
	GetByID(ctx context.Context, id string) (*domain.Transport, error)

	// List는 Transport 목록을 조회합니다
	List(ctx context.Context, offset, limit int) ([]domain.Transport, int, error)

	// Update는 Transport를 수정합니다
	Update(ctx context.Context, transport *domain.Transport) error

	// Delete는 Transport를 삭제합니다
	Delete(ctx context.Context, id string) error

	// UpdateStatus는 Transport 상태를 변경합니다
	UpdateStatus(ctx context.Context, id string, status domain.TransportStatus) error
}
