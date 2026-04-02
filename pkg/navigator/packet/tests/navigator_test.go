package tests

import (
	"testing"

	navpacket "github.com/momlesstomato/pixel-server/pkg/navigator/packet"
)

// TestNavigatorMetaDataPacketEncode verifies metadata packet serialization.
func TestNavigatorMetaDataPacketEncode(t *testing.T) {
	p := navpacket.NavigatorMetaDataPacket{TopLevelContexts: []string{"official", "my"}}
	if p.PacketID() != navpacket.NavigatorMetaDataPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestNavigatorCollapsedPacketEncode verifies collapsed packet serialization.
func TestNavigatorCollapsedPacketEncode(t *testing.T) {
	p := navpacket.NavigatorCollapsedPacket{Categories: []string{"hotel_view"}}
	if p.PacketID() != navpacket.NavigatorCollapsedPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestNavigatorSettingsPacketEncode verifies settings packet serialization.
func TestNavigatorSettingsPacketEncode(t *testing.T) {
	p := navpacket.NavigatorSettingsPacket{X: 10, Y: 20, Width: 425, Height: 535}
	if p.PacketID() != navpacket.NavigatorSettingsPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestNavigatorSavedSearchesPacketEncode verifies saved searches packet serialization.
func TestNavigatorSavedSearchesPacketEncode(t *testing.T) {
	p := navpacket.NavigatorSavedSearchesPacket{
		Searches: []navpacket.SavedSearchEntry{{ID: 1, SearchCode: "hotel_view", Filter: ""}},
	}
	if p.PacketID() != navpacket.NavigatorSavedSearchesPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestNavigatorMetaDataPacketEncodeEmpty verifies empty metadata packet.
func TestNavigatorMetaDataPacketEncodeEmpty(t *testing.T) {
	p := navpacket.NavigatorMetaDataPacket{}
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error %v", err)
	}
	if len(body) != 4 {
		t.Fatalf("expected 4 bytes for empty list, got %d", len(body))
	}
}
