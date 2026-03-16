package message

import "github.com/momlesstomato/pixel-server/core/codec"

// MessengerSendMsgPacket defines client messenger.send_msg payload.
type MessengerSendMsgPacket struct {
	// UserID stores the recipient user identifier.
	UserID int32
	// Message stores the message content.
	Message string
}

// PacketID returns protocol packet identifier.
func (p MessengerSendMsgPacket) PacketID() uint16 { return MessengerSendMsgPacketID }

// Encode serializes packet body payload.
func (p MessengerSendMsgPacket) Encode() ([]byte, error) { return []byte{}, nil }

// Decode parses packet body payload.
func (p *MessengerSendMsgPacket) Decode(payload []byte) error {
	r := codec.NewReader(payload)
	userID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = userID
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}

// MessengerNewMessageComposer defines server messenger.new_message payload.
type MessengerNewMessageComposer struct {
	// SenderID stores the sender user identifier.
	SenderID int32
	// Message stores the message content.
	Message string
	// SecondsSinceSent stores elapsed seconds since original send time.
	SecondsSinceSent int32
}

// PacketID returns protocol packet identifier.
func (p MessengerNewMessageComposer) PacketID() uint16 { return MessengerNewMessageComposerID }

// Encode serializes packet body payload.
func (p MessengerNewMessageComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.SenderID)
	if err := w.WriteString(p.Message); err != nil {
		return nil, err
	}
	w.WriteInt32(p.SecondsSinceSent)
	return w.Bytes(), nil
}

// MessengerMessageErrorComposer defines server messenger.message_error payload.
type MessengerMessageErrorComposer struct {
	// ErrorCode stores message delivery error code.
	ErrorCode int32
	// UserID stores the affected user identifier.
	UserID int32
}

// PacketID returns protocol packet identifier.
func (p MessengerMessageErrorComposer) PacketID() uint16 { return MessengerMessageErrorComposerID }

// Encode serializes packet body payload.
func (p MessengerMessageErrorComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
		w.WriteInt32(p.ErrorCode)
	w.WriteInt32(p.UserID)
		return w.Bytes(), nil
}
