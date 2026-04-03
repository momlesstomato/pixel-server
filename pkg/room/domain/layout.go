package domain

// Layout defines the spatial configuration of a room.
type Layout struct {
	// Slug stores the room model identifier.
	Slug string
	// DoorX stores the door tile horizontal coordinate.
	DoorX int
	// DoorY stores the door tile vertical coordinate.
	DoorY int
	// DoorZ stores the door tile height.
	DoorZ float64
	// DoorDir stores the door facing direction (0-7).
	DoorDir int
	// WallHeight stores the custom wall height (-1 for auto).
	WallHeight int
	// Grid stores the parsed tile grid [y][x].
	Grid [][]Tile
}

// Width returns the number of columns in the grid.
func (l Layout) Width() int {
	if len(l.Grid) == 0 {
		return 0
	}
	return len(l.Grid[0])
}

// Height returns the number of rows in the grid.
func (l Layout) Height() int {
	return len(l.Grid)
}

// TileAt returns the tile at the given position.
func (l Layout) TileAt(x, y int) (Tile, bool) {
	if y < 0 || y >= len(l.Grid) || x < 0 || x >= len(l.Grid[y]) {
		return Tile{}, false
	}
	return l.Grid[y][x], true
}

// IsWalkable reports whether the given position is walkable.
func (l Layout) IsWalkable(x, y int) bool {
	tile, ok := l.TileAt(x, y)
	if !ok {
		return false
	}
	return tile.State == TileOpen
}
