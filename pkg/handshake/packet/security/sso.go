package security

import "github.com/momlesstomato/pixel-server/core/codec"

// SSOTicketPacketID identifies security.sso_ticket packet.
const SSOTicketPacketID uint16 = 2419

// SSOTicketPacket carries single-sign-on ticket payload.
type SSOTicketPacket struct {
	// Ticket stores authentication ticket string.
	Ticket string
	// Timestamp stores optional ticket timestamp when provided.
	Timestamp *int32
}

// PacketID returns protocol packet id.
func (packet SSOTicketPacket) PacketID() uint16 { return SSOTicketPacketID }

// Decode parses packet body into struct fields.
func (packet *SSOTicketPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	ticket, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.Ticket = ticket
	if reader.Remaining() >= 4 {
		timestamp, tsErr := reader.ReadInt32()
		if tsErr != nil {
			return tsErr
		}
		packet.Timestamp = &timestamp
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet SSOTicketPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Ticket); err != nil {
		return nil, err
	}
	if packet.Timestamp != nil {
		writer.WriteInt32(*packet.Timestamp)
	}
	return writer.Bytes(), nil
}
