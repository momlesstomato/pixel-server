package packet

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// DiceValuePacket broadcasts one dice state change.
type DiceValuePacket struct {
	// ItemID stores the placed item identifier.
	ItemID int32
	// Value stores the visible dice state.
	Value int32
}

// PacketID returns the outbound packet identifier.
func (packet DiceValuePacket) PacketID() uint16 { return DiceValuePacketID }

// Encode serialises the packet body.
func (packet DiceValuePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ItemID)
	writer.WriteInt32(packet.Value)
	return writer.Bytes(), nil
}

// StackHeightUpdatePacket confirms one stack-helper height change.
type StackHeightUpdatePacket struct {
	// ItemID stores the placed item identifier.
	ItemID int32
	// Height stores the updated item height in hundredths.
	Height int32
}

// PacketID returns the outbound packet identifier.
func (packet StackHeightUpdatePacket) PacketID() uint16 { return StackHeightUpdatePacketID }

// Encode serialises the packet body.
func (packet StackHeightUpdatePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ItemID)
	writer.WriteInt32(packet.Height)
	return writer.Bytes(), nil
}

// WallItemAddPacket encodes one placed wall item event (s2c 2187).
type WallItemAddPacket struct {
	// Item stores the placed wall item payload.
	Item FurnitureWallItem
	// Username stores the placing user display name.
	Username string
}

// PacketID returns the outbound packet identifier.
func (packet WallItemAddPacket) PacketID() uint16 { return WallItemAddPacketID }

// Encode serialises the packet body.
func (packet WallItemAddPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := encodeWallItem(writer, packet.Item); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Username); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// WallItemUpdatePacket encodes one wall item change event (s2c 2009).
type WallItemUpdatePacket struct {
	// Item stores the changed wall item payload.
	Item FurnitureWallItem
}

// PacketID returns the outbound packet identifier.
func (packet WallItemUpdatePacket) PacketID() uint16 { return WallItemUpdatePacketID }

