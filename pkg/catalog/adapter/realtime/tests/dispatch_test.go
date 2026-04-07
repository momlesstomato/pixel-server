package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/catalog/adapter/realtime"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogpacket "github.com/momlesstomato/pixel-server/pkg/catalog/packet"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	inventorypkt "github.com/momlesstomato/pixel-server/pkg/inventory/packet"
)

// repoStub provides deterministic catalog data.
type repoStub struct {
	offer      domain.CatalogOffer
	pageLayout string
}

func (r repoStub) ListPages(_ context.Context) ([]domain.CatalogPage, error) { return nil, nil }
func (r repoStub) FindPageByID(_ context.Context, _ int) (domain.CatalogPage, error) {
	pageLayout := r.pageLayout
	if pageLayout == "" {
		pageLayout = "default_3x3"
	}
	return domain.CatalogPage{ID: 1, PageLayout: pageLayout, Enabled: true, Visible: true}, nil
}
func (r repoStub) CreatePage(_ context.Context, p domain.CatalogPage) (domain.CatalogPage, error) {
	return p, nil
}
func (r repoStub) UpdatePage(_ context.Context, _ int, _ domain.PagePatch) (domain.CatalogPage, error) {
	return domain.CatalogPage{}, nil
}
func (r repoStub) DeletePage(_ context.Context, _ int) error { return nil }
func (r repoStub) ListOffersByPageID(_ context.Context, _ int) ([]domain.CatalogOffer, error) {
	return []domain.CatalogOffer{r.offer}, nil
}
func (r repoStub) FindOfferByID(_ context.Context, _ int) (domain.CatalogOffer, error) {
	return r.offer, nil
}
func (r repoStub) CreateOffer(_ context.Context, o domain.CatalogOffer) (domain.CatalogOffer, error) {
	return o, nil
}
func (r repoStub) UpdateOffer(_ context.Context, _ int, _ domain.OfferPatch) (domain.CatalogOffer, error) {
	return domain.CatalogOffer{}, nil
}
func (r repoStub) DeleteOffer(_ context.Context, _ int) error                   { return nil }
func (r repoStub) IncrementLimitedSells(_ context.Context, _ int) (bool, error) { return true, nil }
func (r repoStub) FindVoucherByCode(_ context.Context, _ string) (domain.Voucher, error) {
	return domain.Voucher{}, nil
}
func (r repoStub) CreateVoucher(_ context.Context, v domain.Voucher) (domain.Voucher, error) {
	return v, nil
}
func (r repoStub) DeleteVoucher(_ context.Context, _ int) error             { return nil }
func (r repoStub) ListVouchers(_ context.Context) ([]domain.Voucher, error) { return nil, nil }
func (r repoStub) RedeemVoucher(_ context.Context, _ int, _ int) error      { return nil }
func (r repoStub) HasUserRedeemedVoucher(_ context.Context, _ int, _ int) (bool, error) {
	return false, nil
}

// sessionStub provides deterministic session lookup.
type sessionStub struct{}

