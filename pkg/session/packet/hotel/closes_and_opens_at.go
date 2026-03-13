package hotel

import "github.com/momlesstomato/pixel-server/core/codec"

// ClosesAndOpensAtPacketID identifies hotel.closes_and_opens_at packet.
const ClosesAndOpensAtPacketID uint16 = 2771

// ClosesAndOpensAtPacket carries close and reopen metadata.
type ClosesAndOpensAtPacket struct {
	// OpenHour stores scheduled reopen hour in UTC.
	OpenHour int32
	// OpenMinute stores scheduled reopen minute in UTC.
	OpenMinute int32
	// UserThrownOutAtClose stores whether connected users are kicked at close.
	UserThrownOutAtClose bool
}

// PacketID returns protocol packet identifier.
func (packet ClosesAndOpensAtPacket) PacketID() uint16 { return ClosesAndOpensAtPacketID }

// Decode parses packet body into struct fields.
func (packet *ClosesAndOpensAtPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	hour, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	minute, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	thrownOut, err := reader.ReadBool()
	if err != nil {
		return err
	}
	packet.OpenHour = hour
	packet.OpenMinute = minute
	packet.UserThrownOutAtClose = thrownOut
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClosesAndOpensAtPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.OpenHour)
	writer.WriteInt32(packet.OpenMinute)
	writer.WriteBool(packet.UserThrownOutAtClose)
	return writer.Bytes(), nil
}
