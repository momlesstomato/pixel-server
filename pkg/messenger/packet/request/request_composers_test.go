package request

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// TestMessengerFriendUpdateComposer_EncodeAddedLayout verifies added entry layout includes lastAccess.
func TestMessengerFriendUpdateComposer_EncodeAddedLayout(t *testing.T) {
	composer := MessengerFriendUpdateComposer{Entries: []FriendUpdateEntry{{
		Action: 1, FriendID: 9, Username: "bob", Gender: 0, Online: true,
		Figure: "hd-180-1", Motto: "hey", Relationship: 1,
	}}}
	body, err := composer.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	r := codec.NewReader(body)
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read categories failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read updates count failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read action failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read id failed: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("read username failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read gender failed: %v", err)
	}
	if _, err = r.ReadBool(); err != nil {
		t.Fatalf("read online failed: %v", err)
	}
	if _, err = r.ReadBool(); err != nil {
		t.Fatalf("read following allowed failed: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("read figure failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read category failed: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("read motto failed: %v", err)
	}
	realName, err := r.ReadString()
	if err != nil {
		t.Fatalf("read real name failed: %v", err)
	}
	if realName != "" {
		t.Fatalf("expected empty real name, got %q", realName)
	}
	lastAccess, err := r.ReadString()
	if err != nil {
		t.Fatalf("read last access failed: %v", err)
	}
	if lastAccess != "" {
		t.Fatalf("expected empty last access, got %q", lastAccess)
	}
	if _, err = r.ReadBool(); err != nil {
		t.Fatalf("read persisted message flag failed: %v", err)
	}
	if _, err = r.ReadBool(); err != nil {
		t.Fatalf("read vip flag failed: %v", err)
	}
	if _, err = r.ReadBool(); err != nil {
		t.Fatalf("read pocket flag failed: %v", err)
	}
	if _, err = r.ReadUint16(); err != nil {
		t.Fatalf("read relationship failed: %v", err)
	}
	if r.Remaining() != 0 {
		t.Fatalf("expected fully consumed payload, remaining=%d", r.Remaining())
	}
}
