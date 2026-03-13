package notification

import "github.com/momlesstomato/pixel-server/core/codec"

// ModerationCautionPacketID identifies session.moderation_caution packet.
const ModerationCautionPacketID uint16 = 1890

// ModerationCautionPacket carries one moderation caution payload.
type ModerationCautionPacket struct {
	// Message stores primary moderation warning message.
	Message string
	// Detail stores moderation warning detail message.
	Detail string
}

// PacketID returns protocol packet identifier.
func (packet ModerationCautionPacket) PacketID() uint16 { return ModerationCautionPacketID }

// Decode parses packet body into struct fields.
func (packet *ModerationCautionPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	message, err := reader.ReadString()
	if err != nil {
		return err
	}
	detail, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.Message = message
	packet.Detail = detail
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ModerationCautionPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteString(packet.Message)
	writer.WriteString(packet.Detail)
	return writer.Bytes(), nil
}
