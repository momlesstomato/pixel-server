package packet

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// FlatCategoriesPacket defines navigator.flat_categories (s2c 1562) payload.
type FlatCategoriesPacket struct {
	// Categories stores navigable room categories.
	Categories []domain.Category
}

// PacketID returns the wire protocol packet identifier.
func (p FlatCategoriesPacket) PacketID() uint16 { return FlatCategoriesPacketID }

// Encode serializes flat categories into packet body.
func (p FlatCategoriesPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Categories)))
	for _, cat := range p.Categories {
		w.WriteInt32(int32(cat.ID))
		if err := w.WriteString(cat.Caption); err != nil {
			return nil, err
		}
		w.WriteBool(cat.Visible)
		w.WriteBool(false)
		if err := w.WriteString(cat.CategoryType); err != nil {
			return nil, err
		}
		if err := w.WriteString(""); err != nil {
			return nil, err
		}
		w.WriteBool(false)
	}
	return w.Bytes(), nil
}

// CanCreateRoomResponsePacket defines navigator.can_create_room (s2c 378) payload.
type CanCreateRoomResponsePacket struct {
	// ResultCode stores whether room creation is allowed (0=ok, 1=limit).
	ResultCode int32
	// MaxRooms stores the room limit for the user.
	MaxRooms int32
}

// PacketID returns the wire protocol packet identifier.
func (p CanCreateRoomResponsePacket) PacketID() uint16 { return CanCreateRoomResponsePacketID }

// Encode serializes can create room result into packet body.
func (p CanCreateRoomResponsePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ResultCode)
	w.WriteInt32(p.MaxRooms)
	return w.Bytes(), nil
}

// RoomCreatedPacket defines navigator.room_created (s2c 1304) payload.
type RoomCreatedPacket struct {
	// RoomID stores the newly created room identifier.
	RoomID int32
	// Name stores the room display name.
	Name string
}

// PacketID returns the wire protocol packet identifier.
func (p RoomCreatedPacket) PacketID() uint16 { return RoomCreatedPacketID }

// Encode serializes room created response into packet body.
func (p RoomCreatedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	if err := w.WriteString(p.Name); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// FavouriteChangedPacket defines navigator.favourite_changed (s2c 2524) payload.
type FavouriteChangedPacket struct {
	// RoomID stores the affected room identifier.
	RoomID int32
	// Added stores whether the room was added (true) or removed (false).
	Added bool
}

// PacketID returns the wire protocol packet identifier.
func (p FavouriteChangedPacket) PacketID() uint16 { return FavouriteChangedPacketID }

// Encode serializes favourite changed notification into packet body.
func (p FavouriteChangedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteBool(p.Added)
	return w.Bytes(), nil
}

// FavouritesListPacket defines navigator.favourites_list (s2c 151) payload.
type FavouritesListPacket struct {
	// MaxFavourites stores the per-user limit.
	MaxFavourites int32
	// RoomIDs stores the favourite room identifiers.
	RoomIDs []int32
}

// PacketID returns the wire protocol packet identifier.
func (p FavouritesListPacket) PacketID() uint16 { return FavouritesListPacketID }

// Encode serializes favourites list into packet body.
func (p FavouritesListPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.MaxFavourites)
	w.WriteInt32(int32(len(p.RoomIDs)))
	for _, id := range p.RoomIDs {
		w.WriteInt32(id)
	}
	return w.Bytes(), nil
}
