// Package usecase는 비즈니스 로직을 구현하는 서비스 레이어입니다.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository"
)

// TransportService는 Transport 비즈니스 로직을 처리합니다
type TransportService struct {
	repo repository.TransportRepository
}

// NewTransportService는 새로운 TransportService를 생성합니다
func NewTransportService(repo repository.TransportRepository) *TransportService {
	return &TransportService{
		repo: repo,
	}
}

// Create는 새로운 Transport를 생성합니다
func (s *TransportService) Create(ctx context.Context, req domain.CreateTransportRequest) (*domain.Transport, error) {
	// 유효성 검사
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// ID 생성
	id := domain.GenerateTransportID(uuid.New().String())

	// Transport 엔티티 생성
	transport := domain.NewTransport(id, req.Name, req.Description, req.Tables)

	// 저장
	if err := s.repo.Create(ctx, transport); err != nil {
		return nil, err
	}

	return transport, nil
}

// GetByID는 ID로 Transport를 조회합니다
func (s *TransportService) GetByID(ctx context.Context, id string) (*domain.Transport, error) {
	return s.repo.GetByID(ctx, id)
}

// List는 Transport 목록을 조회합니다
func (s *TransportService) List(ctx context.Context, offset, limit int) (*domain.TransportListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	transports, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	return &domain.TransportListResponse{
		Transports: transports,
		Total:      total,
		Offset:     offset,
		Limit:      limit,
	}, nil
}

// Delete는 Transport를 삭제합니다
func (s *TransportService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// UpdateStatus는 Transport 상태를 변경합니다
func (s *TransportService) UpdateStatus(ctx context.Context, id string, status domain.TransportStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// Update는 Transport를 수정합니다
func (s *TransportService) Update(ctx context.Context, transport *domain.Transport) error {
	return s.repo.Update(ctx, transport)
}
