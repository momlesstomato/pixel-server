package heightmap

import (
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// EncodeFloorMap converts a tile grid to the client floor heightmap format.
func EncodeFloorMap(grid [][]domain.Tile) string {
	if len(grid) == 0 {
		return ""
	}
	rows := make([]string, len(grid))
	for y, row := range grid {
		builder := strings.Builder{}
		builder.Grow(len(row))
		for _, tile := range row {
			builder.WriteByte(heightToChar(tile))
		}
		rows[y] = builder.String()
	}
	return strings.Join(rows, "\r")
}

// EncodeStackingMap converts a tile grid to RoomHeightMap stacking short array.
func EncodeStackingMap(grid [][]domain.Tile) []int16 {
	if len(grid) == 0 {
		return nil
	}
	height := len(grid)
	width := len(grid[0])
	result := make([]int16, height*width)
	for y, row := range grid {
		for x, tile := range row {
			idx := y*width + x
			if tile.State == domain.TileBlocked {
				result[idx] = int16(0x4000)
				continue
			}
			result[idx] = int16(int(tile.Z*256) & 16383)
		}
	}
	return result
}

// heightToChar converts a tile to its base-36 heightmap character.
func heightToChar(tile domain.Tile) byte {
	if tile.State == domain.TileBlocked {
		return 'x'
	}
	h := int(tile.Z)
	if h >= 0 && h <= 9 {
		return byte('0' + h)
	}
	if h >= 10 && h <= 35 {
		return byte('a' + h - 10)
	}
	return 'x'
}
