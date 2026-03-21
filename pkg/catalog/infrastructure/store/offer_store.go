package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// ListOffersByPageID resolves all offers for one catalog page.
func (store *Store) ListOffersByPageID(ctx context.Context, pageID int) ([]domain.CatalogOffer, error) {
	var rows []catalogmodel.Offer
	if err := store.database.WithContext(ctx).Where("page_id = ?", pageID).Order("order_num ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CatalogOffer, len(rows))
	for i, row := range rows {
		result[i] = mapOffer(row)
	}
	return result, nil
}

// FindOfferByID resolves one catalog offer by identifier.
func (store *Store) FindOfferByID(ctx context.Context, id int) (domain.CatalogOffer, error) {
	var row catalogmodel.Offer
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CatalogOffer{}, domain.ErrOfferNotFound
		}
		return domain.CatalogOffer{}, err
	}
	return mapOffer(row), nil
}

// CreateOffer persists one catalog offer row.
func (store *Store) CreateOffer(ctx context.Context, offer domain.CatalogOffer) (domain.CatalogOffer, error) {
	row := catalogmodel.Offer{
		PageID: uint(offer.PageID), ItemDefinitionID: uint(offer.ItemDefinitionID),
		CatalogName: offer.CatalogName,
		CostPrimary: offer.CostPrimary, CostPrimaryType: offer.CostPrimaryType,
		CostSecondary: offer.CostSecondary, CostSecondaryType: offer.CostSecondaryType,
		Amount: offer.Amount, LimitedTotal: offer.LimitedTotal,
		OfferActive: offer.OfferActive, ExtraData: offer.ExtraData,
		BadgeID: offer.BadgeID, ClubOnly: offer.ClubOnly, OrderNum: offer.OrderNum,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.CatalogOffer{}, err
	}
	return mapOffer(row), nil
}

// UpdateOffer applies partial offer update.
func (store *Store) UpdateOffer(ctx context.Context, id int, patch domain.OfferPatch) (domain.CatalogOffer, error) {
	updates := map[string]any{}
	if patch.CostPrimary != nil {
		updates["cost_primary"] = *patch.CostPrimary
	}
	if patch.CostPrimaryType != nil {
		updates["cost_primary_type"] = *patch.CostPrimaryType
	}
	if patch.CostSecondary != nil {
		updates["cost_secondary"] = *patch.CostSecondary
	}
	if patch.CostSecondaryType != nil {
		updates["cost_secondary_type"] = *patch.CostSecondaryType
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
