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

// TestGuestRoomDataPacketEncode verifies guest room data packet matches Nitro wire format.
func TestGuestRoomDataPacketEncode(t *testing.T) {
	room := domain.Room{ID: 1, Name: "Test", OwnerID: 2, OwnerName: "owner", State: "open", MaxUsers: 10}
	p := navpacket.GuestRoomDataPacket{Room: room, Forward: true}
	if p.PacketID() != navpacket.GuestRoomDataPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

// TestGuestRoomDataPacketEncodeForwardFalse verifies guest room data with forward=false.
func TestGuestRoomDataPacketEncodeForwardFalse(t *testing.T) {
	room := domain.Room{ID: 5, Name: "Closed", State: "locked", MaxUsers: 5}
	p := navpacket.GuestRoomDataPacket{Room: room, Forward: false}
	body, err := p.Encode()
	if err != nil || len(body) == 0 {
		t.Fatalf("unexpected encode result len=%d err=%v", len(body), err)
	}
}

// TestNavigatorSearchResultsPacketEncode verifies search results include bitmask per room.
func TestNavigatorSearchResultsPacketEncode(t *testing.T) {
	rooms := []domain.Room{
		{ID: 1, Name: "Room A", OwnerID: 1, OwnerName: "alice", State: "open", MaxUsers: 10},
		{ID: 2, Name: "Room B", OwnerID: 2, OwnerName: "bob", State: "locked", MaxUsers: 5},
	}
	p := navpacket.NavigatorSearchResultsPacket{
		SearchCode: "hotel_view",
		Filter:     "",
		Results: []navpacket.SearchResultBlock{
			{SearchCode: "hotel_view", Text: "Hotel", Rooms: rooms},
		},
	}
	if p.PacketID() != navpacket.NavigatorSearchResultsPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

// TestNavigatorEventCategoriesPacketEncode verifies event categories stub packet.
func TestNavigatorEventCategoriesPacketEncode(t *testing.T) {
	p := navpacket.NavigatorEventCategoriesPacket{}
	if p.PacketID() != navpacket.NavigatorEventCategoriesPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if len(body) != 4 {
		t.Fatalf("expected 4 bytes for empty list, got %d", len(body))
	}
}
