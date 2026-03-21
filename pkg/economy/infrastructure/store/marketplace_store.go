package store

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
	economymodel "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/model"
	"gorm.io/gorm"
)

// ListOpenOffers resolves paginated open marketplace offers.
func (store *Store) ListOpenOffers(ctx context.Context, filter domain.OfferFilter) ([]domain.MarketplaceOffer, int, error) {
	query := store.database.WithContext(ctx).Model(&economymodel.MarketplaceOffer{}).Where("state = ?", string(domain.OfferStateOpen))
	if filter.MinPrice > 0 {
		query = query.Where("asking_price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("asking_price <= ?", filter.MaxPrice)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []economymodel.MarketplaceOffer
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]domain.MarketplaceOffer, len(rows))
	for i, row := range rows {
		result[i] = mapOffer(row)
	}
	return result, int(total), nil
}

// ListOffersBySellerID resolves all offers for one seller.
func (store *Store) ListOffersBySellerID(ctx context.Context, sellerID int) ([]domain.MarketplaceOffer, error) {
	var rows []economymodel.MarketplaceOffer
	if err := store.database.WithContext(ctx).Where("seller_id = ?", sellerID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.MarketplaceOffer, len(rows))
	for i, row := range rows {
		result[i] = mapOffer(row)
	}
	return result, nil
}

// FindOfferByID resolves one marketplace offer by identifier.
func (store *Store) FindOfferByID(ctx context.Context, id int) (domain.MarketplaceOffer, error) {
	var row economymodel.MarketplaceOffer
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.MarketplaceOffer{}, domain.ErrOfferNotFound
		}
		return domain.MarketplaceOffer{}, err
	}
	return mapOffer(row), nil
}

// CreateOffer persists one marketplace listing atomically.
func (store *Store) CreateOffer(ctx context.Context, offer domain.MarketplaceOffer) (domain.MarketplaceOffer, error) {
	row := economymodel.MarketplaceOffer{
		SellerID: uint(offer.SellerID), ItemID: uint(offer.ItemID),
		DefinitionID: uint(offer.DefinitionID), AskingPrice: offer.AskingPrice,
		State: string(domain.OfferStateOpen), ExpireAt: offer.ExpireAt,
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.MarketplaceOffer{}, err
	}
	return mapOffer(row), nil
}

// PurchaseOffer atomically marks offer as sold.
func (store *Store) PurchaseOffer(ctx context.Context, offerID int, buyerID int) (domain.MarketplaceOffer, error) {
	now := time.Now().UTC()
	result := store.database.WithContext(ctx).Model(&economymodel.MarketplaceOffer{}).
		Where("id = ? AND state = ?", offerID, string(domain.OfferStateOpen)).
		Updates(map[string]any{"state": string(domain.OfferStateSold), "buyer_id": buyerID, "sold_at": now})
	if result.Error != nil {
		return domain.MarketplaceOffer{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.MarketplaceOffer{}, domain.ErrOfferNotFound
	}
	return store.FindOfferByID(ctx, offerID)
}

// CancelOffer atomically marks offer as cancelled.
func (store *Store) CancelOffer(ctx context.Context, offerID int) error {
	result := store.database.WithContext(ctx).Model(&economymodel.MarketplaceOffer{}).
		Where("id = ? AND state = ?", offerID, string(domain.OfferStateOpen)).
		Update("state", string(domain.OfferStateCancelled))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOfferNotFound
	}
	return nil
}

// ExpireOffers marks all expired offers and returns affected rows.
func (store *Store) ExpireOffers(ctx context.Context, maxAgeHours int) ([]domain.MarketplaceOffer, error) {
	var rows []economymodel.MarketplaceOffer
	err := store.database.WithContext(ctx).
		Where("state = ? AND expire_at < NOW()", string(domain.OfferStateOpen)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		ids := make([]uint, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}
		store.database.WithContext(ctx).Model(&economymodel.MarketplaceOffer{}).
			Where("id IN ?", ids).Update("state", string(domain.OfferStateExpired))
	}
	result := make([]domain.MarketplaceOffer, len(rows))
	for i, row := range rows {
		result[i] = mapOffer(row)
	}
	return result, nil
}

// CountActiveOffers returns active offer count for one seller.
func (store *Store) CountActiveOffers(ctx context.Context, sellerID int) (int, error) {
	var count int64
	err := store.database.WithContext(ctx).Model(&economymodel.MarketplaceOffer{}).
		Where("seller_id = ? AND state = ?", sellerID, string(domain.OfferStateOpen)).
		Count(&count).Error
	return int(count), err
}

// mapOffer converts one GORM model into domain marketplace offer.
func mapOffer(row economymodel.MarketplaceOffer) domain.MarketplaceOffer {
	var buyerID *int
	if row.BuyerID != nil {
		v := int(*row.BuyerID)
		buyerID = &v
	}
	return domain.MarketplaceOffer{
		ID: int(row.ID), SellerID: int(row.SellerID), ItemID: int(row.ItemID),
		DefinitionID: int(row.DefinitionID), AskingPrice: row.AskingPrice,
		State: domain.OfferState(row.State), BuyerID: buyerID,
		SoldAt: row.SoldAt, ExpireAt: row.ExpireAt, CreatedAt: row.CreatedAt,
	}
}
