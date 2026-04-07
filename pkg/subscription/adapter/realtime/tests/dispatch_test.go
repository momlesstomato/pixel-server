package tests

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/subscription/adapter/realtime"
	subapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	subpacket "github.com/momlesstomato/pixel-server/pkg/subscription/packet"
)

// repoStub provides deterministic subscription repository responses for tests.
type repoStub struct {
	// sub stores the subscription returned by FindActiveSubscription.
	sub domain.Subscription
	// offers stores the offers returned by ListClubOffers.
	offers []domain.ClubOffer
	// findErr stores the error returned by FindActiveSubscription.
	findErr error
	// paydayConfig stores deterministic payday config.
	paydayConfig domain.PaydayConfig
	// benefitsState stores deterministic benefits progress.
	benefitsState domain.BenefitsState
	// gifts stores deterministic club gifts.
	gifts []domain.ClubGift
}

// FindActiveSubscription returns the stub subscription or error.
func (r repoStub) FindActiveSubscription(_ context.Context, _ int) (domain.Subscription, error) {
	return r.sub, r.findErr
}

// CreateSubscription returns noop create.
func (r repoStub) CreateSubscription(_ context.Context, s domain.Subscription) (domain.Subscription, error) {
	s.ID = 1
	return s, nil
}

// ExtendSubscription returns noop extend.
func (r repoStub) ExtendSubscription(_ context.Context, _ int, days int) (domain.Subscription, error) {
	s := r.sub
	s.DurationDays += days
	return s, nil
}

// DeactivateSubscription returns noop deactivate.
func (r repoStub) DeactivateSubscription(_ context.Context, _ int) error { return nil }

// FindExpiredActive returns empty expired list.
func (r repoStub) FindExpiredActive(_ context.Context) ([]domain.Subscription, error) {
	return nil, nil
}

// ListClubOffers returns stub offers list.
func (r repoStub) ListClubOffers(_ context.Context) ([]domain.ClubOffer, error) {
	return r.offers, nil
}

// FindClubOfferByID returns first stub offer.
func (r repoStub) FindClubOfferByID(_ context.Context, _ int) (domain.ClubOffer, error) {
	if len(r.offers) == 0 {
		return domain.ClubOffer{}, domain.ErrClubOfferNotFound
	}
	return r.offers[0], nil
}

// CreateClubOffer returns noop create.
func (r repoStub) CreateClubOffer(_ context.Context, o domain.ClubOffer) (domain.ClubOffer, error) {
	o.ID = 1
	return o, nil
}

// DeleteClubOffer returns noop delete.
func (r repoStub) DeleteClubOffer(_ context.Context, _ int) error { return nil }

// FindPaydayConfig returns deterministic payday config.
func (r repoStub) FindPaydayConfig(_ context.Context) (domain.PaydayConfig, error) {
	if r.paydayConfig.IntervalDays == 0 {
		return domain.PaydayConfig{}, domain.ErrPaydayConfigNotFound
	}
	return r.paydayConfig, nil
}

// SavePaydayConfig returns deterministic payday config.
func (r repoStub) SavePaydayConfig(_ context.Context, cfg domain.PaydayConfig) (domain.PaydayConfig, error) {
	return cfg, nil
}

// FindBenefitsState returns deterministic benefits state.
func (r repoStub) FindBenefitsState(_ context.Context, _ int) (domain.BenefitsState, error) {
	if r.benefitsState.UserID == 0 {
		return domain.BenefitsState{}, domain.ErrBenefitsStateNotFound
	}
	return r.benefitsState, nil
}

// SaveBenefitsState returns deterministic benefits state.
func (r repoStub) SaveBenefitsState(_ context.Context, state domain.BenefitsState) (domain.BenefitsState, error) {
	return state, nil
}

// ListClubGifts returns deterministic club gifts.
func (r repoStub) ListClubGifts(_ context.Context) ([]domain.ClubGift, error) {
	return r.gifts, nil
}

// FindClubGiftByName resolves one deterministic club gift by name.
func (r repoStub) FindClubGiftByName(_ context.Context, name string) (domain.ClubGift, error) {
	for _, gift := range r.gifts {
		if gift.Name == name {
			return gift, nil
		}
	}
	return domain.ClubGift{}, domain.ErrClubGiftNotFound
}

// sessionStub provides deterministic session lookup.
type sessionStub struct{}

// Register is a noop for test sessions.
func (sessionStub) Register(coreconnection.Session) error { return nil }

// FindByConnID returns user 1 for conn1 and miss otherwise.
func (sessionStub) FindByConnID(id string) (coreconnection.Session, bool) {
	if id == "conn1" {
		return coreconnection.Session{UserID: 1}, true
	}
	return coreconnection.Session{}, false
}

