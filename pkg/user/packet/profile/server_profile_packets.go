package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserSettingsPacketID defines packet identifier for user.settings.
const UserSettingsPacketID uint16 = 513

// UserHomeRoomPacketID defines packet identifier for user.home_room.
const UserHomeRoomPacketID uint16 = 2875

// UserFigurePacketID defines packet identifier for user.figure.
const UserFigurePacketID uint16 = 2429

// UserRespectReceivedPacketID defines packet identifier for user.respect_received.
const UserRespectReceivedPacketID uint16 = 2815

// UserSettingsPacket defines user.settings packet payload.
type UserSettingsPacket struct {
	VolumeSystem, VolumeFurni, VolumeTrax int32
	OldChat, RoomInvites, CameraFollow    bool
	Flags, ChatType                       int32
}

// PacketID returns protocol packet identifier.
func (packet UserSettingsPacket) PacketID() uint16 { return UserSettingsPacketID }

// Encode serializes packet body payload.
func (packet UserSettingsPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.VolumeSystem)
	writer.WriteInt32(packet.VolumeFurni)
	writer.WriteInt32(packet.VolumeTrax)
	writer.WriteBool(packet.OldChat)
	writer.WriteBool(packet.RoomInvites)
	writer.WriteBool(packet.CameraFollow)
	writer.WriteInt32(packet.Flags)
	writer.WriteInt32(packet.ChatType)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSettingsPacket) Decode(payload []byte) error {
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
	flag, boolErr := reader.ReadBool()
	if boolErr != nil {
		return boolErr
	}
	packet.OldChat = flag
	flag, boolErr = reader.ReadBool()
	if boolErr != nil {
		return boolErr
	}
	packet.RoomInvites = flag
	flag, boolErr = reader.ReadBool()
	if boolErr != nil {
		return boolErr
	}
	packet.CameraFollow = flag
	value, err = reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.Flags = value
	value, err = reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.ChatType = value
	return nil
}

// UserHomeRoomPacket defines user.home_room packet payload.
type UserHomeRoomPacket struct {
	HomeRoomID    int32
	RoomIDToEnter int32
}

// PacketID returns protocol packet identifier.
func (packet UserHomeRoomPacket) PacketID() uint16 { return UserHomeRoomPacketID }

// Encode serializes packet body payload.
func (packet UserHomeRoomPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.HomeRoomID)
	writer.WriteInt32(packet.RoomIDToEnter)
	return writer.Bytes(), nil
}

// UserFigurePacket defines user.figure packet payload.
type UserFigurePacket struct {
	Figure string
	Gender string
}

// PacketID returns protocol packet identifier.
func (packet UserFigurePacket) PacketID() uint16 { return UserFigurePacketID }

// Encode serializes packet body payload.
func (packet UserFigurePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Figure); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Gender); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// UserRespectReceivedPacket defines user.respect_received packet payload.
type UserRespectReceivedPacket struct {
	UserID           int32
	RespectsReceived int32
}

// PacketID returns protocol packet identifier.
func (packet UserRespectReceivedPacket) PacketID() uint16 { return UserRespectReceivedPacketID }

// Encode serializes packet body payload.
func (packet UserRespectReceivedPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	writer.WriteInt32(packet.RespectsReceived)
	return writer.Bytes(), nil
}
