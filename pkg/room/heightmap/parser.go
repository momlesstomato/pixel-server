package heightmap

import (
	"strings"
	"unicode"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// Parse converts a raw heightmap string into a tile grid.
func Parse(raw string) ([][]domain.Tile, error) {
	normalized := normalize(raw)
	rows := strings.Split(normalized, "\r")
	if len(rows) == 0 {
		return nil, domain.ErrInvalidHeightmap
	}
	rows = trimEmpty(rows)
	if len(rows) == 0 {
		return nil, domain.ErrInvalidHeightmap
	}
	width := len(rows[0])
	if width == 0 {
		return nil, domain.ErrInvalidHeightmap
	}
	grid := make([][]domain.Tile, len(rows))
	for y, row := range rows {
		if len(row) != width {
			return nil, domain.ErrInvalidHeightmap
		}
		grid[y] = make([]domain.Tile, width)
		for x, ch := range row {
			tile, err := parseTileChar(ch, x, y)
			if err != nil {
				return nil, err
			}
			grid[y][x] = tile
		}
	}
	return grid, nil
}

// normalize cleans raw heightmap input by standardizing line separators.
func normalize(raw string) string {
	s := strings.ReplaceAll(raw, "\\r\\n", "\r")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\n", "")
	s = strings.ReplaceAll(s, "\r\n", "\r")
	s = strings.ReplaceAll(s, "\n", "\r")
	return s
}

// trimEmpty removes leading and trailing empty rows.
func trimEmpty(rows []string) []string {
	start, end := 0, len(rows)
	for start < end && rows[start] == "" {
		start++
	}
	for end > start && rows[end-1] == "" {
		end--
	}
	return rows[start:end]
}

// parseTileChar converts a single heightmap character to a Tile.
func parseTileChar(ch rune, x int, y int) (domain.Tile, error) {
	lower := unicode.ToLower(ch)
	if lower == 'x' {
		return domain.Tile{X: x, Y: y, Z: 0, State: domain.TileBlocked}, nil
	}
	height, ok := charToHeight(lower)
	if !ok {
		return domain.Tile{}, domain.ErrInvalidHeightmap
	}
	return domain.Tile{X: x, Y: y, Z: float64(height), State: domain.TileOpen}, nil
}

// charToHeight converts a base-36 character to a height value.
func charToHeight(ch rune) (int, bool) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), true
	}
	if ch >= 'a' && ch <= 'z' {
		return int(ch-'a') + 10, true
	}
	return 0, false
}
