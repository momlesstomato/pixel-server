package packet

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModeratorInitPacketEncode verifies moderator init encoding matches the ModeratorInitData wire format.
func TestModeratorInitPacketEncode(t *testing.T) {
	pkt := ModeratorInitPacket{
		MessageTemplates:    []string{"Scam", "Harassment"},
		CfhPermission:       true,
		ChatlogsPermission:  false,
		AlertPermission:     true,
		KickPermission:      false,
		BanPermission:       false,
		RoomAlertPermission: true,
		RoomKickPermission:  false,
	}
	assert.Equal(t, ModeratorInitPacketID, pkt.PacketID())
	body, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(body)
	issues, _ := r.ReadInt32()
	assert.Equal(t, int32(0), issues)
	tplCount, _ := r.ReadInt32()
	assert.Equal(t, int32(2), tplCount)
	t1, _ := r.ReadString()
	assert.Equal(t, "Scam", t1)
	t2, _ := r.ReadString()
	assert.Equal(t, "Harassment", t2)
	reserved, _ := r.ReadInt32()
	assert.Equal(t, int32(0), reserved)
	cfh, _ := r.ReadBool()
	assert.True(t, cfh)
	chatlogs, _ := r.ReadBool()
	assert.False(t, chatlogs)
	alert, _ := r.ReadBool()
	assert.True(t, alert)
	kick, _ := r.ReadBool()
	assert.False(t, kick)
	ban, _ := r.ReadBool()
	assert.False(t, ban)
	roomAlert, _ := r.ReadBool()
	assert.True(t, roomAlert)
	roomKick, _ := r.ReadBool()
	assert.False(t, roomKick)
	roomTpl, _ := r.ReadInt32()
	assert.Equal(t, int32(0), roomTpl)
}

// TestCallForHelpPacketDecode verifies CFH decoding.
func TestCallForHelpPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	_ = w.WriteString("Help me!")
	w.WriteInt32(3)
	w.WriteInt32(42)
	w.WriteInt32(100)
	pkt := &CallForHelpPacket{}
	require.NoError(t, pkt.Decode(w.Bytes()))
	assert.Equal(t, "Help me!", pkt.Message)
	assert.Equal(t, int32(3), pkt.Category)
	assert.Equal(t, int32(42), pkt.ReportedID)
	assert.Equal(t, int32(100), pkt.RoomID)
}

// TestSanctionTradeLockPacketDecode verifies trade lock decoding.
func TestSanctionTradeLockPacketDecode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(99)
	_ = w.WriteString("scam reason")
	w.WriteInt32(24)
	pkt := &SanctionTradeLockPacket{}
	require.NoError(t, pkt.Decode(w.Bytes()))
	assert.Equal(t, int32(99), pkt.UserID)
	assert.Equal(t, "scam reason", pkt.Message)
	assert.Equal(t, int32(24), pkt.Duration)
}

// TestCFHPendingPacketEncode verifies pending packet encoding.
func TestCFHPendingPacketEncode(t *testing.T) {
	pkt := CFHPendingPacket{Count: 5}
	assert.Equal(t, CFHPendingPacketID, pkt.PacketID())
	body, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(body)
	count, _ := r.ReadInt32()
	assert.Equal(t, int32(5), count)
}
