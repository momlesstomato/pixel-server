package codec

import (
	"errors"
	"testing"
)

// TestWriterReaderRoundTrip validates primitive encode/decode compatibility.
func TestWriterReaderRoundTrip(t *testing.T) {
	w := NewWriter(64)
	w.WriteBool(true)
	w.WriteInt32(-7)
	w.WriteUint16(9)
	w.WriteUint32(11)
	w.WriteString("ok")
	w.WriteBytes([]byte{1, 2, 3})
	r := NewReader(w.Bytes())
	b, err := r.ReadBool()
	if err != nil || !b {
		t.Fatalf("unexpected bool decode: %v %v", b, err)
	}
	i, err := r.ReadInt32()
	if err != nil || i != -7 {
		t.Fatalf("unexpected int32 decode: %d %v", i, err)
	}
	u16, err := r.ReadUint16()
	if err != nil || u16 != 9 {
		t.Fatalf("unexpected uint16 decode: %d %v", u16, err)
	}
	u32, err := r.ReadUint32()
	if err != nil || u32 != 11 {
		t.Fatalf("unexpected uint32 decode: %d %v", u32, err)
	}
	text, err := r.ReadString()
	if err != nil || text != "ok" {
		t.Fatalf("unexpected string decode: %s %v", text, err)
	}
	body, err := r.ReadBytes(3)
	if err != nil || len(body) != 3 {
		t.Fatalf("unexpected bytes decode: %v %v", body, err)
	}
}

// TestReaderUnexpectedEOF validates EOF protection.
func TestReaderUnexpectedEOF(t *testing.T) {
	r := NewReader([]byte{1})
	_, err := r.ReadUint16()
	if !errors.Is(err, ErrUnexpectedEOF) {
		t.Fatalf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestSplitFrames validates multi-packet websocket payload splitting.
func TestSplitFrames(t *testing.T) {
	first := EncodeFrame(4000, []byte{1, 2})
	second := EncodeFrame(295, []byte{3})
	joined := append(first, second...)
	frames, err := SplitFrames(joined)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(frames) != 2 {
		t.Fatalf("expected two frames, got %d", len(frames))
	}
	if frames[0].Header != 4000 || len(frames[0].Payload) != 2 {
		t.Fatalf("unexpected first frame: %+v", frames[0])
	}
	if frames[1].Header != 295 || len(frames[1].Payload) != 1 {
		t.Fatalf("unexpected second frame: %+v", frames[1])
	}
}

// TestSplitFramesInvalid validates invalid frame protection.
func TestSplitFramesInvalid(t *testing.T) {
	if _, err := SplitFrames([]byte{0, 0, 0, 1, 0}); !errors.Is(err, ErrInvalidFrame) {
		t.Fatalf("expected ErrInvalidFrame, got %v", err)
	}
}
