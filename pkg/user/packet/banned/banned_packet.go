package banned

import "github.com/momlesstomato/pixel-server/core/codec"

// UserBannedPacketID defines packet identifier for user.banned.
const UserBannedPacketID uint16 = 1683

// UserBannedPacket defines user.banned payload.
type UserBannedPacket struct{ Message string }

// PacketID returns protocol packet identifier.
func (packet UserBannedPacket) PacketID() uint16 { return UserBannedPacketID }

// Encode serializes packet body payload.
func (packet UserBannedPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Message); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserBannedPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadString()
	if err == nil {
		packet.Message = value
	}
	return err
}
