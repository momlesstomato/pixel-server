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

// GuideSessionCreatePacket decodes a guide assistance request.
type GuideSessionCreatePacket struct {
	// RequestType stores the guide-session request category.
	RequestType int32
	// Message stores the requester message.
	Message string
}

// Decode reads fields from the packet body.
func (p *GuideSessionCreatePacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	requestType, err := r.ReadInt32()
	if err != nil {
		return err
	}
	message, err := r.ReadString()
	if err != nil {
		return err
	}
	p.RequestType = requestType
	p.Message = message
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

// CFHSanctionStatusPacket composes sanction-status details for help flows.
type CFHSanctionStatusPacket struct {
	// IsSanctionNew stores whether the current sanction is recent.
	IsSanctionNew bool
	// IsSanctionActive stores whether the sanction is currently active.
	IsSanctionActive bool
	// SanctionName stores the current sanction label.
	SanctionName string
	// SanctionLengthHours stores the sanction length in hours.
	SanctionLengthHours int32
	// SanctionReason stores the moderation reason key or text.
	SanctionReason string
	// SanctionCreationTime stores the sanction creation timestamp string.
	SanctionCreationTime string
	// ProbationHoursLeft stores remaining active hours for the sanction.
	ProbationHoursLeft int32
	// NextSanctionName stores the next escalation label.
	NextSanctionName string
	// NextSanctionLengthHours stores the next escalation duration in hours.
	NextSanctionLengthHours int32
	// HasCustomMute stores whether the user is actively muted.
	HasCustomMute bool
	// TradeLockExpiryTime stores the trade-lock expiry timestamp string.
	TradeLockExpiryTime string
}

// PacketID returns protocol packet identifier.
func (p CFHSanctionStatusPacket) PacketID() uint16 { return CFHSanctionStatusPacketID }

// Encode serializes sanction-status details.
func (p CFHSanctionStatusPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.IsSanctionNew)
	w.WriteBool(p.IsSanctionActive)
	if err := w.WriteString(p.SanctionName); err != nil {
		return nil, err
	}
	w.WriteInt32(p.SanctionLengthHours)
	w.WriteInt32(30)
	if err := w.WriteString(p.SanctionReason); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.SanctionCreationTime); err != nil {
		return nil, err
	}
	w.WriteInt32(p.ProbationHoursLeft)
	if err := w.WriteString(p.NextSanctionName); err != nil {
		return nil, err
	}
	w.WriteInt32(p.NextSanctionLengthHours)
	w.WriteInt32(30)
	w.WriteBool(p.HasCustomMute)
	if err := w.WriteString(p.TradeLockExpiryTime); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GuideSessionErrorPacket composes one guide-session error response.
type GuideSessionErrorPacket struct {
	// ErrorCode stores the guide-session error code.
	ErrorCode int32
}

// PacketID returns protocol packet identifier.
func (p GuideSessionErrorPacket) PacketID() uint16 { return GuideSessionErrorPacketID }

// Encode serializes the guide-session error.
func (p GuideSessionErrorPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ErrorCode)
	return w.Bytes(), nil
}

// PendingGuideTicketData stores guide/reporting pending ticket details.
type PendingGuideTicketData struct {
	// Type stores the pending ticket kind.
	Type int32
	// SecondsAgo stores the age of the pending ticket.
	SecondsAgo int32
	// IsGuide stores whether the other party is a guide helper.
	IsGuide bool
	// OtherPartyName stores the other party display name.
	OtherPartyName string
	// OtherPartyFigure stores the other party figure string.
	OtherPartyFigure string
	// Description stores the pending description.
	Description string
	// RoomName stores the room context name.
	RoomName string
}

// GuideReportingStatusPacket composes guide/reporting status data.
type GuideReportingStatusPacket struct {
	// StatusCode stores the guide/reporting status code.
	StatusCode int32
	// PendingTicket stores optional pending-ticket details.
	PendingTicket PendingGuideTicketData
}

// PacketID returns protocol packet identifier.
func (p GuideReportingStatusPacket) PacketID() uint16 { return GuideReportingStatusPacketID }

// Encode serializes the guide/reporting status payload.
func (p GuideReportingStatusPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.StatusCode)
	w.WriteInt32(p.PendingTicket.Type)
	w.WriteInt32(p.PendingTicket.SecondsAgo)
	w.WriteBool(p.PendingTicket.IsGuide)
	for _, value := range []string{p.PendingTicket.OtherPartyName, p.PendingTicket.OtherPartyFigure, p.PendingTicket.Description, p.PendingTicket.RoomName} {
		if err := w.WriteString(value); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
