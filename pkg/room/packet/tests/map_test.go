package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFloorHeightMapComposer_Encode verifies heightmap encoding.
func TestFloorHeightMapComposer_Encode(t *testing.T) {
	pkt := packet.FloorHeightMapComposer{Scale: true, WallHeight: -1, Heightmap: "000\r000"}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.FloorHeightMapComposerID, pkt.PacketID())
}

// TestHeightMapComposer_Encode verifies stacking map encoding.
func TestHeightMapComposer_Encode(t *testing.T) {
	pkt := packet.HeightMapComposer{Width: 3, TotalTiles: 6, Heights: []int16{0, 256, 512, 0, 256, 512}}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.HeightMapComposerID, pkt.PacketID())
}

// TestHeightMapComposer_EmptyHeights verifies zero-tile map.
func TestHeightMapComposer_EmptyHeights(t *testing.T) {
	pkt := packet.HeightMapComposer{Width: 0, TotalTiles: 0}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

// TestRoomEntryInfoComposer_Encode verifies entry info encoding.
func TestRoomEntryInfoComposer_Encode(t *testing.T) {
	pkt := packet.RoomEntryInfoComposer{RoomID: 5, IsOwner: true}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomEntryInfoComposerID, pkt.PacketID())
}

// TestRoomVisualizationComposer_Encode verifies visualization settings.
func TestRoomVisualizationComposer_Encode(t *testing.T) {
	pkt := packet.RoomVisualizationComposer{WallsHidden: true, WallThickness: 1, FloorThickness: 2}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomVisualizationComposerID, pkt.PacketID())
}

// TestFurnitureAliasesComposer_Encode verifies empty alias map.
func TestFurnitureAliasesComposer_Encode(t *testing.T) {
	pkt := packet.FurnitureAliasesComposer{}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.FurnitureAliasesComposerID, pkt.PacketID())
}

// TestCantConnectComposer_Encode verifies error code encoding.
func TestCantConnectComposer_Encode(t *testing.T) {
	pkt := packet.CantConnectComposer{ErrorCode: 4}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.CantConnectComposerID, pkt.PacketID())
}

// TestGetRoomSettingsPacket_Decode verifies room settings request decoding.
func TestGetRoomSettingsPacket_Decode(t *testing.T) {
	src := packet.RoomSettingsSavedComposer{RoomID: 42}
	body, err := src.Encode()
	require.NoError(t, err)
	var pkt packet.GetRoomSettingsPacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, int32(42), pkt.RoomID)
	assert.Equal(t, packet.GetRoomSettingsPacketID, pkt.PacketID())
}

// TestRoomSettingsSavedComposer_Encode verifies save confirmation encoding.
func TestRoomSettingsSavedComposer_Encode(t *testing.T) {
	pkt := packet.RoomSettingsSavedComposer{RoomID: 7}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomSettingsSavedComposerID, pkt.PacketID())
}

// TestRoomSettingsComposer_Encode verifies full settings payload encoding.
func TestRoomSettingsComposer_Encode(t *testing.T) {
	room := domain.Room{
		ID: 1, Name: "Test", Description: "Desc",
		State: domain.AccessPassword, Password: "hash",
		MaxUsers: 25, CategoryID: 2, Tags: []string{"fun"},
		TradeMode: 1, AllowPets: true, AllowTrading: false,
		WallThickness: 1, FloorThickness: 0, WallHeight: -1,
	}
	pkt := packet.RoomSettingsComposer{Room: room}
	body, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(body)
	roomID, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	name, readErr := r.ReadString()
	require.NoError(t, readErr)
	description, readErr := r.ReadString()
	require.NoError(t, readErr)
	doorMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	categoryID, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	maxUsers, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	maxUsersLimit, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	tagCount, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	tag, readErr := r.ReadString()
	require.NoError(t, readErr)
	tradeMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	allowPets, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	allowFoodConsume, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	allowWalkThrough, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	hideWalls, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	wallThickness, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	floorThickness, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	chatMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	chatWeight, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	chatSpeed, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	chatDistance, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	chatProtection, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	allowNavigatorDynamicCats, readErr := r.ReadBool()
	require.NoError(t, readErr)
	muteMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	kickMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	banMode, readErr := r.ReadInt32()
	require.NoError(t, readErr)
	assert.Equal(t, int32(1), roomID)
	assert.Equal(t, "Test", name)
	assert.Equal(t, "Desc", description)
	assert.Equal(t, int32(2), doorMode)
	assert.Equal(t, int32(2), categoryID)
	assert.Equal(t, int32(25), maxUsers)
	assert.Equal(t, int32(25), maxUsersLimit)
	assert.Equal(t, int32(1), tagCount)
	assert.Equal(t, "fun", tag)
	assert.Equal(t, int32(1), tradeMode)
	assert.Equal(t, int32(1), allowPets)
	assert.Equal(t, int32(0), allowFoodConsume)
	assert.Equal(t, int32(0), allowWalkThrough)
	assert.Equal(t, int32(0), hideWalls)
	assert.Equal(t, int32(1), wallThickness)
	assert.Equal(t, int32(0), floorThickness)
	assert.Equal(t, int32(0), chatMode)
	assert.Equal(t, int32(0), chatWeight)
	assert.Equal(t, int32(0), chatSpeed)
	assert.Equal(t, int32(0), chatDistance)
	assert.Equal(t, int32(0), chatProtection)
	assert.False(t, allowNavigatorDynamicCats)
	assert.Equal(t, int32(1), muteMode)
	assert.Equal(t, int32(1), kickMode)
	assert.Equal(t, int32(1), banMode)
	assert.Equal(t, packet.RoomSettingsComposerID, pkt.PacketID())
}

