package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// ModeratorInitPacket composes the moderator tool init payload (s2c 2696).
type ModeratorInitPacket struct {
	// MessageTemplates stores flat list of moderation preset messages.
	MessageTemplates []string
	// RoomMessageTemplates stores flat list of room-alert preset messages.
	RoomMessageTemplates []string
	// CfhPermission indicates whether the user can handle CFH reports.
	CfhPermission bool
	// ChatlogsPermission indicates whether the user can view chatlogs.
	ChatlogsPermission bool
	// AlertPermission indicates whether the user can send user alerts.
	AlertPermission bool
	// KickPermission indicates whether the user can kick users.
	KickPermission bool
	// BanPermission indicates whether the user can ban users.
	BanPermission bool
	// RoomAlertPermission indicates whether the user can send room alerts.
	RoomAlertPermission bool
	// RoomKickPermission indicates whether the user can kick users from rooms.
	RoomKickPermission bool
}

// PacketID returns protocol packet identifier.
func (p ModeratorInitPacket) PacketID() uint16 { return ModeratorInitPacketID }

// Encode serializes the moderator init payload matching the ModeratorInitData wire format.
func (p ModeratorInitPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(0)
	w.WriteInt32(int32(len(p.MessageTemplates)))
	for _, t := range p.MessageTemplates {
		if err := w.WriteString(t); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(0)
	w.WriteBool(p.CfhPermission)
	w.WriteBool(p.ChatlogsPermission)
	w.WriteBool(p.AlertPermission)
	w.WriteBool(p.KickPermission)
	w.WriteBool(p.BanPermission)
	w.WriteBool(p.RoomAlertPermission)
	w.WriteBool(p.RoomKickPermission)
	w.WriteInt32(int32(len(p.RoomMessageTemplates)))
	for _, t := range p.RoomMessageTemplates {
		if err := w.WriteString(t); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}
