package messenger

import (
	"context"
	"fmt"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
	messengerstore "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/store"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	sdk "github.com/momlesstomato/pixel-sdk"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// e2eRegistry is a no-op session registry for tests.
type e2eRegistry struct{}

func (e2eRegistry) Register(_ coreconnection.Session) error              { return nil }
func (e2eRegistry) FindByUserID(_ int) (coreconnection.Session, bool)    { return coreconnection.Session{}, false }
func (e2eRegistry) FindByConnID(_ string) (coreconnection.Session, bool) { return coreconnection.Session{}, false }
func (e2eRegistry) Touch(_ string) error                                 { return nil }
func (e2eRegistry) Remove(_ string)                                      {}
func (e2eRegistry) ListAll() ([]coreconnection.Session, error)           { return nil, nil }

// e2eBroadcaster is a no-op broadcaster for tests.
type e2eBroadcaster struct{}

func (e2eBroadcaster) Publish(_ context.Context, _ string, _ []byte) error { return nil }
func (e2eBroadcaster) Subscribe(_ context.Context, _ string) (<-chan []byte, coreconnection.Disposable, error) {
	ch := make(chan []byte)
	close(ch)
	return ch, coreconnection.DisposeFunc(func() error { return nil }), nil
}

// setupE2E opens an in-memory SQLite database, seeds two users, and returns a messenger service.
func setupE2E(t *testing.T) (*messengerapplication.Service, int, int) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil { t.Fatalf("open db: %v", err) }
	if err = db.AutoMigrate(&usermodel.Record{}); err != nil { t.Fatalf("migrate users: %v", err) }
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS messenger_friendships (user_one_id INTEGER NOT NULL, user_two_id INTEGER NOT NULL, relationship INTEGER NOT NULL DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_one_id, user_two_id))`,
		`CREATE TABLE IF NOT EXISTS friend_requests (id INTEGER PRIMARY KEY AUTOINCREMENT, from_user_id INTEGER NOT NULL, to_user_id INTEGER NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE (from_user_id, to_user_id))`,
		`CREATE TABLE IF NOT EXISTS offline_messages (id INTEGER PRIMARY KEY AUTOINCREMENT, from_user_id INTEGER NOT NULL, to_user_id INTEGER NOT NULL, message TEXT NOT NULL DEFAULT '', sent_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS messenger_message_log (id INTEGER PRIMARY KEY AUTOINCREMENT, from_user_id INTEGER NOT NULL, to_user_id INTEGER NOT NULL, message TEXT NOT NULL DEFAULT '', sent_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, s := range sqls {
		if err = db.Exec(s).Error; err != nil { t.Fatalf("create table: %v", err) }
	}
	uA, uB := usermodel.Record{Username: "alice"}, usermodel.Record{Username: "bob"}
	if err = db.Create(&uA).Error; err != nil { t.Fatalf("seed alice: %v", err) }
	if err = db.Create(&uB).Error; err != nil { t.Fatalf("seed bob: %v", err) }
	repo, err := messengerstore.NewRepository(db)
	if err != nil { t.Fatalf("new repo: %v", err) }
	svc, err := messengerapplication.NewService(repo, e2eRegistry{}, e2eBroadcaster{}, messengerapplication.Config{})
	if err != nil { t.Fatalf("new service: %v", err) }
	var noop func(sdk.Event)
	svc.SetEventFirer(noop)
	return svc, int(uA.ID), int(uB.ID)
}

// Test10MessengerFriendRequestFlow verifies end-to-end send and accept request.
func Test10MessengerFriendRequestFlow(t *testing.T) {
	svc, aliceID, bobID := setupE2E(t)
	ctx := context.Background()
	req, accepted, err := svc.SendRequest(ctx, "conn1", aliceID, "bob")
	if err != nil || accepted {
		t.Fatalf("unexpected send result err=%v accepted=%v", err, accepted)
	}
	if err := svc.AcceptRequest(ctx, bobID, req.ID); err != nil {
		t.Fatalf("accept request: %v", err)
	}
	if ok, err := svc.AreFriends(ctx, aliceID, bobID); err != nil || !ok {
		t.Fatalf("expected friends after accept err=%v ok=%v", err, ok)
	}
}

// Test10MessengerFriendshipManagement verifies add, list, count, and remove.
func Test10MessengerFriendshipManagement(t *testing.T) {
	svc, aliceID, bobID := setupE2E(t)
	ctx := context.Background()
	if err := svc.AddFriendship(ctx, aliceID, bobID); err != nil { t.Fatalf("add: %v", err) }
	if friends, err := svc.ListFriends(ctx, aliceID); err != nil || len(friends) == 0 {
		t.Fatalf("list friends err=%v len=%d", err, len(friends))
	}
	if n, err := svc.FriendCount(ctx, aliceID); err != nil || n != 1 {
		t.Fatalf("count err=%v n=%d", err, n)
	}
	if err := svc.RemoveFriendship(ctx, aliceID, bobID); err != nil { t.Fatalf("remove: %v", err) }
	if n, err := svc.FriendCount(ctx, aliceID); err != nil || n != 0 {
		t.Fatalf("count after remove err=%v n=%d", err, n)
	}
}

// Test10MessengerOfflineMessageFlow verifies offline message save and delivery.
func Test10MessengerOfflineMessageFlow(t *testing.T) {
	svc, aliceID, bobID := setupE2E(t)
	ctx := context.Background()
	if err := svc.AddFriendship(ctx, aliceID, bobID); err != nil { t.Fatalf("add friends: %v", err) }
	if err := svc.SendMessage(ctx, "conn1", aliceID, bobID, "hello bob"); err != nil {
		t.Fatalf("send message: %v", err)
	}
	msgs, err := svc.DeliverOfflineMessages(ctx, bobID)
	if err != nil || len(msgs) == 0 {
		t.Fatalf("deliver offline err=%v len=%d", err, len(msgs))
	}
	if msgs[0].Message != "hello bob" {
		t.Fatalf("unexpected message: %s", msgs[0].Message)
	}
}

// Test10MessengerSearchUsersFlow verifies username-based user search.
func Test10MessengerSearchUsersFlow(t *testing.T) {
	svc, _, _ := setupE2E(t)
	results, err := svc.SearchUsers(context.Background(), "ali", 10)
	if err != nil || len(results) == 0 {
		t.Fatalf("search err=%v len=%d", err, len(results))
	}
	if results[0].Username != "alice" {
		t.Fatalf("unexpected username: %s", results[0].Username)
	}
}

// Test10MessengerRelationshipFlow verifies setting and reporting relationship types.
func Test10MessengerRelationshipFlow(t *testing.T) {
	svc, aliceID, bobID := setupE2E(t)
	ctx := context.Background()
	if err := svc.AddFriendship(ctx, aliceID, bobID); err != nil { t.Fatalf("add: %v", err) }
	if err := svc.SetRelationship(ctx, aliceID, bobID, domain.RelationshipHeart); err != nil {
		t.Fatalf("set relationship: %v", err)
	}
	counts, err := svc.GetRelationshipCounts(ctx, aliceID)
	if err != nil || len(counts) == 0 {
		t.Fatalf("counts err=%v len=%d", err, len(counts))
	}
	if counts[0].Type != domain.RelationshipHeart || counts[0].Count != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}
