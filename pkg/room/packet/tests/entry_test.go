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

// TestDoorbellComposer_Encode verifies doorbell notification encoding.
func TestDoorbellComposer_Encode(t *testing.T) {
	pkt := packet.DoorbellComposer{Username: "Alice"}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.DoorbellComposerID, pkt.PacketID())
}

// TestFlatAccessibleComposer_Approved verifies approved entry encoding.
func TestFlatAccessibleComposer_Approved(t *testing.T) {
	pkt := packet.FlatAccessibleComposer{Username: "Alice", Accessible: true}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.FlatAccessibleComposerID, pkt.PacketID())
}

// TestFlatAccessibleComposer_Denied verifies denied entry encoding.
func TestFlatAccessibleComposer_Denied(t *testing.T) {
	pkt := packet.FlatAccessibleComposer{Username: "Bob", Accessible: false}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

// TestLetUserInPacket_Decode verifies doorbell approval decoding.
func TestLetUserInPacket_Decode(t *testing.T) {
	source := packet.FlatAccessibleComposer{Username: "Alice", Accessible: true}
	body, err := source.Encode()
	require.NoError(t, err)
	var pkt packet.LetUserInPacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, "Alice", pkt.Username)
	assert.True(t, pkt.Let)
	assert.Equal(t, packet.LetUserInPacketID, pkt.PacketID())
}

// TestLetUserInPacket_DecodeDeclined verifies doorbell denial decoding.
func TestLetUserInPacket_DecodeDeclined(t *testing.T) {
	source := packet.FlatAccessibleComposer{Username: "Bob", Accessible: false}
	body, err := source.Encode()
	require.NoError(t, err)
	var pkt packet.LetUserInPacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, "Bob", pkt.Username)
	assert.False(t, pkt.Let)
}

// TestRoomForwardComposer_Encode verifies room forward encoding.
func TestRoomForwardComposer_Encode(t *testing.T) {
	pkt := packet.RoomForwardComposer{RoomID: 77}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomForwardComposerID, pkt.PacketID())
}

// TestDeleteRoomPacket_RoundTrip verifies delete room encode/decode.
func TestDeleteRoomPacket_RoundTrip(t *testing.T) {
	enc := packet.RoomForwardComposer{RoomID: 55}
	body, err := enc.Encode()
	require.NoError(t, err)
	var pkt packet.DeleteRoomPacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, int32(55), pkt.RoomID)
	assert.Equal(t, packet.DeleteRoomPacketID, pkt.PacketID())
}

// TestGiveRoomScorePacket_Decode verifies vote packet decoding.
func TestGiveRoomScorePacket_Decode(t *testing.T) {
	enc := packet.RoomScoreComposer{Score: 1, CanVote: true}
	body, err := enc.Encode()
	require.NoError(t, err)
	var pkt packet.GiveRoomScorePacket
	require.NoError(t, pkt.Decode(body))
	assert.Equal(t, int32(1), pkt.Score)
	assert.Equal(t, packet.GiveRoomScorePacketID, pkt.PacketID())
}

// TestRoomScoreComposer_Encode verifies room score encoding.
func TestRoomScoreComposer_Encode(t *testing.T) {
	pkt := packet.RoomScoreComposer{Score: 42, CanVote: false}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.RoomScoreComposerID, pkt.PacketID())
}
