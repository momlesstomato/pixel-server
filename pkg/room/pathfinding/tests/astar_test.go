package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindPath_StraightLine verifies a straight horizontal path.
func TestFindPath_StraightLine(t *testing.T) {
	tiles := makeGrid([]string{"00000"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 4, 0, opts)
	require.NotNil(t, path)
	assert.Len(t, path, 4)
	assert.Equal(t, 4, path[len(path)-1].X)
}

// TestFindPath_Diagonal verifies diagonal movement.
func TestFindPath_Diagonal(t *testing.T) {
	tiles := makeGrid([]string{"000", "000", "000"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 2, 2, opts)
	require.NotNil(t, path)
	assert.Equal(t, 2, path[len(path)-1].X)
	assert.Equal(t, 2, path[len(path)-1].Y)
}

// TestFindPath_CardinalOnly verifies cardinal-only movement.
func TestFindPath_CardinalOnly(t *testing.T) {
	tiles := makeGrid([]string{"000", "000", "000"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 2, 2, opts)
	require.NotNil(t, path)
	assert.Len(t, path, 4)
}

// TestFindPath_Blocked verifies path around obstacle.
func TestFindPath_Blocked(t *testing.T) {
	tiles := makeGrid([]string{
		"000",
		"0x0",
		"000",
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 2, 2, opts)
	require.NotNil(t, path)
	for _, tile := range path {
		assert.False(t, tile.X == 1 && tile.Y == 1)
	}
}

// TestFindPath_NoPath verifies nil returned when no path exists.
func TestFindPath_NoPath(t *testing.T) {
	tiles := makeGrid([]string{
		"0x0",
		"xxx",
		"0x0",
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 2, 2, opts)
	assert.Nil(t, path)
}

// TestFindPath_DestinationBlocked verifies nil when destination blocked.
func TestFindPath_DestinationBlocked(t *testing.T) {
	tiles := makeGrid([]string{"00x"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 2, 0, opts)
	assert.Nil(t, path)
}

// TestFindPath_SamePosition verifies nil when start equals end.
func TestFindPath_SamePosition(t *testing.T) {
	tiles := makeGrid([]string{"00"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 0, 0, opts)
	assert.Nil(t, path)
}

// TestFindPath_MaxIterations verifies iteration limit.
func TestFindPath_MaxIterations(t *testing.T) {
	tiles := makeGrid([]string{"00000", "00000", "00000", "00000", "00000"})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.MaxIterations = 1
	path := pathfinding.FindPath(g, 0, 0, 4, 4, opts)
	assert.Nil(t, path)
}

// TestFindPath_DiagonalBlocking verifies diagonal requires open cardinals.
func TestFindPath_DiagonalBlocking(t *testing.T) {
	tiles := makeGrid([]string{
		"0x",
		"x0",
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 1, 1, opts)
	assert.Nil(t, path)
}

// TestFindPath_LargeGrid verifies performance on a large open map.
func TestFindPath_LargeGrid(t *testing.T) {
	row := ""
	for i := 0; i < 50; i++ {
		row += "0"
	}
	rows := make([]string, 50)
	for i := range rows {
		rows[i] = row
	}
	tiles := makeGrid(rows)
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(g, 0, 0, 49, 49, opts)
	require.NotNil(t, path)
	assert.Equal(t, 49, path[len(path)-1].X)
	assert.Equal(t, 49, path[len(path)-1].Y)
}
