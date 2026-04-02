package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// repositoryStub defines deterministic catalog repository behavior.
type repositoryStub struct {
	// page stores deterministic page return.
	page domain.CatalogPage
	// offer stores deterministic offer return.
	offer domain.CatalogOffer
	// voucher stores deterministic voucher return.
	voucher domain.Voucher
	// findErr stores deterministic find error.
	findErr error
	// deleteErr stores deterministic delete error.
	deleteErr error
	// redeemed stores whether voucher is marked as redeemed.
	redeemed bool
}

// ListPages returns deterministic page list.
func (s repositoryStub) ListPages(_ context.Context) ([]domain.CatalogPage, error) {
	return []domain.CatalogPage{s.page}, nil
}

// FindPageByID returns deterministic page.
func (s repositoryStub) FindPageByID(_ context.Context, _ int) (domain.CatalogPage, error) {
	return s.page, s.findErr
}

// CreatePage returns deterministic page.
func (s repositoryStub) CreatePage(_ context.Context, p domain.CatalogPage) (domain.CatalogPage, error) {
	p.ID = 1
	return p, nil
}

// UpdatePage returns deterministic page.
func (s repositoryStub) UpdatePage(_ context.Context, _ int, _ domain.PagePatch) (domain.CatalogPage, error) {
	return s.page, s.findErr
}

// DeletePage returns deterministic error.
func (s repositoryStub) DeletePage(_ context.Context, _ int) error {
	return s.deleteErr
}

// ListOffersByPageID returns deterministic offer list.
func (s repositoryStub) ListOffersByPageID(_ context.Context, _ int) ([]domain.CatalogOffer, error) {
	return []domain.CatalogOffer{s.offer}, nil
}

// FindOfferByID returns deterministic offer.
func (s repositoryStub) FindOfferByID(_ context.Context, _ int) (domain.CatalogOffer, error) {
	return s.offer, s.findErr
}

// CreateOffer returns deterministic offer.
func (s repositoryStub) CreateOffer(_ context.Context, o domain.CatalogOffer) (domain.CatalogOffer, error) {
	o.ID = 1
	return o, nil
}

// UpdateOffer returns deterministic offer.
func (s repositoryStub) UpdateOffer(_ context.Context, _ int, _ domain.OfferPatch) (domain.CatalogOffer, error) {
	return s.offer, nil
}

// DeleteOffer returns deterministic error.
func (s repositoryStub) DeleteOffer(_ context.Context, _ int) error {
	return s.deleteErr
}

// IncrementLimitedSells returns deterministic success.
func (s repositoryStub) IncrementLimitedSells(_ context.Context, _ int) (bool, error) {
	return true, nil
}

// FindVoucherByCode returns deterministic voucher.
func (s repositoryStub) FindVoucherByCode(_ context.Context, _ string) (domain.Voucher, error) {
	return s.voucher, s.findErr
}

// CreateVoucher returns deterministic voucher.
func (s repositoryStub) CreateVoucher(_ context.Context, v domain.Voucher) (domain.Voucher, error) {
	v.ID = 1
	return v, nil
}

// DeleteVoucher returns deterministic error.
func (s repositoryStub) DeleteVoucher(_ context.Context, _ int) error {
	return s.deleteErr
}

// ListVouchers returns deterministic voucher list.
func (s repositoryStub) ListVouchers(_ context.Context) ([]domain.Voucher, error) {
	return []domain.Voucher{s.voucher}, nil
}

// RedeemVoucher returns deterministic error.
func (s repositoryStub) RedeemVoucher(_ context.Context, _ int, _ int) error {
	return nil
}

// HasUserRedeemedVoucher returns deterministic result.
func (s repositoryStub) HasUserRedeemedVoucher(_ context.Context, _ int, _ int) (bool, error) {
	return s.redeemed, nil
}

// spenderStub implements domain.Spender with configurable balances and errors.
type spenderStub struct {
	// credits stores stubbed credit balance.
	credits int
	// currency stores stubbed activity-point balance.
	currency int
	// err stores optional error to return.
	err error
}

// GetCredits returns stubbed credit balance.
func (s spenderStub) GetCredits(_ context.Context, _ int) (int, error) { return s.credits, s.err }

// AddCredits adjusts stubbed credit balance.
func (s spenderStub) AddCredits(_ context.Context, _ int, delta int) (int, error) {
	return s.credits + delta, s.err
}

// GetCurrencyBalance returns stubbed activity-point balance.
func (s spenderStub) GetCurrencyBalance(_ context.Context, _ int, _ int) (int, error) {
	return s.currency, s.err
}

// AddCurrencyBalance adjusts stubbed activity-point balance.
func (s spenderStub) AddCurrencyBalance(_ context.Context, _ int, _ int, delta int) (int, error) {
	return s.currency + delta, s.err
}

// recipientFinderStub implements domain.RecipientFinder with configurable results.
type recipientFinderStub struct {
	// info stores the recipient info to return.
	info domain.RecipientInfo
	// err stores optional error to return.
	err error
}

// FindRecipientByUsername returns stubbed recipient info.
func (s recipientFinderStub) FindRecipientByUsername(_ context.Context, _ string) (domain.RecipientInfo, error) {
	return s.info, s.err
}
