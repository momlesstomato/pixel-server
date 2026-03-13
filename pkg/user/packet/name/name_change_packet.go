package name

import "github.com/momlesstomato/pixel-server/core/codec"

// UserNameChangePacket defines user.name_change packet payload.
type UserNameChangePacket struct {
	// WebID stores web identifier payload.
	WebID int32
	// UserID stores user identifier payload.
	UserID int32
	// NewName stores changed username payload.
	NewName string
}

// PacketID returns protocol packet identifier.
func (packet UserNameChangePacket) PacketID() uint16 { return UserNameChangePacketID }

// Encode serializes packet body payload.
func (packet UserNameChangePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.WebID)
	writer.WriteInt32(packet.UserID)
	if err := writer.WriteString(packet.NewName); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
