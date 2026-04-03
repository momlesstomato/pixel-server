package packet

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// FurnitureFloorItem stores one floor item entry for the room floor list.
type FurnitureFloorItem struct {
	// ItemID stores the item instance identifier.
	ItemID int
	// SpriteID stores the client-side sprite identifier.
	SpriteID int
	// X stores the tile horizontal coordinate.
	X int
	// Y stores the tile vertical coordinate.
	Y int
	// Dir stores the rotation direction (0-7).
	Dir int
	// Z stores the tile height offset.
	Z float64
	// ExtraData stores item-specific state data.
	ExtraData string
	// UserID stores the owner user identifier.
	UserID int
}

// FurnitureFloorComposer encodes the full room floor item list (s2c 1778).
// Wire format matches FurnitureFloorParser in Nitro: owners map then item list.
type FurnitureFloorComposer struct {
	// Items stores all placed floor items in the room.
	Items []FurnitureFloorItem
	// Owners maps user identifier to display name for all item owners.
	Owners map[int]string
}

// PacketID returns the wire protocol packet identifier.
func (p FurnitureFloorComposer) PacketID() uint16 { return FurnitureFloorPacketID }

// Encode serializes the floor list per Nitro FurnitureFloorParser.
func (p FurnitureFloorComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Owners)))
	for uid, name := range p.Owners {
		w.WriteInt32(int32(uid))
		if err := w.WriteString(name); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Items)))
	for _, item := range p.Items {
		if err := encodeFloorItem(w, item); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeFloorItem writes one item per FurnitureFloorDataParser field order.
func encodeFloorItem(w *codec.Writer, item FurnitureFloorItem) error {
	w.WriteInt32(int32(item.ItemID))
	w.WriteInt32(int32(item.SpriteID))
	w.WriteInt32(int32(item.X))
	w.WriteInt32(int32(item.Y))
	w.WriteInt32(int32(item.Dir))
	if err := w.WriteString(fmt.Sprintf("%.2f", item.Z)); err != nil {
		return err
	}
	if err := w.WriteString("0.00"); err != nil {
		return err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	extraData := item.ExtraData
	if extraData == "" {
		extraData = "0"
	}
	if err := w.WriteString(extraData); err != nil {
		return err
	}
	w.WriteInt32(-1)
	w.WriteInt32(0)
	w.WriteInt32(int32(item.UserID))
	return nil
}
