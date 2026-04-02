package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// FurniListNotificationType defines item type codes used in FurniListNotificationPacket.
// Type 1 = floor furniture, type 2 = wall furniture.
type FurniListNotificationType int32

const (
	// FurniListFloor indicates a floor furniture item type.
	FurniListFloor FurniListNotificationType = 1
	// FurniListWall indicates a wall furniture item type.
	FurniListWall FurniListNotificationType = 2
)

// FurniListNotificationPacket encodes a server-side notification for one newly
// delivered inventory item. Wire format: int32(1) int32(type) int32(1) int32(itemID).
type FurniListNotificationPacket struct {
	// ItemID stores the item instance identifier.
	ItemID int
	// ItemType stores the furniture category (floor or wall).
	ItemType FurniListNotificationType
}

// PacketID returns the wire protocol packet identifier.
func (p FurniListNotificationPacket) PacketID() uint16 { return FurniListNotificationPacketID }

// Encode serializes the notification per the Habbo wire protocol.
func (p FurniListNotificationPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(1)
	w.WriteInt32(int32(p.ItemType))
	w.WriteInt32(1)
	w.WriteInt32(int32(p.ItemID))
	return w.Bytes(), nil
}
