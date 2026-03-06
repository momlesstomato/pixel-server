package transport

import "testing"

// TestTopicBuilders validates concrete topic builders.
func TestTopicBuilders(t *testing.T) {
	if got := PacketC2STopic("handshake-security", "abc"); got != "packet.c2s.handshake-security.abc" {
		t.Fatalf("unexpected packet topic: %s", got)
	}
}

// TestParsePacketC2STopic validates packet ingress topic parsing.
func TestParsePacketC2STopic(t *testing.T) {
	realm, sessionID, ok := ParsePacketC2STopic("packet.c2s.handshake-security.s1")
	if !ok || realm != "handshake-security" || sessionID != "s1" {
		t.Fatalf("unexpected parse result: %q %q %v", realm, sessionID, ok)
	}
	if _, _, ok := ParsePacketC2STopic("packet.c2s.handshake-security"); ok {
		t.Fatalf("expected parse failure")
	}
	if _, _, ok := ParsePacketC2STopic("session.output.s1"); ok {
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
