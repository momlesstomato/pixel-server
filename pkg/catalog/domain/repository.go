package domain

import "context"

// Repository defines catalog persistence behavior.
type Repository interface {
	// ListPages resolves all catalog page rows.
	ListPages(context.Context) ([]CatalogPage, error)
	// FindPageByID resolves one catalog page by identifier.
	FindPageByID(context.Context, int) (CatalogPage, error)
	// CreatePage persists one catalog page row.
	CreatePage(context.Context, CatalogPage) (CatalogPage, error)
	// UpdatePage applies partial page update.
	UpdatePage(context.Context, int, PagePatch) (CatalogPage, error)
	// DeletePage removes one catalog page by identifier.
	DeletePage(context.Context, int) error
	// ListOffersByPageID resolves all offers for one catalog page.
	ListOffersByPageID(context.Context, int) ([]CatalogOffer, error)
	// FindOfferByID resolves one catalog offer by identifier.
	FindOfferByID(context.Context, int) (CatalogOffer, error)
	// CreateOffer persists one catalog offer row.
	CreateOffer(context.Context, CatalogOffer) (CatalogOffer, error)
	// UpdateOffer applies partial offer update.
	UpdateOffer(context.Context, int, OfferPatch) (CatalogOffer, error)
	// DeleteOffer removes one catalog offer by identifier.
	DeleteOffer(context.Context, int) error
	// IncrementLimitedSells atomically increments sold count and returns success.
	IncrementLimitedSells(ctx context.Context, offerID int) (bool, error)
	// FindVoucherByCode resolves one voucher by unique code.
	FindVoucherByCode(context.Context, string) (Voucher, error)
	// CreateVoucher persists one voucher row.
	CreateVoucher(context.Context, Voucher) (Voucher, error)
	// DeleteVoucher removes one voucher by identifier.
	DeleteVoucher(context.Context, int) error
	// ListVouchers resolves all voucher rows.
	ListVouchers(context.Context) ([]Voucher, error)
	// RedeemVoucher atomically increments use count and records redemption.
	RedeemVoucher(ctx context.Context, voucherID int, userID int) error
	// HasUserRedeemedVoucher checks per-user voucher redemption.
	HasUserRedeemedVoucher(ctx context.Context, voucherID int, userID int) (bool, error)
}

// PagePatch defines partial catalog page update payload.
type PagePatch struct {
	// Caption stores optional page title update.
	Caption *string
	// Visible stores optional visibility update.
	Visible *bool
	// Enabled stores optional availability update.
	Enabled *bool
	// MinPermission stores optional dotted permission restriction update.
	MinPermission *string
	// OrderNum stores optional sort position update.
	OrderNum *int
	// PageLayout stores optional layout key update.
	PageLayout *string
}

// OfferPatch defines partial catalog offer update payload.
type OfferPatch struct {
	// CostCredits stores optional credits price update.
	CostCredits *int
	// CostActivityPoints stores optional activity-point price update.
	CostActivityPoints *int
	// ActivityPointType stores optional activity-point currency type update.
	ActivityPointType *int
	// OfferActive stores optional active flag update.
	OfferActive *bool
	// ClubOnly stores optional club restriction update.
	ClubOnly *bool
	// OrderNum stores optional sort position update.
	OrderNum *int
}
