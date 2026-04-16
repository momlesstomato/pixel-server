package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
	"github.com/stretchr/testify/assert"
)

// makeGrid creates a simple grid from a string map.
func makeGrid(rows []string) [][]domain.Tile {
	grid := make([][]domain.Tile, len(rows))
	for y, row := range rows {
		grid[y] = make([]domain.Tile, len(row))
		for x, ch := range row {
			t := domain.Tile{X: x, Y: y, Z: 0, State: domain.TileOpen}
			if ch == 'x' {
				t.State = domain.TileBlocked
			}
			grid[y][x] = t
		}
	}
	return grid
}

// makeGridWithHeight creates a grid with height values.
func makeGridWithHeight(rows []string, heights map[[2]int]float64) [][]domain.Tile {
	grid := makeGrid(rows)
	for pos, h := range heights {
		grid[pos[1]][pos[0]].Z = h
	}
	return grid
}

// TestNewGrid_Dimensions verifies grid dimensions are correct.
func TestNewGrid_Dimensions(t *testing.T) {
	tiles := makeGrid([]string{"000", "000"})
	g := pathfinding.NewGrid(tiles)
	assert.Equal(t, 3, g.Width())
	assert.Equal(t, 2, g.Height())
}

// TestNewGrid_Empty verifies empty grid dimensions.
func TestNewGrid_Empty(t *testing.T) {
	g := pathfinding.NewGrid(nil)
	assert.Equal(t, 0, g.Width())
	assert.Equal(t, 0, g.Height())
}

// TestGrid_InBounds verifies boundary checks.
func TestGrid_InBounds(t *testing.T) {
	tiles := makeGrid([]string{"00", "00"})
	g := pathfinding.NewGrid(tiles)
	assert.True(t, g.InBounds(0, 0))
	assert.True(t, g.InBounds(1, 1))
	assert.False(t, g.InBounds(-1, 0))
	assert.False(t, g.InBounds(2, 0))
	assert.False(t, g.InBounds(0, 2))
}

// TestGrid_IsWalkable verifies walkability detection.
func TestGrid_IsWalkable(t *testing.T) {
	tiles := makeGrid([]string{"0x", "00"})
	g := pathfinding.NewGrid(tiles)
	assert.True(t, g.IsWalkable(0, 0))
	assert.False(t, g.IsWalkable(1, 0))
	assert.False(t, g.IsWalkable(5, 5))
}

// TestGrid_IsWalkableWithDynamicBlockChecker verifies dynamic blocker callbacks participate in walkability checks.
func TestGrid_IsWalkableWithDynamicBlockChecker(t *testing.T) {
	tiles := makeGrid([]string{"00", "00"})
	g := pathfinding.NewGridWithBlockersAndChecker(tiles, nil, func(x, y int) bool {
		return x == 1 && y == 0
	})
	assert.True(t, g.IsWalkable(0, 0))
	assert.False(t, g.IsWalkable(1, 0))
}

// TestGrid_HeightAt verifies height retrieval.
func TestGrid_HeightAt(t *testing.T) {
	tiles := makeGridWithHeight([]string{"00"}, map[[2]int]float64{{1, 0}: 3.5})
	g := pathfinding.NewGrid(tiles)
	assert.Equal(t, 0.0, g.HeightAt(0, 0))
	assert.Equal(t, 3.5, g.HeightAt(1, 0))
	assert.Equal(t, 0.0, g.HeightAt(99, 99))
}