func (sessionStub) Register(coreconnection.Session) error { return nil }
func (sessionStub) FindByConnID(id string) (coreconnection.Session, bool) {
	if id == "conn1" {
		return coreconnection.Session{UserID: 1}, true
	}
	return coreconnection.Session{}, false
}
func (sessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}
func (sessionStub) Touch(string) error                         { return nil }
func (sessionStub) Remove(string)                              {}
func (sessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// transportStub captures sent packets.
type transportStub struct{ sent []uint16 }

func (t *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	t.sent = append(t.sent, packetID)
	return nil
}

// buildService creates a catalog service using the given repo stub.
func buildService(repo domain.Repository) *catalogapplication.Service {
	svc, _ := catalogapplication.NewService(repo)
	return svc
}

// buildPurchaseBody encodes a minimal purchase request body (pageID=a, offerID=b, extraData="", amount=1).
func buildPurchaseBody(pageID, offerID int32) []byte {
	buf := make([]byte, 8)
	buf[0] = byte(pageID >> 24)
	buf[1] = byte(pageID >> 16)
	buf[2] = byte(pageID >> 8)
	buf[3] = byte(pageID)
	buf[4] = byte(offerID >> 24)
	buf[5] = byte(offerID >> 16)
	buf[6] = byte(offerID >> 8)
	buf[7] = byte(offerID)
	return buf
}

// zeroSpender returns zero balances simulating an empty wallet.
type zeroSpender struct{}

func (zeroSpender) GetCredits(_ context.Context, _ int) (int, error)                { return 0, nil }
func (zeroSpender) AddCredits(_ context.Context, _ int, d int) (int, error)         { return d, nil }
func (zeroSpender) GetCurrencyBalance(_ context.Context, _ int, _ int) (int, error) { return 0, nil }
func (zeroSpender) AddCurrencyBalance(_ context.Context, _ int, _ int, d int) (int, error) {
	return d, nil
}

// itemDelivererStub returns a deterministic delivered item identifier.
type itemDelivererStub struct{ itemID int }

// DeliverItem returns the configured delivered item identifier.
func (d itemDelivererStub) DeliverItem(_ context.Context, _ int, _ int, _ string, _ int, _ int) (int, error) {
	return d.itemID, nil
}

// recipientFinderStub returns deterministic gift recipient data.
type recipientFinderStub struct{ info domain.RecipientInfo }

// FindRecipientByUsername returns the configured recipient info.
func (f recipientFinderStub) FindRecipientByUsername(_ context.Context, _ string) (domain.RecipientInfo, error) {
	return f.info, nil
}

// buildGiftBody encodes a minimal gift purchase request body.
func buildGiftBody(pageID, offerID int32, recipient string) []byte {
	w := codec.NewWriter()
	w.WriteInt32(pageID)
	w.WriteInt32(offerID)
	_ = w.WriteString("")
	_ = w.WriteString(recipient)
	return w.Bytes()
}

// TestHandleGiftWrappingConfigSendsResponse verifies 418 triggers gift config packet 2234.
func TestHandleGiftWrappingConfigSendsResponse(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.GetGiftWrappingConfigPacketID, nil)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != catalogpacket.GiftWrappingConfigResponsePacketID {
		t.Fatalf("expected gift config packet 2234, got %v", transport.sent)
	}
}

// TestHandlePurchaseFreeOfferSendsPurchaseOK verifies purchase of free offer sends 869.
func TestHandlePurchaseFreeOfferSendsPurchaseOK(t *testing.T) {
	offer := domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 0, ItemType: "s"}
	transport := &transportStub{}
	svc := buildService(repoStub{offer: offer})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	body := buildPurchaseBody(1, 1)
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.PurchasePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) < 2 || transport.sent[0] != catalogpacket.PurchaseOKPacketID || transport.sent[1] != inventorypkt.CreditsResponsePacketID {
		t.Fatalf("expected [%d %d] packets, got %v", catalogpacket.PurchaseOKPacketID, inventorypkt.CreditsResponsePacketID, transport.sent)
	}
}

// TestHandlePurchaseBoughtFurniturePushesInventoryDelta verifies purchases send the live inventory add packet before the unseen marker.
func TestHandlePurchaseBoughtFurniturePushesInventoryDelta(t *testing.T) {
	offer := domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 0, ItemType: "s", ItemDefinitionID: 9}
	transport := &transportStub{}
	svc := buildService(repoStub{offer: offer})
	svc.SetItemDeliverer(itemDelivererStub{itemID: 42})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	rt.SetInventoryItemSender(func(_ context.Context, connID string, userID int, itemID int) error {
		if connID != "conn1" || userID != 1 || itemID != 42 {
			t.Fatalf("unexpected inventory sender args conn=%s user=%d item=%d", connID, userID, itemID)
		}
		return transport.Send(connID, furnipacket.InventoryAddPacketID, nil)
	})
	body := buildPurchaseBody(1, 1)
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.PurchasePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	expected := []uint16{
		catalogpacket.PurchaseOKPacketID,
		inventorypkt.CreditsResponsePacketID,
		furnipacket.InventoryAddPacketID,
		catalogpacket.FurniListNotificationPacketID,
	}
	if len(transport.sent) != len(expected) {
		t.Fatalf("expected packets %v, got %v", expected, transport.sent)
	}
	for i, packetID := range expected {
		if transport.sent[i] != packetID {
			t.Fatalf("expected packets %v, got %v", expected, transport.sent)
		}
	}
}

