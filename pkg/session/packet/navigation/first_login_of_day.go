package navigation

import "github.com/momlesstomato/pixel-server/core/codec"

// FirstLoginOfDayPacketID identifies session.first_login_of_day packet.
const FirstLoginOfDayPacketID uint16 = 793

// FirstLoginOfDayPacket carries first-login marker flag.
type FirstLoginOfDayPacket struct {
	// IsFirstLogin stores whether login is first in UTC day.
	IsFirstLogin bool
}

// PacketID returns protocol packet identifier.
func (packet FirstLoginOfDayPacket) PacketID() uint16 { return FirstLoginOfDayPacketID }

// Decode parses packet body into struct fields.
func (packet *FirstLoginOfDayPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	value, err := reader.ReadBool()
	if err != nil {
		return err
	}
	packet.IsFirstLogin = value
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet FirstLoginOfDayPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteBool(packet.IsFirstLogin)
	return writer.Bytes(), nil
}
