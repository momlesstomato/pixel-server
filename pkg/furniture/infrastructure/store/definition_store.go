package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnituremodel "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/model"
	"gorm.io/gorm"
)

// FindDefinitionByID resolves one item definition by identifier.
func (store *Store) FindDefinitionByID(ctx context.Context, id int) (domain.Definition, error) {
	var row furnituremodel.Definition
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Definition{}, domain.ErrDefinitionNotFound
		}
		return domain.Definition{}, err
	}
	return mapDefinition(row), nil
}

// FindDefinitionByName resolves one item definition by internal name.
func (store *Store) FindDefinitionByName(ctx context.Context, name string) (domain.Definition, error) {
	var row furnituremodel.Definition
	if err := store.database.WithContext(ctx).Where("item_name = ?", name).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Definition{}, domain.ErrDefinitionNotFound
		}
		return domain.Definition{}, err
	}
	return mapDefinition(row), nil
}

// ListDefinitions resolves all item definition rows.
func (store *Store) ListDefinitions(ctx context.Context) ([]domain.Definition, error) {
	var rows []furnituremodel.Definition
	if err := store.database.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Definition, len(rows))
	for i, row := range rows {
		result[i] = mapDefinition(row)
	}
	return result, nil
}

// CreateDefinition persists one item definition row.
func (store *Store) CreateDefinition(ctx context.Context, def domain.Definition) (domain.Definition, error) {
	row := furnituremodel.Definition{
		ItemName: def.ItemName, PublicName: def.PublicName,
		ItemType: string(def.ItemType), Width: int16(def.Width), Length: int16(def.Length),
		StackHeight: def.StackHeight, CanStack: def.CanStack, CanSit: def.CanSit, CanLay: def.CanLay,
		IsWalkable: def.IsWalkable, SpriteID: def.SpriteID,
		AllowRecycle: def.AllowRecycle, AllowTrade: def.AllowTrade,
		AllowMarketplaceSell: def.AllowMarketplaceSell, AllowGift: def.AllowGift,
		AllowInventoryStack:   def.AllowInventoryStack,
		InteractionType:       string(def.InteractionType),
		InteractionModesCount: int16(def.InteractionModesCount),
		EffectID:              def.EffectID,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Definition{}, err
	}
	return mapDefinition(row), nil
}

// UpdateDefinition applies partial definition update.
func (store *Store) UpdateDefinition(ctx context.Context, id int, patch domain.DefinitionPatch) (domain.Definition, error) {
	updates := map[string]any{}
	if patch.PublicName != nil {
		updates["public_name"] = *patch.PublicName
	}
	if patch.StackHeight != nil {
		updates["stack_height"] = *patch.StackHeight
	}
	if patch.CanStack != nil {
		updates["can_stack"] = *patch.CanStack
	}
	if patch.CanLay != nil {
		updates["can_lay"] = *patch.CanLay
	}
	if patch.AllowTrade != nil {
		updates["allow_trade"] = *patch.AllowTrade
	}
	if patch.AllowMarketplaceSell != nil {
		updates["allow_marketplace_sell"] = *patch.AllowMarketplaceSell
	}
	if patch.AllowGift != nil {
		updates["allow_gift"] = *patch.AllowGift
	}
	if patch.InteractionType != nil {
		updates["interaction_type"] = *patch.InteractionType
	}
	if len(updates) > 0 {
		result := store.database.WithContext(ctx).Model(&furnituremodel.Definition{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return domain.Definition{}, result.Error
		}
		if result.RowsAffected == 0 {
			return domain.Definition{}, domain.ErrDefinitionNotFound
		}
	}
	return store.FindDefinitionByID(ctx, id)
}

// DeleteDefinition removes one item definition by identifier.
func (store *Store) DeleteDefinition(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&furnituremodel.Definition{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDefinitionNotFound
	}
	return nil
}
