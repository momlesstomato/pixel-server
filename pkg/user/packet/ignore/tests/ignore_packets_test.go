package tests

import (
	"testing"

	packetignore "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
)

// TestUserIgnorePacketEncodeDecode verifies ignore packet serialization.
func TestUserIgnorePacketEncodeDecode(t *testing.T) {
	packet := packetignore.UserIgnorePacket{Username: "target"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetignore.UserIgnorePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Username != "target" {
		t.Fatalf("unexpected decoded username %q", decoded.Username)
	}
}

// TestUserIgnoreByIDPacketEncodeDecode verifies ignore-by-id packet serialization.
func TestUserIgnoreByIDPacketEncodeDecode(t *testing.T) {
	packet := packetignore.UserIgnoreByIDPacket{UserID: 7}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetignore.UserIgnoreByIDPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.UserID != 7 {
		t.Fatalf("unexpected decoded user id %d", decoded.UserID)
	}
}

// TestServerIgnorePacketsEncode verifies server ignore packet serialization.
func TestServerIgnorePacketsEncode(t *testing.T) {
	users := packetignore.UserIgnoredUsersPacket{Usernames: []string{"a", "b"}}
	if body, err := users.Encode(); err != nil || len(body) == 0 {
		t.Fatalf("expected ignored users packet encode success, got len=%d err=%v", len(body), err)
	}
	result := packetignore.UserIgnoreResultPacket{Result: 1, Name: "target"}
	if body, err := result.Encode(); err != nil || len(body) == 0 {
		t.Fatalf("expected ignore result packet encode success, got len=%d err=%v", len(body), err)
	}
}
