package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserNoobnessLevelPacketID defines packet identifier for user.noobness_level.
const UserNoobnessLevelPacketID uint16 = 3738

// UserBannedPacketID defines packet identifier for user.banned.
const UserBannedPacketID uint16 = 1683

// UserGetInfoPacketID defines packet identifier for user.get_info.
const UserGetInfoPacketID uint16 = 357

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

// UserGetInfoPacket defines user.get_info packet payload.
type UserGetInfoPacket struct{}

// PacketID returns protocol packet identifier.
func (packet UserGetInfoPacket) PacketID() uint16 { return UserGetInfoPacketID }

// Encode serializes packet body payload.
func (packet UserGetInfoPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (packet *UserGetInfoPacket) Decode(payload []byte) error { _ = payload; return nil }
