package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// SubscriptionResponsePacket defines user.subscription (s2c 954) payload.
type SubscriptionResponsePacket struct {
	// ProductName stores club product identifier, e.g. "club_habbo".
	ProductName string
	// DaysToPeriodEnd stores days remaining in the current period.
	DaysToPeriodEnd int32
	// MemberPeriods stores number of complete subscription periods held.
	MemberPeriods int32
	// PeriodsAhead stores number of future periods already paid for.
	PeriodsAhead int32
	// ResponseType stores trigger code: 1=login, 2=purchase, 3=discount, 4=citizenship.
	ResponseType int32
	// HasEverBeenMember stores whether the user has ever held a subscription.
	HasEverBeenMember bool
	// IsVIP stores whether the user currently has a VIP subscription.
	IsVIP bool
	// PastClubDays stores total days as a club member historically.
	PastClubDays int32
	// PastVIPDays stores total days as a VIP member historically.
	PastVIPDays int32
	// MinutesUntilExpiration stores minutes until the current period expires.
	MinutesUntilExpiration int32
}

// PacketID returns protocol packet identifier.
func (p SubscriptionResponsePacket) PacketID() uint16 { return SubscriptionResponsePacketID }

// Encode serializes subscription state into packet body.
func (p SubscriptionResponsePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ProductName); err != nil {
		return nil, err
	}
	w.WriteInt32(p.DaysToPeriodEnd)
	w.WriteInt32(p.MemberPeriods)
	w.WriteInt32(p.PeriodsAhead)
	w.WriteInt32(p.ResponseType)
	w.WriteBool(p.HasEverBeenMember)
	w.WriteBool(p.IsVIP)
	w.WriteInt32(p.PastClubDays)
	w.WriteInt32(p.PastVIPDays)
	w.WriteInt32(p.MinutesUntilExpiration)
	return w.Bytes(), nil
}
