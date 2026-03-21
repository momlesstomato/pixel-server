package tests

import (
	"context"
	"errors"
	"testing"

	economyapplication "github.com/momlesstomato/pixel-server/pkg/economy/application"
	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// TestNewServiceRejectsNilRepository verifies constructor precondition validation.
func TestNewServiceRejectsNilRepository(t *testing.T) {
	if _, err := economyapplication.NewService(nil); err == nil {
		t.Fatalf("expected nil repository validation failure")
	}
}

// TestServiceMarketplaceOfferCRUD verifies marketplace offer operations.
func TestServiceMarketplaceOfferCRUD(t *testing.T) {
	stub := repositoryStub{offer: domain.MarketplaceOffer{ID: 1, SellerID: 1, AskingPrice: 100}}
	service, _ := economyapplication.NewService(stub)
	if _, err := service.FindOfferByID(context.Background(), 0); err == nil {
		t.Fatalf("expected find failure for invalid id")
	}
	offer, err := service.FindOfferByID(context.Background(), 1)
	if err != nil || offer.ID != 1 {
		t.Fatalf("unexpected find result %+v err=%v", offer, err)
	}
	if _, err := service.ListOffersBySellerID(context.Background(), 0); err == nil {
		t.Fatalf("expected list failure for invalid seller id")
	}
	if _, err := service.CreateOffer(context.Background(), domain.MarketplaceOffer{}); err == nil {
		t.Fatalf("expected create failure for missing seller id")
	}
	if _, err := service.CreateOffer(context.Background(), domain.MarketplaceOffer{SellerID: 1}); err == nil {
		t.Fatalf("expected create failure for missing price")
	}
	created, err := service.CreateOffer(context.Background(), domain.MarketplaceOffer{SellerID: 1, AskingPrice: 50})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result %+v err=%v", created, err)
	}
}

// TestServicePurchaseAndCancel verifies purchase and cancel validation.
func TestServicePurchaseAndCancel(t *testing.T) {
	stub := repositoryStub{offer: domain.MarketplaceOffer{ID: 1, State: domain.OfferStateOpen}}
	service, _ := economyapplication.NewService(stub)
	if _, err := service.PurchaseOffer(context.Background(), 0, 1); err == nil {
		t.Fatalf("expected purchase failure for invalid offer id")
	}
	if _, err := service.PurchaseOffer(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected purchase failure for invalid buyer id")
	}
	if err := service.CancelOffer(context.Background(), 0); err == nil {
		t.Fatalf("expected cancel failure for invalid id")
	}
}

// TestServiceOpenOffersDefaultLimit verifies default limit for open offers.
func TestServiceOpenOffersDefaultLimit(t *testing.T) {
	stub := repositoryStub{offer: domain.MarketplaceOffer{ID: 1}}
	service, _ := economyapplication.NewService(stub)
	offers, count, err := service.ListOpenOffers(context.Background(), domain.OfferFilter{})
	if err != nil || count != 1 || len(offers) != 1 {
		t.Fatalf("unexpected open offers result len=%d count=%d err=%v", len(offers), count, err)
	}
}

// TestServiceTradeLog verifies trade log creation validation.
func TestServiceTradeLog(t *testing.T) {
	service, _ := economyapplication.NewService(repositoryStub{})
	if _, err := service.CreateTradeLog(context.Background(), domain.TradeLog{}); err == nil {
		t.Fatalf("expected create failure for missing user ids")
	}
	if _, err := service.CreateTradeLog(context.Background(), domain.TradeLog{UserOneID: 1, UserTwoID: 1}); err == nil {
		t.Fatalf("expected create failure for self-trade")
	}
	log, err := service.CreateTradeLog(context.Background(), domain.TradeLog{UserOneID: 1, UserTwoID: 2})
	if err != nil || log.ID != 1 {
		t.Fatalf("unexpected trade log result %+v err=%v", log, err)
	}
}

// TestServicePriceHistory verifies price history operations.
func TestServicePriceHistory(t *testing.T) {
	service, _ := economyapplication.NewService(repositoryStub{})
	if _, err := service.GetPriceHistory(context.Background(), 0); err == nil {
		t.Fatalf("expected get failure for invalid sprite id")
	}
	if err := service.RecordPriceHistory(context.Background(), domain.PriceHistory{}); err == nil {
		t.Fatalf("expected record failure for invalid sprite id")
	}
}

// TestServiceCountActiveOffers verifies active offer counting.
func TestServiceCountActiveOffers(t *testing.T) {
	service, _ := economyapplication.NewService(repositoryStub{})
	if _, err := service.CountActiveOffers(context.Background(), 0); err == nil {
		t.Fatalf("expected count failure for invalid seller id")
	}
	count, err := service.CountActiveOffers(context.Background(), 1)
	if err != nil || count != 3 {
		t.Fatalf("unexpected count result %d err=%v", count, err)
	}
}

// TestServicePropagatesErrors verifies repository error propagation.
func TestServicePropagatesErrors(t *testing.T) {
	service, _ := economyapplication.NewService(repositoryStub{findErr: errors.New("boom")})
	if _, err := service.FindOfferByID(context.Background(), 1); err == nil {
		t.Fatalf("expected find failure")
	}
	if _, err := service.ListOffersBySellerID(context.Background(), 1); err == nil {
		t.Fatalf("expected list failure")
	}
}
