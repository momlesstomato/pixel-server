package application

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// VisitService implements room visit tracking business logic.
type VisitService struct {
	// repo stores the visit persistence layer.
	repo domain.VisitRepository
}

// NewVisitService creates a visit tracking service.
func NewVisitService(repo domain.VisitRepository) (*VisitService, error) {
	if repo == nil {
		return nil, domain.ErrMissingTarget
	}
	return &VisitService{repo: repo}, nil
}

// RecordVisit persists a room visit for one user.
func (s *VisitService) RecordVisit(ctx context.Context, userID int, roomID int) error {
	record := &domain.VisitRecord{UserID: userID, RoomID: roomID}
	return s.repo.Record(ctx, record)
}

// ListByUser returns recent visits for one user.
func (s *VisitService) ListByUser(ctx context.Context, userID int, limit int) ([]domain.VisitRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListByUser(ctx, userID, limit)
}

// ListByRoom returns recent visits for one room.
func (s *VisitService) ListByRoom(ctx context.Context, roomID int, limit int) ([]domain.VisitRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListByRoom(ctx, roomID, limit)
}
