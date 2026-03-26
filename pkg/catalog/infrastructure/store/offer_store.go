package store

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// offerWithName is the base SQL for loading catalog offers. The display name,
// sprite ID, and item type are always resolved from the linked item definition
// so no separate columns are stored on the offer itself.
const offerWithName = `
	SELECT ci.*, id.public_name AS effective_name,
	       id.sprite_id AS effective_sprite_id,
	       id.item_type AS effective_item_type
	FROM catalog_items ci
	LEFT JOIN item_definitions id ON id.id = ci.item_definition_id
`

// resolvedOffer extends the catalog offer model with display name, sprite ID,
// and item type joined from the item_definitions table.
type resolvedOffer struct {
	catalogmodel.Offer
	// EffectiveName carries the public_name from the linked item definition.
	EffectiveName string `gorm:"column:effective_name"`
	// EffectiveSpriteID carries the sprite_id from the linked item definition.
	EffectiveSpriteID int `gorm:"column:effective_sprite_id"`
	// EffectiveItemType carries the item_type from the linked item definition.
	EffectiveItemType string `gorm:"column:effective_item_type"`
}

// mapResolvedOffer converts one resolved offer scan into a domain catalog offer
// with CatalogName, SpriteID, and ItemType set from the joined item definition.
func mapResolvedOffer(row resolvedOffer) domain.CatalogOffer {
	o := mapOffer(row.Offer)
	o.CatalogName = row.EffectiveName
	o.SpriteID = row.EffectiveSpriteID
	o.ItemType = row.EffectiveItemType
	return o
}

// ListOffersByPageID resolves all offers for one catalog page.
func (store *Store) ListOffersByPageID(ctx context.Context, pageID int) ([]domain.CatalogOffer, error) {
	var rows []resolvedOffer
	if err := store.database.WithContext(ctx).Raw(offerWithName+"WHERE ci.page_id = ? ORDER BY ci.order_num ASC", pageID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CatalogOffer, len(rows))
	for i, row := range rows {
		result[i] = mapResolvedOffer(row)
	}
	return result, nil
}

// FindOfferByID resolves one catalog offer by identifier.
func (store *Store) FindOfferByID(ctx context.Context, id int) (domain.CatalogOffer, error) {
	var row resolvedOffer
	res := store.database.WithContext(ctx).Raw(offerWithName+"WHERE ci.id = ?", id).Scan(&row)
	if res.Error != nil {
		return domain.CatalogOffer{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.CatalogOffer{}, domain.ErrOfferNotFound
	}
	return mapResolvedOffer(row), nil
}

// CreateOffer persists one catalog offer row.
func (store *Store) CreateOffer(ctx context.Context, offer domain.CatalogOffer) (domain.CatalogOffer, error) {
	row := catalogmodel.Offer{
		PageID: uint(offer.PageID), ItemDefinitionID: uint(offer.ItemDefinitionID),
		CostCredits: offer.CostCredits,
		CostActivityPoints: offer.CostActivityPoints, ActivityPointType: offer.ActivityPointType,
		Amount: offer.Amount, LimitedTotal: offer.LimitedTotal,
		OfferActive: offer.OfferActive, ExtraData: offer.ExtraData,
		BadgeID: offer.BadgeID, ClubOnly: offer.ClubOnly, OrderNum: offer.OrderNum,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.CatalogOffer{}, err
	}
	return store.FindOfferByID(ctx, int(row.ID))
}

// UpdateOffer applies partial offer update.
func (store *Store) UpdateOffer(ctx context.Context, id int, patch domain.OfferPatch) (domain.CatalogOffer, error) {
	updates := map[string]any{}
	if patch.CostCredits != nil {
		updates["cost_credits"] = *patch.CostCredits
	}
	if patch.CostActivityPoints != nil {
		updates["cost_activity_points"] = *patch.CostActivityPoints
	}
	if patch.ActivityPointType != nil {
		updates["activity_point_type"] = *patch.ActivityPointType
	}
	if patch.OfferActive != nil {
		updates["offer_active"] = *patch.OfferActive
	}
	if patch.ClubOnly != nil {
		updates["club_only"] = *patch.ClubOnly
	}
	if patch.OrderNum != nil {
		updates["order_num"] = *patch.OrderNum
	}
	if len(updates) > 0 {
		result := store.database.WithContext(ctx).Model(&catalogmodel.Offer{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return domain.CatalogOffer{}, result.Error
		}
		if result.RowsAffected == 0 {
			return domain.CatalogOffer{}, domain.ErrOfferNotFound
		}
	}
	return store.FindOfferByID(ctx, id)
}

// DeleteOffer removes one catalog offer by identifier.
func (store *Store) DeleteOffer(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&catalogmodel.Offer{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOfferNotFound
	}
	return nil
}

// IncrementLimitedSells atomically increments sold count and returns success.
func (store *Store) IncrementLimitedSells(ctx context.Context, offerID int) (bool, error) {
	result := store.database.WithContext(ctx).Model(&catalogmodel.Offer{}).
		Where("id = ? AND limited_total > 0 AND limited_sells < limited_total", offerID).
		Update("limited_sells", gorm.Expr("limited_sells + 1"))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
