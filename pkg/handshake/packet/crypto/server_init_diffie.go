package crypto

import "github.com/momlesstomato/pixel-server/core/codec"

// ServerInitDiffiePacketID identifies handshake.init_diffie S2C packet.
const ServerInitDiffiePacketID uint16 = 1347

// ServerInitDiffiePacket carries server diffie parameters payload.
type ServerInitDiffiePacket struct {
	// EncryptedPrime stores encoded prime value.
	EncryptedPrime string
	// EncryptedGenerator stores encoded generator value.
	EncryptedGenerator string
}

// PacketID returns protocol packet id.
func (packet ServerInitDiffiePacket) PacketID() uint16 { return ServerInitDiffiePacketID }

// Decode parses packet body into struct fields.
func (packet *ServerInitDiffiePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	prime, err := reader.ReadString()
	if err != nil {
		return err
	}
	generator, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.EncryptedPrime = prime
	packet.EncryptedGenerator = generator
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ServerInitDiffiePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.EncryptedPrime); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.EncryptedGenerator); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
