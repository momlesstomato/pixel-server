package tests

import (
	"context"
	"testing"
	"time"

	subscriptionapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

type creditSpenderStub struct{ credits int }

func (s creditSpenderStub) AddCredits(_ context.Context, _ int, amount int) (int, error) {
	return s.credits + amount, nil
}

type itemDelivererStub struct{ itemID int }

func (s itemDelivererStub) DeliverItem(_ context.Context, _ int, _ int, _ string, _ int, _ int) (int, error) {
	return s.itemID, nil
}

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := subscriptionapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceSubscriptionCRUD verifies subscription operations behavior.
func TestServiceSubscriptionCRUD(t *testing.T) {
	stub := repositoryStub{subscription: domain.Subscription{ID: 1, UserID: 1, DurationDays: 30, Active: true}}
	service, _ := subscriptionapplication.NewService(stub)
	if _, err := service.FindActiveSubscription(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	sub, err := service.FindActiveSubscription(context.Background(), 1)
	if err != nil || sub.ID != 1 {
		t.Fatalf("unexpected find result %+v err=%v", sub, err)
	}
	if _, err := service.CreateSubscription(context.Background(), domain.Subscription{}); err == nil {
		t.Fatalf("expected create failure for missing user id")
	}
	if _, err := service.CreateSubscription(context.Background(), domain.Subscription{UserID: 1}); err == nil {
		t.Fatalf("expected create failure for missing duration")
	}
	created, err := service.CreateSubscription(context.Background(), domain.Subscription{UserID: 1, DurationDays: 30})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
}

// TestServiceExtendSubscription verifies extension validation behavior.
func TestServiceExtendSubscription(t *testing.T) {
	stub := repositoryStub{subscription: domain.Subscription{ID: 1, DurationDays: 30}}
	service, _ := subscriptionapplication.NewService(stub)
	if _, err := service.ExtendSubscription(context.Background(), 0, 30); err == nil {
		t.Fatalf("expected extend failure for invalid user id")
	}
	if _, err := service.ExtendSubscription(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected extend failure for invalid days")
	}
	extended, err := service.ExtendSubscription(context.Background(), 1, 30)
	if err != nil || extended.DurationDays != 60 {
		t.Fatalf("unexpected extend result %+v err=%v", extended, err)
	}
}

// TestServiceDeactivateSubscription verifies deactivation validation behavior.
func TestServiceDeactivateSubscription(t *testing.T) {
	service, _ := subscriptionapplication.NewService(repositoryStub{})
	if err := service.DeactivateSubscription(context.Background(), 0); err == nil {
		t.Fatalf("expected deactivate failure for invalid id")
	}
}

// TestServiceExpireSubscriptions verifies batch expiration behavior.
func TestServiceExpireSubscriptions(t *testing.T) {
	stub := repositoryStub{expired: []domain.Subscription{{ID: 1, Active: true}, {ID: 2, Active: true}}}
	service, _ := subscriptionapplication.NewService(stub)
	expired, err := service.ExpireSubscriptions(context.Background())
	if err != nil || len(expired) != 2 {
		t.Fatalf("unexpected expire result len=%d err=%v", len(expired), err)
	}
}

// TestServiceClubOfferCRUD verifies club offer operations behavior.
func TestServiceClubOfferCRUD(t *testing.T) {
	stub := repositoryStub{clubOffer: domain.ClubOffer{ID: 1, Name: "HC 1 Month", Days: 31}}
	service, _ := subscriptionapplication.NewService(stub)
	if _, err := service.FindClubOfferByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	offer, err := service.FindClubOfferByID(context.Background(), 1)
	if err != nil || offer.Name != "HC 1 Month" {
		t.Fatalf("unexpected find result %+v err=%v", offer, err)
	}
	if _, err := service.CreateClubOffer(context.Background(), domain.ClubOffer{}); err == nil {
		t.Fatalf("expected create failure for empty name")
	}
	if _, err := service.CreateClubOffer(context.Background(), domain.ClubOffer{Name: "Test"}); err == nil {
		t.Fatalf("expected create failure for missing days")
	}
	created, err := service.CreateClubOffer(context.Background(), domain.ClubOffer{Name: "Test", Days: 30})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
	if err := service.DeleteClubOffer(context.Background(), 0); err == nil {
		t.Fatalf("expected delete failure for invalid id")
	}
	offers, err := service.ListClubOffers(context.Background())
	if err != nil || len(offers) != 1 {
		t.Fatalf("unexpected list result len=%d err=%v", len(offers), err)
	}
}

// TestServiceGetPaydayStatus verifies payday status is derived from config and benefits state.
func TestServiceGetPaydayStatus(t *testing.T) {
	stub := repositoryStub{
		subscription: domain.Subscription{ID: 1, UserID: 1, StartedAt: time.Now().Add(-48 * time.Hour), DurationDays: 365, Active: true},
		paydayConfig: domain.PaydayConfig{IntervalDays: 31, KickbackPercentage: 10},
		benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-48 * time.Hour), NextPaydayAt: time.Now().Add(24 * time.Hour), CycleCreditsSpent: 50},
	}
	service, _ := subscriptionapplication.NewService(stub)
	status, err := service.GetPaydayStatus(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected payday status error: %v", err)
	}
	if status.TotalRewardCredits != 5 {
		t.Fatalf("expected spend reward 5, got %d", status.TotalRewardCredits)
	}
	if status.CurrentHCStreakDays < 2 {
		t.Fatalf("expected HC streak days >= 2, got %d", status.CurrentHCStreakDays)
	}
}

