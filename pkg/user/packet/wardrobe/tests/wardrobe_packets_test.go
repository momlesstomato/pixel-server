package tests

import (
	"testing"

	packetwardrobe "github.com/momlesstomato/pixel-server/pkg/user/packet/wardrobe"
)

// TestUserGetWardrobePacketEncodeDecode verifies get wardrobe packet serialization.
func TestUserGetWardrobePacketEncodeDecode(t *testing.T) {
	packet := packetwardrobe.UserGetWardrobePacket{PageID: 2}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetwardrobe.UserGetWardrobePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.PageID != 2 {
		t.Fatalf("unexpected page id %d", decoded.PageID)
	}
}

// TestUserSaveWardrobeOutfitPacketEncodeDecode verifies save wardrobe packet serialization.
func TestUserSaveWardrobeOutfitPacketEncodeDecode(t *testing.T) {
	packet := packetwardrobe.UserSaveWardrobeOutfitPacket{SlotID: 3, Figure: "hr-100", Gender: "F"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetwardrobe.UserSaveWardrobeOutfitPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.SlotID != 3 || decoded.Figure != "hr-100" || decoded.Gender != "F" {
		t.Fatalf("unexpected decoded packet %+v", decoded)
	}
}

// TestUserWardrobePagePacketEncode verifies wardrobe page packet serialization.
func TestUserWardrobePagePacketEncode(t *testing.T) {
	packet := packetwardrobe.UserWardrobePagePacket{
		PageID: 1,
		Slots:  []packetwardrobe.SlotEntry{{SlotID: 1, Figure: "hr-1", Gender: "M"}, {SlotID: 2, Figure: "hr-2", Gender: "F"}},
	}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty payload")
	}
}
