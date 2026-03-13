package ignore

import "github.com/momlesstomato/pixel-server/core/codec"

// UserIgnorePacket defines user.ignore and user.unignore packet payload.
type UserIgnorePacket struct {
	// Username stores target username payload.
	Username string
}

// PacketID returns protocol packet identifier.
func (packet UserIgnorePacket) PacketID() uint16 { return UserIgnorePacketID }

// Encode serializes packet body payload.
func (packet UserIgnorePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Username); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserIgnorePacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadString()
	if err == nil {
		packet.Username = value
	}
	return err
}
