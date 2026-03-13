package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserSettingsVolumePacketID defines packet identifier for user.settings_volume.
const UserSettingsVolumePacketID uint16 = 1367

// UserSettingsRoomInvitesPacketID defines packet identifier for user.settings_room_invites.
const UserSettingsRoomInvitesPacketID uint16 = 65534

// UserSettingsOldChatPacketID defines packet identifier for user.settings_old_chat.
const UserSettingsOldChatPacketID uint16 = 65535

// UserSettingsVolumePacket defines user.settings_volume packet payload.
type UserSettingsVolumePacket struct {
	VolumeSystem int32
	VolumeFurni  int32
	VolumeTrax   int32
}

// PacketID returns protocol packet identifier.
func (packet UserSettingsVolumePacket) PacketID() uint16 { return UserSettingsVolumePacketID }

// Encode serializes packet body payload.
func (packet UserSettingsVolumePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.VolumeSystem)
	writer.WriteInt32(packet.VolumeFurni)
	writer.WriteInt32(packet.VolumeTrax)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSettingsVolumePacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	value, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.VolumeSystem = value
	value, err = reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.VolumeFurni = value
	value, err = reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.VolumeTrax = value
	return nil
}

// UserSettingsRoomInvitesPacket defines user.settings_room_invites payload.
type UserSettingsRoomInvitesPacket struct{ Enabled bool }

// PacketID returns protocol packet identifier.
func (packet UserSettingsRoomInvitesPacket) PacketID() uint16 { return UserSettingsRoomInvitesPacketID }

// Encode serializes packet body payload.
func (packet UserSettingsRoomInvitesPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteBool(packet.Enabled)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSettingsRoomInvitesPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadBool()
	if err == nil {
		packet.Enabled = value
	}
	return err
}

// UserSettingsOldChatPacket defines user.settings_old_chat payload.
type UserSettingsOldChatPacket struct{ Enabled bool }

// PacketID returns protocol packet identifier.
func (packet UserSettingsOldChatPacket) PacketID() uint16 { return UserSettingsOldChatPacketID }

// Encode serializes packet body payload.
func (packet UserSettingsOldChatPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteBool(packet.Enabled)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSettingsOldChatPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadBool()
	if err == nil {
		packet.Enabled = value
	}
	return err
}
