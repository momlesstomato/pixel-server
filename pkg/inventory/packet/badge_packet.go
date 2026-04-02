package packet

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// BadgesListPacket encodes user.badges (s2c 1087) with full badge inventory.
type BadgesListPacket struct {
	// Badges stores all user badge entries.
	Badges []domain.Badge
	// Slots stores equipped badge slot assignments.
	Slots []domain.BadgeSlot
}

// PacketID returns the wire protocol packet identifier.
func (p BadgesListPacket) PacketID() uint16 { return BadgesResponsePacketID }

// Encode serializes badge inventory and slot assignments into packet body.
func (p BadgesListPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Badges)))
	for _, b := range p.Badges {
		w.WriteInt32(int32(b.ID))
		if err := w.WriteString(b.BadgeCode); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Slots)))
	for _, s := range p.Slots {
		w.WriteInt32(int32(s.SlotID))
		if err := w.WriteString(s.BadgeCode); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// CurrentBadgesPacket encodes user.current_badges (s2c 2091) with equipped slots.
type CurrentBadgesPacket struct {
	// UserID stores the badge owner identifier.
	UserID int
	// Slots stores currently equipped badge slots.
	Slots []domain.BadgeSlot
}

// PacketID returns the wire protocol packet identifier.
func (p CurrentBadgesPacket) PacketID() uint16 { return CurrentBadgesPacketID }

// Encode serializes equipped badge slot state into packet body.
func (p CurrentBadgesPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.UserID))
	w.WriteInt32(int32(len(p.Slots)))
	for _, s := range p.Slots {
		w.WriteInt32(int32(s.SlotID))
		if err := w.WriteString(s.BadgeCode); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
