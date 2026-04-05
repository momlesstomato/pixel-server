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
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomSettingsComposerID, pkt.PacketID())
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
