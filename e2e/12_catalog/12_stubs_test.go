package catalog

import (
	"context"

	catalogdomain "github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// activeOfferRepo returns a single active free offer with ID=1.
type activeOfferRepo struct{}

func (activeOfferRepo) ListPages(context.Context) ([]catalogdomain.CatalogPage, error) {
	return nil, nil
}
func (activeOfferRepo) FindPageByID(context.Context, int) (catalogdomain.CatalogPage, error) {
	return catalogdomain.CatalogPage{}, nil
}
func (activeOfferRepo) CreatePage(_ context.Context, p catalogdomain.CatalogPage) (catalogdomain.CatalogPage, error) {
	p.ID = 1
	return p, nil
}
func (activeOfferRepo) UpdatePage(context.Context, int, catalogdomain.PagePatch) (catalogdomain.CatalogPage, error) {
	return catalogdomain.CatalogPage{}, nil
}
func (activeOfferRepo) DeletePage(context.Context, int) error { return nil }
func (activeOfferRepo) ListOffersByPageID(context.Context, int) ([]catalogdomain.CatalogOffer, error) {
	return nil, nil
}
func (activeOfferRepo) FindOfferByID(_ context.Context, id int) (catalogdomain.CatalogOffer, error) {
	return catalogdomain.CatalogOffer{ID: id, OfferActive: true}, nil
}
func (activeOfferRepo) CreateOffer(_ context.Context, o catalogdomain.CatalogOffer) (catalogdomain.CatalogOffer, error) {
	o.ID = 1
	return o, nil
}
func (activeOfferRepo) UpdateOffer(context.Context, int, catalogdomain.OfferPatch) (catalogdomain.CatalogOffer, error) {
	return catalogdomain.CatalogOffer{}, nil
}
func (activeOfferRepo) DeleteOffer(context.Context, int) error                   { return nil }
func (activeOfferRepo) IncrementLimitedSells(context.Context, int) (bool, error) { return true, nil }
func (activeOfferRepo) FindVoucherByCode(context.Context, string) (catalogdomain.Voucher, error) {
	return catalogdomain.Voucher{}, nil
}
func (activeOfferRepo) CreateVoucher(_ context.Context, v catalogdomain.Voucher) (catalogdomain.Voucher, error) {
	v.ID = 1
	return v, nil
}
func (activeOfferRepo) DeleteVoucher(context.Context, int) error { return nil }
func (activeOfferRepo) ListVouchers(context.Context) ([]catalogdomain.Voucher, error) {
	return nil, nil
}
func (activeOfferRepo) RedeemVoucher(context.Context, int, int) error { return nil }
func (activeOfferRepo) HasUserRedeemedVoucher(context.Context, int, int) (bool, error) {
	return false, nil
}

// blockedRecipientFinder always reports AllowGifts=false.
type blockedRecipientFinder struct{}

func (blockedRecipientFinder) FindRecipientByUsername(_ context.Context, _ string) (catalogdomain.RecipientInfo, error) {
	return catalogdomain.RecipientInfo{UserID: 99, AllowGifts: false}, nil
}
