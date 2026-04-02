package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// furnitureCategoryID is the unseen-items category code for furniture.
const furnitureCategoryID int32 = 1

// FurniListNotificationPacket encodes an unseen-items notification for one newly
// delivered furniture item. Wire format matches UnseenItemsParser:
// int32(1 category) int32(category=1) int32(1 item) int32(itemID).
type FurniListNotificationPacket struct {
	// ItemID stores the item instance identifier.
	ItemID int
}

// PacketID returns the wire protocol packet identifier.
func (p FurniListNotificationPacket) PacketID() uint16 { return FurniListNotificationPacketID }

// Encode serializes the notification per the Habbo unseen-items wire protocol.
func (p FurniListNotificationPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(1)
	w.WriteInt32(furnitureCategoryID)
	w.WriteInt32(1)
	w.WriteInt32(int32(p.ItemID))
	return w.Bytes(), nil
}