// FindByUserID returns a miss for test sessions.
func (sessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}

// Touch is a noop for test sessions.
func (sessionStub) Touch(string) error { return nil }

// Remove is a noop for test sessions.
func (sessionStub) Remove(string) {}

// ListAll returns empty for test sessions.
func (sessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// transportStub captures sent packet IDs.
type transportStub struct {
	// sent records packet identifiers in send order.
	sent []uint16
	// bodies records encoded packet bodies in send order.
	bodies [][]byte
}

type giftDelivererStub struct{ itemID int }

func (d giftDelivererStub) DeliverItem(_ context.Context, _ int, _ int, _ string, _ int, _ int) (int, error) {
	return d.itemID, nil
}

// Send records the packet identifier and discards the payload.
func (t *transportStub) Send(_ string, packetID uint16, body []byte) error {
	t.sent = append(t.sent, packetID)
	t.bodies = append(t.bodies, append([]byte(nil), body...))
	return nil
}

// buildService creates a subscription service backed by the given stub.
func buildService(repo domain.Repository) *subapplication.Service {
	svc, _ := subapplication.NewService(repo)
	return svc
}

// buildStringBody encodes a string body for get_subscription request.
func buildStringBody(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.BigEndian.PutUint16(b[:2], uint16(len(s)))
	copy(b[2:], s)
	return b
}

// activeOffer returns a minimal club offer suitable for packet encoding tests.
func activeOffer() domain.ClubOffer {
	return domain.ClubOffer{ID: 1, Name: "HC 1 Month", Days: 31, Credits: 25, Enabled: true}
}

// activeSub returns an active subscription with 365 days duration.
func activeSub() domain.Subscription {
	return domain.Subscription{
		ID: 1, UserID: 1, SubscriptionType: domain.SubscriptionHabboClub,
		StartedAt: time.Now(), DurationDays: 365, Active: true,
	}
}

// TestHandleGetSubscriptionSendsSubscriptionPacket verifies 3166 returns 954 when subscription exists.
func TestHandleGetSubscriptionSendsSubscriptionPacket(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{sub: activeSub()})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetSubscriptionPacketID, buildStringBody("club_habbo"))
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.SubscriptionResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.SubscriptionResponsePacketID, transport.sent)
	}
}

// TestHandleGetSubscriptionSendsFallbackWhenNotFound verifies 3166 returns 954 with type 1 for missing subscription.
func TestHandleGetSubscriptionSendsFallbackWhenNotFound(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{findErr: domain.ErrSubscriptionNotFound})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetSubscriptionPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.SubscriptionResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.SubscriptionResponsePacketID, transport.sent)
	}
}

// TestHandleGetClubOffersReturnsClubOffersPacket verifies 3285 returns 2405 with available offers.
func TestHandleGetClubOffersReturnsClubOffersPacket(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{offers: []domain.ClubOffer{activeOffer()}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetClubOffersPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 2 || transport.sent[0] != subpacket.ClubOffersResponsePacketID || transport.sent[1] != subpacket.DirectClubBuyAvailableResponsePacketID {
		t.Fatalf("expected packets [%d %d], got %v", subpacket.ClubOffersResponsePacketID, subpacket.DirectClubBuyAvailableResponsePacketID, transport.sent)
	}
	r := codec.NewReader(transport.bodies[0])
	count, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("expected offers count, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one club offer, got %d", count)
	}
	_, _ = r.ReadInt32()
	_, _ = r.ReadString()
	_, _ = r.ReadBool()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadBool()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadBool()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	if _, err = r.ReadInt32(); err == nil {
		t.Fatal("expected no trailing club offers window id")
	}
	direct := codec.NewReader(transport.bodies[1])
	url, err := direct.ReadString()
	if err != nil {
		t.Fatalf("expected direct club buy url, got %v", err)
	}
	if url != "" {
		t.Fatalf("expected empty direct club buy url, got %q", url)
	}
	market, err := direct.ReadString()
	if err != nil {
		t.Fatalf("expected direct club buy market, got %v", err)
	}
	if market != "" {
		t.Fatalf("expected empty direct club buy market, got %q", market)
	}
	days, err := direct.ReadInt32()
	if err != nil {
		t.Fatalf("expected direct club buy days, got %v", err)
	}
	if days != int32(activeOffer().Days) {
		t.Fatalf("expected direct club buy days %d, got %d", activeOffer().Days, days)
	}
}

// TestHandleGetHCExtendOfferReturnsExtendPacketWhenSubscribed verifies 2462 returns 3964 when user has active subscription.
func TestHandleGetHCExtendOfferReturnsExtendPacketWhenSubscribed(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{sub: activeSub(), offers: []domain.ClubOffer{activeOffer()}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetHCExtendOfferPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.HCExtendOfferResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.HCExtendOfferResponsePacketID, transport.sent)
	}
}

