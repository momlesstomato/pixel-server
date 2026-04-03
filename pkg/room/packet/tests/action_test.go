package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func encodeInt32Body(v int32) []byte {
	w := codec.NewWriter()
	w.WriteInt32(v)
	return w.Bytes()
}

func encodeLookToBody(x, y int32) []byte {
	w := codec.NewWriter()
	w.WriteInt32(x)
	w.WriteInt32(y)
	return w.Bytes()
}

// TestDancePacket_Decode verifies dance packet body decoding.
func TestDancePacket_Decode(t *testing.T) {
	p := &packet.DancePacket{}
	require.NoError(t, p.Decode(encodeInt32Body(3)))
	assert.Equal(t, int32(3), p.DanceID)
}

// TestDancePacket_PacketID verifies the protocol identifier.
func TestDancePacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.DancePacketID, packet.DancePacket{}.PacketID())
}

// TestActionPacket_Decode verifies action packet body decoding.
func TestActionPacket_Decode(t *testing.T) {
	p := &packet.ActionPacket{}
	require.NoError(t, p.Decode(encodeInt32Body(1)))
	assert.Equal(t, int32(1), p.ActionID)
}

// TestActionPacket_PacketID verifies the protocol identifier.
func TestActionPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.ActionPacketID, packet.ActionPacket{}.PacketID())
}

// TestSignPacket_Decode verifies sign packet body decoding.
func TestSignPacket_Decode(t *testing.T) {
	p := &packet.SignPacket{}
	require.NoError(t, p.Decode(encodeInt32Body(5)))
	assert.Equal(t, int32(5), p.SignID)
}

// TestSignPacket_PacketID verifies the protocol identifier.
func TestSignPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.SignPacketID, packet.SignPacket{}.PacketID())
}

// TestLookToPacket_Decode verifies look-to packet body decoding.
func TestLookToPacket_Decode(t *testing.T) {
	p := &packet.LookToPacket{}
	require.NoError(t, p.Decode(encodeLookToBody(4, 7)))
	assert.Equal(t, int32(4), p.X)
	assert.Equal(t, int32(7), p.Y)
}

// TestLookToPacket_PacketID verifies the protocol identifier.
func TestLookToPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.LookToPacketID, packet.LookToPacket{}.PacketID())
}

// TestDanceComposer_Encode verifies dance composer serialization.
func TestDanceComposer_Encode(t *testing.T) {
	pkt := packet.DanceComposer{VirtualID: 2, DanceID: 3}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.DanceComposerID, pkt.PacketID())
}

// TestUserTypingComposer_Encode verifies typing composer serialization.
func TestUserTypingComposer_Encode(t *testing.T) {
	pkt := packet.UserTypingComposer{VirtualID: 1, IsTyping: true}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.UserTypingComposerID, pkt.PacketID())
}

// TestSleepComposer_Encode verifies sleep composer serialization.
func TestSleepComposer_Encode(t *testing.T) {
	pkt := packet.SleepComposer{VirtualID: 1, IsAsleep: true}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.SleepComposerID, pkt.PacketID())
}

// TestActionComposer_Encode verifies action composer serialization.
func TestActionComposer_Encode(t *testing.T) {
	pkt := packet.ActionComposer{VirtualID: 5, ActionID: 1}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.ActionComposerID, pkt.PacketID())
}

// TestActionComposerID verifies the protocol constant.
func TestActionComposerID(t *testing.T) {
	assert.Equal(t, uint16(1631), packet.ActionComposerID)
}
