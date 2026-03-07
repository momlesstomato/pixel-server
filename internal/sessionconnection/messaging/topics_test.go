package messaging

import "testing"

// TestOutputTopics validates output topic builders and parser.
func TestOutputTopics(t *testing.T) {
	if got := OutputTopic("abc"); got != "session.output.abc" {
		t.Fatalf("unexpected output topic: %s", got)
	}
	if got := OutputWildcardTopic(); got != "session.output.>" {
		t.Fatalf("unexpected wildcard topic: %s", got)
	}
	value, ok := ParseOutputTopic("session.output.abc")
	if !ok || value != "abc" {
		t.Fatalf("unexpected parse result: %q %v", value, ok)
	}
	if _, ok := ParseOutputTopic("session.output"); ok {
		t.Fatalf("expected parse failure")
	}
}

// TestDisconnectTopics validates disconnect topic builders and parser.
func TestDisconnectTopics(t *testing.T) {
	if got := DisconnectTopic("abc"); got != "session.disconnect.abc" {
		t.Fatalf("unexpected disconnect topic: %s", got)
	}
	if got := DisconnectWildcardTopic(); got != "session.disconnect.>" {
		t.Fatalf("unexpected wildcard topic: %s", got)
	}
	value, ok := ParseDisconnectTopic("session.disconnect.abc")
	if !ok || value != "abc" {
		t.Fatalf("unexpected parse result: %q %v", value, ok)
	}
	if _, ok := ParseDisconnectTopic("session.disconnect"); ok {
		t.Fatalf("expected parse failure")
	}
}

// TestPacketTopics validates session-connection packet topic contracts.
func TestPacketTopics(t *testing.T) {
	if got := PacketIngressTopic("abc"); got != "packet.c2s.session-connection.abc" {
		t.Fatalf("unexpected ingress topic: %s", got)
	}
	if got := PacketIngressWildcardTopic(); got != "packet.c2s.session-connection.*" {
		t.Fatalf("unexpected ingress wildcard topic: %s", got)
	}
	value, ok := ParsePacketIngressTopic("packet.c2s.session-connection.abc")
	if !ok || value != "abc" {
		t.Fatalf("unexpected parse result: %q %v", value, ok)
	}
	if _, ok := ParsePacketIngressTopic("packet.c2s.handshake-security.abc"); ok {
		t.Fatalf("expected parse failure")
	}
}
