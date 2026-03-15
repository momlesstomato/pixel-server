package tests

import (
	"testing"

	bannedpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/banned"
	userpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
)

// TestUserInfoPacketEncodeDecode verifies user.info encode/decode symmetry.
func TestUserInfoPacketEncodeDecode(t *testing.T) {
	packet := userpacket.UserInfoPacket{
		UserID: 7, Username: "player", Figure: "hd-180-1", Gender: "M", Motto: "hi",
		RealName: "Player", DirectMail: false, RespectsReceived: 12, RespectsRemaining: 3,
		RespectsPetRemaining: 2, StreamPublishingAllowed: false, LastAccessDate: "2026-03-13T12:00:00Z",
		CanChangeName: true, SafetyLocked: false,
	}
	payload, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := userpacket.UserInfoPacket{}
	if err := decoded.Decode(payload); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Username != packet.Username || decoded.Figure != packet.Figure || decoded.RespectsReceived != packet.RespectsReceived {
		t.Fatalf("unexpected decoded payload %+v", decoded)
	}
}

// TestIdentityPacketEncoders verifies basic identity packet encoders.
func TestIdentityPacketEncoders(t *testing.T) {
	if body, err := (userpacket.UserGetInfoPacket{}).Encode(); err != nil || len(body) != 0 {
		t.Fatalf("expected get info empty payload, got %v and %d", err, len(body))
	}
	noobness := userpacket.UserNoobnessLevelPacket{NoobnessLevel: 2}
	if _, err := noobness.Encode(); err != nil {
		t.Fatalf("expected noobness encode success, got %v", err)
	}
	banned := bannedpacket.UserBannedPacket{Message: "banned"}
	encoded, err := banned.Encode()
	if err != nil {
		t.Fatalf("expected banned encode success, got %v", err)
	}
	decoded := bannedpacket.UserBannedPacket{}
	if err := decoded.Decode(encoded); err != nil || decoded.Message != "banned" {
		t.Fatalf("unexpected banned decode payload %+v err=%v", decoded, err)
	}
}
