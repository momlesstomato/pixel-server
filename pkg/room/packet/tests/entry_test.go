package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenFlatConnection_RoundTrip verifies encode/decode symmetry.
func TestOpenFlatConnection_RoundTrip(t *testing.T) {
	original := packet.OpenFlatConnectionPacket{RoomID: 42, Password: "secret"}
	body, err := original.Encode()
	require.NoError(t, err)
	var decoded packet.OpenFlatConnectionPacket
	err = decoded.Decode(body)
	require.NoError(t, err)
	assert.Equal(t, int32(42), decoded.RoomID)
	assert.Equal(t, "secret", decoded.Password)
}

// TestOpenFlatConnection_NoPassword verifies empty password handling.
func TestOpenFlatConnection_NoPassword(t *testing.T) {
	original := packet.OpenFlatConnectionPacket{RoomID: 10, Password: ""}
	body, err := original.Encode()
	require.NoError(t, err)
	var decoded packet.OpenFlatConnectionPacket
	err = decoded.Decode(body)
	require.NoError(t, err)
	assert.Equal(t, int32(10), decoded.RoomID)
}

// TestOpenFlatConnection_PacketID verifies protocol identifier.
func TestOpenFlatConnection_PacketID(t *testing.T) {
	pkt := packet.OpenFlatConnectionPacket{}
	assert.Equal(t, packet.OpenFlatConnectionPacketID, pkt.PacketID())
}

// TestRoomReadyComposer_Encode verifies room ready encoding.
func TestRoomReadyComposer_Encode(t *testing.T) {
	pkt := packet.RoomReadyComposer{ModelSlug: "model_a", RoomID: 1}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomReadyComposerID, pkt.PacketID())
}

// TestOpenConnectionComposer_Encode verifies empty connection ack.
func TestOpenConnectionComposer_Encode(t *testing.T) {
	pkt := packet.OpenConnectionComposer{}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.Empty(t, body)
	assert.Equal(t, packet.OpenConnectionComposerID, pkt.PacketID())
}
