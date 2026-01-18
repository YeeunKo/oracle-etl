// Package memory는 개발 및 테스트용 인메모리 저장소 구현을 제공합니다.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"oracle-etl/internal/domain"
	"oracle-etl/internal/repository"
)

// TransportRepository는 인메모리 Transport 저장소 구현입니다
type TransportRepository struct {
	mu         sync.RWMutex
	transports map[string]*domain.Transport
}

// NewTransportRepository는 새로운 인메모리 Transport 저장소를 생성합니다
func NewTransportRepository() repository.TransportRepository {
	return &TransportRepository{
		transports: make(map[string]*domain.Transport),
	}
}

// Create는 새로운 Transport를 생성합니다
func (r *TransportRepository) Create(ctx context.Context, transport *domain.Transport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transports[transport.ID]; exists {
		return fmt.Errorf("transport ID '%s'가 이미 존재합니다", transport.ID)
	}

	// 복사본 저장
	copied := *transport
	r.transports[transport.ID] = &copied

	return nil
}

// GetByID는 ID로 Transport를 조회합니다
func (r *TransportRepository) GetByID(ctx context.Context, id string) (*domain.Transport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transport, exists := r.transports[id]
	if !exists {
		return nil, fmt.Errorf("transport ID '%s'를 찾을 수 없습니다", id)
	}

	// 복사본 반환
	copied := *transport
	return &copied, nil
}

// List는 Transport 목록을 조회합니다
func (r *TransportRepository) List(ctx context.Context, offset, limit int) ([]domain.Transport, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 전체 목록을 슬라이스로 변환
	list := make([]domain.Transport, 0, len(r.transports))
	for _, t := range r.transports {
		list = append(list, *t)
	}

	// 생성 시간 기준 정렬 (최신순)
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)

	// offset/limit 적용
	if offset >= len(list) {
		return []domain.Transport{}, total, nil
	}

	end := offset + limit
	if end > len(list) {
		end = len(list)
	}

	return list[offset:end], total, nil
}

// Update는 Transport를 수정합니다
func (r *TransportRepository) Update(ctx context.Context, transport *domain.Transport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transports[transport.ID]; !exists {
		return fmt.Errorf("transport ID '%s'를 찾을 수 없습니다", transport.ID)
	}

	// UpdatedAt 갱신
	transport.UpdatedAt = time.Now().UTC()

	// 복사본 저장
	copied := *transport
	r.transports[transport.ID] = &copied

	return nil
}

// Delete는 Transport를 삭제합니다
func (r *TransportRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transports[id]; !exists {
		return fmt.Errorf("transport ID '%s'를 찾을 수 없습니다", id)
	}

	delete(r.transports, id)
	return nil
}

// UpdateStatus는 Transport 상태를 변경합니다
func (r *TransportRepository) UpdateStatus(ctx context.Context, id string, status domain.TransportStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transport, exists := r.transports[id]
	if !exists {
		return fmt.Errorf("transport ID '%s'를 찾을 수 없습니다", id)
	}

	transport.Status = status
	transport.UpdatedAt = time.Now().UTC()

	return nil
}
