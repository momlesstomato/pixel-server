package tests

import (
	"testing"

	packetname "github.com/momlesstomato/pixel-server/pkg/user/packet/name"
)

// TestUserNameInputPacketEncodeDecode verifies name input packet serialization.
func TestUserNameInputPacketEncodeDecode(t *testing.T) {
	packet := packetname.UserNameInputPacket{Name: "Alpha"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetname.UserNameInputPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Name != "Alpha" {
		t.Fatalf("unexpected decoded name %q", decoded.Name)
	}
}

// TestUserNameResultPacketEncode verifies name result packet serialization.
func TestUserNameResultPacketEncode(t *testing.T) {
	packet := packetname.UserNameResultPacket{ResultCode: 1, Name: "Alpha", Suggestions: []string{"alpha1", "alpha_2"}}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty payload")
	}
}

// TestUserNameChangePacketEncode verifies name change packet serialization.
func TestUserNameChangePacketEncode(t *testing.T) {
	packet := packetname.UserNameChangePacket{WebID: 1, UserID: 1, NewName: "Bravo"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty payload")
	}
}
