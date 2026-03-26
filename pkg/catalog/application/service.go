package application

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	redislib "github.com/redis/go-redis/v9"
)

// Service defines catalog application use-cases.
type Service struct {
	// repository stores catalog persistence contract implementation.
	repository domain.Repository
	// currencyValidator stores optional activity-currency type validation port.
	currencyValidator domain.ActivityCurrencyValidator
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
	// redis stores optional Redis client for cache operations.
	redis *redislib.Client
	// cachePrefix stores Redis key namespace prefix.
	cachePrefix string
	// cacheTTL stores cache entry time-to-live duration.
	cacheTTL time.Duration
}

// NewService creates one catalog service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("catalog repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetCurrencyValidator configures the activity-currency type validator.
// When set, CreateOffer and UpdateOffer will reject unknown or disabled type IDs.
func (service *Service) SetCurrencyValidator(v domain.ActivityCurrencyValidator) {
	service.currencyValidator = v
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// ListPages resolves all catalog page rows, returning from cache when available.
func (service *Service) ListPages(ctx context.Context) ([]domain.CatalogPage, error) {
	if pages, ok := service.loadCachedPages(ctx); ok {
		return pages, nil
	}
	pages, err := service.repository.ListPages(ctx)
	if err != nil {
		return nil, err
	}
	service.storeCachedPages(ctx, pages)
	return pages, nil
}

// FindPageByID resolves one catalog page by identifier.
func (service *Service) FindPageByID(ctx context.Context, id int) (domain.CatalogPage, error) {
	if id <= 0 {
		return domain.CatalogPage{}, fmt.Errorf("page id must be positive")
	}
	return service.repository.FindPageByID(ctx, id)
}

// CreatePage persists one validated catalog page.
func (service *Service) CreatePage(ctx context.Context, page domain.CatalogPage) (domain.CatalogPage, error) {
	if page.Caption == "" {
		return domain.CatalogPage{}, fmt.Errorf("page caption is required")
	}
	result, err := service.repository.CreatePage(ctx, page)
	if err == nil {
		service.invalidatePages(ctx)
	}
	return result, err
}

// UpdatePage applies partial page update.
func (service *Service) UpdatePage(ctx context.Context, id int, patch domain.PagePatch) (domain.CatalogPage, error) {
	if id <= 0 {
		return domain.CatalogPage{}, fmt.Errorf("page id must be positive")
	}
	result, err := service.repository.UpdatePage(ctx, id, patch)
	if err == nil {
		service.invalidatePages(ctx)
	}
	return result, err
}

// DeletePage removes one catalog page by identifier.
func (service *Service) DeletePage(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("page id must be positive")
	}
	err := service.repository.DeletePage(ctx, id)
	if err == nil {
		service.invalidatePages(ctx)
		service.invalidateOffers(ctx, id)
	}
	return err
}

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
	result, err := service.repository.CreateOffer(ctx, offer)
	if err == nil {
		service.invalidateOffers(ctx, offer.PageID)
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
