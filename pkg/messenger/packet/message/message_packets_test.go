package message

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// TestMessengerNewMessageComposer_EncodeLayout verifies sender|message|seconds field order.
func TestMessengerNewMessageComposer_EncodeLayout(t *testing.T) {
	composer := MessengerNewMessageComposer{SenderID: 9, Message: "hello", SecondsSinceSent: 3}
	body, err := composer.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	r := codec.NewReader(body)
	senderID, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("read sender id failed: %v", err)
	}
	if senderID != 9 {
		t.Fatalf("unexpected sender id: %d", senderID)
	}
	message, err := r.ReadString()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}
	if message != "hello" {
		t.Fatalf("unexpected message: %q", message)
	}
	seconds, err := r.ReadInt32()
	if err != nil {
		t.Fatalf("read seconds failed: %v", err)
	}
	if seconds != 3 {
		t.Fatalf("unexpected seconds value: %d", seconds)
	}
	if r.Remaining() != 0 {
		t.Fatalf("expected fully consumed payload, remaining=%d", r.Remaining())
	}
}
