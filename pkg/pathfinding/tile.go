package pathfinding

// TileState describes the passability of a tile.
type TileState uint8

const (
	TileOpen    TileState = iota // passable floor
	TileBlocked                  // wall or occupied
	TileSeat                     // passable but terminal (sit)
	TileBed                      // passable but terminal (lay)
)

// Tile represents a single cell in the room grid.
type Tile struct {
	X, Y  int16
	Z     float32 // stack height
	State TileState
}

// PathStep is a waypoint returned by FindPath.
type PathStep struct {
	X, Y int16
	Z    float32
}
