package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// ListCurrencyTypes resolves all registered activity-point currency definitions.
func (store *Store) ListCurrencyTypes(ctx context.Context) ([]domain.ActivityCurrencyType, error) {
	var rows []inventorymodel.CurrencyType
	if err := store.database.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ActivityCurrencyType, len(rows))
	for i, row := range rows {
		result[i] = mapCurrencyType(row)
	}
	return result, nil
}

// FindCurrencyTypeByID resolves one currency type definition by its wire-protocol ID.
func (store *Store) FindCurrencyTypeByID(ctx context.Context, id int) (domain.ActivityCurrencyType, error) {
	var row inventorymodel.CurrencyType
	err := store.database.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ActivityCurrencyType{}, fmt.Errorf("currency type %d not found", id)
	}
	if err != nil {
		return domain.ActivityCurrencyType{}, err
	}
	return mapCurrencyType(row), nil
}

// IsValidActivityPointType reports whether typeID is registered and enabled in currency_types.
func (store *Store) IsValidActivityPointType(ctx context.Context, typeID int) (bool, error) {
	var count int64
	err := store.database.WithContext(ctx).Model(&inventorymodel.CurrencyType{}).
		Where("id = ? AND enabled = true", typeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// mapCurrencyType converts one GORM model into a domain currency type definition.
func mapCurrencyType(row inventorymodel.CurrencyType) domain.ActivityCurrencyType {
	return domain.ActivityCurrencyType{
		ID: row.ID, Name: row.Name, DisplayName: row.DisplayName,
		Trackable: row.Trackable, Enabled: row.Enabled,
	}
}
