package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserUpdateMottoPacketID defines packet identifier for user.update_motto.
const UserUpdateMottoPacketID uint16 = 2228

// UserUpdateFigurePacketID defines packet identifier for user.update_figure.
const UserUpdateFigurePacketID uint16 = 2730

// UserSetHomeRoomPacketID defines packet identifier for user.set_home_room.
const UserSetHomeRoomPacketID uint16 = 1740

// UserRespectPacketID defines packet identifier for user.respect.
const UserRespectPacketID uint16 = 2694

// UserUpdateMottoPacket defines user.update_motto packet payload.
type UserUpdateMottoPacket struct{ Motto string }

// PacketID returns protocol packet identifier.
func (packet UserUpdateMottoPacket) PacketID() uint16 { return UserUpdateMottoPacketID }

// Encode serializes packet body payload.
func (packet UserUpdateMottoPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Motto); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserUpdateMottoPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadString()
	if err == nil {
		packet.Motto = value
	}
	return err
}

// UserUpdateFigurePacket defines user.update_figure packet payload.
type UserUpdateFigurePacket struct {
	Gender string
	Figure string
}

// PacketID returns protocol packet identifier.
func (packet UserUpdateFigurePacket) PacketID() uint16 { return UserUpdateFigurePacketID }

// Encode serializes packet body payload.
func (packet UserUpdateFigurePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := writer.WriteString(packet.Gender); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Figure); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserUpdateFigurePacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	value, err := reader.ReadString()
	if err != nil {
		return err
	}
	packet.Gender = value
	packet.Figure, err = reader.ReadString()
	return err
}

// UserSetHomeRoomPacket defines user.set_home_room packet payload.
type UserSetHomeRoomPacket struct{ RoomID int32 }

// PacketID returns protocol packet identifier.
func (packet UserSetHomeRoomPacket) PacketID() uint16 { return UserSetHomeRoomPacketID }

// Encode serializes packet body payload.
func (packet UserSetHomeRoomPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.RoomID)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSetHomeRoomPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		packet.RoomID = value
	}
	return err
}

// UserRespectPacket defines user.respect packet payload.
type UserRespectPacket struct{ UserID int32 }

// PacketID returns protocol packet identifier.
func (packet UserRespectPacket) PacketID() uint16 { return UserRespectPacketID }

// Encode serializes packet body payload.
func (packet UserRespectPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserRespectPacket) Decode(payload []byte) error {
	value, err := codec.NewReader(payload).ReadInt32()
	if err == nil {
		packet.UserID = value
	}
	return err
}
