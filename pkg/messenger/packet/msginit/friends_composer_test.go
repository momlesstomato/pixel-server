package msginit

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// TestMessengerFriendsComposer_EncodeFriendEntryLayout verifies friend entry field order and lastAccess presence.
func TestMessengerFriendsComposer_EncodeFriendEntryLayout(t *testing.T) {
	composer := MessengerFriendsComposer{
		TotalFragments: 1,
		FragmentNumber: 0,
		Friends: []FriendEntry{{
			ID: 12, Username: "alice", Gender: 1, Online: true,
			Figure: "hr-165-45", Motto: "hello", Relationship: 2,
		}},
	}
	body, err := composer.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	r := codec.NewReader(body)
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read total fragments failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read fragment number failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read friend count failed: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read friend id failed: %v", err)
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
		t.Fatalf("read followingAllowed failed: %v", err)
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