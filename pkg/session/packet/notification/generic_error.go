package notification

import "github.com/momlesstomato/pixel-server/core/codec"

// GenericErrorPacketID identifies session.generic_error packet.
const GenericErrorPacketID uint16 = 1600

// GenericErrorPacket carries one generic error code payload.
type GenericErrorPacket struct {
	// ErrorCode stores numeric generic error code.
	ErrorCode int32
}

// PacketID returns protocol packet identifier.
func (packet GenericErrorPacket) PacketID() uint16 { return GenericErrorPacketID }

// Decode parses packet body into struct fields.
func (packet *GenericErrorPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	errorCode, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.ErrorCode = errorCode
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet GenericErrorPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ErrorCode)
	return writer.Bytes(), nil
}
