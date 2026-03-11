package authentication

import "fmt"

// AuthenticationOKPacketID identifies authentication.ok packet.
const AuthenticationOKPacketID uint16 = 2491

// AuthenticationOKPacket indicates successful SSO authentication.
type AuthenticationOKPacket struct{}

// PacketID returns protocol packet id.
func (packet AuthenticationOKPacket) PacketID() uint16 { return AuthenticationOKPacketID }

// Decode validates packet body as empty payload.
func (packet *AuthenticationOKPacket) Decode(body []byte) error {
	if len(body) != 0 {
		return fmt.Errorf("authentication.ok body must be empty")
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet AuthenticationOKPacket) Encode() ([]byte, error) {
	return []byte{}, nil
}
