package profileview

import "github.com/momlesstomato/pixel-server/core/codec"

// UserProfilePacket defines user.profile packet payload.
type UserProfilePacket struct {
	// UserID stores target user identifier.
	UserID int32
	// Username stores target username payload.
	Username string
	// Figure stores target figure payload.
	Figure string
	// Motto stores target motto payload.
	Motto string
	// Registration stores registration date payload.
	Registration string
	// AchievementPoints stores achievement points value.
	AchievementPoints int32
	// FriendsCount stores friend count value.
	FriendsCount int32
	// IsMyFriend stores friend status marker.
	IsMyFriend bool
	// RequestSent stores pending friend request marker.
	RequestSent bool
	// IsOnline stores online marker.
	IsOnline bool
	// SecondsSinceLastVisit stores elapsed seconds since last visit.
	SecondsSinceLastVisit int32
	// OpenProfileWindow stores profile view marker.
	OpenProfileWindow bool
}

// PacketID returns protocol packet identifier.
func (packet UserProfilePacket) PacketID() uint16 { return UserProfilePacketID }

// Encode serializes packet body payload.
func (packet UserProfilePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	for _, value := range []string{packet.Username, packet.Figure, packet.Motto, packet.Registration} {
		if err := writer.WriteString(value); err != nil {
			return nil, err
		}
	}
	writer.WriteInt32(packet.AchievementPoints)
	writer.WriteInt32(packet.FriendsCount)
	writer.WriteBool(packet.IsMyFriend)
	writer.WriteBool(packet.RequestSent)
	writer.WriteBool(packet.IsOnline)
	writer.WriteInt32(0)
	writer.WriteInt32(packet.SecondsSinceLastVisit)
	writer.WriteBool(packet.OpenProfileWindow)
	return writer.Bytes(), nil
}
