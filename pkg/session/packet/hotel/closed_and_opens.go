package hotel

import "github.com/momlesstomato/pixel-server/core/codec"

// ClosedAndOpensPacketID identifies hotel.closed_and_opens packet.
const ClosedAndOpensPacketID uint16 = 3728

// ClosedAndOpensPacket carries reopen schedule for closed hotels.
type ClosedAndOpensPacket struct {
	// OpenHour stores scheduled reopen hour in UTC.
	OpenHour int32
	// OpenMinute stores scheduled reopen minute in UTC.
	OpenMinute int32
}

// PacketID returns protocol packet identifier.
func (packet ClosedAndOpensPacket) PacketID() uint16 { return ClosedAndOpensPacketID }

// Decode parses packet body into struct fields.
func (packet *ClosedAndOpensPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	hour, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	minute, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.OpenHour = hour
	packet.OpenMinute = minute
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClosedAndOpensPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.OpenHour)
	writer.WriteInt32(packet.OpenMinute)
	return writer.Bytes(), nil
}
