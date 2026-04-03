package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// ChatPacket decodes client talk message (c2s 744).
type ChatPacket struct {
	// Message stores the chat text.
	Message string
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p ChatPacket) PacketID() uint16 { return ChatPacketID }

// Decode parses packet body.
func (p *ChatPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	style, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	p.BubbleStyle = style
	return nil
}

// ShoutPacket decodes client shout message (c2s 697).
type ShoutPacket struct {
	// Message stores the shout text.
	Message string
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p ShoutPacket) PacketID() uint16 { return ShoutPacketID }

// Decode parses packet body.
func (p *ShoutPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	style, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	p.BubbleStyle = style
	return nil
}

// WhisperPacket decodes client whisper message (c2s 3003).
type WhisperPacket struct {
	// TargetUsername stores the whisper recipient display name.
	TargetUsername string
	// Message stores the whisper text.
	Message string
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p WhisperPacket) PacketID() uint16 { return WhisperPacketID }

// Decode parses packet body.
func (p *WhisperPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	target, _ := r.ReadString()
	msg, _ := r.ReadString()
	style, _ := r.ReadInt32()
	p.TargetUsername, p.Message, p.BubbleStyle = target, msg, style
	return nil
}

// ChatComposer sends talk message to proximate entities (s2c 2785).
type ChatComposer struct {
	// VirtualID stores the sender entity virtual identifier.
	VirtualID int32
	// Message stores the chat text payload.
	Message string
	// GestureID stores the gesture animation identifier.
	GestureID int32
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p ChatComposer) PacketID() uint16 { return ChatComposerID }

// Encode serializes the chat message.
func (p ChatComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	if err := w.WriteString(p.Message); err != nil {
		return nil, err
	}
	w.WriteInt32(p.GestureID)
	w.WriteInt32(p.BubbleStyle)
	w.WriteInt32(0)
	w.WriteInt32(-1)
	return w.Bytes(), nil
}

// ShoutComposer sends room-wide shout to all entities (s2c 2888).
type ShoutComposer struct {
	// VirtualID stores the sender entity virtual identifier.
	VirtualID int32
	// Message stores the shout text payload.
	Message string
	// GestureID stores the gesture animation identifier.
	GestureID int32
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p ShoutComposer) PacketID() uint16 { return ShoutComposerID }

// Encode serializes the shout message.
func (p ShoutComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	if err := w.WriteString(p.Message); err != nil {
		return nil, err
	}
	w.WriteInt32(p.GestureID)
	w.WriteInt32(p.BubbleStyle)
	w.WriteInt32(0)
	w.WriteInt32(-1)
	return w.Bytes(), nil
}

// WhisperComposer sends targeted whisper to sender and recipient (s2c 1400).
type WhisperComposer struct {
	// VirtualID stores the sender entity virtual identifier.
	VirtualID int32
	// SenderName stores the sender display name.
	SenderName string
	// Message stores the whisper text payload.
	Message string
	// GestureID stores the gesture animation identifier.
	GestureID int32
	// BubbleStyle stores the chat bubble style identifier.
	BubbleStyle int32
}

// PacketID returns the protocol packet identifier.
func (p WhisperComposer) PacketID() uint16 { return WhisperComposerID }

// Encode serializes the whisper message.
func (p WhisperComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	if err := w.WriteString(p.SenderName); err != nil {
		return nil, err
	}
	w.WriteInt32(0)
	if err := w.WriteString(p.Message); err != nil {
		return nil, err
	}
	w.WriteInt32(p.GestureID)
	w.WriteInt32(p.BubbleStyle)
	w.WriteInt32(0)
	w.WriteInt32(-1)
	return w.Bytes(), nil
}
