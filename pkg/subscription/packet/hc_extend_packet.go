package packet

import (
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// HCExtendOfferPacket defines catalog.hc_extend_offer (s2c 3964) payload.
type HCExtendOfferPacket struct {
	// Offer stores the subscription offer presented for extension.
	Offer domain.ClubOffer
	// SubscriptionDaysLeft stores remaining days in the user's current subscription.
	SubscriptionDaysLeft int32
}

// PacketID returns the wire protocol packet identifier.
func (p HCExtendOfferPacket) PacketID() uint16 { return HCExtendOfferResponsePacketID }

// Encode serializes the HC extend offer into packet body matching ClubOfferExtendData.
func (p HCExtendOfferPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	o := p.Offer
	now := time.Now().UTC()
	totalSeconds := int64(o.Days) * 86400
	months := totalSeconds / (86400 * 31)
	remainder := totalSeconds - months*(86400*31)
	extraDays := remainder / 86400
	remainder -= extraDays * 86400
	expiry := now.Add(time.Duration(totalSeconds) * time.Second)
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
	w.WriteInt32(int32(remainder))
	w.WriteInt32(int32(expiry.Year()))
	w.WriteInt32(int32(expiry.Month()))
	w.WriteInt32(int32(expiry.Day()))
	w.WriteInt32(int32(o.Credits))
	w.WriteInt32(int32(o.Points))
	w.WriteInt32(int32(o.PointsType))
	w.WriteInt32(p.SubscriptionDaysLeft)
	return w.Bytes(), nil
}
