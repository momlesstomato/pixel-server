package social

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerFriendNotificationComposer defines server messenger.friend_notification payload.
type MessengerFriendNotificationComposer struct {
	// FriendID stores the notifying friend identifier as a string.
	FriendID string
	// TypeCode stores the notification type code.
	TypeCode int32
	// Data stores supplementary notification data.
	Data string
}

// PacketID returns protocol packet identifier.
func (p MessengerFriendNotificationComposer) PacketID() uint16 {
	return MessengerFriendNotificationComposerID
}

// Encode serializes packet body payload.
func (p MessengerFriendNotificationComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.FriendID); err != nil {
		return nil, err
	}
	w.WriteInt32(p.TypeCode)
	return w.Bytes(), w.WriteString(p.Data)
}