// TestSaveRoomSettingsPacket_Decode verifies Nitro room settings save decoding.
func TestSaveRoomSettingsPacket_Decode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(1)
	require.NoError(t, w.WriteString("Test"))
	require.NoError(t, w.WriteString("Desc"))
	w.WriteInt32(2)
	require.NoError(t, w.WriteString("secret"))
	w.WriteInt32(20)
	w.WriteInt32(6)
	w.WriteInt32(2)
	require.NoError(t, w.WriteString("fun"))
	require.NoError(t, w.WriteString("build"))
	w.WriteInt32(1)
	w.WriteBool(true)
	w.WriteBool(false)
	w.WriteBool(true)
	w.WriteBool(false)
	w.WriteInt32(1)
	w.WriteInt32(-1)
	w.WriteInt32(1)
	w.WriteInt32(2)
	w.WriteInt32(0)
	w.WriteInt32(3)
	w.WriteInt32(1)
	w.WriteInt32(2)
	w.WriteInt32(50)
	w.WriteInt32(1)
	var pkt packet.SaveRoomSettingsPacket
	require.NoError(t, pkt.Decode(w.Bytes()))
	assert.Equal(t, int32(1), pkt.RoomID)
	assert.Equal(t, "Test", pkt.Name)
	assert.Equal(t, "Desc", pkt.Description)
	assert.Equal(t, int32(2), pkt.State)
	assert.Equal(t, "secret", pkt.Password)
	assert.Equal(t, int32(20), pkt.MaxUsers)
	assert.Equal(t, int32(6), pkt.CategoryID)
	assert.Equal(t, []string{"fun", "build"}, pkt.Tags)
	assert.Equal(t, int32(1), pkt.TradeMode)
	assert.True(t, pkt.AllowPets)
	assert.False(t, pkt.AllowFoodConsume)
	assert.True(t, pkt.AllowWalkThrough)
	assert.False(t, pkt.HideWalls)
	assert.Equal(t, int32(1), pkt.WallThickness)
	assert.Equal(t, int32(-1), pkt.FloorThickness)
	assert.Equal(t, int32(1), pkt.MuteMode)
	assert.Equal(t, int32(2), pkt.KickMode)
	assert.Equal(t, int32(0), pkt.BanMode)
	assert.Equal(t, int32(3), pkt.ChatMode)
	assert.Equal(t, int32(1), pkt.ChatWeight)
	assert.Equal(t, int32(2), pkt.ChatSpeed)
	assert.Equal(t, int32(50), pkt.ChatDistance)
	assert.Equal(t, int32(1), pkt.ChatProtection)
}

// TestIntToAccessState_RoundTrip verifies integer to AccessState conversion.
func TestIntToAccessState_RoundTrip(t *testing.T) {
	assert.Equal(t, domain.AccessOpen, packet.IntToAccessState(0))
	assert.Equal(t, domain.AccessLocked, packet.IntToAccessState(1))
	assert.Equal(t, domain.AccessPassword, packet.IntToAccessState(2))
	assert.Equal(t, domain.AccessInvisible, packet.IntToAccessState(3))
	assert.Equal(t, domain.AccessOpen, packet.IntToAccessState(99))
}

// TestGetBannedUsersPacket_Decode verifies ban list request decoding.
func TestGetBannedUsersPacket_Decode(t *testing.T) {
	enc := packet.RoomForwardComposer{RoomID: 5}
	body, err := enc.Encode()
	require.NoError(t, err)
	var pkt packet.GetBannedUsersPacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, int32(5), pkt.RoomID)
	assert.Equal(t, packet.GetBannedUsersPacketID, pkt.PacketID())
}

// TestBannedUsersComposer_EncodeEmpty verifies empty ban list encoding.
func TestBannedUsersComposer_EncodeEmpty(t *testing.T) {
	pkt := packet.BannedUsersComposer{RoomID: 3, Bans: nil}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.BannedUsersComposerID, pkt.PacketID())
}

// TestBannedUsersComposer_EncodeWithEntries verifies encoding with entries.
func TestBannedUsersComposer_EncodeWithEntries(t *testing.T) {
	pkt := packet.BannedUsersComposer{
		RoomID: 1,
		Bans:   []packet.BannedUserEntry{{UserID: 10, Username: "alice"}, {UserID: 20, Username: "bob"}},
	}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

// TestUnbanUserPacket_Decode verifies unban request decoding.
func TestUnbanUserPacket_Decode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(7)
	w.WriteInt32(99)
	body := w.Bytes()
	var dec packet.UnbanUserPacket
	require.NoError(t, dec.Decode(body))
	assert.Equal(t, int32(7), dec.UserID)
	assert.Equal(t, int32(99), dec.RoomID)
	assert.Equal(t, packet.UnbanUserPacketID, dec.PacketID())
}

// TestBanUserPacket_Decode verifies ban request decoding with ban type.
func TestBanUserPacket_Decode(t *testing.T) {
	w := codec.NewWriter()
	w.WriteInt32(3)
	w.WriteInt32(10)
	_ = w.WriteString("RWUAM_BAN_USER_HOUR")
	body := w.Bytes()
	var dec packet.BanUserPacket
	require.NoError(t, dec.Decode(body))
	assert.Equal(t, int32(3), dec.UserID)
	assert.Equal(t, int32(10), dec.RoomID)
	assert.Equal(t, "RWUAM_BAN_USER_HOUR", dec.BanType)
	assert.Equal(t, packet.BanUserPacketID, dec.PacketID())
}
