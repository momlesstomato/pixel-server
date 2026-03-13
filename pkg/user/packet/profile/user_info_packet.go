package profile

import "github.com/momlesstomato/pixel-server/core/codec"

// UserInfoPacketID defines packet identifier for user.info.
const UserInfoPacketID uint16 = 2725

// UserInfoPacket defines user.info packet payload.
type UserInfoPacket struct {
	// UserID stores user identifier.
	UserID int32
	// Username stores account username.
	Username string
	// Figure stores avatar figure string.
	Figure string
	// Gender stores avatar gender marker.
	Gender string
	// Motto stores profile motto.
	Motto string
	// RealName stores profile real name.
	RealName string
	// DirectMail stores direct mail preference.
	DirectMail bool
	// RespectsReceived stores total received respects.
	RespectsReceived int32
	// RespectsRemaining stores remaining user respects for UTC day.
	RespectsRemaining int32
	// RespectsPetRemaining stores remaining pet respects for UTC day.
	RespectsPetRemaining int32
	// StreamPublishingAllowed stores stream publishing permission.
	StreamPublishingAllowed bool
	// LastAccessDate stores formatted last access date.
	LastAccessDate string
	// CanChangeName stores account rename capability.
	CanChangeName bool
	// SafetyLocked stores account safety lock marker.
	SafetyLocked bool
}

// PacketID returns protocol packet identifier.
func (packet UserInfoPacket) PacketID() uint16 { return UserInfoPacketID }

// Encode serializes packet body payload.
func (packet UserInfoPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.UserID)
	for _, value := range []string{packet.Username, packet.Figure, packet.Gender, packet.Motto, packet.RealName} {
		if err := writer.WriteString(value); err != nil {
			return nil, err
		}
	}
	writer.WriteBool(packet.DirectMail)
	writer.WriteInt32(packet.RespectsReceived)
	writer.WriteInt32(packet.RespectsRemaining)
	writer.WriteInt32(packet.RespectsPetRemaining)
	writer.WriteBool(packet.StreamPublishingAllowed)
	if err := writer.WriteString(packet.LastAccessDate); err != nil {
		return nil, err
	}
	writer.WriteBool(packet.CanChangeName)
	writer.WriteBool(packet.SafetyLocked)
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserInfoPacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	userID, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.UserID = userID
	fields := []*string{&packet.Username, &packet.Figure, &packet.Gender, &packet.Motto, &packet.RealName}
	for _, field := range fields {
		value, readErr := reader.ReadString()
		if readErr != nil {
			return readErr
		}
		*field = value
	}
	if packet.DirectMail, err = reader.ReadBool(); err != nil {
		return err
	}
	if packet.RespectsReceived, err = reader.ReadInt32(); err != nil {
		return err
	}
	if packet.RespectsRemaining, err = reader.ReadInt32(); err != nil {
		return err
	}
	if packet.RespectsPetRemaining, err = reader.ReadInt32(); err != nil {
		return err
	}
	if packet.StreamPublishingAllowed, err = reader.ReadBool(); err != nil {
		return err
	}
	if packet.LastAccessDate, err = reader.ReadString(); err != nil {
		return err
	}
	if packet.CanChangeName, err = reader.ReadBool(); err != nil {
		return err
	}
	packet.SafetyLocked, err = reader.ReadBool()
	return err
}
