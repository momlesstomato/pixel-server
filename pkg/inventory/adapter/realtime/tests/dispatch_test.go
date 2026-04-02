package realtime_test

import (
	"context"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/inventory/adapter/realtime"
	inventoryapp "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorypkt "github.com/momlesstomato/pixel-server/pkg/inventory/packet"
)

// transportStub captures sent packets for assertion.
type transportStub struct {
	// sent stores all packets transmitted by the runtime.
	sent []uint16
}

// Send appends the packet ID to the sent list.
func (t *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	t.sent = append(t.sent, packetID)
	return nil
}

// sessionStub provides deterministic session lookup backed by conn1.
type sessionStub struct{}

// Register is a no-op stub implementation.
func (sessionStub) Register(coreconnection.Session) error { return nil }

// FindByConnID returns a session for conn1 only.
func (sessionStub) FindByConnID(id string) (coreconnection.Session, bool) {
	if id == "conn1" {
		return coreconnection.Session{UserID: 1}, true
	}
	return coreconnection.Session{}, false
}

// FindByUserID is a no-op stub implementation.
func (sessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}

// Touch is a no-op stub implementation.
func (sessionStub) Touch(string) error { return nil }

// Remove is a no-op stub implementation.
func (sessionStub) Remove(string) {}

// ListAll is a no-op stub implementation.
func (sessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// repoStub returns deterministic inventory data for all Repository methods.
type repoStub struct{}

// ListCurrencyTypes returns empty result.
func (repoStub) ListCurrencyTypes(context.Context) ([]domain.ActivityCurrencyType, error) {
	return nil, nil
}

// FindCurrencyTypeByID returns zero value.
func (repoStub) FindCurrencyTypeByID(context.Context, int) (domain.ActivityCurrencyType, error) {
	return domain.ActivityCurrencyType{}, nil
}

// ListBadges returns empty result.
func (repoStub) ListBadges(context.Context, int) ([]domain.Badge, error) { return nil, nil }

// AwardBadge returns a badge with the given code.
func (repoStub) AwardBadge(_ context.Context, _ int, code string) (domain.Badge, error) {
	return domain.Badge{BadgeCode: code}, nil
}

// RevokeBadge is a no-op stub.
func (repoStub) RevokeBadge(context.Context, int, string) error { return nil }

// UpdateBadgeSlots is a no-op stub.
func (repoStub) UpdateBadgeSlots(context.Context, int, []domain.BadgeSlot) error { return nil }

// GetEquippedBadges returns empty result.
func (repoStub) GetEquippedBadges(context.Context, int) ([]domain.BadgeSlot, error) {
	return nil, nil
}

// GetCredits returns 200 for any user.
func (repoStub) GetCredits(context.Context, int) (int, error) { return 200, nil }

// SetCredits is a no-op stub.
func (repoStub) SetCredits(context.Context, int, int) error { return nil }

// AddCredits returns the incremented value.
func (repoStub) AddCredits(_ context.Context, _ int, amount int) (int, error) {
	return 200 + amount, nil
}

// GetCurrency returns 100 for any user and type.
func (repoStub) GetCurrency(context.Context, int, domain.CurrencyType) (int, error) {
	return 100, nil
}

// ListCurrencies returns one Duckets entry.
func (repoStub) ListCurrencies(context.Context, int) ([]domain.Currency, error) {
	return []domain.Currency{{Type: domain.CurrencyDuckets, Amount: 100}}, nil
}

// SetCurrency is a no-op stub.
func (repoStub) SetCurrency(context.Context, int, domain.CurrencyType, int) error { return nil }

// AddCurrency returns the incremented value.
func (repoStub) AddCurrency(_ context.Context, _ int, _ domain.CurrencyType, amount int) (int, error) {
	return 100 + amount, nil
}

// RecordTransaction is a no-op stub.
func (repoStub) RecordTransaction(context.Context, domain.CurrencyTransaction) error { return nil }

// ListTransactions returns empty result.
func (repoStub) ListTransactions(context.Context, int, domain.CurrencyType, int) ([]domain.CurrencyTransaction, error) {
	return nil, nil
}

// ListEffects returns empty result.
func (repoStub) ListEffects(context.Context, int) ([]domain.Effect, error) { return nil, nil }

// AwardEffect returns an effect with the given id.
func (repoStub) AwardEffect(_ context.Context, _ int, id int, _ int, _ bool) (domain.Effect, error) {
	return domain.Effect{EffectID: id}, nil
}

// ActivateEffect returns zero value.
func (repoStub) ActivateEffect(context.Context, int, int) (domain.Effect, error) {
	return domain.Effect{}, nil
}

// RemoveExpiredEffects returns empty result.
func (repoStub) RemoveExpiredEffects(context.Context) ([]domain.ExpiredEffect, error) {
	return nil, nil
}

// TestHandleGetCurrencySendsBothPackets verifies handleGetCurrency sends credits and currency packets.
func TestHandleGetCurrencySendsBothPackets(t *testing.T) {
	transport := &transportStub{}
	service, _ := inventoryapp.NewService(repoStub{})
	rt, _ := realtime.NewRuntime(service, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", inventorypkt.GetCurrencyPacketID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Fatal("expected GetCurrencyPacketID to be handled")
	}
	if len(transport.sent) != 2 {
		t.Fatalf("expected 2 packets sent, got %d", len(transport.sent))
	}
	ids := map[uint16]bool{}
	for _, id := range transport.sent {
		ids[id] = true
	}
	if !ids[inventorypkt.CreditsResponsePacketID] {
		t.Fatal("expected credits response packet to be sent")
	}
	if !ids[inventorypkt.CurrencyResponsePacketID] {
		t.Fatal("expected currency response packet to be sent")
	}
}

// TestHandleUnknownPacketReturnsFalse verifies unknown packet IDs are not handled.
func TestHandleUnknownPacketReturnsFalse(t *testing.T) {
	transport := &transportStub{}
	service, _ := inventoryapp.NewService(repoStub{})
	rt, _ := realtime.NewRuntime(service, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", 9999, nil)
	if err != nil || handled {
		t.Fatalf("expected unhandled: handled=%v err=%v", handled, err)
	}
}
