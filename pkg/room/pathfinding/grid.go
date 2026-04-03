package pathfinding

import "github.com/momlesstomato/pixel-server/pkg/room/domain"

// Grid adapts a room layout for pathfinding queries.
type Grid struct {
	// tiles stores the room tile grid [y][x].
	tiles [][]domain.Tile
	// width stores the grid column count.
	width int
	// height stores the grid row count.
	height int
}

// NewGrid creates a pathfinding grid from a tile array.
func NewGrid(tiles [][]domain.Tile) Grid {
	h := len(tiles)
	w := 0
	if h > 0 {
		w = len(tiles[0])
	}
	return Grid{tiles: tiles, width: w, height: h}
}

// InBounds reports whether the position is inside the grid.
func (g Grid) InBounds(x, y int) bool {
	return x >= 0 && x < g.width && y >= 0 && y < g.height
}

// IsWalkable reports whether the tile at (x,y) can be traversed.
func (g Grid) IsWalkable(x, y int) bool {
	if !g.InBounds(x, y) {
		return false
	}
	return g.tiles[y][x].State == domain.TileOpen
}

// HeightAt returns the Z-height at the given position.
func (g Grid) HeightAt(x, y int) float64 {
	if !g.InBounds(x, y) {
		return 0
	}
	return g.tiles[y][x].Z
}

// Width returns the grid column count.
func (g Grid) Width() int { return g.width }

// Height returns the grid row count.
func (g Grid) Height() int { return g.height }
