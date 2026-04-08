package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
	roomdomain "github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// RoomAmbassadorAlertPacket decodes a targeted ambassador alert request.
type RoomAmbassadorAlertPacket struct {
	// UserID stores the target user identifier.
	UserID int32
}

// Decode reads fields from the packet body.
func (p *RoomAmbassadorAlertPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	userID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = userID
	return nil
}

// ModToolRequestRoomInfoPacket decodes a moderator room info request.
type ModToolRequestRoomInfoPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
}

// Decode reads fields from the packet body.
func (p *ModToolRequestRoomInfoPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	return nil
}

// ModToolChangeRoomSettingsPacket decodes a moderator room settings action request.
type ModToolChangeRoomSettingsPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
	// LockDoor indicates whether the room should be locked.
	LockDoor int32
	// ChangeTitle indicates whether the room title should be changed.
	ChangeTitle int32
	// KickUsers indicates whether room users should be removed.
	KickUsers int32
}

// Decode reads fields from the packet body.
func (p *ModToolChangeRoomSettingsPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	lockDoor, err := r.ReadInt32()
	if err != nil {
		return err
	}
	changeTitle, err := r.ReadInt32()
	if err != nil {
		return err
	}
	kickUsers, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	p.LockDoor = lockDoor
	p.ChangeTitle = changeTitle
	p.KickUsers = kickUsers
	return nil
}

// ModToolRequestRoomChatlogPacket decodes a moderator room chatlog request.
type ModToolRequestRoomChatlogPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
}

// Decode reads fields from the packet body.
func (p *ModToolRequestRoomChatlogPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	if _, err := r.ReadInt32(); err != nil {
		return err
	}
	roomID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.RoomID = roomID
	return nil
}

// ModToolUserInfoPacket decodes a moderator user info request.
type ModToolUserInfoPacket struct {
	// UserID stores the target user identifier.
	UserID int32
}

// Decode reads fields from the packet body.
func (p *ModToolUserInfoPacket) Decode(body []byte) error {
	r := codec.NewReader(body)
	userID, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.UserID = userID
	return nil
}

// ModToolRoomInfoPacket composes moderator room info payload.
type ModToolRoomInfoPacket struct {
	// RoomID stores the target room identifier.
	RoomID int32
	// UserCount stores live room occupants.
	UserCount int32
	// OwnerInRoom stores whether the owner is present.
	OwnerInRoom bool
	// OwnerID stores the owner user identifier.
	OwnerID int32
	// OwnerName stores the owner username.
	OwnerName string
	// Exists stores whether room metadata is available.
	Exists bool
	// Name stores the room name.
	Name string
	// Description stores the room description.
	Description string
	// Tags stores room tags.
	Tags []string
}

// PacketID returns protocol packet identifier.
func (p ModToolRoomInfoPacket) PacketID() uint16 { return ModToolRoomInfoComposerID }

