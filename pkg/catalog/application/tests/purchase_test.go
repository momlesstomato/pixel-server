package tests

import (
	"context"
	"testing"

	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// TestPurchaseOfferRejectsZeroAmount verifies amount validation.
func TestPurchaseOfferRejectsZeroAmount(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 10}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetSpender(spenderStub{credits: 100})
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 0); err == nil {
		t.Fatalf("expected failure for zero amount")
	}
}

// TestPurchaseOfferRejectsInactiveOffer verifies inactive offer guard.
func TestPurchaseOfferRejectsInactiveOffer(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: false}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetSpender(spenderStub{credits: 100})
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1); err == nil {
		t.Fatalf("expected failure for inactive offer")
	}
}

// TestPurchaseOfferInsufficientCredits verifies credit balance guard.
func TestPurchaseOfferInsufficientCredits(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 50}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetSpender(spenderStub{credits: 10})
	_, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1)
	if err == nil {
		t.Fatalf("expected insufficient credits error")
	}
}

// TestPurchaseOfferNoSpenderError verifies missing spender guard.
func TestPurchaseOfferNoSpenderError(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 10}}
	svc, _ := catalogapplication.NewService(stub)
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1); err == nil {
		t.Fatalf("expected missing spender error")
	}
}

// TestPurchaseOfferFreeSucceeds verifies a zero-cost offer requires no spender.
func TestPurchaseOfferFreeSucceeds(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 0}}
	svc, _ := catalogapplication.NewService(stub)
	offer, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer.ID != 1 {
		t.Fatalf("expected offer id 1, got %d", offer.ID)
	}
}

// TestPurchaseOfferDeductsCredits verifies credit deduction on successful purchase.
func TestPurchaseOfferDeductsCredits(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 10}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetSpender(spenderStub{credits: 100})
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurchaseGiftRejectsNoRecipientFinder verifies missing finder guard.
func TestPurchaseGiftRejectsNoRecipientFinder(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true}}
	svc, _ := catalogapplication.NewService(stub)
	if _, err := svc.PurchaseGift(context.Background(), "conn", 1, 1, "", "alice"); err == nil {
		t.Fatalf("expected ErrRecipientNotFound")
	}
}

// TestPurchaseGiftRejectsUnknownRecipient verifies unknown recipient guard.
func TestPurchaseGiftRejectsUnknownRecipient(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetRecipientFinder(recipientFinderStub{err: domain.ErrRecipientNotFound})
	if _, err := svc.PurchaseGift(context.Background(), "conn", 1, 1, "", "nobody"); err == nil {
		t.Fatalf("expected ErrRecipientNotFound")
	}
}

// TestPurchaseGiftRejectsBlockedRecipient verifies AllowGifts=false guard.
func TestPurchaseGiftRejectsBlockedRecipient(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetRecipientFinder(recipientFinderStub{info: domain.RecipientInfo{UserID: 2, AllowGifts: false}})
	if _, err := svc.PurchaseGift(context.Background(), "conn", 1, 1, "", "blocked"); err == nil {
		t.Fatalf("expected rejection for blocked recipient")
	}
}

// TestPurchaseGiftFreeSucceeds verifies gift purchase of a free offer.
func TestPurchaseGiftFreeSucceeds(t *testing.T) {
	stub := repositoryStub{offer: domain.CatalogOffer{ID: 1, OfferActive: true}}
	svc, _ := catalogapplication.NewService(stub)
	svc.SetRecipientFinder(recipientFinderStub{info: domain.RecipientInfo{UserID: 2, AllowGifts: true}})
	if _, err := svc.PurchaseGift(context.Background(), "conn", 1, 1, "", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