// TestServiceTriggerPayday verifies payday triggering grants credits.
func TestServiceTriggerPayday(t *testing.T) {
	stub := repositoryStub{
		subscription: domain.Subscription{ID: 1, UserID: 1, StartedAt: time.Now().Add(-40 * 24 * time.Hour), DurationDays: 365, Active: true},
		paydayConfig: domain.PaydayConfig{IntervalDays: 31, KickbackPercentage: 10, FlatCredits: 2},
		benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-40 * 24 * time.Hour), NextPaydayAt: time.Now().Add(-1 * time.Hour), CycleCreditsSpent: 50},
	}
	service, _ := subscriptionapplication.NewService(stub)
	service.SetCreditSpender(creditSpenderStub{credits: 100})
	result, err := service.TriggerPayday(context.Background(), "", 1, false)
	if err != nil {
		t.Fatalf("unexpected trigger error: %v", err)
	}
	if result.RewardCredits != 7 {
		t.Fatalf("expected reward 7, got %d", result.RewardCredits)
	}
	if result.NewCredits != 107 {
		t.Fatalf("expected new credits 107, got %d", result.NewCredits)
	}
	}

// TestServiceClaimClubGift verifies club gift claims deliver items.
func TestServiceClaimClubGift(t *testing.T) {
	stub := repositoryStub{
		subscription: domain.Subscription{ID: 1, UserID: 1, StartedAt: time.Now().Add(-70 * 24 * time.Hour), DurationDays: 365, Active: true},
		paydayConfig: domain.PaydayConfig{IntervalDays: 31},
		benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-70 * 24 * time.Hour), NextPaydayAt: time.Now().Add(24 * time.Hour)},
		clubGifts: []domain.ClubGift{{ID: 1, Name: "Gray Dining Chair", ItemDefinitionID: 26, DaysRequired: 31}},
	}
	service, _ := subscriptionapplication.NewService(stub)
	service.SetItemDeliverer(itemDelivererStub{itemID: 77})
	result, err := service.ClaimClubGift(context.Background(), "", 1, "Gray Dining Chair")
	if err != nil {
		t.Fatalf("unexpected club gift claim error: %v", err)
	}
	if result.ItemID != 77 {
		t.Fatalf("expected delivered item 77, got %d", result.ItemID)
	}
}
