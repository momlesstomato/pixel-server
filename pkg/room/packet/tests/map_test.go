package tests

import (
	"testing"

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
