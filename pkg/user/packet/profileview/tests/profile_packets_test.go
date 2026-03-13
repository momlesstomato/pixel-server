package tests

import (
	"testing"

	packetprofileview "github.com/momlesstomato/pixel-server/pkg/user/packet/profileview"
)

// TestUserGetProfilePacketEncodeDecode verifies get profile packet serialization.
func TestUserGetProfilePacketEncodeDecode(t *testing.T) {
	packet := packetprofileview.UserGetProfilePacket{UserID: 9, OpenProfileWindow: true}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := packetprofileview.UserGetProfilePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.UserID != 9 || !decoded.OpenProfileWindow {
		t.Fatalf("unexpected decoded packet %+v", decoded)
	}
}

// TestUserProfilePacketEncode verifies profile packet serialization.
func TestUserProfilePacketEncode(t *testing.T) {
	packet := packetprofileview.UserProfilePacket{
		UserID: 1, Username: "alpha", Figure: "hr-1", Motto: "hello", Registration: "2026-03-13",
		AchievementPoints: 0, FriendsCount: 0, IsOnline: true, OpenProfileWindow: true,
	}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty payload")
	}
}
