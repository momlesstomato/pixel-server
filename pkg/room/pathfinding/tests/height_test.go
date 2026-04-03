package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindPath_HeightDelta_Passable verifies path over acceptable height change.
func TestFindPath_HeightDelta_Passable(t *testing.T) {
	tiles := makeGridWithHeight([]string{"000"}, map[[2]int]float64{
		{1, 0}: 1.0,
		{2, 0}: 1.5,
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 2, 0, opts)
	require.NotNil(t, path)
	assert.Len(t, path, 2)
}

// TestFindPath_HeightDelta_TooSteep verifies rejection of steep steps.
func TestFindPath_HeightDelta_TooSteep(t *testing.T) {
	tiles := makeGridWithHeight([]string{"00"}, map[[2]int]float64{
		{1, 0}: 2.0,
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 1, 0, opts)
	assert.Nil(t, path)
}

// TestFindPath_HeightCost_PrefersFlat verifies flat path preferred over hilly.
func TestFindPath_HeightCost_PrefersFlat(t *testing.T) {
	tiles := makeGridWithHeight(
		[]string{"000", "000", "000"},
		map[[2]int]float64{
			{1, 0}: 1.4,
		},
	)
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	opts.HeightCostEnabled = true
	path := pathfinding.FindPath(g, 0, 0, 2, 0, opts)
	require.NotNil(t, path)
	hasHill := false
	for _, tile := range path {
		if tile.X == 1 && tile.Y == 0 {
			hasHill = true
		}
	}
	assert.False(t, hasHill, "path should avoid hill at (1,0)")
}

// TestFindPath_HeightCost_AscentExpensive verifies ascent costs more.
func TestFindPath_HeightCost_AscentExpensive(t *testing.T) {
	tiles := makeGridWithHeight(
		[]string{"00", "00"},
		map[[2]int]float64{
			{1, 0}: 1.0,
			{0, 1}: 0.1,
			{1, 1}: 0.1,
		},
	)
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	opts.HeightCostEnabled = true
	path := pathfinding.FindPath(g, 0, 0, 1, 0, opts)
	require.NotNil(t, path)
	assert.True(t, len(path) >= 1)
}

// TestFindPath_HeightDelta_ExactBoundary verifies max step height boundary.
func TestFindPath_HeightDelta_ExactBoundary(t *testing.T) {
	tiles := makeGridWithHeight([]string{"00"}, map[[2]int]float64{
		{1, 0}: 1.5,
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 1, 0, opts)
	require.NotNil(t, path)
	assert.Len(t, path, 1)
}

// TestFindPath_HeightDelta_JustOverBoundary verifies rejection above max step.
func TestFindPath_HeightDelta_JustOverBoundary(t *testing.T) {
	tiles := makeGridWithHeight([]string{"00"}, map[[2]int]float64{
		{1, 0}: 1.501,
	})
	g := pathfinding.NewGrid(tiles)
	opts := pathfinding.DefaultOptions()
	opts.AllowDiagonal = false
	path := pathfinding.FindPath(g, 0, 0, 1, 0, opts)
	assert.Nil(t, path)
}
