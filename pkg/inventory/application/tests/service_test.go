package tests

import (
	"context"
	"errors"
	"testing"

	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := inventoryapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceCreditsFlow verifies credit operations behavior.
func TestServiceCreditsFlow(t *testing.T) {
	stub := repositoryStub{credits: 500}
	service, _ := inventoryapplication.NewService(stub)
	if _, err := service.GetCredits(context.Background(), 0); err == nil {
		t.Fatalf("expected get failure for invalid id")
	}
	credits, err := service.GetCredits(context.Background(), 1)
	if err != nil || credits != 500 {
		t.Fatalf("unexpected credits %d err=%v", credits, err)
	}
	newBalance, err := service.AddCredits(context.Background(), 1, 100)
	if err != nil || newBalance != 600 {
		t.Fatalf("unexpected add credits result %d err=%v", newBalance, err)
	}
}

// TestServiceCurrencyFlow verifies currency operations behavior.
func TestServiceCurrencyFlow(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{credits: 100})
	if _, err := service.ListCurrencies(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid id")
	}
	currencies, err := service.ListCurrencies(context.Background(), 1)
	if err != nil || len(currencies) != 1 {
		t.Fatalf("unexpected currencies len=%d err=%v", len(currencies), err)
	}
	newBal, err := service.AddCurrencyTracked(context.Background(), 1, domain.CurrencyDuckets, 50, domain.SourceAdmin, "test", "1")
	if err != nil || newBal != 150 {
		t.Fatalf("unexpected tracked add result %d err=%v", newBal, err)
	}
	if _, err := service.AddCurrencyTracked(context.Background(), 0, domain.CurrencyDuckets, 50, domain.SourceAdmin, "", ""); err == nil {
		t.Fatalf("expected tracked add failure for invalid id")
	}
}

// TestServiceBadgeFlow verifies badge operations behavior.
func TestServiceBadgeFlow(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{})
	if _, err := service.ListBadges(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid id")
	}
	if _, err := service.AwardBadge(context.Background(), 0, "ACH1"); err == nil {
		t.Fatalf("expected award failure for invalid id")
	}
	if _, err := service.AwardBadge(context.Background(), 1, ""); err == nil {
		t.Fatalf("expected award failure for empty code")
	}
	badge, err := service.AwardBadge(context.Background(), 1, "ACH1")
	if err != nil || badge.BadgeCode != "ACH1" {
		t.Fatalf("unexpected badge %+v err=%v", badge, err)
	}
	if err := service.UpdateBadgeSlots(context.Background(), 0, nil); err == nil {
		t.Fatalf("expected update slots failure for invalid id")
	}
	tooMany := make([]domain.BadgeSlot, domain.MaxBadgeSlots+1)
	if err := service.UpdateBadgeSlots(context.Background(), 1, tooMany); err == nil {
		t.Fatalf("expected update slots failure for too many slots")
	}
}

// TestServiceEffectFlow verifies effect operations behavior.
func TestServiceEffectFlow(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{})
	if _, err := service.ListEffects(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid id")
	}
	if _, err := service.AwardEffect(context.Background(), 0, 1, 60, false); err == nil {
		t.Fatalf("expected award failure for invalid user id")
	}
	if _, err := service.AwardEffect(context.Background(), 1, 0, 60, false); err == nil {
		t.Fatalf("expected award failure for invalid effect id")
	}
	if _, err := service.ActivateEffect(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected activate failure for invalid user id")
	}
	if _, err := service.ActivateEffect(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected activate failure for invalid effect id")
	}
}

// TestServicePropagatesErrors verifies repository error propagation.
func TestServicePropagatesErrors(t *testing.T) {
	service, _ := inventoryapplication.NewService(repositoryStub{findErr: errors.New("boom")})
	if _, err := service.GetCredits(context.Background(), 1); err == nil {
		t.Fatalf("expected credits error propagation")
	}
	if _, err := service.GetCurrency(context.Background(), 1, domain.CurrencyDuckets); err == nil {
		t.Fatalf("expected currency error propagation")
	}
}
