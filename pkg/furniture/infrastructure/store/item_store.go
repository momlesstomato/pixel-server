package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnituremodel "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/model"
	"gorm.io/gorm"
)

// FindItemByID resolves one item instance by identifier.
func (store *Store) FindItemByID(ctx context.Context, id int) (domain.Item, error) {
	var row furnituremodel.Item
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Item{}, domain.ErrItemNotFound
		}
		return domain.Item{}, err
	}
	return mapItem(row), nil
}

// ListItemsByUserID resolves all inventory items for one user.
func (store *Store) ListItemsByUserID(ctx context.Context, userID int) ([]domain.Item, error) {
	var rows []furnituremodel.Item
	if err := store.database.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Item, len(rows))
	for i, row := range rows {
		result[i] = mapItem(row)
	}
	return result, nil
}

// CreateItem persists one item instance.
func (store *Store) CreateItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	row := toItemRecord(item)
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Item{}, err
	}
	return mapItem(row), nil
}

// DeleteItem removes one item instance by identifier.
func (store *Store) DeleteItem(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&furnituremodel.Item{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// TransferItem changes item ownership atomically.
func (store *Store) TransferItem(ctx context.Context, itemID int, newUserID int) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).Update("user_id", newUserID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// UpdateItemData updates the custom data payload for one item.
func (store *Store) UpdateItemData(ctx context.Context, itemID int, extraData string) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).Update("extra_data", extraData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// UpdateItemInteractionData updates the hidden interaction payload for one item.
func (store *Store) UpdateItemInteractionData(ctx context.Context, itemID int, interactionData string) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).Update("interaction_data", interactionData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// PlaceItem updates item room placement coordinates.
func (store *Store) PlaceItem(ctx context.Context, itemID int, roomID int, x int, y int, z float64, dir int) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).
		Updates(map[string]interface{}{"room_id": roomID, "x": x, "y": y, "z": z, "dir": dir, "wall_position": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// PlaceWallItem updates item wall placement coordinates.
func (store *Store) PlaceWallItem(ctx context.Context, itemID int, roomID int, wallPosition string) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).
		Updates(map[string]interface{}{"room_id": roomID, "x": 0, "y": 0, "z": 0, "dir": 0, "wall_position": wallPosition})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// UpdateItemDefinition transforms one item into another definition payload.
func (store *Store) UpdateItemDefinition(ctx context.Context, itemID int, definitionID int, extraData string, interactionData string) error {
	result := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("id = ?", itemID).
		Updates(map[string]interface{}{"definition_id": definitionID, "extra_data": extraData, "interaction_data": interactionData})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// ListItemsByRoomID resolves all placed items in one room.
func (store *Store) ListItemsByRoomID(ctx context.Context, roomID int) ([]domain.Item, error) {
	var rows []furnituremodel.Item
	if err := store.database.WithContext(ctx).Where("room_id = ? AND room_id > 0", roomID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Item, len(rows))
	for i, row := range rows {
		result[i] = mapItem(row)
	}
	return result, nil
}

// CountItemsByUserID returns item count for one user inventory.
func (store *Store) CountItemsByUserID(ctx context.Context, userID int) (int, error) {
	var count int64
	if err := store.database.WithContext(ctx).Model(&furnituremodel.Item{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
