package availability

import "github.com/momlesstomato/pixel-server/core/codec"

// StatusPacketID identifies availability.status packet.
const StatusPacketID uint16 = 2033

// StatusPacket carries hotel availability flags.
type StatusPacket struct {
	// IsOpen stores whether hotel is currently open.
	IsOpen bool
	// OnShutdown stores whether hotel is in shutdown countdown.
	OnShutdown bool
	// IsAuthentic stores whether session is authenticated.
	IsAuthentic bool
}

// PacketID returns protocol packet identifier.
func (packet StatusPacket) PacketID() uint16 { return StatusPacketID }

// Decode parses packet body into struct fields.
func (packet *StatusPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	isOpen, err := reader.ReadBool()
	if err != nil {
		return err
	}
	onShutdown, err := reader.ReadBool()
	if err != nil {
		return err
	}
	isAuthentic, err := reader.ReadBool()
	if err != nil {
		return err
	}
	packet.IsOpen = isOpen
	packet.OnShutdown = onShutdown
	packet.IsAuthentic = isAuthentic
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet StatusPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteBool(packet.IsOpen)
	writer.WriteBool(packet.OnShutdown)
	writer.WriteBool(packet.IsAuthentic)
	return writer.Bytes(), nil
}
