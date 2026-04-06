package packet

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubscriptionResponsePacketEncodeIncludesMinutesSinceModified verifies that
// MinutesSinceLastModified is always written so the Nitro client's bytesAvailable
// check succeeds and the HC Center can resolve %joindate%.
func TestSubscriptionResponsePacketEncodeIncludesMinutesSinceModified(t *testing.T) {
	pkt := SubscriptionResponsePacket{
		ProductName: "club_habbo", DaysToPeriodEnd: 365, MemberPeriods: 0,
		PeriodsAhead: 0, ResponseType: 1, HasEverBeenMember: true, IsVIP: true,
		PastClubDays: 0, PastVIPDays: 0, MinutesUntilExpiration: 525600,
		MinutesSinceLastModified: 42,
	}
	data, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(data)
	_, _ = r.ReadString()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadBool()
	_, _ = r.ReadBool()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	minutesSince, err := r.ReadInt32()
	require.NoError(t, err)
	assert.Equal(t, int32(42), minutesSince)
}

// TestSubscriptionResponsePacketID verifies the packet identifier.
func TestSubscriptionResponsePacketID(t *testing.T) {
	pkt := SubscriptionResponsePacket{}
	assert.Equal(t, uint16(954), pkt.PacketID())
}

// TestHCExtendOfferPacketID verifies packet ID is set correctly.
func TestHCExtendOfferPacketID(t *testing.T) {
	pkt := HCExtendOfferPacket{}
	assert.Equal(t, uint16(3964), pkt.PacketID())
}

// TestHCExtendOfferPacketEncode verifies the extend offer wire format.
func TestHCExtendOfferPacketEncode(t *testing.T) {
	pkt := HCExtendOfferPacket{
		Offer: domain.ClubOffer{
			ID: 1, Name: "club_habbo", Credits: 10, Points: 0,
			PointsType: 0, Days: 31, Giftable: false, OfferType: "HC",
		},
		SubscriptionDaysLeft: 14,
	}
	data, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(data)
	offerID, _ := r.ReadInt32()
	assert.Equal(t, int32(1), offerID)
	name, _ := r.ReadString()
	assert.Equal(t, "club_habbo", name)
	_, _ = r.ReadBool()
	credits, _ := r.ReadInt32()
	assert.Equal(t, int32(10), credits)
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadBool()
	months, _ := r.ReadInt32()
	assert.Equal(t, int32(1), months)
	extraDays, _ := r.ReadInt32()
	assert.Equal(t, int32(0), extraDays)
	_, _ = r.ReadBool()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	origPrice, _ := r.ReadInt32()
	assert.Equal(t, int32(10), origPrice)
	_, _ = r.ReadInt32()
	_, _ = r.ReadInt32()
	daysLeft, _ := r.ReadInt32()
	assert.Equal(t, int32(14), daysLeft)
}
