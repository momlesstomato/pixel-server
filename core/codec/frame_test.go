package codec

import "testing"

// TestEncodeDecodeFrame verifies frame round-trip behavior.
func TestEncodeDecodeFrame(t *testing.T) {
	encoded := EncodeFrame(2419, []byte{1, 2, 3})
	decoded, consumed, err := DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if consumed != len(encoded) || decoded.PacketID != 2419 || len(decoded.Body) != 3 {
		t.Fatalf("unexpected decode result: %+v consumed=%d", decoded, consumed)
	}
}

// TestDecodeFramesParsesConcatenatedPayload verifies multiplexed frame parsing.
func TestDecodeFramesParsesConcatenatedPayload(t *testing.T) {
	first := EncodeFrame(10, []byte{})
	second := EncodeFrame(2491, []byte{9})
	payload := append(first, second...)
	frames, err := DecodeFrames(payload)
	if err != nil {
		t.Fatalf("expected decode frames success, got %v", err)
	}
	if len(frames) != 2 || frames[0].PacketID != 10 || frames[1].PacketID != 2491 {
		t.Fatalf("unexpected frame decode result: %+v", frames)
	}
}

// TestDecodeFrameRejectsInvalidPayload verifies decode validation behavior.
func TestDecodeFrameRejectsInvalidPayload(t *testing.T) {
	if _, _, err := DecodeFrame([]byte{0, 0, 0}); err == nil {
		t.Fatalf("expected error for short header")
	}
	if _, _, err := DecodeFrame([]byte{0, 0, 0, 1, 0, 1}); err == nil {
		t.Fatalf("expected error for invalid length")
	}
}
