package heightmap

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/heightmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEncodeFloorMap_SimpleGrid validates floor map encoding.
func TestEncodeFloorMap_SimpleGrid(t *testing.T) {
	grid := [][]domain.Tile{
		{{X: 0, Y: 0, Z: 0, State: domain.TileOpen}, {X: 1, Y: 0, Z: 1, State: domain.TileOpen}},
		{{X: 0, Y: 1, Z: 2, State: domain.TileOpen}, {X: 1, Y: 1, Z: 0, State: domain.TileBlocked}},
	}
	result := heightmap.EncodeFloorMap(grid)
	assert.Equal(t, "01\r2x", result)
}

// TestEncodeFloorMap_Base36Heights validates base-36 encoding.
func TestEncodeFloorMap_Base36Heights(t *testing.T) {
	grid := [][]domain.Tile{
		{{Z: 10, State: domain.TileOpen}, {Z: 25, State: domain.TileOpen}, {Z: 35, State: domain.TileOpen}},
	}
	result := heightmap.EncodeFloorMap(grid)
	assert.Equal(t, "apz", result)
}

// TestEncodeFloorMap_Empty validates empty grid encoding.
func TestEncodeFloorMap_Empty(t *testing.T) {
	assert.Equal(t, "", heightmap.EncodeFloorMap(nil))
}

// TestEncodeFloorMap_RoundTrip validates parse-encode round trip.
func TestEncodeFloorMap_RoundTrip(t *testing.T) {
	original := "x0x\r012\rxax"
	grid, err := heightmap.Parse(original)
	require.NoError(t, err)
	encoded := heightmap.EncodeFloorMap(grid)
	assert.Equal(t, original, encoded)
}

// TestEncodeStackingMap_SimpleGrid validates stacking map encoding.
func TestEncodeStackingMap_SimpleGrid(t *testing.T) {
	grid := [][]domain.Tile{
		{{Z: 0, State: domain.TileOpen}, {Z: 1, State: domain.TileOpen}},
		{{Z: 0, State: domain.TileBlocked}, {Z: 0.5, State: domain.TileOpen}},
	}
	result := heightmap.EncodeStackingMap(grid)
	require.Len(t, result, 4)
	assert.Equal(t, int16(0), result[0])
	assert.Equal(t, int16(256), result[1])
	assert.Equal(t, int16(0x4000), result[2])
	assert.Equal(t, int16(128), result[3])
}

// TestEncodeStackingMap_Empty validates empty grid stacking map.
func TestEncodeStackingMap_Empty(t *testing.T) {
	assert.Nil(t, heightmap.EncodeStackingMap(nil))
}

// TestEncodeStackingMap_BlockedBit validates stacking blocked bit.
func TestEncodeStackingMap_BlockedBit(t *testing.T) {
	grid := [][]domain.Tile{
		{{Z: 0, State: domain.TileBlocked}},
	}
	result := heightmap.EncodeStackingMap(grid)
	assert.Equal(t, int16(0x4000), result[0])
}
