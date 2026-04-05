package packet

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModKickUserPacketDecode verifies kick packet decode round-trip.
func TestModKickUserPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(42)
	_ = w.WriteString("bad behavior")
	var pkt ModKickUserPacket
	err := pkt.Decode(w.Bytes())
	require.NoError(t, err)
	assert.Equal(t, int32(42), pkt.UserID)
	assert.Equal(t, "bad behavior", pkt.Message)
}

// TestModMuteUserPacketDecode verifies mute packet decode round-trip.
func TestModMuteUserPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(10)
	_ = w.WriteString("spamming")
	w.WriteInt32(60)
	var pkt ModMuteUserPacket
	err := pkt.Decode(w.Bytes())
	require.NoError(t, err)
	assert.Equal(t, int32(10), pkt.UserID)
	assert.Equal(t, "spamming", pkt.Message)
	assert.Equal(t, int32(60), pkt.Minutes)
}

// TestModBanUserPacketDecode verifies ban packet decode round-trip.
func TestModBanUserPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(99)
	_ = w.WriteString("cheating")
	w.WriteInt32(2)
	_ = w.WriteString("topic")
	w.WriteInt32(24)
	var pkt ModBanUserPacket
	err := pkt.Decode(w.Bytes())
	require.NoError(t, err)
	assert.Equal(t, int32(99), pkt.UserID)
	assert.Equal(t, "cheating", pkt.Message)
	assert.Equal(t, int32(2), pkt.BanType)
	assert.Equal(t, "topic", pkt.CfhTopic)
	assert.Equal(t, int32(24), pkt.Duration)
}

// TestModWarnUserPacketDecode verifies warn packet decode round-trip.
func TestModWarnUserPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(7)
	_ = w.WriteString("warning message")
	var pkt ModWarnUserPacket
	err := pkt.Decode(w.Bytes())
	require.NoError(t, err)
	assert.Equal(t, int32(7), pkt.UserID)
	assert.Equal(t, "warning message", pkt.Message)
}

// TestModKickUserPacketDecodeEmpty verifies empty body returns error.
func TestModKickUserPacketDecodeEmpty(t *testing.T) {
	var pkt ModKickUserPacket
	err := pkt.Decode(nil)
	assert.Error(t, err)
}

// TestPacketConstants verifies packet ID values.
func TestPacketConstants(t *testing.T) {
	assert.Equal(t, uint16(2582), ModKickUserPacketID)
	assert.Equal(t, uint16(1945), ModMuteUserPacketID)
	assert.Equal(t, uint16(2766), ModBanUserPacketID)
	assert.Equal(t, uint16(1840), ModWarnUserPacketID)
}
