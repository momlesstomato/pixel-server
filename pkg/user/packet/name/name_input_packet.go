package name

import "github.com/momlesstomato/pixel-server/core/codec"

// UserNameInputPacket defines user.check_name and user.change_name packet payload.
type UserNameInputPacket struct {
	// Name stores target username payload.
	Name string
}

// PacketID returns protocol packet identifier.
func (packet UserNameInputPacket) PacketID() uint16 { return UserCheckNamePacketID }

// Encode serializes packet body payload.
func (packet UserNameInputPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Name); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserNameInputPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadString()
	if err == nil {
		packet.Name = value
	}
	return err
}