// TestHandlePurchaseGiftSkipsBuyerInventoryDelta verifies gift purchases do not mutate the buyer inventory view.
func TestHandlePurchaseGiftSkipsBuyerInventoryDelta(t *testing.T) {
	offer := domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 0, ItemType: "s", ItemDefinitionID: 9}
	transport := &transportStub{}
	svc := buildService(repoStub{offer: offer})
	svc.SetItemDeliverer(itemDelivererStub{itemID: 51})
	svc.SetRecipientFinder(recipientFinderStub{info: domain.RecipientInfo{UserID: 2, AllowGifts: true}})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	inventoryPushed := false
	rt.SetInventoryItemSender(func(_ context.Context, _ string, _ int, _ int) error {
		inventoryPushed = true
		return nil
	})
	body := buildGiftBody(1, 1, "target")
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.PurchaseGiftPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if inventoryPushed {
		t.Fatal("expected no buyer inventory push for gift purchase")
	}
	expected := []uint16{catalogpacket.PurchaseOKPacketID, inventorypkt.CreditsResponsePacketID}
	if len(transport.sent) != len(expected) {
		t.Fatalf("expected packets %v, got %v", expected, transport.sent)
	}
	for i, packetID := range expected {
		if transport.sent[i] != packetID {
			t.Fatalf("expected packets %v, got %v", expected, transport.sent)
		}
	}
}

// TestHandlePurchaseInsufficientCredits verifies insufficient credits sends error packet 1404.
func TestHandlePurchaseInsufficientCredits(t *testing.T) {
	offer := domain.CatalogOffer{ID: 1, OfferActive: true, CostCredits: 100, ItemType: "s"}
	transport := &transportStub{}
	svc := buildService(repoStub{offer: offer})
	svc.SetSpender(zeroSpender{})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	body := buildPurchaseBody(1, 1)
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.PurchasePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != catalogpacket.PurchaseErrorPacketID {
		t.Fatalf("expected purchase_error packet 1404, got %v", transport.sent)
	}
}

// TestHandleGetPageVipBuyTriggersClubOffers verifies vip_buy pages trigger the HC offers side-channel.
func TestHandleGetPageVipBuyTriggersClubOffers(t *testing.T) {
	transport := &transportStub{}
	svc := buildService(repoStub{pageLayout: "vip_buy"})
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	triggered := false
	rt.SetClubOffersSender(func(_ context.Context, connID string, userID int) error {
		triggered = true
		if connID != "conn1" || userID != 1 {
			t.Fatalf("unexpected club offers sender args conn=%s user=%d", connID, userID)
		}
		return nil
	})
	body := buildPurchaseBody(1, 0)
	handled, err := rt.Handle(context.Background(), "conn1", catalogpacket.GetPagePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if !triggered {
		t.Fatal("expected vip_buy page to trigger club offers sender")
	}
	if len(transport.sent) != 1 || transport.sent[0] != catalogpacket.PageResponsePacketID {
		t.Fatalf("expected only catalog page packet before side-channel send, got %v", transport.sent)
	}
}

// TestHandleUnknownPacketNotHandled verifies unknown packet ID returns false.
func TestHandleUnknownPacketNotHandled(t *testing.T) {
	svc := buildService(repoStub{})
	transport := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, transport, nil)
	handled, err := rt.Handle(context.Background(), "conn1", 9999, nil)
	if err != nil || handled {
		t.Fatalf("expected not handled, got handled=%v err=%v", handled, err)
	}
}