// Encode serialises the packet body.
func (packet WallItemUpdatePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := encodeWallItem(writer, packet.Item); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// WallItemRemovePacket encodes one wall item removal event (s2c 3208).
type WallItemRemovePacket struct {
	// ItemID stores the removed wall item identifier.
	ItemID int
	// UserID stores the identifier of the user who removed the item.
	UserID int
}

// PacketID returns the outbound packet identifier.
func (packet WallItemRemovePacket) PacketID() uint16 { return WallItemRemovePacketID }

// Encode serialises the packet body.
func (packet WallItemRemovePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(fmt.Sprintf("%d", packet.ItemID)); err != nil {
		return nil, err
	}
	writer.WriteInt32(int32(packet.UserID))
	return writer.Bytes(), nil
}

// ItemDataUpdatePacket encodes one hidden wall item data update (s2c 2202).
type ItemDataUpdatePacket struct {
	// ItemID stores the wall item identifier.
	ItemID int
	// Data stores the wall item hidden text payload.
	Data string
}

// PacketID returns the outbound packet identifier.
func (packet ItemDataUpdatePacket) PacketID() uint16 { return ItemDataUpdatePacketID }

// Encode serialises the packet body.
func (packet ItemDataUpdatePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(fmt.Sprintf("%d", packet.ItemID)); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Data); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// RollingItem stores one floor item movement inside a rolling packet.
type RollingItem struct {
	// ItemID stores the moved item identifier.
	ItemID int
	// Height stores the starting item height.
	Height float64
	// NextHeight stores the ending item height.
	NextHeight float64
}

// RollingUnit stores one avatar movement inside a rolling packet.
type RollingUnit struct {
	// MovementType stores the Nitro rolling movement marker.
	MovementType int
	// UnitID stores the moved virtual user identifier.
	UnitID int
	// Height stores the starting avatar height.
	Height float64
	// NextHeight stores the ending avatar height.
	NextHeight float64
}

// RoomRollingPacket encodes one room rolling update bundle (s2c 3207).
type RoomRollingPacket struct {
	// SourceX stores the source tile horizontal coordinate.
	SourceX int
	// SourceY stores the source tile vertical coordinate.
	SourceY int
	// TargetX stores the target tile horizontal coordinate.
	TargetX int
	// TargetY stores the target tile vertical coordinate.
	TargetY int
	// Items stores moved floor items.
	Items []RollingItem
	// RollerID stores the triggering roller item identifier.
	RollerID int
	// Unit stores the optional moved avatar payload.
	Unit *RollingUnit
}

// PacketID returns the outbound packet identifier.
func (packet RoomRollingPacket) PacketID() uint16 { return RoomRollingPacketID }

// Encode serialises the packet body.
func (packet RoomRollingPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(int32(packet.SourceX))
	writer.WriteInt32(int32(packet.SourceY))
	writer.WriteInt32(int32(packet.TargetX))
	writer.WriteInt32(int32(packet.TargetY))
	writer.WriteInt32(int32(len(packet.Items)))
	for _, item := range packet.Items {
		writer.WriteInt32(int32(item.ItemID))
		if err := writer.WriteString(fmt.Sprintf("%.2f", item.Height)); err != nil {
			return nil, err
		}
		if err := writer.WriteString(fmt.Sprintf("%.2f", item.NextHeight)); err != nil {
			return nil, err
		}
	}
	writer.WriteInt32(int32(packet.RollerID))
	if packet.Unit == nil {
		return writer.Bytes(), nil
	}
	writer.WriteInt32(int32(packet.Unit.MovementType))
	writer.WriteInt32(int32(packet.Unit.UnitID))
	if err := writer.WriteString(fmt.Sprintf("%.2f", packet.Unit.Height)); err != nil {
		return nil, err
	}
	if err := writer.WriteString(fmt.Sprintf("%.2f", packet.Unit.NextHeight)); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// DimmerPreset stores one room dimmer preset payload.
type DimmerPreset struct {
	// PresetID stores the stable preset slot identifier.
	PresetID int
	// Type stores the room dimmer effect identifier.
	Type int
	// Color stores the RGB color payload with leading hash.
	Color string
	// Brightness stores the preset brightness value.
	Brightness int
}

// DimmerPresetsPacket encodes the room dimmer preset list (s2c 2710).
type DimmerPresetsPacket struct {
	// SelectedPresetID stores the currently selected preset slot.
	SelectedPresetID int
	// Presets stores all available presets.
	Presets []DimmerPreset
}

// PacketID returns the outbound packet identifier.
func (packet DimmerPresetsPacket) PacketID() uint16 { return DimmerPresetsPacketID }

// Encode serialises the packet body.
func (packet DimmerPresetsPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(int32(len(packet.Presets)))
	writer.WriteInt32(int32(packet.SelectedPresetID))
	for _, preset := range packet.Presets {
		writer.WriteInt32(int32(preset.PresetID))
		writer.WriteInt32(int32(preset.Type))
		if err := writer.WriteString(preset.Color); err != nil {
			return nil, err
		}
		writer.WriteInt32(int32(preset.Brightness))
	}
	return writer.Bytes(), nil
}

// GiftOpenedPacket encodes the present-opened response payload (s2c 56).
type GiftOpenedPacket struct {
	// ItemType stores the revealed item type marker.
	ItemType string
	// ClassID stores the revealed sprite or class identifier.
	ClassID int
	// ProductCode stores the revealed product code.
	ProductCode string
	// PlacedItemID stores the revealed placed item identifier.
	PlacedItemID int
	// PlacedItemType stores the placed item type marker.
	PlacedItemType string
	// PlacedInRoom reports whether the gift was revealed in-room.
	PlacedInRoom bool
	// PetFigureString stores the optional revealed pet figure string.
	PetFigureString string
}

// PacketID returns the outbound packet identifier.
func (packet GiftOpenedPacket) PacketID() uint16 { return GiftOpenedPacketID }

// Encode serialises the packet body.
func (packet GiftOpenedPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.ItemType); err != nil {
		return nil, err
	}
	writer.WriteInt32(int32(packet.ClassID))
	if err := writer.WriteString(packet.ProductCode); err != nil {
		return nil, err
	}
	writer.WriteInt32(int32(packet.PlacedItemID))
	if err := writer.WriteString(packet.PlacedItemType); err != nil {
		return nil, err
	}
	writer.WriteBool(packet.PlacedInRoom)
	if err := writer.WriteString(packet.PetFigureString); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
