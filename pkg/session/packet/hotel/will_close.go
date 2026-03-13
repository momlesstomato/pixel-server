package hotel

import "github.com/momlesstomato/pixel-server/core/codec"

// WillClosePacketID identifies hotel.will_close packet.
const WillClosePacketID uint16 = 1050

// WillClosePacket carries minutes-to-close value.
type WillClosePacket struct {
	// Minutes stores remaining minutes before closure.
	Minutes int32
}

// PacketID returns protocol packet identifier.
func (packet WillClosePacket) PacketID() uint16 { return WillClosePacketID }

// Decode parses packet body into struct fields.
func (packet *WillClosePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	minutes, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.Minutes = minutes
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet WillClosePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.Minutes)
	return writer.Bytes(), nil
}
