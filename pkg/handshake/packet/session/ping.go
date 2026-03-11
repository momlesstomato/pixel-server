package session

import "fmt"

// ClientPingPacketID identifies client.ping packet.
const ClientPingPacketID uint16 = 3928

// ClientPingPacket represents server heartbeat probe packet.
type ClientPingPacket struct{}

// PacketID returns protocol packet id.
func (packet ClientPingPacket) PacketID() uint16 { return ClientPingPacketID }

// Decode validates packet body as empty payload.
func (packet *ClientPingPacket) Decode(body []byte) error {
	if len(body) != 0 {
		return fmt.Errorf("client.ping body must be empty")
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientPingPacket) Encode() ([]byte, error) {
	return []byte{}, nil
}
