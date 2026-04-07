package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// DirectClubBuyAvailablePacket defines catalog.direct_sms_club_buy (s2c 195) payload.
type DirectClubBuyAvailablePacket struct {
	// PricePointURL stores the external SMS/direct-buy purchase URL.
	PricePointURL string
	// Market stores the market code returned to the client.
	Market string
	// LengthInDays stores the offer duration referenced by the availability check.
	LengthInDays int32
}

// PacketID returns the wire protocol packet identifier.
func (p DirectClubBuyAvailablePacket) PacketID() uint16 { return DirectClubBuyAvailableResponsePacketID }

// Encode serializes the direct-club-buy availability payload.
func (p DirectClubBuyAvailablePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.PricePointURL); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.Market); err != nil {
		return nil, err
	}
	w.WriteInt32(p.LengthInDays)
	return w.Bytes(), nil
}