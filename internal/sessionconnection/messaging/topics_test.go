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
