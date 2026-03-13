package error

import "github.com/momlesstomato/pixel-server/core/codec"

// ConnectionErrorPacketID identifies connection.error packet.
const ConnectionErrorPacketID uint16 = 1004

// ConnectionErrorPacket carries one protocol-level connection error payload.
type ConnectionErrorPacket struct {
	// MessageID stores offending packet identifier.
	MessageID int32
	// ErrorCode stores server error classification.
	ErrorCode int32
	// Timestamp stores server-side UTC timestamp value.
	Timestamp string
}

// PacketID returns protocol packet identifier.
func (packet ConnectionErrorPacket) PacketID() uint16 { return ConnectionErrorPacketID }

// Decode parses packet body into struct fields.
func (packet *ConnectionErrorPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	messageID, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	errorCode, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	timestamp, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.MessageID = messageID
	packet.ErrorCode = errorCode
	packet.Timestamp = timestamp
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ConnectionErrorPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.MessageID)
	writer.WriteInt32(packet.ErrorCode)
	writer.WriteString(packet.Timestamp)
	return writer.Bytes(), nil
}
