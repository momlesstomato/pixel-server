package codec

import "testing"

// TestReaderWriterRoundTrip verifies primitive payload round-trip behavior.
func TestReaderWriterRoundTrip(t *testing.T) {
	writer := NewWriter()
	writer.WriteInt32(42)
	writer.WriteUint16(7)
	writer.WriteBool(true)
	if err := writer.WriteString("hello"); err != nil {
		t.Fatalf("expected write string success, got %v", err)
	}
	reader := NewReader(writer.Bytes())
	number, err := reader.ReadInt32()
	if err != nil || number != 42 {
		t.Fatalf("unexpected int32 read: %d %v", number, err)
	}
	u16, err := reader.ReadUint16()
	if err != nil || u16 != 7 {
		t.Fatalf("unexpected uint16 read: %d %v", u16, err)
	}
	flag, err := reader.ReadBool()
	if err != nil || !flag {
		t.Fatalf("unexpected bool read: %v %v", flag, err)
	}
	text, err := reader.ReadString()
	if err != nil || text != "hello" {
		t.Fatalf("unexpected string read: %s %v", text, err)
	}
	if reader.Remaining() != 0 {
		t.Fatalf("expected fully consumed payload, got %d", reader.Remaining())
	}
}

// TestReaderWriterValidation verifies primitive validation behavior.
func TestReaderWriterValidation(t *testing.T) {
	reader := NewReader([]byte{0})
	if _, err := reader.ReadInt32(); err == nil {
		t.Fatalf("expected int32 read failure")
	}
	if err := NewWriter().WriteString(string(make([]byte, 70000))); err == nil {
		t.Fatalf("expected write string length failure")
	}
}
