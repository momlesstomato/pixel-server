package tests

import (
	"context"
	"testing"

	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := catalogapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServicePageCRUD verifies page create, find, update, and delete behavior.
func TestServicePageCRUD(t *testing.T) {
	stub := repositoryStub{page: domain.CatalogPage{ID: 1, Caption: "Frontpage"}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.FindPageByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	page, err := service.FindPageByID(context.Background(), 1)
	if err != nil || page.Caption != "Frontpage" {
		t.Fatalf("unexpected find result %+v err=%v", page, err)
	}
	if _, err := service.CreatePage(context.Background(), domain.CatalogPage{}); err == nil {
		t.Fatalf("expected create failure for empty caption")
	}
	created, err := service.CreatePage(context.Background(), domain.CatalogPage{Caption: "Test"})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if err := service.DeletePage(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
	if _, err := service.UpdatePage(context.Background(), 0, domain.PagePatch{}); err == nil {
		t.Fatalf("expected update failure for invalid id")
	}
}

// TestServiceOfferCRUD verifies offer operations behavior.
func TestServiceOfferCRUD(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, PageID: 1}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.FindOfferByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	if _, err := service.ListOffersByPageID(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid page id")
	}
	if _, err := service.CreateOffer(context.Background(), domain.CatalogOffer{}); err == nil {
		t.Fatalf("expected create failure for missing page id")
	}
	if _, err := service.UpdateOffer(context.Background(), 0, domain.OfferPatch{}); err == nil {
		t.Fatalf("expected update failure for invalid id")
	}
	if err := service.DeleteOffer(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
}

// TestServiceVoucherFlow verifies voucher operations behavior.
func TestServiceVoucherFlow(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "TEST", Enabled: true, MaxUses: 10}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.FindVoucherByCode(context.Background(), ""); err == nil {
		t.Fatalf("expected find failure for empty code")
	}
	if _, err := service.CreateVoucher(context.Background(), domain.Voucher{}); err == nil {
		t.Fatalf("expected create failure for empty code")
	}
	if err := service.DeleteVoucher(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
}

// TestServiceRedeemVoucherValidation verifies voucher redemption validation.
func TestServiceRedeemVoucherValidation(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "TEST", Enabled: true, MaxUses: 10}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.RedeemVoucher(context.Background(), "", 1); err == nil {
		t.Fatalf("expected redeem failure for empty code")
	}
	if _, err := service.RedeemVoucher(context.Background(), "TEST", 0); err == nil {
		t.Fatalf("expected redeem failure for invalid user id")
	}
	v, err := service.RedeemVoucher(context.Background(), "TEST", 1)
	if err != nil || v.CurrentUses != 1 {
		t.Fatalf("unexpected redeem result %+v err=%v", v, err)
	}
}

// TestServiceRedeemDisabledVoucher verifies disabled voucher rejection.
func TestServiceRedeemDisabledVoucher(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "OFF", Enabled: false, MaxUses: 10}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.RedeemVoucher(context.Background(), "OFF", 1); err != domain.ErrVoucherDisabled {
		t.Fatalf("expected disabled voucher error, got %v", err)
	}
}

// TestServiceRedeemExhaustedVoucher verifies exhausted voucher rejection.
func TestServiceRedeemExhaustedVoucher(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "FULL", Enabled: true, MaxUses: 5, CurrentUses: 5}}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.RedeemVoucher(context.Background(), "FULL", 1); err != domain.ErrVoucherExhausted {
		t.Fatalf("expected exhausted voucher error, got %v", err)
	}
}

// TestServiceRedeemAlreadyRedeemedVoucher verifies duplicate redemption rejection.
func TestServiceRedeemAlreadyRedeemedVoucher(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "DUP", Enabled: true, MaxUses: 10}, redeemed: true}
	service, _ := catalogapplication.NewService(stub)
	if _, err := service.RedeemVoucher(context.Background(), "DUP", 1); err != domain.ErrVoucherAlreadyRedeemed {
		t.Fatalf("expected already redeemed error, got %v", err)
	}
}
