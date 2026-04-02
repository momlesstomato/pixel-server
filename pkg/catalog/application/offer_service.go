package application

import (
	"context"
	"fmt"

	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// FindOfferByID resolves one catalog offer by identifier.
func (service *Service) FindOfferByID(ctx context.Context, id int) (domain.CatalogOffer, error) {
	if id <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("offer id must be positive")
	}
	return service.repository.FindOfferByID(ctx, id)
}

// ListOffersByPageID resolves all offers for one catalog page, returning from cache when available.
func (service *Service) ListOffersByPageID(ctx context.Context, pageID int) ([]domain.CatalogOffer, error) {
	if pageID <= 0 {
		return nil, fmt.Errorf("page id must be positive")
	}
	if offers, ok := service.loadCachedOffers(ctx, pageID); ok {
		return offers, nil
	}
	offers, err := service.repository.ListOffersByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	service.storeCachedOffers(ctx, pageID, offers)
	return offers, nil
}

// CreateOffer persists one validated catalog offer.
func (service *Service) CreateOffer(ctx context.Context, offer domain.CatalogOffer) (domain.CatalogOffer, error) {
	if offer.PageID <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("page id must be positive")
	}
	if err := service.validateActivityPointType(ctx, offer.CostActivityPoints, offer.ActivityPointType); err != nil {
		return domain.CatalogOffer{}, err
	}
	if service.fire != nil {
		event := &sdkcatalog.OfferCreating{PageID: offer.PageID}
		service.fire(event)
		if event.Cancelled() {
			return domain.CatalogOffer{}, fmt.Errorf("offer creation cancelled by plugin")
		}
	}
	result, err := service.repository.CreateOffer(ctx, offer)
	if err == nil {
		service.invalidateOffers(ctx, offer.PageID)
		if service.fire != nil {
			service.fire(&sdkcatalog.OfferCreated{PageID: offer.PageID, OfferID: result.ID})
		}
	}
	return result, err
}

// UpdateOffer applies partial offer update.
func (service *Service) UpdateOffer(ctx context.Context, id int, patch domain.OfferPatch) (domain.CatalogOffer, error) {
	if id <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("offer id must be positive")
	}
	if patch.ActivityPointType != nil {
		if err := service.validateActivityPointType(ctx, 1, *patch.ActivityPointType); err != nil {
			return domain.CatalogOffer{}, err
		}
	}
	result, err := service.repository.UpdateOffer(ctx, id, patch)
	if err == nil {
		service.invalidateOffers(ctx, result.PageID)
	}
	return result, err
}

// validateActivityPointType returns an error when activityPoints > 0 and the
// typeID is not registered as an enabled currency type.
func (service *Service) validateActivityPointType(ctx context.Context, activityPoints int, typeID int) error {
	if activityPoints <= 0 || service.currencyValidator == nil {
		return nil
	}
	valid, err := service.currencyValidator.IsValidActivityPointType(ctx, typeID)
	if err != nil {
		return fmt.Errorf("currency type lookup failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("activity point type %d is not registered or disabled", typeID)
	}
	return nil
}

// DeleteOffer removes one catalog offer by identifier.
func (service *Service) DeleteOffer(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("offer id must be positive")
	}
	offer, _ := service.repository.FindOfferByID(ctx, id)
	err := service.repository.DeleteOffer(ctx, id)
	if err == nil {
		service.invalidateOffers(ctx, offer.PageID)
	}
	return err
}
