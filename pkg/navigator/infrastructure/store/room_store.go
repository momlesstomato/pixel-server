package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	model "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	"gorm.io/gorm"
)

// ListRooms resolves paginated rooms with optional filter.
func (s *Store) ListRooms(ctx context.Context, filter domain.RoomFilter) ([]domain.Room, int, error) {
	query := s.database.WithContext(ctx).Model(&model.Room{})
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.OwnerID != nil {
		query = query.Where("owner_id = ?", *filter.OwnerID)
	}
	if filter.SearchQuery != "" {
		like := "%" + filter.SearchQuery + "%"
		query = query.Where("name ILIKE ? OR tags ILIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Room
	if err := query.Order("score DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]domain.Room, len(rows))
	for i, row := range rows {
		result[i] = mapRoom(row)
	}
	return result, int(total), nil
}

// FindRoomByID resolves one room by identifier.
func (s *Store) FindRoomByID(ctx context.Context, id int) (domain.Room, error) {
	var row model.Room
	if err := s.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Room{}, domain.ErrRoomNotFound
		}
		return domain.Room{}, err
	}
	return mapRoom(row), nil
}

// CreateRoom persists one room row.
func (s *Store) CreateRoom(ctx context.Context, room domain.Room) (domain.Room, error) {
	row := model.Room{
		OwnerID: uint(room.OwnerID), OwnerName: room.OwnerName,
		Name: room.Name, Description: room.Description, State: room.State,
		CategoryID: uint(room.CategoryID), MaxUsers: room.MaxUsers,
		Tags: joinTags(room.Tags), TradeMode: room.TradeMode,
	}
	if err := s.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Room{}, err
	}
	return mapRoom(row), nil
}

// UpdateRoom applies partial room update.
func (s *Store) UpdateRoom(ctx context.Context, id int, patch domain.RoomPatch) (domain.Room, error) {
	updates := buildRoomUpdates(patch)
	if len(updates) == 0 {
		return s.FindRoomByID(ctx, id)
	}
	result := s.database.WithContext(ctx).Model(&model.Room{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return domain.Room{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	return s.FindRoomByID(ctx, id)
}

// DeleteRoom removes one room by identifier.
func (s *Store) DeleteRoom(ctx context.Context, id int) error {
	result := s.database.WithContext(ctx).Delete(&model.Room{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

// buildRoomUpdates constructs GORM update map from patch.
func buildRoomUpdates(patch domain.RoomPatch) map[string]any {
	updates := map[string]any{}
	if patch.Name != nil {
		updates["name"] = *patch.Name
	}
	if patch.Description != nil {
		updates["description"] = *patch.Description
	}
	if patch.State != nil {
		updates["state"] = *patch.State
	}
	if patch.CategoryID != nil {
		updates["category_id"] = *patch.CategoryID
	}
	if patch.MaxUsers != nil {
		updates["max_users"] = *patch.MaxUsers
	}
	return updates
}
