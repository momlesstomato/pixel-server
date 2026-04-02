package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkinventory "github.com/momlesstomato/pixel-sdk/events/inventory"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// TestBadgeAwardingEventCancelsAward verifies BadgeAwarding cancellation aborts badge award.
func TestBadgeAwardingEventCancelsAward(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkinventory.BadgeAwarding); ok {
			value.Cancel()
		}
	})
	if _, err := service.AwardBadge(context.Background(), 1, "ACH1"); err == nil {
		t.Fatalf("expected badge award to be cancelled")
	}
}

// TestBadgeAwardingEventAllowsAward verifies BadgeAwarding passes through and fires BadgeAwarded after.
func TestBadgeAwardingEventAllowsAward(t *testing.T) {
	var afterFired bool
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkinventory.BadgeAwarded); ok {
			afterFired = true
		}
	})
	badge, err := service.AwardBadge(context.Background(), 1, "ACH1")
	if err != nil || badge.BadgeCode != "ACH1" {
		t.Fatalf("unexpected badge %+v err=%v", badge, err)
	}
	if !afterFired {
		t.Fatalf("expected BadgeAwarded event to fire")
	}
}

// TestBadgeRevokingEventCancelsRevoke verifies BadgeRevoking cancellation aborts badge revoke.
func TestBadgeRevokingEventCancelsRevoke(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkinventory.BadgeRevoking); ok {
			value.Cancel()
		}
	})
	if err := service.RevokeBadge(context.Background(), 1, "ACH1"); err == nil {
		t.Fatalf("expected badge revoke to be cancelled")
	}
}

// TestBadgeRevokingEventAllowsRevoke verifies BadgeRevoking passes through and fires BadgeRevoked after.
func TestBadgeRevokingEventAllowsRevoke(t *testing.T) {
	var afterFired bool
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkinventory.BadgeRevoked); ok {
			afterFired = true
		}
	})
	if err := service.RevokeBadge(context.Background(), 1, "ACH1"); err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}
	if !afterFired {
		t.Fatalf("expected BadgeRevoked event to fire")
	}
}

// TestCreditsUpdatingEventCancelsAddCredits verifies CreditsUpdating cancellation aborts credit add.
func TestCreditsUpdatingEventCancelsAddCredits(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{credits: 500})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkinventory.CreditsUpdating); ok {
			value.Cancel()
		}
	})
	if _, err := service.AddCredits(context.Background(), 1, 100); err == nil {
		t.Fatalf("expected credits add to be cancelled")
	}
}

// TestCreditsUpdatingEventAllowsAddCredits verifies CreditsUpdating passes and fires CreditsUpdated after.
func TestCreditsUpdatingEventAllowsAddCredits(t *testing.T) {
	var afterFired bool
	service, _ := inventoryapplication.NewService(repositoryStub{credits: 500})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkinventory.CreditsUpdated); ok {
			afterFired = true
		}
	})
	balance, err := service.AddCredits(context.Background(), 1, 100)
	if err != nil || balance != 600 {
		t.Fatalf("unexpected balance %d err=%v", balance, err)
	}
	if !afterFired {
		t.Fatalf("expected CreditsUpdated event to fire")
	}
}

// TestCurrencyUpdatingEventCancelsAddCurrency verifies CurrencyUpdating cancellation aborts currency add.
func TestCurrencyUpdatingEventCancelsAddCurrency(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkinventory.CurrencyUpdating); ok {
			value.Cancel()
		}
	})
	if _, err := service.AddCurrencyTracked(context.Background(), 1, domain.CurrencyDuckets, 50, domain.SourceAdmin, "test", "1"); err == nil {
		t.Fatalf("expected currency add to be cancelled")
	}
}

// TestCurrencyUpdatingEventAllowsAddCurrency verifies CurrencyUpdating passes and fires CurrencyUpdated.
func TestCurrencyUpdatingEventAllowsAddCurrency(t *testing.T) {
	var afterFired bool
	service, _ := inventoryapplication.NewService(repositoryStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkinventory.CurrencyUpdated); ok {
			afterFired = true
		}
	})
	bal, err := service.AddCurrencyTracked(context.Background(), 1, domain.CurrencyDuckets, 50, domain.SourceAdmin, "test", "1")
	if err != nil || bal != 150 {
		t.Fatalf("unexpected balance %d err=%v", bal, err)
	}
	if !afterFired {
		t.Fatalf("expected CurrencyUpdated event to fire")
	}
}
