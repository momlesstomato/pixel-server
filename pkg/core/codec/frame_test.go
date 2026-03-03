package codec_test

import (
	"encoding/binary"
	"testing"

	"pixel-server/pkg/core/codec"
)

func TestParseFrameSingle(t *testing.T) {
	// Build a frame: length=6 (2 header + 4 payload), headerID=4000, payload=int32(42)
	frame := make([]byte, 10)
	binary.BigEndian.PutUint32(frame[0:4], 6)    // length
	binary.BigEndian.PutUint16(frame[4:6], 4000)  // headerID
	binary.BigEndian.PutUint32(frame[6:10], 42)   // payload (int32)

	headerID, payload, rest, err := codec.ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if headerID != 4000 {
		t.Fatalf("expected header 4000, got %d", headerID)
	}
	if len(payload) != 4 {
		t.Fatalf("expected 4 bytes payload, got %d", len(payload))
	}
	if len(rest) != 0 {
		t.Fatalf("expected no rest, got %d bytes", len(rest))
	}

	// Verify payload contains int32(42)
	r := codec.NewReader(payload)
	v, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("ReadInt32: %v", err)
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestParseFramesConcatenated(t *testing.T) {
	// Build two frames back to back
	w1 := codec.NewWriter(0)
	w1.WriteString("hello")
	f1 := w1.Frame(100)

	w2 := codec.NewWriter(0)
	w2.WriteInt32(99)
	f2 := w2.Frame(200)

	data := append(f1, f2...)

	frames, err := codec.ParseFrames(data)
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	if frames[0].HeaderID != 100 {
		t.Fatalf("frame 0: expected header 100, got %d", frames[0].HeaderID)
	}
	if frames[1].HeaderID != 200 {
		t.Fatalf("frame 1: expected header 200, got %d", frames[1].HeaderID)
	}

	// Verify frame 0 payload
	r := codec.NewReader(frames[0].Payload)
	s, err := r.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Fatalf("expected %q, got %q", "hello", s)
	}

	// Verify frame 1 payload
	r = codec.NewReader(frames[1].Payload)
	v, err := r.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if v != 99 {
		t.Fatalf("expected 99, got %d", v)
	}
}

func TestParseFrameErrors(t *testing.T) {
	// Too short for length
	if _, _, _, err := codec.ParseFrame([]byte{0, 0}); err == nil {
		t.Fatal("expected error for short data")
	}

	// Length says 100 but only header bytes present
	bad := make([]byte, 6)
	binary.BigEndian.PutUint32(bad[0:4], 100)
	binary.BigEndian.PutUint16(bad[4:6], 1)
	if _, _, _, err := codec.ParseFrame(bad); err == nil {
		t.Fatal("expected error for incomplete frame")
	}

	// Frame length < 2 (invalid)
	tiny := make([]byte, 4)
	binary.BigEndian.PutUint32(tiny[0:4], 1) // too small
	if _, _, _, err := codec.ParseFrame(tiny); err == nil {
		t.Fatal("expected error for frame length < 2")
	}
}

func TestFrameRoundTrip(t *testing.T) {
	// Write → Frame → ParseFrame → Read should produce original values
	w := codec.NewWriter(0)
	w.WriteString("NITRO-1-6-6")
	w.WriteInt32(2)
	framed := w.Frame(4000)

	headerID, payload, rest, err := codec.ParseFrame(framed)
	if err != nil {
		t.Fatalf("ParseFrame: %v", err)
	}
	if headerID != 4000 {
		t.Fatalf("expected 4000, got %d", headerID)
	}
	if len(rest) != 0 {
		t.Fatalf("expected no rest, got %d", len(rest))
	}

	r := codec.NewReader(payload)
	s, err := r.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "NITRO-1-6-6" {
		t.Fatalf("expected NITRO-1-6-6, got %q", s)
	}
	v, err := r.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}
