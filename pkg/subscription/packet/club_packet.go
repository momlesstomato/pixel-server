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
}

// PacketID returns the wire protocol packet identifier.
func (p ClubOffersPacket) PacketID() uint16 { return ClubOffersResponsePacketID }

// Encode serializes club offers list into packet body.
func (p ClubOffersPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	now := time.Now().UTC()
	w.WriteInt32(int32(len(p.Offers)))
	for _, o := range p.Offers {
		months := o.Days / 31
		extraDays := o.Days % 31
		w.WriteInt32(int32(o.ID))
		if err := w.WriteString(o.Name); err != nil {
			return nil, err
		}
		w.WriteBool(false)
		w.WriteInt32(int32(o.Credits))
		w.WriteInt32(int32(o.Points))
		w.WriteInt32(int32(o.PointsType))
		w.WriteBool(o.OfferType == "VIP")
		w.WriteInt32(int32(months))
		w.WriteInt32(int32(extraDays))
		w.WriteBool(o.Giftable)
		w.WriteInt32(int32(o.Days))
		expiry := now.AddDate(0, 0, o.Days)
		w.WriteInt32(int32(expiry.Year()))
		w.WriteInt32(int32(expiry.Month()))
		w.WriteInt32(int32(expiry.Day()))
	}
	return w.Bytes(), nil
}

// ClubGiftInfoPacket defines catalog.club_gift_info (s2c 619) payload.
type ClubGiftInfoPacket struct {
	// DaysUntilNextGift stores the remaining days.
	DaysUntilNextGift int32
	// GiftsAvailable stores the number of gifts available.
	GiftsAvailable int32
	// ActiveDays stores the user's current HC age in days.
	ActiveDays int32
	// Gifts stores all configured gift options.
	Gifts []domain.ClubGift
}

// PacketID returns the wire protocol packet identifier.
func (p ClubGiftInfoPacket) PacketID() uint16 { return ClubGiftInfoResponsePacketID }

// Encode serializes club gift eligibility info into packet body.
func (p ClubGiftInfoPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.DaysUntilNextGift)
	w.WriteInt32(p.GiftsAvailable)
	w.WriteInt32(int32(len(p.Gifts)))
	for _, gift := range p.Gifts {
		w.WriteInt32(int32(gift.ID))
		if err := w.WriteString(gift.Name); err != nil {
			return nil, err
		}
		w.WriteBool(false)
		w.WriteInt32(0)
		w.WriteInt32(0)
		w.WriteInt32(0)
		w.WriteBool(false)
		w.WriteInt32(1)
		if err := w.WriteString("s"); err != nil {
			return nil, err
		}
		w.WriteInt32(int32(gift.SpriteID))
		if err := w.WriteString(gift.ExtraData); err != nil {
			return nil, err
		}
		w.WriteInt32(1)
		w.WriteBool(false)
		w.WriteInt32(0)
		w.WriteBool(false)
		w.WriteBool(false)
		if err := w.WriteString(""); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Gifts)))
	for _, gift := range p.Gifts {
		w.WriteInt32(int32(gift.ID))
		w.WriteBool(gift.VIPOnly)
		w.WriteInt32(int32(gift.DaysRequired))
		w.WriteBool(p.GiftsAvailable > 0 && p.ActiveDays >= int32(gift.DaysRequired))
	}
	return w.Bytes(), nil
}

// ClubGiftSelectedPacket defines catalog.club_gift_selected (s2c) payload.
type ClubGiftSelectedPacket struct {
	// ProductCode stores the selected club gift name.
	ProductCode string
	// SpriteID stores the delivered furniture sprite identifier.
	SpriteID int32
	// ExtraData stores delivered item custom data.
	ExtraData string
}

// PacketID returns the wire protocol packet identifier.
func (p ClubGiftSelectedPacket) PacketID() uint16 { return ClubGiftSelectedResponsePacketID }

// Encode serializes one selected club gift result.
func (p ClubGiftSelectedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ProductCode); err != nil {
		return nil, err
	}
	w.WriteInt32(1)
	if err := w.WriteString("s"); err != nil {
		return nil, err
	}
	w.WriteInt32(p.SpriteID)
	if err := w.WriteString(p.ExtraData); err != nil {
		return nil, err
	}
	w.WriteInt32(1)
	w.WriteBool(false)
	return w.Bytes(), nil
}
