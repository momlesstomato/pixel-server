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
	// blocked stores dynamically blocked tile positions (e.g. occupied by entities).
	blocked map[[2]int]struct{}
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

// NewGridWithBlockers creates a pathfinding grid with additional blocked positions.
func NewGridWithBlockers(tiles [][]domain.Tile, blockers [][2]int) Grid {
	g := NewGrid(tiles)
	if len(blockers) == 0 {
		return g
	}
	g.blocked = make(map[[2]int]struct{}, len(blockers))
	for _, b := range blockers {
		g.blocked[b] = struct{}{}
	}
	return g
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
	if g.blocked != nil {
		if _, blocked := g.blocked[[2]int{x, y}]; blocked {
			return false
		}
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
