package security

import "testing"

// TestClientMachineIDEncodeDecode verifies client machine_id packet round-trip behavior.
func TestClientMachineIDEncodeDecode(t *testing.T) {
	source := ClientMachineIDPacket{MachineID: "m", Fingerprint: "f", Capabilities: "c"}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientMachineIDPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.MachineID != source.MachineID || decoded.Capabilities != source.Capabilities {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}

// TestServerMachineIDEncodeDecode verifies server machine_id packet round-trip behavior.
func TestServerMachineIDEncodeDecode(t *testing.T) {
	source := ServerMachineIDPacket{MachineID: "machine"}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ServerMachineIDPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.MachineID != source.MachineID {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}
