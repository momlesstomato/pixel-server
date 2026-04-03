package heightmap

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/heightmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParse_SimpleGrid validates basic heightmap parsing.
func TestParse_SimpleGrid(t *testing.T) {
	grid, err := heightmap.Parse("000\r000\r000")
	require.NoError(t, err)
	assert.Equal(t, 3, len(grid))
	assert.Equal(t, 3, len(grid[0]))
	assert.Equal(t, float64(0), grid[1][1].Z)
	assert.Equal(t, domain.TileOpen, grid[0][0].State)
}

// TestParse_BlockedTiles validates blocked tile parsing.
func TestParse_BlockedTiles(t *testing.T) {
	grid, err := heightmap.Parse("x0x\r000\rx0x")
	require.NoError(t, err)
	assert.Equal(t, domain.TileBlocked, grid[0][0].State)
	assert.Equal(t, domain.TileOpen, grid[0][1].State)
	assert.Equal(t, domain.TileBlocked, grid[2][2].State)
}

// TestParse_Base36Heights validates all base-36 height values.
func TestParse_Base36Heights(t *testing.T) {
	grid, err := heightmap.Parse("0123456789\rabcdefghij")
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		assert.Equal(t, float64(i), grid[0][i].Z)
	}
	for i := 0; i < 10; i++ {
		assert.Equal(t, float64(i+10), grid[1][i].Z)
	}
}

// TestParse_UppercaseBase36 validates case-insensitive height parsing.
func TestParse_UppercaseBase36(t *testing.T) {
	grid, err := heightmap.Parse("ABCZ")
	require.NoError(t, err)
	assert.Equal(t, float64(10), grid[0][0].Z)
	assert.Equal(t, float64(11), grid[0][1].Z)
	assert.Equal(t, float64(12), grid[0][2].Z)
	assert.Equal(t, float64(35), grid[0][3].Z)
}

// TestParse_NormalizeCRLF validates CRLF normalization.
func TestParse_NormalizeCRLF(t *testing.T) {
	grid, err := heightmap.Parse("00\r\n00")
	require.NoError(t, err)
	assert.Equal(t, 2, len(grid))
}

// TestParse_NormalizeLF validates LF normalization.
func TestParse_NormalizeLF(t *testing.T) {
	grid, err := heightmap.Parse("00\n00")
	require.NoError(t, err)
	assert.Equal(t, 2, len(grid))
}

// TestParse_NormalizeLiteralEscape validates literal escape sequence normalization.
func TestParse_NormalizeLiteralEscape(t *testing.T) {
	grid, err := heightmap.Parse("00\\r\\n00")
	require.NoError(t, err)
	assert.Equal(t, 2, len(grid))
}

// TestParse_EmptyString returns error for empty input.
func TestParse_EmptyString(t *testing.T) {
	_, err := heightmap.Parse("")
	assert.ErrorIs(t, err, domain.ErrInvalidHeightmap)
}

// TestParse_InconsistentWidths returns error for jagged rows.
func TestParse_InconsistentWidths(t *testing.T) {
	_, err := heightmap.Parse("000\r00")
	assert.ErrorIs(t, err, domain.ErrInvalidHeightmap)
}

// TestParse_InvalidCharacter returns error for unsupported characters.
func TestParse_InvalidCharacter(t *testing.T) {
	_, err := heightmap.Parse("0!0")
	assert.ErrorIs(t, err, domain.ErrInvalidHeightmap)
}

// TestParse_CoordinatesAssigned validates tile coordinates.
func TestParse_CoordinatesAssigned(t *testing.T) {
	grid, err := heightmap.Parse("012\r345")
	require.NoError(t, err)
	assert.Equal(t, 0, grid[0][0].X)
	assert.Equal(t, 0, grid[0][0].Y)
	assert.Equal(t, 2, grid[0][2].X)
	assert.Equal(t, 0, grid[0][2].Y)
	assert.Equal(t, 1, grid[1][1].X)
	assert.Equal(t, 1, grid[1][1].Y)
}

// TestParse_ModelA validates parsing of standard model_a heightmap.
func TestParse_ModelA(t *testing.T) {
	hmap := "xxxxxxxxxxxx\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx"
	grid, err := heightmap.Parse(hmap)
	require.NoError(t, err)
	assert.Equal(t, 16, len(grid))
	assert.Equal(t, 12, len(grid[0]))
	assert.Equal(t, domain.TileBlocked, grid[0][0].State)
	assert.Equal(t, domain.TileOpen, grid[1][4].State)
}
