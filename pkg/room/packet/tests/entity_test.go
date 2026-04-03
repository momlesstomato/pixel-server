package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsersComposer_Empty verifies empty entity list encoding.
func TestUsersComposer_Empty(t *testing.T) {
	pkt := packet.UsersComposer{}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.UsersComposerID, pkt.PacketID())
}

// TestUsersComposer_OneEntity verifies single entity encoding.
func TestUsersComposer_OneEntity(t *testing.T) {
	entity := domain.NewPlayerEntity(1, 10, "c1", "Bob", "hr-100", "hi", "M",
		domain.Tile{X: 2, Y: 3, Z: 1.0, State: domain.TileOpen})
	pkt := packet.UsersComposer{Entities: []domain.RoomEntity{entity}}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

// TestUserUpdateComposer_Encode verifies status update encoding.
func TestUserUpdateComposer_Encode(t *testing.T) {
	entity := domain.NewPlayerEntity(1, 10, "c1", "Bob", "", "", "M",
		domain.Tile{X: 0, Y: 0, Z: 0})
	entity.Statuses["mv"] = "1,2,0"
	pkt := packet.UserUpdateComposer{Entities: []domain.RoomEntity{entity}}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.UserUpdateComposerID, pkt.PacketID())
}

// TestUserRemoveComposer_Encode verifies entity removal packet.
func TestUserRemoveComposer_Encode(t *testing.T) {
	pkt := packet.UserRemoveComposer{VirtualID: 5}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.UserRemoveComposerID, pkt.PacketID())
}

// TestDecodeMoveAvatar_Valid verifies walk destination decoding.
func TestDecodeMoveAvatar_Valid(t *testing.T) {
	w := make([]byte, 8)
	w[0], w[1], w[2], w[3] = 0, 0, 0, 5
	w[4], w[5], w[6], w[7] = 0, 0, 0, 10
	result := packet.DecodeMoveAvatar(w)
	require.NotNil(t, result)
	assert.Equal(t, 5, result[0])
	assert.Equal(t, 10, result[1])
}

// TestDecodeMoveAvatar_Short verifies nil on short input.
func TestDecodeMoveAvatar_Short(t *testing.T) {
	result := packet.DecodeMoveAvatar([]byte{0, 0})
	assert.Nil(t, result)
}

// TestDecodeMoveAvatar_Empty verifies nil on empty input.
func TestDecodeMoveAvatar_Empty(t *testing.T) {
	result := packet.DecodeMoveAvatar([]byte{})
	assert.Nil(t, result)
}
