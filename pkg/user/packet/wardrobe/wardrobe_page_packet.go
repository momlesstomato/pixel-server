package wardrobe

import "github.com/momlesstomato/pixel-server/core/codec"

// SlotEntry defines one wardrobe slot packet entry payload.
type SlotEntry struct {
	// SlotID stores slot identifier.
	SlotID int32
	// Figure stores slot figure payload.
	Figure string
	// Gender stores slot gender payload.
	Gender string
}

// UserWardrobePagePacket defines user.wardrobe_page packet payload.
type UserWardrobePagePacket struct {
	// PageID stores requested wardrobe page index.
	PageID int32
	// Slots stores wardrobe slot entries for the page.
	Slots []SlotEntry
}

// PacketID returns protocol packet identifier.
func (packet UserWardrobePagePacket) PacketID() uint16 { return UserWardrobePagePacketID }

// Encode serializes packet body payload.
func (packet UserWardrobePagePacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(packet.PageID)
	writer.WriteInt32(int32(len(packet.Slots)))
	for _, slot := range packet.Slots {
		writer.WriteInt32(slot.SlotID)
		if err := writer.WriteString(slot.Figure); err != nil {
			return nil, err
		}
		if err := writer.WriteString(slot.Gender); err != nil {
			return nil, err
		}
	}
	return writer.Bytes(), nil
}
