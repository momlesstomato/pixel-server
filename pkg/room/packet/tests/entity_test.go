package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
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

// TestUserUpdateComposer_EncodesDeterministicStatusOrder verifies posture status ordering is stable and Nitro-compatible.
func TestUserUpdateComposer_EncodesDeterministicStatusOrder(t *testing.T) {
	entity := domain.NewPlayerEntity(1, 10, "c1", "Bob", "", "", "M",
		domain.Tile{X: 0, Y: 0, Z: 0})
	entity.Statuses["sit"] = "1.10 1"
	entity.Statuses["sign"] = "5"
	entity.Statuses["dance"] = "2"
	pkt := packet.UserUpdateComposer{Entities: []domain.RoomEntity{entity}}
	body, err := pkt.Encode()
	require.NoError(t, err)
	r := codec.NewReader(body)
	count, err := r.ReadInt32()
	require.NoError(t, err)
	assert.Equal(t, int32(1), count)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadString()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	status, err := r.ReadString()
	require.NoError(t, err)
	assert.Equal(t, "/sign 5/dance 2/sit 1.10 1/", status)
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
