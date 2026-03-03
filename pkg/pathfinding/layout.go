package pathfinding

// Layout is the room's pre-computed tile grid.
type Layout struct {
	Width  int
	Height int
	Tiles  [][]Tile
}

// NewLayout creates a Layout from dimensions. All tiles start as TileOpen at Z=0.
func NewLayout(width, height int) *Layout {
	tiles := make([][]Tile, height)
	for y := range tiles {
		row := make([]Tile, width)
		for x := range row {
			row[x] = Tile{X: int16(x), Y: int16(y), State: TileOpen}
		}
		tiles[y] = row
	}
	return &Layout{Width: width, Height: height, Tiles: tiles}
}

// InBounds reports whether (x, y) is within the grid.
func (l *Layout) InBounds(x, y int) bool {
	return x >= 0 && x < l.Width && y >= 0 && y < l.Height
}

// At returns a pointer to the tile at (x, y). Panics if out of bounds.
func (l *Layout) At(x, y int) *Tile {
	return &l.Tiles[y][x]
}

// ParseHeightmap builds a Layout from a multi-line heightmap string.
// Each character encodes one tile: 'x' = blocked, '0'-'9' = height 0.0-9.0,
// 'a'-'z' = height 10.0-35.0. Rows are separated by newlines.
func ParseHeightmap(hm string) *Layout {
	var rows []string
	start := 0
	for i := 0; i <= len(hm); i++ {
		if i == len(hm) || hm[i] == '\n' {
			line := hm[start:i]
			if len(line) > 0 {
				rows = append(rows, line)
			}
			start = i + 1
		}
	}
	if len(rows) == 0 {
		return NewLayout(0, 0)
	}
	height := len(rows)
	width := len(rows[0])
	l := NewLayout(width, height)
	for y, row := range rows {
		for x := 0; x < len(row) && x < width; x++ {
			ch := row[x]
			t := l.At(x, y)
			switch {
			case ch == 'x' || ch == 'X':
				t.State = TileBlocked
				t.Z = 0
			case ch >= '0' && ch <= '9':
				t.Z = float32(ch - '0')
			case ch >= 'a' && ch <= 'z':
				t.Z = float32(ch-'a') + 10
			default:
				t.State = TileBlocked
			}
		}
	}
	return l
}
