package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// Service defines catalog application use-cases.
type Service struct {
	// repository stores catalog persistence contract implementation.
	repository domain.Repository
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one catalog service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("catalog repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// ListPages resolves all catalog page rows.
func (service *Service) ListPages(ctx context.Context) ([]domain.CatalogPage, error) {
	return service.repository.ListPages(ctx)
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
	return service.repository.CreatePage(ctx, page)
}

// UpdatePage applies partial page update.
func (service *Service) UpdatePage(ctx context.Context, id int, patch domain.PagePatch) (domain.CatalogPage, error) {
	if id <= 0 {
		return domain.CatalogPage{}, fmt.Errorf("page id must be positive")
	}
	return service.repository.UpdatePage(ctx, id, patch)
}

// DeletePage removes one catalog page by identifier.
func (service *Service) DeletePage(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("page id must be positive")
	}
	return service.repository.DeletePage(ctx, id)
}

// FindOfferByID resolves one catalog offer by identifier.
func (service *Service) FindOfferByID(ctx context.Context, id int) (domain.CatalogOffer, error) {
	if id <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("offer id must be positive")
	}
	return service.repository.FindOfferByID(ctx, id)
}

// ListOffersByPageID resolves all offers for one catalog page.
func (service *Service) ListOffersByPageID(ctx context.Context, pageID int) ([]domain.CatalogOffer, error) {
	if pageID <= 0 {
		return nil, fmt.Errorf("page id must be positive")
	}
	return service.repository.ListOffersByPageID(ctx, pageID)
}

// CreateOffer persists one validated catalog offer.
func (service *Service) CreateOffer(ctx context.Context, offer domain.CatalogOffer) (domain.CatalogOffer, error) {
	if offer.PageID <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("page id must be positive")
	}
	return service.repository.CreateOffer(ctx, offer)
}

// UpdateOffer applies partial offer update.
func (service *Service) UpdateOffer(ctx context.Context, id int, patch domain.OfferPatch) (domain.CatalogOffer, error) {
	if id <= 0 {
		return domain.CatalogOffer{}, fmt.Errorf("offer id must be positive")
	}
	return service.repository.UpdateOffer(ctx, id, patch)
}

// DeleteOffer removes one catalog offer by identifier.
func (service *Service) DeleteOffer(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("offer id must be positive")
	}
	return service.repository.DeleteOffer(ctx, id)
}
