package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// ModeratorInitPacket composes the moderator tool init payload (s2c 2696).
type ModeratorInitPacket struct {
	// Presets stores available moderation preset categories.
	Presets []PresetCategory
	// TicketPermission indicates whether the user can handle tickets.
	TicketPermission bool
	// ChatlogPermission indicates whether the user can view chatlogs.
	ChatlogPermission bool
}

// PresetCategory groups presets under one category name.
type PresetCategory struct {
	// Name stores the category display name.
	Name string
	// Entries stores preset entries within the category.
	Entries []string
}

// PacketID returns protocol packet identifier.
func (p ModeratorInitPacket) PacketID() uint16 { return ModeratorInitPacketID }

// Encode serializes the moderator init payload.
func (p ModeratorInitPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Presets)))
	for _, cat := range p.Presets {
		if err := w.WriteString(cat.Name); err != nil {
			return nil, err
		}
		w.WriteInt32(int32(len(cat.Entries)))
		for _, entry := range cat.Entries {
			if err := w.WriteString(entry); err != nil {
				return nil, err
			}
		}
	}
	w.WriteBool(p.TicketPermission)
	w.WriteBool(p.ChatlogPermission)
	return w.Bytes(), nil
}
