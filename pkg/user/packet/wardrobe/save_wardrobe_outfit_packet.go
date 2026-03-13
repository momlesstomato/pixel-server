package wardrobe

import "github.com/momlesstomato/pixel-server/core/codec"

// UserSaveWardrobeOutfitPacket defines user.save_wardrobe_outfit packet payload.
type UserSaveWardrobeOutfitPacket struct {
	// SlotID stores target slot identifier.
	SlotID int32
	// Figure stores saved avatar figure value.
	Figure string
	// Gender stores saved avatar gender value.
	Gender string
}

// PacketID returns protocol packet identifier.
func (packet UserSaveWardrobeOutfitPacket) PacketID() uint16 { return UserSaveWardrobeOutfitPacketID }

// Encode serializes packet body payload.
func (packet UserSaveWardrobeOutfitPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.SlotID)
	if err := writer.WriteString(packet.Figure); err != nil {
		return nil, err
	}
	if err := writer.WriteString(packet.Gender); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// Decode parses packet body payload.
func (packet *UserSaveWardrobeOutfitPacket) Decode(payload []byte) error {
	reader := codec.NewReader(payload)
	value, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	packet.SlotID = value
	if packet.Figure, err = reader.ReadString(); err != nil {
		return err
	}
	packet.Gender, err = reader.ReadString()
	return err
}
