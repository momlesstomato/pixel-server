package domain

// TileState defines the walkability state of a single tile.
type TileState int

const (
	// TileOpen represents a walkable tile.
	TileOpen TileState = iota
	// TileBlocked represents a non-walkable tile.
	TileBlocked
	// TileOccupied represents a tile occupied by an entity.
	TileOccupied
)

// Tile represents a single position in the room grid.
type Tile struct {
	// X stores the horizontal coordinate.
	X int
	// Y stores the vertical coordinate.
	Y int
	// Z stores the height value (0-35).
	Z float64
	// State stores the tile walkability state.
	State TileState
}

// Coordinate represents a 2D grid position.
type Coordinate struct {
	// X stores the horizontal coordinate.
	X int
	// Y stores the vertical coordinate.
	Y int
}
