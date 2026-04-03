package packet

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// InventoryRemovePacket encodes a single item removal from the inventory (s2c 159).
// Wire format matches FurnitureListRemovedParser in Nitro: a single int32 item ID.
type InventoryRemovePacket struct {
	// ItemID stores the item instance identifier to remove from inventory.
	ItemID int
}

// PacketID returns the wire protocol packet identifier.
func (p InventoryRemovePacket) PacketID() uint16 { return InventoryRemovePacketID }

// Encode serializes the inventory remove packet.
func (p InventoryRemovePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.ItemID))
	return w.Bytes(), nil
}

// InventoryAddPacket encodes a single floor item addition to the inventory (s2c 104).
// Wire format matches FurnitureListAddOrUpdateParser + FurnitureListItemParser in Nitro.
type InventoryAddPacket struct {
	// ItemID stores the item instance identifier.
	ItemID int
	// SpriteID stores the client-side sprite identifier.
	SpriteID int
	// ExtraData stores item-specific state data.
	ExtraData string
	// AllowRecycle stores whether the item can be recycled.
	AllowRecycle bool
	// AllowTrade stores whether the item can be traded.
	AllowTrade bool
	// AllowInventoryStack stores whether the item can stack in inventory.
	AllowInventoryStack bool
	// AllowMarketplaceSell stores whether the item can be sold on marketplace.
	AllowMarketplaceSell bool
}

// PacketID returns the wire protocol packet identifier.
func (p InventoryAddPacket) PacketID() uint16 { return InventoryAddPacketID }

// Encode serializes the inventory add packet per Nitro FurnitureListItemParser.
func (p InventoryAddPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.ItemID))
	if err := w.WriteString("S"); err != nil {
		return nil, err
	}
	w.WriteInt32(int32(p.ItemID))
	w.WriteInt32(int32(p.SpriteID))
	w.WriteInt32(1)
	w.WriteInt32(0)
	extraData := p.ExtraData
	if extraData == "" {
		extraData = "0"
	}
	if err := w.WriteString(fmt.Sprintf("%s", extraData)); err != nil {
		return nil, err
	}
	w.WriteBool(p.AllowRecycle)
	w.WriteBool(p.AllowTrade)
	w.WriteBool(p.AllowInventoryStack)
	w.WriteBool(p.AllowMarketplaceSell)
	w.WriteInt32(-1)
	w.WriteBool(false)
	w.WriteInt32(0)
	if err := w.WriteString(""); err != nil {
		return nil, err
	}
	w.WriteInt32(0)
	return w.Bytes(), nil
}
