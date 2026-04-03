package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func encodeChatBody(msg string, style int32) []byte {
	w := codec.NewWriter()
	_ = w.WriteString(msg)
	w.WriteInt32(style)
	return w.Bytes()
}

func encodeWhisperBody(target, msg string, style int32) []byte {
	w := codec.NewWriter()
	_ = w.WriteString(target)
	_ = w.WriteString(msg)
	w.WriteInt32(style)
	return w.Bytes()
}

// TestChatPacket_Decode verifies talk packet body decoding.
func TestChatPacket_Decode(t *testing.T) {
	p := &packet.ChatPacket{}
	require.NoError(t, p.Decode(encodeChatBody("hello", 0)))
	assert.Equal(t, "hello", p.Message)
	assert.Equal(t, int32(0), p.BubbleStyle)
}

// TestChatPacket_PacketID verifies the protocol identifier.
func TestChatPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.ChatPacketID, packet.ChatPacket{}.PacketID())
}

// TestShoutPacket_Decode verifies shout packet body decoding.
func TestShoutPacket_Decode(t *testing.T) {
	p := &packet.ShoutPacket{}
	require.NoError(t, p.Decode(encodeChatBody("hey", 2)))
	assert.Equal(t, "hey", p.Message)
	assert.Equal(t, int32(2), p.BubbleStyle)
}

// TestShoutPacket_PacketID verifies the protocol identifier.
func TestShoutPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.ShoutPacketID, packet.ShoutPacket{}.PacketID())
}

// TestWhisperPacket_Decode verifies whisper packet body decoding.
func TestWhisperPacket_Decode(t *testing.T) {
	p := &packet.WhisperPacket{}
	require.NoError(t, p.Decode(encodeWhisperBody("Bob", "secret", 1)))
	assert.Equal(t, "Bob", p.TargetUsername)
	assert.Equal(t, "secret", p.Message)
	assert.Equal(t, int32(1), p.BubbleStyle)
}

// TestWhisperPacket_PacketID verifies the protocol identifier.
func TestWhisperPacket_PacketID(t *testing.T) {
	assert.Equal(t, packet.WhisperPacketID, packet.WhisperPacket{}.PacketID())
}

// TestChatComposer_Encode verifies chat composer serialization.
func TestChatComposer_Encode(t *testing.T) {
	pkt := packet.ChatComposer{VirtualID: 1, Message: "hi", GestureID: 0, BubbleStyle: 0}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.ChatComposerID, pkt.PacketID())
}

// TestShoutComposer_Encode verifies shout composer serialization.
func TestShoutComposer_Encode(t *testing.T) {
	pkt := packet.ShoutComposer{VirtualID: 1, Message: "hey all", GestureID: 0, BubbleStyle: 0}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.ShoutComposerID, pkt.PacketID())
}

// TestWhisperComposer_Encode verifies whisper composer serialization.
func TestWhisperComposer_Encode(t *testing.T) {
	pkt := packet.WhisperComposer{VirtualID: 2, SenderName: "Alice", Message: "psst", GestureID: 0, BubbleStyle: 0}
	body, err := pkt.Encode()
	require.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, packet.WhisperComposerID, pkt.PacketID())
}
