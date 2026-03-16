package request

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerNewRequestComposer defines server messenger.new_request payload.
type MessengerNewRequestComposer struct {
	// RequestID stores the pending request identifier.
	RequestID int32
	// FromUsername stores the requesting user display name.
	FromUsername string
	// FromFigure stores the requesting user figure string.
	FromFigure string
}

// PacketID returns protocol packet identifier.
func (p MessengerNewRequestComposer) PacketID() uint16 { return MessengerNewRequestComposerID }

// Encode serializes packet body payload.
func (p MessengerNewRequestComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RequestID)
	if err := w.WriteString(p.FromUsername); err != nil {
		return nil, err
	}
	return w.Bytes(), w.WriteString(p.FromFigure)
}

// MessengerRequestErrorComposer defines server messenger.request_error payload.
type MessengerRequestErrorComposer struct {
	// ClientMsgID stores the originating client message identifier.
	ClientMsgID int32
	// ErrorCode stores the request failure code.
	ErrorCode int32
}

// PacketID returns protocol packet identifier.
func (p MessengerRequestErrorComposer) PacketID() uint16 { return MessengerRequestErrorComposerID }

// Encode serializes packet body payload.
func (p MessengerRequestErrorComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ClientMsgID)
	w.WriteInt32(p.ErrorCode)
	return w.Bytes(), nil
}

// FriendUpdateEntry defines one update action entry in messenger.friend_update.
type FriendUpdateEntry struct {
	// Action stores -1=removed, 0=updated, 1=added.
	Action int32
	// FriendID stores the affected user identifier.
	FriendID int32
	// Username stores the display name (used for action>=0).
	Username string
	// Gender stores gender value (used for action>=0).
	Gender int32
	// Online stores online status (used for action>=0).
	Online bool
	// Figure stores the avatar figure string (used for action>=0).
	Figure string
	// Motto stores the player motto (used for action>=0).
	Motto string
	// Relationship stores the relationship label (used for action>=0).
	Relationship int16
}

// MessengerFriendUpdateComposer defines server messenger.friend_update payload.
type MessengerFriendUpdateComposer struct {
	// Entries stores all update action entries.
	Entries []FriendUpdateEntry
}

// PacketID returns protocol packet identifier.
func (p MessengerFriendUpdateComposer) PacketID() uint16 { return MessengerFriendUpdateComposerID }

// Encode serializes packet body payload.
func (p MessengerFriendUpdateComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(0)
	w.WriteInt32(int32(len(p.Entries)))
	for _, e := range p.Entries {
		w.WriteInt32(e.Action)
		if e.Action < 0 {
			w.WriteInt32(e.FriendID)
			continue
		}
		w.WriteInt32(e.FriendID)
		if err := w.WriteString(e.Username); err != nil {
			return nil, err
		}
		w.WriteInt32(e.Gender)
		w.WriteBool(e.Online)
		w.WriteBool(false)
		figure := ""
		if e.Online {
			figure = e.Figure
		}
		if err := w.WriteString(figure); err != nil {
			return nil, err
		}
		w.WriteInt32(0)
		if err := w.WriteString(e.Motto); err != nil {
			return nil, err
		}
		if err := w.WriteString(""); err != nil {
			return nil, err
		}
		if err := w.WriteString(""); err != nil {
			return nil, err
		}
		w.WriteBool(true)
		w.WriteBool(false)
		w.WriteBool(false)
		w.WriteUint16(uint16(e.Relationship))
	}
	return w.Bytes(), nil
}
