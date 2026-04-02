package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	navpacket "github.com/momlesstomato/pixel-server/pkg/navigator/packet"
)

// TestFlatCategoriesPacketEncode verifies flat categories packet serialization.
func TestFlatCategoriesPacketEncode(t *testing.T) {
	p := navpacket.FlatCategoriesPacket{Categories: []domain.Category{
		{ID: 1, Caption: "Public", Visible: true, CategoryType: "public"},
	}}
	if p.PacketID() != navpacket.FlatCategoriesPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestCanCreateRoomResponsePacketEncode verifies can create room response packet.
func TestCanCreateRoomResponsePacketEncode(t *testing.T) {
	p := navpacket.CanCreateRoomResponsePacket{ResultCode: 0, MaxRooms: 25}
	if p.PacketID() != navpacket.CanCreateRoomResponsePacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) != 8 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestRoomCreatedPacketEncode verifies room created packet serialization.
func TestRoomCreatedPacketEncode(t *testing.T) {
	p := navpacket.RoomCreatedPacket{RoomID: 42, Name: "My Room"}
	if p.PacketID() != navpacket.RoomCreatedPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestFavouriteChangedPacketEncode verifies favourite changed packet serialization.
func TestFavouriteChangedPacketEncode(t *testing.T) {
	p := navpacket.FavouriteChangedPacket{RoomID: 1, Added: true}
	if p.PacketID() != navpacket.FavouriteChangedPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestFavouritesListPacketEncode verifies favourites list packet serialization.
func TestFavouritesListPacketEncode(t *testing.T) {
	p := navpacket.FavouritesListPacket{MaxFavourites: 30, RoomIDs: []int32{1, 2, 3}}
	if p.PacketID() != navpacket.FavouritesListPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}
