package packet

import (
	"strings"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// limitedFlag is OR-ed into the stuffData type int for limited-edition items.
const limitedFlag int32 = 0xFF00

// FurniListItem holds one encoded inventory item entry for FurniListPacket.
type FurniListItem struct {
	// ID stores the item instance identifier.
	ID int
	// ItemType stores the furniture category marker (floor=s, wall=i).
	ItemType domain.ItemType
	// SpriteID stores the client-side sprite identifier.
	SpriteID int
	// ExtraData stores item-specific custom data.
	ExtraData string
	// LimitedNumber stores limited-edition serial number (0 = not limited).
	LimitedNumber int
	// LimitedTotal stores limited-edition total print run (0 = not limited).
	LimitedTotal int
	// AllowRecycle stores whether the item can be recycled.
	AllowRecycle bool
	// AllowTrade stores whether the item can be traded.
	AllowTrade bool
	// AllowInventoryStack stores whether inventory stacking is allowed.
	AllowInventoryStack bool
	// AllowMarketplaceSell stores whether marketplace listing is allowed.
	AllowMarketplaceSell bool
}

// FurniListPacket encodes a paginated furniture inventory response.
// Wire format matches FurnitureListParser: totalFragments, fragmentIndex, count, items.
type FurniListPacket struct {
	// TotalFragments stores the total number of pages in the response.
	TotalFragments int
	// FragmentIndex stores the current page index (zero-based).
	FragmentIndex int
	// Items stores the inventory entries to encode.
	Items []FurniListItem
}

// PacketID returns the wire protocol packet identifier.
func (p FurniListPacket) PacketID() uint16 { return FurniListPacketID }

// Encode serializes the furniture list per the FurnitureListParser wire protocol.
func (p FurniListPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.TotalFragments))
	w.WriteInt32(int32(p.FragmentIndex))
	w.WriteInt32(int32(len(p.Items)))
	for _, item := range p.Items {
		if err := encodeItem(w, item); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeItem writes one FurnitureListItemParser-compatible entry.
func encodeItem(w *codec.Writer, item FurniListItem) error {
	typeStr := strings.ToUpper(string(item.ItemType))
	isFloor := item.ItemType == domain.ItemTypeFloor

	w.WriteInt32(int32(item.ID))
	if err := w.WriteString(typeStr); err != nil {
		return err
	}
	w.WriteInt32(int32(item.ID))
	w.WriteInt32(int32(item.SpriteID))
	w.WriteInt32(0)

	stuffType := int32(0)
	if item.LimitedTotal > 0 {
		stuffType |= limitedFlag
	}
	w.WriteInt32(stuffType)
	if err := w.WriteString(item.ExtraData); err != nil {
		return err
	}
	if item.LimitedTotal > 0 {
		w.WriteInt32(int32(item.LimitedNumber))
		w.WriteInt32(int32(item.LimitedTotal))
	}

	w.WriteBool(item.AllowRecycle)
	w.WriteBool(item.AllowTrade)
	w.WriteBool(item.AllowInventoryStack)
	w.WriteBool(item.AllowMarketplaceSell)
	w.WriteInt32(-1)
	w.WriteBool(false)
	w.WriteInt32(-1)
	if isFloor {
		if err := w.WriteString(""); err != nil {
			return err
		}
		w.WriteInt32(0)
	}
	return nil
}

