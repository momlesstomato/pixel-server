package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserNoobnessLevelPacketID defines packet identifier for user.noobness_level.
const UserNoobnessLevelPacketID uint16 = 3738

// UserNoobnessLevelPacket defines user.noobness_level payload.
type UserNoobnessLevelPacket struct{ NoobnessLevel int32 }

// PacketID returns protocol packet identifier.
func (packet UserNoobnessLevelPacket) PacketID() uint16 { return UserNoobnessLevelPacketID }

// Encode serializes packet body payload.
func (packet UserNoobnessLevelPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.NoobnessLevel)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserNoobnessLevelPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		packet.NoobnessLevel = value
	}
	return err
}