// Encode serializes the moderator room info payload.
func (p ModToolRoomInfoPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.RoomID)
	w.WriteInt32(p.UserCount)
	w.WriteBool(p.OwnerInRoom)
	w.WriteInt32(p.OwnerID)
	if err := w.WriteString(p.OwnerName); err != nil {
		return nil, err
	}
	w.WriteBool(p.Exists)
	if !p.Exists {
		return w.Bytes(), nil
	}
	for _, value := range []string{p.Name, p.Description} {
		if err := w.WriteString(value); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Tags)))
	for _, tag := range p.Tags {
		if err := w.WriteString(tag); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// ModeratorUserInfoPacket composes moderator user info payload.
type ModeratorUserInfoPacket struct {
	// UserID stores the user identifier.
	UserID int32
	// Username stores the username.
	Username string
	// Figure stores the avatar figure.
	Figure string
	// RegistrationAgeMinutes stores account age in minutes.
	RegistrationAgeMinutes int32
	// MinutesSinceLastLogin stores minutes since last login.
	MinutesSinceLastLogin int32
	// Online stores whether the user is online.
	Online bool
	// CFHCount stores number of submitted calls for help.
	CFHCount int32
	// AbusiveCFHCount stores number of abusive calls for help.
	AbusiveCFHCount int32
	// CautionCount stores number of warning actions.
	CautionCount int32
	// BanCount stores number of ban actions.
	BanCount int32
	// TradingLockCount stores number of trade locks.
	TradingLockCount int32
	// TradingExpiryDate stores the trade lock expiry text.
	TradingExpiryDate string
	// LastPurchaseDate stores the last purchase text.
	LastPurchaseDate string
	// IdentityID stores the account identity identifier.
	IdentityID int32
	// IdentityRelatedBanCount stores linked-account ban count.
	IdentityRelatedBanCount int32
	// PrimaryEmailAddress stores the email or empty string.
	PrimaryEmailAddress string
	// UserClassification stores the role summary.
	UserClassification string
	// LastSanctionTime stores the last sanction text.
	LastSanctionTime string
	// SanctionAgeHours stores age in hours of the last sanction.
	SanctionAgeHours int32
}

// PacketID returns protocol packet identifier.
func (p ModeratorUserInfoPacket) PacketID() uint16 { return ModeratorUserInfoComposerID }

// Encode serializes the moderator user info payload.
func (p ModeratorUserInfoPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.UserID)
	for _, value := range []string{p.Username, p.Figure} {
		if err := w.WriteString(value); err != nil {
			return nil, err
		}
	}
	for _, value := range []int32{p.RegistrationAgeMinutes, p.MinutesSinceLastLogin} {
		w.WriteInt32(value)
	}
	w.WriteBool(p.Online)
	for _, value := range []int32{p.CFHCount, p.AbusiveCFHCount, p.CautionCount, p.BanCount, p.TradingLockCount} {
		w.WriteInt32(value)
	}
	for _, value := range []string{p.TradingExpiryDate, p.LastPurchaseDate} {
		if err := w.WriteString(value); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(p.IdentityID)
	w.WriteInt32(p.IdentityRelatedBanCount)
	for _, value := range []string{p.PrimaryEmailAddress, p.UserClassification, p.LastSanctionTime} {
		if err := w.WriteString(value); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(p.SanctionAgeHours)
	return w.Bytes(), nil
}

// ModeratorToolPreferencesPacket composes moderator tool preferences payload.
type ModeratorToolPreferencesPacket struct {
	// WindowX stores the left position.
	WindowX int32
	// WindowY stores the top position.
	WindowY int32
	// WindowWidth stores the width.
	WindowWidth int32
	// WindowHeight stores the height.
	WindowHeight int32
}

// PacketID returns protocol packet identifier.
func (p ModeratorToolPreferencesPacket) PacketID() uint16 { return ModeratorToolPreferencesComposerID }

// Encode serializes the moderator tool preferences payload.
func (p ModeratorToolPreferencesPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	for _, value := range []int32{p.WindowX, p.WindowY, p.WindowWidth, p.WindowHeight} {
		w.WriteInt32(value)
	}
	return w.Bytes(), nil
}

// RoomMutedPacket composes room muted state payload.
type RoomMutedPacket struct {
	// Muted stores whether the room is globally muted.
	Muted bool
}

// PacketID returns protocol packet identifier.
func (p RoomMutedPacket) PacketID() uint16 { return RoomMutedComposerID }

// Encode serializes the room muted state.
func (p RoomMutedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.Muted)
	return w.Bytes(), nil
}

// RoomChatlogLine stores one moderation chat log line.
type RoomChatlogLine struct {
	// Timestamp stores the display timestamp.
	Timestamp string
	// UserID stores the sender user identifier.
	UserID int32
	// Username stores the sender username.
	Username string
	// Message stores the chat message text.
	Message string
	// Highlighted stores whether Nitro should highlight the line.
	Highlighted bool
}

// ModToolRoomChatlogPacket composes moderator room chatlog payload.
type ModToolRoomChatlogPacket struct {
	// RoomID stores the room identifier.
	RoomID int32
	// RoomName stores the room name.
	RoomName string
	// Chatlog stores room chat lines.
	Chatlog []RoomChatlogLine
}

// PacketID returns protocol packet identifier.
func (p ModToolRoomChatlogPacket) PacketID() uint16 { return ModToolRoomChatlogComposerID }

// Encode serializes the moderator room chatlog payload.
func (p ModToolRoomChatlogPacket) Encode() ([]byte, error) {
	context := []moderationContextEntry{
		{Key: "roomName", Type: moderationContextString, StringValue: p.RoomName},
		{Key: "roomId", Type: moderationContextInt, IntValue: p.RoomID},
	}
	return encodeChatRecordPayload(moderationChatRecordRoom, context, p.Chatlog)
}

// ModeratorCFHChatlogPacket composes moderator CFH chatlog payload.
type ModeratorCFHChatlogPacket struct {
	// TicketID stores the ticket identifier.
	TicketID int32
	// ReporterID stores the caller identifier.
	ReporterID int32
	// ReportedID stores the reported user identifier.
	ReportedID int32
	// ChatRecordID stores the chat record identifier.
	ChatRecordID int32
	// RoomID stores the room identifier.
	RoomID int32
	// RoomName stores the room name.
	RoomName string
	// Chatlog stores room chat lines.
	Chatlog []RoomChatlogLine
}

// PacketID returns protocol packet identifier.
func (p ModeratorCFHChatlogPacket) PacketID() uint16 { return ModeratorCFHChatlogPacketID }

// Encode serializes the moderator CFH chatlog payload.
func (p ModeratorCFHChatlogPacket) Encode() ([]byte, error) {
	record, err := encodeChatRecordPayload(moderationChatRecordRoom, []moderationContextEntry{
		{Key: "roomName", Type: moderationContextString, StringValue: p.RoomName},
		{Key: "roomId", Type: moderationContextInt, IntValue: p.RoomID},
	}, p.Chatlog)
	if err != nil {
		return nil, err
	}
	w := codec.NewWriter()
	for _, value := range []int32{p.TicketID, p.ReporterID, p.ReportedID, p.ChatRecordID} {
		w.WriteInt32(value)
	}
	return append(w.Bytes(), record...), nil
}

type moderationContextType byte

const (
	moderationContextBool    moderationContextType = 0
	moderationContextInt     moderationContextType = 1
	moderationContextString  moderationContextType = 2
	moderationChatRecordRoom byte                  = 1
)

type moderationContextEntry struct {
	Key         string
	Type        moderationContextType
	BoolValue   bool
	IntValue    int32
	StringValue string
}

func encodeChatRecordPayload(recordType byte, context []moderationContextEntry, lines []RoomChatlogLine) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(recordType)
	writeUint16(buf, uint16(len(context)))
	for _, entry := range context {
		if err := writeString(buf, entry.Key); err != nil {
			return nil, err
		}
		buf.WriteByte(byte(entry.Type))
		switch entry.Type {
		case moderationContextBool:
			if entry.BoolValue {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		case moderationContextInt:
			writeInt32(buf, entry.IntValue)
		case moderationContextString:
			if err := writeString(buf, entry.StringValue); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown moderation context type %d", entry.Type)
		}
	}
	writeUint16(buf, uint16(len(lines)))
	for _, line := range lines {
		if err := writeString(buf, line.Timestamp); err != nil {
			return nil, err
		}
		writeInt32(buf, line.UserID)
		if err := writeString(buf, line.Username); err != nil {
			return nil, err
		}
		if err := writeString(buf, line.Message); err != nil {
			return nil, err
		}
		if line.Highlighted {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
	}
	return buf.Bytes(), nil
}

func writeInt32(buf *bytes.Buffer, value int32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(value))
	buf.Write(data)
}

func writeUint16(buf *bytes.Buffer, value uint16) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, value)
	buf.Write(data)
}

func writeString(buf *bytes.Buffer, value string) error {
	if len(value) > 65535 {
		return fmt.Errorf("string length exceeds uint16 max")
	}
	writeUint16(buf, uint16(len(value)))
	buf.WriteString(value)
	return nil
}

// _ suppresses unused import warnings for generated moderation packet helpers.
var _ roomdomain.ChatLogEntry
