package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// Service defines catalog API behavior required by HTTP routes.
type Service interface {
	// ListPages resolves all catalog page rows.
	ListPages(context.Context) ([]domain.CatalogPage, error)
	// FindPageByID resolves one catalog page by identifier.
	FindPageByID(context.Context, int) (domain.CatalogPage, error)
	// CreatePage persists one validated catalog page.
	CreatePage(context.Context, domain.CatalogPage) (domain.CatalogPage, error)
	// UpdatePage applies partial page update.
	UpdatePage(context.Context, int, domain.PagePatch) (domain.CatalogPage, error)
	// DeletePage removes one catalog page by identifier.
	DeletePage(context.Context, int) error
	// ListOffersByPageID resolves all offers for one catalog page.
	ListOffersByPageID(context.Context, int) ([]domain.CatalogOffer, error)
	// FindOfferByID resolves one catalog offer by identifier.
	FindOfferByID(context.Context, int) (domain.CatalogOffer, error)
	// CreateOffer persists one validated catalog offer.
	CreateOffer(context.Context, domain.CatalogOffer) (domain.CatalogOffer, error)
	// RedeemVoucher validates and redeems one voucher.
	RedeemVoucher(context.Context, string, int) (domain.Voucher, error)
	// ListVouchers resolves all voucher rows.
	ListVouchers(context.Context) ([]domain.Voucher, error)
}
