package telemetry

import "testing"

// TestClientLatencyTestEncodeDecode verifies client.latency_test packet round-trip behavior.
func TestClientLatencyTestEncodeDecode(t *testing.T) {
	source := ClientLatencyTestPacket{RequestID: 89}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientLatencyTestPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.RequestID != source.RequestID {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}

// TestClientLatencyResponseEncodeDecode verifies client.latency_response packet round-trip behavior.
func TestClientLatencyResponseEncodeDecode(t *testing.T) {
	source := ClientLatencyResponsePacket{RequestID: 54}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientLatencyResponsePacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.RequestID != source.RequestID {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}

// TestClientLatencyDecodeRejectsTrailingBytes verifies strict body validation behavior.
func TestClientLatencyDecodeRejectsTrailingBytes(t *testing.T) {
	packet := ClientLatencyTestPacket{}
	if err := packet.Decode([]byte{0, 0, 0, 1, 2}); err == nil {
		t.Fatalf("expected decode failure for trailing bytes")
	}
}
