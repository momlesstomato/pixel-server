package packet

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// EffectsListPacket encodes user.effects (s2c 340) with full effect inventory.
type EffectsListPacket struct {
	// Effects stores all user effect entries.
	Effects []domain.Effect
}

// PacketID returns the wire protocol packet identifier.
func (p EffectsListPacket) PacketID() uint16 { return EffectsResponsePacketID }

// Encode serializes effect inventory into packet body.
func (p EffectsListPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Effects)))
	for _, e := range p.Effects {
		w.WriteInt32(int32(e.EffectID))
		w.WriteInt32(0)
		w.WriteInt32(int32(e.Duration))
		w.WriteInt32(int32(e.Quantity))
		activated := e.ActivatedAt != nil
		w.WriteBool(activated)
		secondsLeft := int32(0)
		if activated && !e.IsPermanent {
			secondsLeft = int32(e.Duration)
		}
		w.WriteInt32(secondsLeft)
		w.WriteBool(e.IsPermanent)
	}
	return w.Bytes(), nil
}

// EffectActivatedPacket encodes user.effect_activated (s2c 1959).
type EffectActivatedPacket struct {
	// EffectID stores the activated effect type identifier.
	EffectID int
	// Duration stores the total duration in seconds.
	Duration int
	// IsPermanent stores whether the effect never expires.
	IsPermanent bool
}

// PacketID returns the wire protocol packet identifier.
func (p EffectActivatedPacket) PacketID() uint16 { return EffectActivatedPacketID }

// Encode serializes effect activation confirmation into packet body.
func (p EffectActivatedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.EffectID))
	w.WriteInt32(int32(p.Duration))
	w.WriteBool(p.IsPermanent)
	return w.Bytes(), nil
}
