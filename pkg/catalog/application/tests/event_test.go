package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// TestPageCreatingEventCancelsCreation verifies PageCreating cancellation aborts page creation.
func TestPageCreatingEventCancelsCreation(t *testing.T) {
	service, _ := catalogapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkcatalog.PageCreating); ok {
			value.Cancel()
		}
	})
	if _, err := service.CreatePage(context.Background(), domain.CatalogPage{Caption: "Test"}); err == nil {
		t.Fatalf("expected page creation to be cancelled")
	}
}

// TestPageCreatingEventAllowsCreation verifies PageCreating passes through without cancellation.
func TestPageCreatingEventAllowsCreation(t *testing.T) {
	var fired bool
	service, _ := catalogapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkcatalog.PageCreating); ok {
			fired = true
		}
	})
	page, err := service.CreatePage(context.Background(), domain.CatalogPage{Caption: "Test"})
	if err != nil || page.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", page, err)
	}
	if !fired {
		t.Fatalf("expected PageCreating event to fire")
	}
}

// TestOfferCreatingEventCancelsCreation verifies OfferCreating cancellation aborts offer creation.
func TestOfferCreatingEventCancelsCreation(t *testing.T) {
	service, _ := catalogapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkcatalog.OfferCreating); ok {
			value.Cancel()
		}
	})
	if _, err := service.CreateOffer(context.Background(), domain.CatalogOffer{PageID: 1}); err == nil {
		t.Fatalf("expected offer creation to be cancelled")
	}
}

// TestOfferCreatingEventAllowsCreation verifies OfferCreating passes through without cancellation.
func TestOfferCreatingEventAllowsCreation(t *testing.T) {
	var afterFired bool
	service, _ := catalogapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkcatalog.OfferCreated); ok {
			afterFired = true
		}
	})
	offer, err := service.CreateOffer(context.Background(), domain.CatalogOffer{PageID: 1})
	if err != nil || offer.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", offer, err)
	}
	if !afterFired {
		t.Fatalf("expected OfferCreated event to fire")
	}
}

// TestVoucherRedeemingEventCancelsRedemption verifies VoucherRedeeming cancellation aborts redemption.
func TestVoucherRedeemingEventCancelsRedemption(t *testing.T) {
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "TEST", Enabled: true, MaxUses: 10}}
	service, _ := catalogapplication.NewService(stub)
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkcatalog.VoucherRedeeming); ok {
			value.Cancel()
		}
	})
	if _, err := service.RedeemVoucher(context.Background(), "TEST", 1); err == nil {
		t.Fatalf("expected voucher redemption to be cancelled")
	}
}

// TestVoucherRedeemingEventAllowsRedemption verifies VoucherRedeeming passes through without cancellation.
func TestVoucherRedeemingEventAllowsRedemption(t *testing.T) {
	var afterFired bool
	stub := repositoryStub{voucher: domain.Voucher{ID: 1, Code: "PASS", Enabled: true, MaxUses: 10}}
	service, _ := catalogapplication.NewService(stub)
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkcatalog.VoucherRedeemed); ok {
			afterFired = true
		}
	})
	v, err := service.RedeemVoucher(context.Background(), "PASS", 1)
	if err != nil {
		t.Fatalf("expected redemption success, got %v", err)
	}
	if v.CurrentUses != 1 {
		t.Fatalf("unexpected current uses %d", v.CurrentUses)
	}
	if !afterFired {
		t.Fatalf("expected VoucherRedeemed event to fire")
	}
}
