package codec_test

import (
	"math"
	"testing"

	"pixel-server/pkg/core/codec"
)

// Primitive round-trips

func TestBoolRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		val  bool
	}{
		{"true", true},
		{"false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := codec.NewWriter(0)
			w.WriteBool(tt.val)
			r := codec.NewReader(w.Bytes())
			got, err := r.ReadBool()
			if err != nil {
				t.Fatalf("ReadBool: %v", err)
			}
			if got != tt.val {
				t.Fatalf("expected %v, got %v", tt.val, got)
			}
			if r.Remaining() != 0 {
				t.Fatalf("expected 0 remaining, got %d", r.Remaining())
			}
		})
	}
}

func TestInt16RoundTrip(t *testing.T) {
	tests := []int16{0, 1, -1, 32767, -32768, 256}
	for _, v := range tests {
		w := codec.NewWriter(0)
		w.WriteInt16(v)
		r := codec.NewReader(w.Bytes())
		got, err := r.ReadInt16()
		if err != nil {
			t.Fatalf("ReadInt16(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	}
}

func TestUint16RoundTrip(t *testing.T) {
	tests := []uint16{0, 1, 65535, 4000, 1347}
	for _, v := range tests {
		w := codec.NewWriter(0)
		w.WriteUint16(v)
		r := codec.NewReader(w.Bytes())
		got, err := r.ReadUint16()
		if err != nil {
			t.Fatalf("ReadUint16(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	}
}

func TestInt32RoundTrip(t *testing.T) {
	tests := []int32{0, 1, -1, 2147483647, -2147483648, 42}
	for _, v := range tests {
		w := codec.NewWriter(0)
		w.WriteInt32(v)
		r := codec.NewReader(w.Bytes())
		got, err := r.ReadInt32()
		if err != nil {
			t.Fatalf("ReadInt32(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	}
}

func TestUint32RoundTrip(t *testing.T) {
	tests := []uint32{0, 1, 4294967295, 100}
	for _, v := range tests {
		w := codec.NewWriter(0)
		w.WriteUint32(v)
		r := codec.NewReader(w.Bytes())
		got, err := r.ReadUint32()
		if err != nil {
			t.Fatalf("ReadUint32(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	}
}

func TestFloat64RoundTrip(t *testing.T) {
	tests := []float64{0, 1.5, -1.5, math.Pi, math.MaxFloat64, math.SmallestNonzeroFloat64}
	for _, v := range tests {
		w := codec.NewWriter(0)
		w.WriteFloat64(v)
		r := codec.NewReader(w.Bytes())
		got, err := r.ReadFloat64()
		if err != nil {
			t.Fatalf("ReadFloat64(%f): %v", v, err)
		}
		if got != v {
			t.Fatalf("expected %f, got %f", v, got)
		}
	}
}

func TestStringRoundTrip(t *testing.T) {
	tests := []string{"", "hello", "NITRO-1-6-6", "日本語テスト", "a string with spaces"}
	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			w := codec.NewWriter(0)
			w.WriteString(v)
			r := codec.NewReader(w.Bytes())
			got, err := r.ReadString()
			if err != nil {
				t.Fatalf("ReadString: %v", err)
			}
			if got != v {
				t.Fatalf("expected %q, got %q", v, got)
			}
		})
	}
}

// List round-trips

func TestListInt32RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		vals []int32
	}{
		{"empty", []int32{}},
		{"single", []int32{42}},
		{"multi", []int32{1, -1, 0, 2147483647}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := codec.NewWriter(0)
			w.WriteListInt32(tt.vals)
			r := codec.NewReader(w.Bytes())
			got, err := r.ReadListInt32()
			if err != nil {
				t.Fatalf("ReadListInt32: %v", err)
			}
			if len(got) != len(tt.vals) {
				t.Fatalf("expected len %d, got %d", len(tt.vals), len(got))
			}
			for i := range tt.vals {
				if got[i] != tt.vals[i] {
					t.Fatalf("index %d: expected %d, got %d", i, tt.vals[i], got[i])
				}
			}
		})
	}
}

func TestListStringRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		vals []string
	}{
		{"empty", []string{}},
		{"single", []string{"hello"}},
		{"multi", []string{"alpha", "beta", "gamma", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := codec.NewWriter(0)
			w.WriteListString(tt.vals)
			r := codec.NewReader(w.Bytes())
			got, err := r.ReadListString()
			if err != nil {
				t.Fatalf("ReadListString: %v", err)
			}
			if len(got) != len(tt.vals) {
				t.Fatalf("expected len %d, got %d", len(tt.vals), len(got))
			}
			for i := range tt.vals {
				if got[i] != tt.vals[i] {
					t.Fatalf("index %d: expected %q, got %q", i, tt.vals[i], got[i])
				}
			}
		})
	}
}

// Composite payloads

func TestMultiFieldPayload(t *testing.T) {
	// Simulates the handshake.release_version packet: string, string, int32, int32
	w := codec.NewWriter(64)
	w.WriteString("NITRO-1-6-6")
	w.WriteString("HTML5")
	w.WriteInt32(2)
	w.WriteInt32(1)

	r := codec.NewReader(w.Bytes())

	s1, err := r.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s1 != "NITRO-1-6-6" {
		t.Fatalf("expected NITRO-1-6-6, got %q", s1)
	}

	s2, err := r.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s2 != "HTML5" {
		t.Fatalf("expected HTML5, got %q", s2)
	}

	v1, err := r.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if v1 != 2 {
		t.Fatalf("expected 2, got %d", v1)
	}

	v2, err := r.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 1 {
		t.Fatalf("expected 1, got %d", v2)
	}

	if r.Remaining() != 0 {
		t.Fatalf("expected 0 remaining, got %d", r.Remaining())
	}
}

// Error cases

func TestReaderUnderflow(t *testing.T) {
	r := codec.NewReader([]byte{})

	if _, err := r.ReadBool(); err == nil {
		t.Fatal("expected error reading bool from empty reader")
	}
	if _, err := r.ReadInt32(); err == nil {
		t.Fatal("expected error reading int32 from empty reader")
	}
	if _, err := r.ReadUint16(); err == nil {
		t.Fatal("expected error reading uint16 from empty reader")
	}
	if _, err := r.ReadString(); err == nil {
		t.Fatal("expected error reading string from empty reader")
	}
}

func TestReaderStringTruncated(t *testing.T) {
	w := codec.NewWriter(0)
	w.WriteUint16(100) // says 100 bytes follow
	w.WriteBytes([]byte("short"))

	r := codec.NewReader(w.Bytes())
	_, err := r.ReadString()
	if err == nil {
		t.Fatal("expected error reading truncated string")
	}
}

// Writer Reset

func TestWriterReset(t *testing.T) {
	w := codec.NewWriter(0)
	w.WriteInt32(42)
	if w.Len() != 4 {
		t.Fatalf("expected len 4, got %d", w.Len())
	}
	w.Reset()
	if w.Len() != 0 {
		t.Fatalf("expected len 0 after reset, got %d", w.Len())
	}
}
