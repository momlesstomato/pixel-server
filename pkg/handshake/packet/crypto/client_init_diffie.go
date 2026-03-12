package crypto

import "fmt"

// ClientInitDiffiePacketID identifies handshake.init_diffie C2S packet.
const ClientInitDiffiePacketID uint16 = 3110

// ClientInitDiffiePacket requests server diffie parameters.
type ClientInitDiffiePacket struct{}

// PacketID returns protocol packet id.
func (packet ClientInitDiffiePacket) PacketID() uint16 { return ClientInitDiffiePacketID }

// Decode validates packet body as empty payload.
func (packet *ClientInitDiffiePacket) Decode(body []byte) error {
	if len(body) != 0 {
		return fmt.Errorf("handshake.init_diffie body must be empty")
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientInitDiffiePacket) Encode() ([]byte, error) {
	return []byte{}, nil
}
