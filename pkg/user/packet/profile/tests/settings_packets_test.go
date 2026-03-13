package tests

import (
	"testing"

	userpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
)

// TestSettingsAndAccessPackets verifies settings and access packet encoding behavior.
func TestSettingsAndAccessPackets(t *testing.T) {
	settings := userpacket.UserSettingsPacket{VolumeSystem: 30, VolumeFurni: 40, VolumeTrax: 50, OldChat: true, RoomInvites: true, CameraFollow: true, Flags: 0, ChatType: 1}
	body, err := settings.Encode()
	if err != nil {
		t.Fatalf("expected settings encode success, got %v", err)
	}
	decoded := userpacket.UserSettingsPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected settings decode success, got %v", err)
	}
	if decoded.VolumeSystem != 30 || decoded.ChatType != 1 {
		t.Fatalf("unexpected settings decode payload %+v", decoded)
	}
	permissions := userpacket.UserPermissionsPacket{ClubLevel: 1, SecurityLevel: 0, IsAmbassador: false}
	if _, err := permissions.Encode(); err != nil {
		t.Fatalf("expected permissions encode success, got %v", err)
	}
	perks := userpacket.UserPerksPacket{Entries: []userpacket.PerkEntry{{Code: "CAMERA", IsAllowed: true}}}
	encodedPerks, err := perks.Encode()
	if err != nil {
		t.Fatalf("expected perks encode success, got %v", err)
	}
	decodedPerks := userpacket.UserPerksPacket{}
	if err := decodedPerks.Decode(encodedPerks); err != nil || len(decodedPerks.Entries) != 1 {
		t.Fatalf("unexpected perks decode payload %+v err=%v", decodedPerks, err)
	}
}

// TestRequestPacketsDecode verifies C2S packet decode behavior.
func TestRequestPacketsDecode(t *testing.T) {
	motto := userpacket.UserUpdateMottoPacket{Motto: "hello"}
	encodedMotto, _ := motto.Encode()
	decodedMotto := userpacket.UserUpdateMottoPacket{}
	if err := decodedMotto.Decode(encodedMotto); err != nil || decodedMotto.Motto != "hello" {
		t.Fatalf("unexpected motto decode payload %+v err=%v", decodedMotto, err)
	}
	figure := userpacket.UserUpdateFigurePacket{Gender: "M", Figure: "hd-180-1"}
	encodedFigure, _ := figure.Encode()
	decodedFigure := userpacket.UserUpdateFigurePacket{}
	if err := decodedFigure.Decode(encodedFigure); err != nil || decodedFigure.Figure != figure.Figure {
		t.Fatalf("unexpected figure decode payload %+v err=%v", decodedFigure, err)
	}
	respect := userpacket.UserRespectPacket{UserID: 9}
	encodedRespect, _ := respect.Encode()
	decodedRespect := userpacket.UserRespectPacket{}
	if err := decodedRespect.Decode(encodedRespect); err != nil || decodedRespect.UserID != 9 {
		t.Fatalf("unexpected respect decode payload %+v err=%v", decodedRespect, err)
	}
}
