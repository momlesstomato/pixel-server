package catalog

import (
	"context"
	"testing"

	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"go.uber.org/zap"
)

// Test12PluginCancelsPurchase verifies a plugin can cancel a purchase via OfferPurchased event.
func Test12PluginCancelsPurchase(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	dispatcher.Subscribe("test", func(ev *sdkcatalog.OfferPurchased) {
		ev.Cancel()
	})
	svc, _ := catalogapplication.NewService(activeOfferRepo{})
	svc.SetEventFirer(dispatcher.Fire)
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1); err == nil {
		t.Fatalf("expected purchase cancelled by plugin")
	}
}

// Test12PluginReceivesConfirmedEvent verifies OfferPurchaseConfirmed fires after commit.
func Test12PluginReceivesConfirmedEvent(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	var confirmedOfferID int
	dispatcher.Subscribe("test", func(ev *sdkcatalog.OfferPurchaseConfirmed) {
		confirmedOfferID = ev.OfferID
	})
	svc, _ := catalogapplication.NewService(activeOfferRepo{})
	svc.SetEventFirer(dispatcher.Fire)
	if _, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1); err != nil {
		t.Fatalf("unexpected purchase error: %v", err)
	}
	if confirmedOfferID != 1 {
		t.Fatalf("expected OfferPurchaseConfirmed with OfferID=1, got %d", confirmedOfferID)
	}
}

// Test12PurchaseFreeOfferNoSpenderNeeded verifies zero-cost purchases need no spender.
func Test12PurchaseFreeOfferNoSpenderNeeded(t *testing.T) {
	svc, _ := catalogapplication.NewService(activeOfferRepo{})
	offer, err := svc.PurchaseOffer(context.Background(), "conn", 1, 1, "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer.ID != 1 {
		t.Fatalf("expected offer ID 1, got %d", offer.ID)
	}
}

// Test12GiftPurchaseRejectsBlockedRecipient verifies blocked recipient prevents gift purchase.
func Test12GiftPurchaseRejectsBlockedRecipient(t *testing.T) {
	svc, _ := catalogapplication.NewService(activeOfferRepo{})
	svc.SetRecipientFinder(blockedRecipientFinder{})
	if _, err := svc.PurchaseGift(context.Background(), "conn", 1, 1, "", "blocked"); err == nil {
		t.Fatalf("expected gift purchase blocked for safety-locked recipient")
	}
}
