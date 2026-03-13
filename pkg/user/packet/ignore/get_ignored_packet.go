package ignore

// UserGetIgnoredPacket defines user.get_ignored packet payload.
type UserGetIgnoredPacket struct{}

// PacketID returns protocol packet identifier.
func (packet UserGetIgnoredPacket) PacketID() uint16 { return UserGetIgnoredPacketID }

// Encode serializes packet body payload.
func (packet UserGetIgnoredPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (packet *UserGetIgnoredPacket) Decode(payload []byte) error { _ = payload; return nil }
