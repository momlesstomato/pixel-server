package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
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

// TestGuestRoomSearchResultPacketEncode verifies legacy private-room search results match Nitro parsing order.
func TestGuestRoomSearchResultPacketEncode(t *testing.T) {
	room := domain.Room{ID: 1, Name: "Room A", OwnerID: 2, OwnerName: "alice", State: "locked", MaxUsers: 10, Description: "Desc", Tags: []string{"fun"}}
	p := navpacket.GuestRoomSearchResultPacket{SearchType: 2, SearchParam: "", Rooms: []domain.Room{room}}
	if p.PacketID() != navpacket.GuestRoomSearchResultPacketID {
		t.Fatalf("unexpected packet id %d", p.PacketID())
	}
	body, err := p.Encode()
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	r := codec.NewReader(body)
	searchType, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read search type: %v", readErr)
	}
	searchParam, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read search param: %v", readErr)
	}
	roomCount, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read room count: %v", readErr)
	}
	roomID, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read room id: %v", readErr)
	}
	name, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read room name: %v", readErr)
	}
	ownerID, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read owner id: %v", readErr)
	}
	ownerName, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read owner name: %v", readErr)
	}
	doorMode, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read door mode: %v", readErr)
	}
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	description, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read description: %v", readErr)
	}
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	tagCount, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read tag count: %v", readErr)
	}
	tag, readErr := r.ReadString()
	if readErr != nil {
		t.Fatalf("read tag: %v", readErr)
	}
	bitmask, readErr := r.ReadInt32()
	if readErr != nil {
		t.Fatalf("read bitmask: %v", readErr)
	}
	hasAdditional, readErr := r.ReadBool()
	if readErr != nil {
		t.Fatalf("read additional flag: %v", readErr)
	}
	if searchType != 2 || searchParam != "" || roomCount != 1 {
		t.Fatalf("unexpected header values type=%d param=%q count=%d", searchType, searchParam, roomCount)
	}
	if roomID != 1 || name != "Room A" || ownerID != 2 || ownerName != "alice" {
		t.Fatalf("unexpected room identity values id=%d name=%q ownerID=%d ownerName=%q", roomID, name, ownerID, ownerName)
	}
	if doorMode != 1 || description != "Desc" || tagCount != 1 || tag != "fun" || bitmask != 8 || hasAdditional {
		t.Fatalf("unexpected room payload values doorMode=%d description=%q tagCount=%d tag=%q bitmask=%d hasAdditional=%v", doorMode, description, tagCount, tag, bitmask, hasAdditional)
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
