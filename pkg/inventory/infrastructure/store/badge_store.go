package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// ListBadges resolves all badge rows for one user.
func (store *Store) ListBadges(ctx context.Context, userID int) ([]domain.Badge, error) {
	var rows []inventorymodel.Badge
	if err := store.database.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Badge, len(rows))
	for i, row := range rows {
		result[i] = domain.Badge{ID: int(row.ID), UserID: int(row.UserID), BadgeCode: row.BadgeCode, SlotID: int(row.SlotID), CreatedAt: row.CreatedAt}
	}
	return result, nil
}

// AwardBadge persists one badge for one user.
func (store *Store) AwardBadge(ctx context.Context, userID int, badgeCode string) (domain.Badge, error) {
	row := inventorymodel.Badge{UserID: uint(userID), BadgeCode: badgeCode}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Badge{}, err
	}
	return domain.Badge{ID: int(row.ID), UserID: int(row.UserID), BadgeCode: row.BadgeCode, CreatedAt: row.CreatedAt}, nil
}

// RevokeBadge removes one badge by user and code.
func (store *Store) RevokeBadge(ctx context.Context, userID int, badgeCode string) error {
	result := store.database.WithContext(ctx).Where("user_id = ? AND badge_code = ?", userID, badgeCode).Delete(&inventorymodel.Badge{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrBadgeNotFound
	}
	return nil
}

// UpdateBadgeSlots replaces equipped badge slot assignments for one user.
func (store *Store) UpdateBadgeSlots(ctx context.Context, userID int, slots []domain.BadgeSlot) error {
	return store.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&inventorymodel.Badge{}).Where("user_id = ? AND slot_id > 0", userID).Update("slot_id", 0).Error; err != nil {
			return err
		}
		for _, slot := range slots {
			result := tx.Model(&inventorymodel.Badge{}).Where("user_id = ? AND badge_code = ?", userID, slot.BadgeCode).Update("slot_id", slot.SlotID)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

// GetEquippedBadges resolves currently equipped badge slots for one user.
func (store *Store) GetEquippedBadges(ctx context.Context, userID int) ([]domain.BadgeSlot, error) {
	var rows []inventorymodel.Badge
	err := store.database.WithContext(ctx).Where("user_id = ? AND slot_id > 0", userID).Order("slot_id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.BadgeSlot, len(rows))
	for i, row := range rows {
		result[i] = domain.BadgeSlot{SlotID: int(row.SlotID), BadgeCode: row.BadgeCode}
	}
	return result, nil
}
