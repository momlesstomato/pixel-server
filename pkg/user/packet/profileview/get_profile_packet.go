package profileview

import "github.com/momlesstomato/pixel-server/core/codec"

// UserGetProfilePacket defines user.get_profile packet payload.
type UserGetProfilePacket struct {
	// UserID stores target user identifier.
	UserID int32
	// OpenProfileWindow stores profile view marker.
	OpenProfileWindow bool
}

// PacketID returns protocol packet identifier.
func (packet UserGetProfilePacket) PacketID() uint16 { return UserGetProfilePacketID }

// Encode serializes packet body payload.
func (packet UserGetProfilePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	writer.WriteBool(packet.OpenProfileWindow)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserGetProfilePacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	value, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.UserID = value
	flag, err := reader.ReadBool()
	if err == nil {
		packet.OpenProfileWindow = flag
	}
	return err
}
