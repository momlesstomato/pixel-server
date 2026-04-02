package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// defaultWrapperIDs holds sprite identifiers for gift wrapping paper items.
var defaultWrapperIDs = []int32{3372, 3373, 3374, 3375, 3376, 3377, 3378, 3379, 3380, 3381}

// defaultBoxTypes holds available gift box style identifiers.
var defaultBoxTypes = []int32{0, 1, 2, 3, 4, 5, 6}

// defaultRibbonTypes holds available ribbon style identifiers.
var defaultRibbonTypes = []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// defaultGiftFurniIDs holds sprite identifiers for giftable furniture items.
var defaultGiftFurniIDs = []int32{187, 188, 189, 190, 191, 192, 193}

// GiftWrappingConfigPacket defines catalog.gift_wrapping_config (s2c 2234) payload.
type GiftWrappingConfigPacket struct {
	// Enabled reports whether gifting is enabled.
	Enabled bool
	// SpecialPrice stores the credit cost for gift wrapping.
	SpecialPrice int32
	// WrapperIDs stores sprite identifiers for wrapping paper items.
	WrapperIDs []int32
	// BoxTypes stores available gift box style identifiers.
	BoxTypes []int32
	// RibbonTypes stores available ribbon style identifiers.
	RibbonTypes []int32
	// GiftFurniIDs stores sprite identifiers for giftable furniture items.
	GiftFurniIDs []int32
}

// DefaultGiftWrappingConfig returns a GiftWrappingConfigPacket with standard defaults.
func DefaultGiftWrappingConfig() GiftWrappingConfigPacket {
	return GiftWrappingConfigPacket{
		Enabled: true, SpecialPrice: 1,
		WrapperIDs:   defaultWrapperIDs,
		BoxTypes:     defaultBoxTypes,
		RibbonTypes:  defaultRibbonTypes,
		GiftFurniIDs: defaultGiftFurniIDs,
	}
}

// PacketID returns protocol packet identifier.
func (p GiftWrappingConfigPacket) PacketID() uint16 { return GiftWrappingConfigResponsePacketID }

// Encode serializes gift wrapping configuration into packet body.
func (p GiftWrappingConfigPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.Enabled)
	w.WriteInt32(p.SpecialPrice)
	w.WriteInt32(int32(len(p.WrapperIDs)))
	for _, id := range p.WrapperIDs {
		w.WriteInt32(id)
	}
	w.WriteInt32(int32(len(p.BoxTypes)))
	for _, bt := range p.BoxTypes {
		w.WriteInt32(bt)
	}
	w.WriteInt32(int32(len(p.RibbonTypes)))
	for _, rt := range p.RibbonTypes {
		w.WriteInt32(rt)
	}
	w.WriteInt32(int32(len(p.GiftFurniIDs)))
	for _, fi := range p.GiftFurniIDs {
		w.WriteInt32(fi)
	}
	return w.Bytes(), nil
}
