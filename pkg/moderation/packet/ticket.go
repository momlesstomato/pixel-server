package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// CallForHelpPacket decodes a call-for-help submit request (c2s 1691).
type CallForHelpPacket struct {
	// Message stores the reporter description.
	Message string
	// Category stores the ticket category identifier.
	Category int32
	// ReportedID stores the user being reported.
	ReportedID int32
	// RoomID stores the room context.
	RoomID int32
}

// Decode reads fields from the packet body.
func (p *CallForHelpPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	cat, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Category = cat
	rid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.ReportedID = rid
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	return nil
}

// SanctionTradeLockPacket decodes a trade lock sanction (c2s 3742).
type SanctionTradeLockPacket struct {
	// UserID stores the target user identifier.
	UserID int32
	// Message stores the lock reason.
	Message string
	// Duration stores the lock duration in hours.
	Duration int32
}

// Decode reads fields from the packet body.
func (p *SanctionTradeLockPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	uid, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = uid
	msg, err := r.ReadString()
	if err != nil {
		return err
	}
	p.Message = msg
	dur, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Duration = dur
	return nil
}

// CFHPendingPacket composes the pending tickets list (s2c 1121).
type CFHPendingPacket struct {
	// Entries stores pending calls displayed in the moderator tool.
	Entries []CFHPendingEntry
}

// CFHPendingEntry stores one pending CFH list row.
type CFHPendingEntry struct {
	// CallID stores the ticket identifier shown by Nitro.
	CallID string
	// Timestamp stores the human-readable submission time.
	Timestamp string
	// Message stores the ticket message preview.
	Message string
}

// PacketID returns protocol packet identifier.
func (p CFHPendingPacket) PacketID() uint16 { return CFHPendingPacketID }

// Encode serializes the pending ticket list.
func (p CFHPendingPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Entries)))
	for _, entry := range p.Entries {
		if err := w.WriteString(entry.CallID); err != nil {
			return nil, err
		}
		if err := w.WriteString(entry.Timestamp); err != nil {
			return nil, err
		}
		if err := w.WriteString(entry.Message); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GetCFHChatlogPacket decodes a moderator CFH chatlog request.
type GetCFHChatlogPacket struct {
	// TicketID stores the target ticket identifier.
	TicketID int32
}

// Decode reads fields from the packet body.
func (p *GetCFHChatlogPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	ticketID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.TicketID = ticketID
	return nil
}
