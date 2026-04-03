package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
)

// BanStore persists room ban data using PostgreSQL via GORM.
type BanStore struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewBanStore creates one room ban repository.
func NewBanStore(database *gorm.DB) (*BanStore, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &BanStore{database: database}, nil
}

// compile-time interface assertion.
var _ domain.BanRepository = (*BanStore)(nil)

// FindActiveBan resolves an active ban for one user in one room.
func (s *BanStore) FindActiveBan(ctx context.Context, roomID int, userID int) (*domain.RoomBan, error) {
	var row model.RoomBan
	query := s.database.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())
	if err := query.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	ban := mapBan(row)
	return &ban, nil
}

// CreateBan persists one room ban entry.
func (s *BanStore) CreateBan(ctx context.Context, ban domain.RoomBan) (domain.RoomBan, error) {
	row := model.RoomBan{
		RoomID: uint(ban.RoomID), UserID: uint(ban.UserID), ExpiresAt: ban.ExpiresAt,
	}
	if err := s.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.RoomBan{}, err
	}
	return mapBan(row), nil
}

// DeleteBan removes one room ban by identifier.
func (s *BanStore) DeleteBan(ctx context.Context, id int) error {
	return s.database.WithContext(ctx).Delete(&model.RoomBan{}, id).Error
}

// ListBansByRoom returns all active bans for one room.
func (s *BanStore) ListBansByRoom(ctx context.Context, roomID int) ([]domain.RoomBan, error) {
	var rows []model.RoomBan
	err := s.database.WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.RoomBan, len(rows))
	for i, row := range rows {
		result[i] = mapBan(row)
	}
	return result, nil
}

// mapBan converts persistence model to domain type.
func mapBan(row model.RoomBan) domain.RoomBan {
	return domain.RoomBan{
		ID: int(row.ID), RoomID: int(row.RoomID), UserID: int(row.UserID),
		ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}
}
