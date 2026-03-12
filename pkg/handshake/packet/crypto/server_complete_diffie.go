package crypto

import "github.com/momlesstomato/pixel-server/core/codec"

// ServerCompleteDiffiePacketID identifies handshake.complete_diffie S2C packet.
const ServerCompleteDiffiePacketID uint16 = 3885

// ServerCompleteDiffiePacket carries server public key handshake completion payload.
type ServerCompleteDiffiePacket struct {
	// EncryptedPublicKey stores encoded server diffie public key value.
	EncryptedPublicKey string
	// ServerClientEncryption indicates whether server-to-client encryption is enabled.
	ServerClientEncryption bool
}

// PacketID returns protocol packet id.
func (packet ServerCompleteDiffiePacket) PacketID() uint16 { return ServerCompleteDiffiePacketID }

// Decode parses packet body into struct fields.
func (packet *ServerCompleteDiffiePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	publicKey, err := reader.ReadString()
	if err != nil {
		return err
	}
	enabled, err := reader.ReadBool()
	if err != nil {
		return err
	}
	packet.EncryptedPublicKey = publicKey
	packet.ServerClientEncryption = enabled
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ServerCompleteDiffiePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.EncryptedPublicKey); err != nil {
		return nil, err
	}
	writer.WriteBool(packet.ServerClientEncryption)
	return writer.Bytes(), nil
}
