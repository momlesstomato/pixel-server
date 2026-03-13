package notification

import "github.com/momlesstomato/pixel-server/core/codec"

// GenericAlertPacketID identifies session.generic_alert packet.
const GenericAlertPacketID uint16 = 3801

// GenericAlertPacket carries one user-visible alert message payload.
type GenericAlertPacket struct {
	// Message stores alert modal text.
	Message string
}

// PacketID returns protocol packet identifier.
func (packet GenericAlertPacket) PacketID() uint16 { return GenericAlertPacketID }

// Decode parses packet body into struct fields.
func (packet *GenericAlertPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	message, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.Message = message
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet GenericAlertPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteString(packet.Message)
	return writer.Bytes(), nil
}
