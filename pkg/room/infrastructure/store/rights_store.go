package store

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RightsStore persists room rights data using PostgreSQL via GORM.
type RightsStore struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRightsStore creates one room rights repository.
func NewRightsStore(database *gorm.DB) (*RightsStore, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &RightsStore{database: database}, nil
}

// compile-time interface assertion.
var _ domain.RightsRepository = (*RightsStore)(nil)

// HasRights reports whether one user holds rights in one room.
func (s *RightsStore) HasRights(ctx context.Context, roomID int, userID int) (bool, error) {
	var count int64
	err := s.database.WithContext(ctx).Model(&model.RoomRight{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count).Error
	return count > 0, err
}

// GrantRights adds rights for one user in one room.
func (s *RightsStore) GrantRights(ctx context.Context, roomID int, userID int) error {
	row := model.RoomRight{RoomID: uint(roomID), UserID: uint(userID)}
	return s.database.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// RevokeRights removes rights for one user in one room.
func (s *RightsStore) RevokeRights(ctx context.Context, roomID int, userID int) error {
	return s.database.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.RoomRight{}).Error
}

// ListRightsByRoom returns all rights holders for one room.
func (s *RightsStore) ListRightsByRoom(ctx context.Context, roomID int) ([]int, error) {
	var rows []model.RoomRight
	if err := s.database.WithContext(ctx).Where("room_id = ?", roomID).Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int, len(rows))
	for i, row := range rows {
		ids[i] = int(row.UserID)
	}
	return ids, nil
}

// RevokeAllRights removes all rights for one room.
func (s *RightsStore) RevokeAllRights(ctx context.Context, roomID int) error {
	return s.database.WithContext(ctx).Where("room_id = ?", roomID).Delete(&model.RoomRight{}).Error
}
