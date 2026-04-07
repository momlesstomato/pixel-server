package packet_test

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
)

// TestPagePacketEncodeUsesZeroOfferIDFooter verifies catalog.page uses the reference zero footer offer id.
func TestPagePacketEncodeUsesZeroOfferIDFooter(t *testing.T) {
	body, err := packet.PagePacket{PageID: 5, CatalogType: "NORMAL", LayoutCode: "vip_buy"}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	r := codec.NewReader(body)
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("expected page id: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("expected catalog type: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("expected layout code: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("expected image count: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("expected text count: %v", err)
	}
	offerCount, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("expected offer count: %v", err)
	}
	if offerCount != 0 {
		t.Fatalf("expected zero offers, got %d", offerCount)
	}
	footerOfferID, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("expected footer offer id: %v", err)
	}
	if footerOfferID != 0 {
		t.Fatalf("expected footer offer id 0, got %d", footerOfferID)
	}
	acceptSeasonCurrency, err := r.ReadBool()
	if err != nil {
		t.Fatalf("expected seasonal currency flag: %v", err)
	}
	if acceptSeasonCurrency {
		t.Fatal("expected seasonal currency flag false")
	}
	frontPageCount, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("expected front page item count: %v", err)
	}
	if frontPageCount != 0 {
		t.Fatalf("expected zero front page items, got %d", frontPageCount)
	}
}
