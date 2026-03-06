package messaging

import "testing"

// TestPacketIngressTopics validates auth packet topic builders and parser.
func TestPacketIngressTopics(t *testing.T) {
	if got := PacketIngressTopic("abc"); got != "packet.c2s.handshake-security.abc" {
		t.Fatalf("unexpected ingress topic: %s", got)
	}
	if got := PacketIngressWildcardTopic(); got != "packet.c2s.handshake-security.*" {
		t.Fatalf("unexpected wildcard topic: %s", got)
	}
	sessionID, ok := ParsePacketIngressTopic("packet.c2s.handshake-security.s1")
	if !ok || sessionID != "s1" {
		t.Fatalf("unexpected parse result: %q %v", sessionID, ok)
	}
	if _, ok := ParsePacketIngressTopic("packet.c2s.room.s1"); ok {
		t.Fatalf("expected parse failure")
	}
}
