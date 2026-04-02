package application

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	redislib "github.com/redis/go-redis/v9"
)

// Service defines catalog application use-cases.
type Service struct {
	// repository stores catalog persistence contract implementation.
	repository domain.Repository
	// currencyValidator stores optional activity-currency type validation port.
	currencyValidator domain.ActivityCurrencyValidator
	// spender stores optional credit and activity-point deduction port.
	spender domain.Spender
	// recipientFinder stores optional gift recipient lookup port.
	recipientFinder domain.RecipientFinder
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

// SetSpender configures the credit and activity-point deduction port.
// When not set, purchases requiring payment will return ErrNoSpender.
func (service *Service) SetSpender(s domain.Spender) {
	service.spender = s
}

// SetRecipientFinder configures the gift recipient lookup port.
// When not set, gift purchases will return ErrRecipientNotFound.
func (service *Service) SetRecipientFinder(rf domain.RecipientFinder) {
	service.recipientFinder = rf
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
	if service.fire != nil {
		event := &sdkcatalog.PageCreating{Caption: page.Caption}
		service.fire(event)
		if event.Cancelled() {
			return domain.CatalogPage{}, fmt.Errorf("page creation cancelled by plugin")
		}
	}
	result, err := service.repository.CreatePage(ctx, page)
	if err == nil {
		service.invalidatePages(ctx)
		if service.fire != nil {
			service.fire(&sdkcatalog.PageCreated{PageID: result.ID})
		}
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