// TestHandleGetHCExtendOfferFallsBackToClubOffersWhenNoSubscription verifies 2462 returns 2405 when user has no subscription.
func TestHandleGetHCExtendOfferFallsBackToClubOffersWhenNoSubscription(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{findErr: domain.ErrSubscriptionNotFound, offers: []domain.ClubOffer{activeOffer()}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetHCExtendOfferPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 2 || transport.sent[0] != subpacket.ClubOffersResponsePacketID || transport.sent[1] != subpacket.DirectClubBuyAvailableResponsePacketID {
		t.Fatalf("expected packets [%d %d], got %v", subpacket.ClubOffersResponsePacketID, subpacket.DirectClubBuyAvailableResponsePacketID, transport.sent)
	}
}

// TestHandleGetClubGiftInfoSendsGiftInfoPacket verifies 487 returns 619.
func TestHandleGetClubGiftInfoSendsGiftInfoPacket(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{
		sub:          domain.Subscription{ID: 1, UserID: 1, SubscriptionType: domain.SubscriptionHabboClub, StartedAt: time.Now().Add(-70 * 24 * time.Hour), DurationDays: 365, Active: true},
		paydayConfig: domain.PaydayConfig{IntervalDays: 31},
		benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-70 * 24 * time.Hour), NextPaydayAt: time.Now().Add(24 * time.Hour)},
		gifts:        []domain.ClubGift{{ID: 1, Name: "Gray Dining Chair", SpriteID: 26, DaysRequired: 31}},
	})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetClubGiftInfoPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.ClubGiftInfoResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.ClubGiftInfoResponsePacketID, transport.sent)
	}
	r := codec.NewReader(transport.bodies[0])
	_, _ = r.ReadInt32()
	available, _ := r.ReadInt32()
	if available < 1 {
		t.Fatalf("expected at least one available gift, got %d", available)
	}
}

// TestHandleGetKickbackInfoSendsPacket verifies 869 returns 3277 with deterministic HC metadata.
func TestHandleGetKickbackInfoSendsPacket(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{sub: activeSub(), paydayConfig: domain.PaydayConfig{IntervalDays: 31}, benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-24 * time.Hour), NextPaydayAt: time.Now().Add(24 * time.Hour)}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.GetKickbackInfoPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.KickbackInfoResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.KickbackInfoResponsePacketID, transport.sent)
	}
	r := codec.NewReader(transport.bodies[0])
	streak, _ := r.ReadInt32()
	if streak != 1 {
		t.Fatalf("expected streak 1, got %d", streak)
	}
	joinedAt, err := r.ReadString()
	if err != nil || joinedAt == "" {
		t.Fatalf("expected joined date string, got %q err=%v", joinedAt, err)
	}
}

// TestHandleSelectClubGiftSendsSelectedPacket verifies 2276 returns 659 and can deliver an item.
func TestHandleSelectClubGiftSendsSelectedPacket(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{
		sub:          domain.Subscription{ID: 1, UserID: 1, StartedAt: time.Now().Add(-70 * 24 * time.Hour), DurationDays: 365, Active: true},
		paydayConfig: domain.PaydayConfig{IntervalDays: 31},
		benefitsState: domain.BenefitsState{UserID: 1, FirstSubscriptionAt: time.Now().Add(-70 * 24 * time.Hour), NextPaydayAt: time.Now().Add(24 * time.Hour)},
		gifts:        []domain.ClubGift{{ID: 1, Name: "Gray Dining Chair", ItemDefinitionID: 26, SpriteID: 26, DaysRequired: 31}},
	})
	svc.SetItemDeliverer(giftDelivererStub{itemID: 99})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", subpacket.SelectClubGiftPacketID, buildStringBody("Gray Dining Chair"))
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != subpacket.ClubGiftSelectedResponsePacketID {
		t.Fatalf("expected packet %d, got %v", subpacket.ClubGiftSelectedResponsePacketID, transport.sent)
	}
}

// TestHandleUnknownPacketIsNotClaimed verifies unrecognized packet IDs return handled=false.
func TestHandleUnknownPacketIsNotClaimed(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", 9999, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handled {
		t.Fatal("expected packet to be unclaimed")
	}
}

// TestHandleDropsUnauthenticatedConnection verifies no packet is sent for unknown conn.
func TestHandleDropsUnauthenticatedConnection(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{offers: []domain.ClubOffer{activeOffer()}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "unknown", subpacket.GetClubOffersPacketID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if handled {
		t.Fatal("expected packet to be unclaimed for unauthenticated connection")
	}
	if len(transport.sent) != 0 {
		t.Fatalf("expected no packets sent, got %v", transport.sent)
	}
}
