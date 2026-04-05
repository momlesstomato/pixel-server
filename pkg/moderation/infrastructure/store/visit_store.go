package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// VisitStore implements domain.VisitRepository using GORM.
type VisitStore struct {
	// db stores the GORM database connection.
	db *gorm.DB
}

// NewVisitStore creates a visit store backed by the given database.
func NewVisitStore(db *gorm.DB) (*VisitStore, error) {
	if db == nil {
		return nil, domain.ErrMissingTarget
	}
	return &VisitStore{db: db}, nil
}

// Record persists a new room visit record.
func (s *VisitStore) Record(ctx context.Context, record *domain.VisitRecord) error {
	m := model.ModerationRoomVisit{UserID: record.UserID, RoomID: record.RoomID, VisitedAt: time.Now()}
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	record.ID = m.ID
	record.VisitedAt = m.VisitedAt
	return nil
}

// ListByUser returns recent room visits for one user.
func (s *VisitStore) ListByUser(ctx context.Context, userID int, limit int) ([]domain.VisitRecord, error) {
	var rows []model.ModerationRoomVisit
	q := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("visited_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return visitsToDomain(rows), nil
}

// ListByRoom returns recent visits for one room.
func (s *VisitStore) ListByRoom(ctx context.Context, roomID int, limit int) ([]domain.VisitRecord, error) {
	var rows []model.ModerationRoomVisit
	q := s.db.WithContext(ctx).Where("room_id = ?", roomID).Order("visited_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return visitsToDomain(rows), nil
}

func visitsToDomain(rows []model.ModerationRoomVisit) []domain.VisitRecord {
	out := make([]domain.VisitRecord, len(rows))
	for i := range rows {
		out[i] = domain.VisitRecord{ID: rows[i].ID, UserID: rows[i].UserID, RoomID: rows[i].RoomID, VisitedAt: rows[i].VisitedAt}
	}
	return out
}
