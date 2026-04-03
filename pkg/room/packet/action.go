package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// DancePacket decodes client dance request (c2s 1225).
type DancePacket struct {
	// DanceID stores the dance animation identifier.
	DanceID int32
}

// PacketID returns the protocol packet identifier.
func (p DancePacket) PacketID() uint16 { return DancePacketID }

// Decode parses packet body.
func (p *DancePacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.DanceID = id
	return nil
}

// ActionPacket decodes client action request (c2s 3268).
type ActionPacket struct {
	// ActionID stores the action animation identifier.
	ActionID int32
}

// PacketID returns the protocol packet identifier.
func (p ActionPacket) PacketID() uint16 { return ActionPacketID }

// Decode parses packet body.
func (p *ActionPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.ActionID = id
	return nil
}

// SignPacket decodes client sign display request (c2s 3555).
type SignPacket struct {
	// SignID stores the sign identifier.
	SignID int32
}

// PacketID returns the protocol packet identifier.
func (p SignPacket) PacketID() uint16 { return SignPacketID }

// Decode parses packet body.
func (p *SignPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	id, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.SignID = id
	return nil
}

// LookToPacket decodes client head rotation request (c2s 1142).
type LookToPacket struct {
	// X stores the target horizontal coordinate.
	X int32
	// Y stores the target vertical coordinate.
	Y int32
}

// PacketID returns the protocol packet identifier.
func (p LookToPacket) PacketID() uint16 { return LookToPacketID }

// Decode parses packet body.
func (p *LookToPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	x, err := r.ReadInt32()
	if err != nil {
		return err
	}
	y, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.X, p.Y = x, y
	return nil
}

// DanceComposer sends dance animation update to all room entities (s2c 130).
type DanceComposer struct {
	// VirtualID stores the dancing entity virtual identifier.
	VirtualID int32
	// DanceID stores the dance animation identifier.
	DanceID int32
}

// PacketID returns the protocol packet identifier.
func (p DanceComposer) PacketID() uint16 { return DanceComposerID }

// Encode serializes the dance update.
func (p DanceComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	w.WriteInt32(p.DanceID)
	return w.Bytes(), nil
}

// UserTypingComposer sends typing indicator update to all entities (s2c 1727).
type UserTypingComposer struct {
	// VirtualID stores the typing entity virtual identifier.
	VirtualID int32
	// IsTyping reports whether the entity is currently typing.
	IsTyping bool
}

// PacketID returns the protocol packet identifier.
func (p UserTypingComposer) PacketID() uint16 { return UserTypingComposerID }

// Encode serializes the typing status update.
func (p UserTypingComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	w.WriteBool(p.IsTyping)
	return w.Bytes(), nil
}

// SleepComposer sends idle state update to all entities (s2c 2306).
type SleepComposer struct {
	// VirtualID stores the idle entity virtual identifier.
	VirtualID int32
	// IsAsleep reports whether the entity is in idle state.
	IsAsleep bool
}

// PacketID returns the protocol packet identifier.
func (p SleepComposer) PacketID() uint16 { return SleepComposerID }

// Encode serializes the idle state update.
func (p SleepComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.VirtualID)
	w.WriteBool(p.IsAsleep)
	return w.Bytes(), nil
}
