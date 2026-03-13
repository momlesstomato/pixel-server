package navigation

// DesktopViewRequestPacketID identifies session.desktop_view C2S packet.
const DesktopViewRequestPacketID uint16 = 105

// DesktopViewRequestPacket carries room-exit desktop view request payload.
type DesktopViewRequestPacket struct{}

// PacketID returns protocol packet identifier.
func (packet DesktopViewRequestPacket) PacketID() uint16 { return DesktopViewRequestPacketID }

// Decode parses packet body into struct fields.
func (packet *DesktopViewRequestPacket) Decode(_ []byte) error { return nil }

// Encode serializes packet fields into protocol body bytes.
func (packet DesktopViewRequestPacket) Encode() ([]byte, error) { return []byte{}, nil }
