package packet

import (
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// ClubOffersPacket defines catalog.club_offers (s2c 2405) payload.
type ClubOffersPacket struct {
	// Offers stores available club membership options.
	Offers []domain.ClubOffer
	// WindowID stores the catalog page identifier context.
	WindowID int32
}

// PacketID returns the wire protocol packet identifier.
func (p ClubOffersPacket) PacketID() uint16 { return ClubOffersResponsePacketID }

// Encode serializes club offers list into packet body.
func (p ClubOffersPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	now := time.Now().UTC()
	w.WriteInt32(int32(len(p.Offers)))
	for _, o := range p.Offers {
		w.WriteInt32(int32(o.ID))
		if err := w.WriteString(o.Name); err != nil {
			return nil, err
		}
		w.WriteBool(false)
		w.WriteInt32(int32(o.Credits))
		w.WriteInt32(int32(o.Points))
		w.WriteInt32(int32(o.PointsType))
		w.WriteBool(o.OfferType == "VIP")
		totalSeconds := int64(o.Days) * 86400
		months := totalSeconds / (86400 * 31)
		remainder := totalSeconds - months*(86400*31)
		extraDays := remainder / 86400
		remainder -= extraDays * 86400
		w.WriteInt32(int32(months))
		w.WriteInt32(int32(extraDays))
		w.WriteBool(o.Giftable)
		w.WriteInt32(int32(remainder))
		expiry := now.Add(time.Duration(totalSeconds) * time.Second)
		w.WriteInt32(int32(expiry.Year()))
		w.WriteInt32(int32(expiry.Month()))
		w.WriteInt32(int32(expiry.Day()))
	}
	w.WriteInt32(p.WindowID)
	return w.Bytes(), nil
}

// ClubGiftInfoPacket defines catalog.club_gift_info (s2c 619) payload.
type ClubGiftInfoPacket struct {
	// DaysUntilNextGift stores the remaining days.
	DaysUntilNextGift int32
	// GiftsAvailable stores the number of gifts available.
	GiftsAvailable int32
}

// PacketID returns the wire protocol packet identifier.
func (p ClubGiftInfoPacket) PacketID() uint16 { return ClubGiftInfoResponsePacketID }

// Encode serializes club gift eligibility info into packet body.
func (p ClubGiftInfoPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.DaysUntilNextGift)
	w.WriteInt32(p.GiftsAvailable)
	w.WriteInt32(0)
	w.WriteInt32(0)
	return w.Bytes(), nil
}
