package ignore

import "github.com/momlesstomato/pixel-server/core/codec"

// UserIgnoreByIDPacket defines user.ignore_id packet payload.
type UserIgnoreByIDPacket struct {
	// UserID stores target user identifier payload.
	UserID int32
}

// PacketID returns protocol packet identifier.
func (packet UserIgnoreByIDPacket) PacketID() uint16 { return UserIgnoreByIDPacketIDDefault }

// Encode serializes packet body payload.
func (packet UserIgnoreByIDPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserIgnoreByIDPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		packet.UserID = value
	}
	return err
}
