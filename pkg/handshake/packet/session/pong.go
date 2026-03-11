package session

import "fmt"

// ClientPongPacketID identifies client.pong packet.
const ClientPongPacketID uint16 = 2596

// ClientPongPacket represents client heartbeat acknowledgment.
type ClientPongPacket struct{}

// PacketID returns protocol packet id.
func (packet ClientPongPacket) PacketID() uint16 { return ClientPongPacketID }

// Decode validates packet body as empty payload.
func (packet *ClientPongPacket) Decode(body []byte) error {
	if len(body) != 0 {
		return fmt.Errorf("client.pong body must be empty")
	}
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet ClientPongPacket) Encode() ([]byte, error) {
	return []byte{}, nil
}
