package packet_test

import (
	"encoding/binary"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
)

// TestGiftWrappingConfigPacketID verifies the packet uses the correct protocol identifier.
func TestGiftWrappingConfigPacketID(t *testing.T) {
	p := packet.DefaultGiftWrappingConfig()
	if p.PacketID() != packet.GiftWrappingConfigResponsePacketID {
		t.Fatalf("expected GiftWrappingConfigResponsePacketID, got %d", p.PacketID())
	}
}

// TestGiftWrappingConfigPacketEncodeEnabled verifies enabled flag is encoded as first byte.
func TestGiftWrappingConfigPacketEncodeEnabled(t *testing.T) {
	p := packet.DefaultGiftWrappingConfig()
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	if body[0] != 1 {
		t.Fatalf("expected enabled=true (1), got %d", body[0])
	}
}

// TestGiftWrappingConfigPacketEncodeWrapperCount verifies wrapper count is encoded after enabled+price.
func TestGiftWrappingConfigPacketEncodeWrapperCount(t *testing.T) {
	p := packet.DefaultGiftWrappingConfig()
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	wrapperCount := int32(binary.BigEndian.Uint32(body[5:9]))
	if int(wrapperCount) != len(p.WrapperIDs) {
		t.Fatalf("expected %d wrappers, got %d", len(p.WrapperIDs), wrapperCount)
	}
}

// TestPurchaseOKPacketID verifies the packet uses the purchase_ok protocol identifier.
func TestPurchaseOKPacketID(t *testing.T) {
	p := packet.PurchaseOKPacket{}
	if p.PacketID() != packet.PurchaseOKPacketID {
		t.Fatalf("expected PurchaseOKPacketID, got %d", p.PacketID())
	}
}

// TestPurchaseErrorPacketID verifies the packet uses the purchase_error protocol identifier.
func TestPurchaseErrorPacketID(t *testing.T) {
	p := packet.PurchaseErrorPacket{Code: packet.PurchaseErrorInsufficientCredits}
	if p.PacketID() != packet.PurchaseErrorPacketID {
		t.Fatalf("expected PurchaseErrorPacketID, got %d", p.PacketID())
	}
}

// TestPurchaseErrorPacketEncodeCode verifies error code is encoded as int32.
func TestPurchaseErrorPacketEncodeCode(t *testing.T) {
	body, err := packet.PurchaseErrorPacket{Code: packet.PurchaseErrorInsufficientCredits}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	code := int32(binary.BigEndian.Uint32(body[:4]))
	if code != int32(packet.PurchaseErrorInsufficientCredits) {
		t.Fatalf("expected code %d, got %d", packet.PurchaseErrorInsufficientCredits, code)
	}
}

// TestPurchaseNotAllowedPacketID verifies the packet uses the not_allowed protocol identifier.
func TestPurchaseNotAllowedPacketID(t *testing.T) {
	p := packet.PurchaseNotAllowedPacket{Code: 0}
	if p.PacketID() != packet.PurchaseNotAllowedPacketID {
		t.Fatalf("expected PurchaseNotAllowedPacketID, got %d", p.PacketID())
	}
}

// TestPurchaseOKPacketEncodesOffer verifies the offer entry is serialized into the body.
func TestPurchaseOKPacketEncodesOffer(t *testing.T) {
	offer := packet.OfferEntry{OfferID: 42, LocalizationID: "test_item", PriceCredits: 10}
	body, err := packet.PurchaseOKPacket{Offer: offer}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty encoded body")
	}
	offerID := int32(binary.BigEndian.Uint32(body[:4]))
	if offerID != 42 {
		t.Fatalf("expected offer id 42, got %d", offerID)
	}
}
