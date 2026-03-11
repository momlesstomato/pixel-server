package session

import "fmt"

// ClientDisconnectPacketID identifies client.disconnect packet.
const ClientDisconnectPacketID uint16 = 2445

// ClientDisconnectPacket represents graceful disconnect signal.
type ClientDisconnectPacket struct{}

// PacketID returns protocol packet id.
func (packet ClientDisconnectPacket) PacketID() uint16 { return ClientDisconnectPacketID }

// Decode validates packet body as empty payload.
func (packet *ClientDisconnectPacket) Decode(body []byte) error {
	if len(body) != 0 {
		return fmt.Errorf("client.disconnect body must be empty")
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientDisconnectPacket) Encode() ([]byte, error) {
	return []byte{}, nil
}
