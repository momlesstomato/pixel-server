package packet

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// FloorItemAddPacket encodes a single placed floor item event (s2c 1534).
// Wire format matches FurnitureFloorDataParser + FurnitureFloorAddParser in Nitro.
type FloorItemAddPacket struct {
	// ItemID stores the unique item instance identifier.
	ItemID int
	// SpriteID stores the client-side sprite identifier.
	SpriteID int
	// X stores the tile horizontal coordinate.
	X int
	// Y stores the tile vertical coordinate.
	Y int
	// Dir stores the rotation direction (0-7; client converts to degrees).
	Dir int
	// Z stores the tile height offset as a formatted string.
	Z float64
	// StackHeight stores the item height used by the client for stacking and seating.
	StackHeight float64
	// ExtraData stores item state data.
	ExtraData string
	// UserID stores the placing user identifier.
	UserID int
	// Username stores the placing user display name.
	Username string
}

// PacketID returns the wire protocol packet identifier.
func (p FloorItemAddPacket) PacketID() uint16 { return FloorItemAddPacketID }

// Encode serializes the floor item add packet per Nitro FurnitureFloorAddParser.
func (p FloorItemAddPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.ItemID))
	w.WriteInt32(int32(p.SpriteID))
	w.WriteInt32(int32(p.X))
	w.WriteInt32(int32(p.Y))
	w.WriteInt32(int32(p.Dir))
	if err := w.WriteString(fmt.Sprintf("%.2f", p.Z)); err != nil {
		return nil, err
	}
	if err := w.WriteString(fmt.Sprintf("%.2f", p.StackHeight)); err != nil {
		return nil, err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	extraData := p.ExtraData
	if extraData == "" {
		extraData = "0"
	}
	if err := w.WriteString(extraData); err != nil {
		return nil, err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	w.WriteInt32(int32(p.UserID))
	if err := w.WriteString(p.Username); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// FloorItemUpdatePacket notifies clients that a placed floor item moved or rotated (s2c 3776).
// Wire format matches FurnitureFloorUpdateParser (FurnitureFloorDataParser) in Nitro — identical to
// FloorItemAddPacket but without the trailing username field.
type FloorItemUpdatePacket struct {
	// ItemID stores the unique item instance identifier.
	ItemID int
	// SpriteID stores the client-side sprite identifier.
	SpriteID int
	// X stores the tile horizontal coordinate.
	X int
	// Y stores the tile vertical coordinate.
	Y int
	// Dir stores the rotation direction (0-7; client converts to degrees).
	Dir int
	// Z stores the tile height offset.
	Z float64
	// StackHeight stores the item height used by the client for stacking and seating.
	StackHeight float64
	// ExtraData stores item state data.
	ExtraData string
	// UserID stores the owning user identifier.
	UserID int
}

// PacketID returns the wire protocol packet identifier.
func (p FloorItemUpdatePacket) PacketID() uint16 { return FloorItemUpdatePacketID }

// Encode serializes the floor item update packet per Nitro FurnitureFloorUpdateParser.
func (p FloorItemUpdatePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.ItemID))
	w.WriteInt32(int32(p.SpriteID))
	w.WriteInt32(int32(p.X))
	w.WriteInt32(int32(p.Y))
	w.WriteInt32(int32(p.Dir))
	if err := w.WriteString(fmt.Sprintf("%.2f", p.Z)); err != nil {
		return nil, err
	}
	if err := w.WriteString(fmt.Sprintf("%.2f", p.StackHeight)); err != nil {
		return nil, err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	extraData := p.ExtraData
	if extraData == "" {
		extraData = "0"
	}
	if err := w.WriteString(extraData); err != nil {
		return nil, err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	w.WriteInt32(int32(p.UserID))
	return w.Bytes(), nil
}

// FloorItemRemovePacket notifies clients that one floor item was removed from the room (s2c 2703).
// Wire format matches FurnitureFloorRemoveParser in Nitro.
type FloorItemRemovePacket struct {
	// ItemID stores the removed item instance identifier.
	ItemID int
	// UserID stores the identifier of the user who removed the item.
	UserID int
}

// PacketID returns the wire protocol packet identifier.
func (p FloorItemRemovePacket) PacketID() uint16 { return FloorItemRemovePacketID }

// Encode serializes the floor item remove notification.
func (p FloorItemRemovePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	itemStr := fmt.Sprintf("%d", p.ItemID)
	if err := w.WriteString(itemStr); err != nil {
		return nil, err
	}
	w.WriteBool(false)
	w.WriteInt32(int32(p.UserID))
	w.WriteInt32(0)
	return w.Bytes(), nil
}
