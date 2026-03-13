package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserPermissionsPacketID defines packet identifier for user.permissions.
const UserPermissionsPacketID uint16 = 411

// UserPerksPacketID defines packet identifier for user.perks.
const UserPerksPacketID uint16 = 2586

// UserPermissionsPacket defines user.permissions packet payload.
type UserPermissionsPacket struct {
	ClubLevel     int32
	SecurityLevel int32
	IsAmbassador  bool
}

// PacketID returns protocol packet identifier.
func (packet UserPermissionsPacket) PacketID() uint16 { return UserPermissionsPacketID }

// Encode serializes packet body payload.
func (packet UserPermissionsPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.ClubLevel)
	writer.WriteInt32(packet.SecurityLevel)
	writer.WriteBool(packet.IsAmbassador)
	return writer.Bytes(), nil
}

// PerkEntry defines one user perk record payload.
type PerkEntry struct {
	Code         string
	ErrorMessage string
	IsAllowed    bool
}

// UserPerksPacket defines user.perks packet payload.
type UserPerksPacket struct{ Entries []PerkEntry }

// PacketID returns protocol packet identifier.
func (packet UserPerksPacket) PacketID() uint16 { return UserPerksPacketID }

// Encode serializes packet body payload.
func (packet UserPerksPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(int32(len(packet.Entries)))
	for _, entry := range packet.Entries {
		if err := writer.WriteString(entry.Code); err != nil {
			return nil, err
		}
		if err := writer.WriteString(entry.ErrorMessage); err != nil {
			return nil, err
		}
		writer.WriteBool(entry.IsAllowed)
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserPerksPacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	count, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	entries := make([]PerkEntry, 0, count)
	for index := int32(0); index < count; index++ {
		code, codeErr := reader.ReadString()
		if codeErr != nil {
			return codeErr
		}
		errorMessage, messageErr := reader.ReadString()
		if messageErr != nil {
			return messageErr
		}
		allowed, allowedErr := reader.ReadBool()
		if allowedErr != nil {
			return allowedErr
		}
		entries = append(entries, PerkEntry{Code: code, ErrorMessage: errorMessage, IsAllowed: allowed})
	}
	packet.Entries = entries
	return nil
}
