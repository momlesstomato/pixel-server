package ignore

import "github.com/momlesstomato/pixel-server/core/codec"

// UserIgnoredUsersPacket defines user.ignored_users packet payload.
type UserIgnoredUsersPacket struct {
	// Usernames stores ignored usernames payload.
	Usernames []string
}

// PacketID returns protocol packet identifier.
func (packet UserIgnoredUsersPacket) PacketID() uint16 { return UserIgnoredUsersPacketID }

// Encode serializes packet body payload.
func (packet UserIgnoredUsersPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(int32(len(packet.Usernames)))
	for _, username := range packet.Usernames {
		if err := writer.WriteString(username); err != nil {
			return nil, err
		}
	}
	return writer.Bytes(), nil
}
