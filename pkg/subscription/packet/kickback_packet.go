package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// KickbackInfoPacket defines user.kickback_info (s2c 3277) payload.
type KickbackInfoPacket struct {
	// CurrentHCStreak stores the user's current HC streak count.
	CurrentHCStreak int32
	// FirstSubscriptionDate stores the first HC join date string.
	FirstSubscriptionDate string
	// KickbackPercentage stores the payday percentage as a float64.
	KickbackPercentage float64
	// TotalCreditsMissed stores credits missed before the current streak.
	TotalCreditsMissed int32
	// TotalCreditsRewarded stores total rewarded credits.
	TotalCreditsRewarded int32
	// TotalCreditsSpent stores total credits spent in the current period.
	TotalCreditsSpent int32
	// CreditRewardForStreakBonus stores reward credits earned from streak bonus.
	CreditRewardForStreakBonus int32
	// CreditRewardForMonthlySpent stores reward credits earned from monthly spend.
	CreditRewardForMonthlySpent int32
	// TimeUntilPayday stores seconds remaining until payday.
	TimeUntilPayday int32
}

// PacketID returns the wire protocol packet identifier.
func (p KickbackInfoPacket) PacketID() uint16 { return KickbackInfoResponsePacketID }

// Encode serializes kickback information into packet body.
func (p KickbackInfoPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.CurrentHCStreak)
	if err := w.WriteString(p.FirstSubscriptionDate); err != nil {
		return nil, err
	}
	w.WriteFloat64(p.KickbackPercentage)
	w.WriteInt32(p.TotalCreditsMissed)
	w.WriteInt32(p.TotalCreditsRewarded)
	w.WriteInt32(p.TotalCreditsSpent)
	w.WriteInt32(p.CreditRewardForStreakBonus)
	w.WriteInt32(p.CreditRewardForMonthlySpent)
	w.WriteInt32(p.TimeUntilPayday)
	return w.Bytes(), nil
}