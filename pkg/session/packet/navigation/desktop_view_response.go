package navigation

// DesktopViewResponsePacketID identifies session.desktop_view S2C packet.
const DesktopViewResponsePacketID uint16 = 3523

// DesktopViewResponsePacket carries room-exit desktop view response payload.
type DesktopViewResponsePacket struct{}

// PacketID returns protocol packet identifier.
func (packet DesktopViewResponsePacket) PacketID() uint16 { return DesktopViewResponsePacketID }

// Decode parses packet body into struct fields.
func (packet *DesktopViewResponsePacket) Decode(_ []byte) error { return nil }

// Encode serializes packet fields into protocol body bytes.
func (packet DesktopViewResponsePacket) Encode() ([]byte, error) { return []byte{}, nil }
