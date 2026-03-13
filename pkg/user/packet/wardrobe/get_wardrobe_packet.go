package wardrobe

import "github.com/momlesstomato/pixel-server/core/codec"

// UserGetWardrobePacket defines user.get_wardrobe packet payload.
type UserGetWardrobePacket struct {
	// PageID stores requested wardrobe page index.
	PageID int32
}

// PacketID returns protocol packet identifier.
func (packet UserGetWardrobePacket) PacketID() uint16 { return UserGetWardrobePacketID }

// Encode serializes packet body payload.
func (packet UserGetWardrobePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.PageID)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserGetWardrobePacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		packet.PageID = value
	}
	return err
}
