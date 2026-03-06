package transport

import "testing"

// TestTopicBuilders validates concrete topic builders.
func TestTopicBuilders(t *testing.T) {
	if got := PacketC2STopic("handshake-security", "abc"); got != "packet.c2s.handshake-security.abc" {
		t.Fatalf("unexpected packet topic: %s", got)
	}
	if got := HandshakeC2STopic("abc"); got != "handshake.c2s.abc" {
		t.Fatalf("unexpected handshake topic: %s", got)
	}
	if got := RoomInputTopic("42"); got != "room.input.42" {
		t.Fatalf("unexpected room input topic: %s", got)
	}
	if got := SessionOutputTopic("abc"); got != "session.output.abc" {
		t.Fatalf("unexpected session output topic: %s", got)
	}
	if got := SocialNotificationTopic("9"); got != "social.notification.9" {
		t.Fatalf("unexpected social topic: %s", got)
	}
	if got := NavigatorRoomUpdatedTopic("42"); got != "navigator.room_updated.42" {
		t.Fatalf("unexpected navigator topic: %s", got)
	}
	if got := ModerationBanIssuedTopic("9"); got != "moderation.ban.issued.9" {
		t.Fatalf("unexpected moderation topic: %s", got)
	}
}

// TestParseSessionOutputTopic validates session output topic parsing.
func TestParseSessionOutputTopic(t *testing.T) {
	if value, ok := ParseSessionOutputTopic("session.output.abc"); !ok || value != "abc" {
		t.Fatalf("expected parsed value, got %q %v", value, ok)
	}
	if _, ok := ParseSessionOutputTopic("session.output"); ok {
		t.Fatalf("expected parse failure")
	}
	if _, ok := ParseSessionOutputTopic("session.output.abc.extra"); ok {
		t.Fatalf("expected parse failure")
	}
}

// TestValidateTopic checks topic validation behavior.
func TestValidateTopic(t *testing.T) {
	if err := ValidateTopic(""); err == nil {
		t.Fatalf("expected empty topic error")
	}
	if err := ValidateTopic("session.authenticated"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
