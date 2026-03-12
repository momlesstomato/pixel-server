package crypto

import "github.com/momlesstomato/pixel-server/core/codec"

// ClientCompleteDiffiePacketID identifies handshake.complete_diffie C2S packet.
const ClientCompleteDiffiePacketID uint16 = 773

// ClientCompleteDiffiePacket carries client public key payload.
type ClientCompleteDiffiePacket struct {
	// EncryptedPublicKey stores encoded client diffie public key value.
	EncryptedPublicKey string
}

// PacketID returns protocol packet id.
func (packet ClientCompleteDiffiePacket) PacketID() uint16 { return ClientCompleteDiffiePacketID }

// Decode parses packet body into struct fields.
func (packet *ClientCompleteDiffiePacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	value, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.EncryptedPublicKey = value
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientCompleteDiffiePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.EncryptedPublicKey); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
