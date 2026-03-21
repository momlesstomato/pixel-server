package store

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// ListEffects resolves all effect rows for one user.
func (store *Store) ListEffects(ctx context.Context, userID int) ([]domain.Effect, error) {
	var rows []inventorymodel.Effect
	if err := store.database.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Effect, len(rows))
	for i, row := range rows {
		result[i] = mapEffect(row)
	}
	return result, nil
}

// AwardEffect persists or increments one effect for one user.
func (store *Store) AwardEffect(ctx context.Context, userID int, effectID int, duration int, permanent bool) (domain.Effect, error) {
	var existing inventorymodel.Effect
	err := store.database.WithContext(ctx).Where("user_id = ? AND effect_id = ?", userID, effectID).First(&existing).Error
	if err == nil {
		existing.Quantity++
		if err := store.database.WithContext(ctx).Save(&existing).Error; err != nil {
			return domain.Effect{}, err
		}
		return mapEffect(existing), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Effect{}, err
	}
	row := inventorymodel.Effect{UserID: uint(userID), EffectID: effectID, Duration: duration, Quantity: 1, IsPermanent: permanent}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Effect{}, err
	}
	return mapEffect(row), nil
}

// ActivateEffect sets activation timestamp for one effect.
func (store *Store) ActivateEffect(ctx context.Context, userID int, effectID int) (domain.Effect, error) {
	now := time.Now().UTC()
	result := store.database.WithContext(ctx).Model(&inventorymodel.Effect{}).
		Where("user_id = ? AND effect_id = ? AND activated_at IS NULL", userID, effectID).
		Updates(map[string]any{"activated_at": now, "quantity": gorm.Expr("quantity - 1")})
	if result.Error != nil {
		return domain.Effect{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Effect{}, domain.ErrEffectNotFound
	}
	var row inventorymodel.Effect
	if err := store.database.WithContext(ctx).Where("user_id = ? AND effect_id = ?", userID, effectID).First(&row).Error; err != nil {
		return domain.Effect{}, err
	}
	return mapEffect(row), nil
}

// RemoveExpiredEffects deletes all expired effects and returns removed IDs.
func (store *Store) RemoveExpiredEffects(ctx context.Context) ([]domain.ExpiredEffect, error) {
	var rows []inventorymodel.Effect
	err := store.database.WithContext(ctx).
		Where("is_permanent = false AND activated_at IS NOT NULL AND activated_at + make_interval(secs => duration) < NOW()").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(rows))
	result := make([]domain.ExpiredEffect, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		result[i] = domain.ExpiredEffect{UserID: int(row.UserID), EffectID: row.EffectID}
	}
	if len(ids) > 0 {
		store.database.WithContext(ctx).Where("id IN ?", ids).Delete(&inventorymodel.Effect{})
	}
	return result, nil
}

// mapEffect converts one GORM model into domain effect.
func mapEffect(row inventorymodel.Effect) domain.Effect {
	return domain.Effect{
		ID: int(row.ID), UserID: int(row.UserID), EffectID: row.EffectID,
		Duration: row.Duration, Quantity: row.Quantity,
		ActivatedAt: row.ActivatedAt, IsPermanent: row.IsPermanent,
		CreatedAt: row.CreatedAt,
	}
}
