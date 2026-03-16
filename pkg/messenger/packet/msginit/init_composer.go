package msginit

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerInitComposer defines server messenger.init response payload.
type MessengerInitComposer struct {
	// UserFriendLimit stores the user-specific friend limit.
	UserFriendLimit int32
	// NormalLimit stores the normal friend list limit.
	NormalLimit int32
	// ExtendedLimit stores the VIP friend list limit.
	ExtendedLimit int32
}

// PacketID returns protocol packet identifier.
func (p MessengerInitComposer) PacketID() uint16 { return MessengerInitComposerID }

// Encode serializes packet body payload.
func (p MessengerInitComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
		w.WriteInt32(p.UserFriendLimit)
		w.WriteInt32(p.NormalLimit)
		w.WriteInt32(p.ExtendedLimit)
		w.WriteInt32(0)
	return w.Bytes(), nil
}
