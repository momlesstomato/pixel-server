package room

import (
	"context"
	"fmt"
	"testing"

	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	roomstore "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/store"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func noopBroadcast(_ int, _ []domain.RoomEntity, _ []byte) {}

func setupE2E(t *testing.T) (*roomapplication.Service, *roomapplication.EntityService, *roomapplication.ChatService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err = db.AutoMigrate(&roommodel.RoomModel{}, &roommodel.RoomBan{}, &roommodel.RoomRight{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seed := roommodel.RoomModel{
		Slug:      "model_a",
		Heightmap: "xxxx\rxxxx\rx00x\rxxxx",
		DoorX:     1, DoorY:    2, DoorDir: 2,
	}
	if err = db.Create(&seed).Error; err != nil {
		t.Fatalf("seed model: %v", err)
	}
	modelStore, err := roomstore.NewModelStore(db)
	if err != nil {
		t.Fatalf("model store: %v", err)
	}
	banStore, err := roomstore.NewBanStore(db)
	if err != nil {
		t.Fatalf("ban store: %v", err)
	}
	rightsStore, err := roomstore.NewRightsStore(db)
	if err != nil {
		t.Fatalf("rights store: %v", err)
	}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcast)
	svc, err := roomapplication.NewService(modelStore, banStore, rightsStore, mgr, zap.NewNop())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	entitySvc, err := roomapplication.NewEntityService(mgr, zap.NewNop())
	if err != nil {
		t.Fatalf("new entity service: %v", err)
	}
	chatSvc, err := roomapplication.NewChatService(zap.NewNop())
	if err != nil {
		t.Fatalf("new chat service: %v", err)
	}
	return svc, entitySvc, chatSvc
}

// Test13RoomLoadAndEntry verifies a room can be loaded and entered by a player.
func Test13RoomLoadAndEntry(t *testing.T) {
	svc, _, _ := setupE2E(t)
	ctx := context.Background()
	room := domain.Room{ID: 1, ModelSlug: "model_a"}
	inst, err := svc.LoadRoom(ctx, room)
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	entity := domain.NewPlayerEntity(0, 1, "conn1", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	if err := svc.EnterRoom(ctx, inst, &entity, 1, 1); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	entities := inst.Entities()
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
}

// Test13RoomEntityWalk verifies a room entity can request a walk destination.
func Test13RoomEntityWalk(t *testing.T) {
	svc, entitySvc, _ := setupE2E(t)
	ctx := context.Background()
	inst, err := svc.LoadRoom(ctx, domain.Room{ID: 2, ModelSlug: "model_a"})
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	entity := domain.NewPlayerEntity(0, 1, "conn2", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	if err := svc.EnterRoom(ctx, inst, &entity, 2, 1); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	if err := entitySvc.Walk(ctx, inst, &entity, 2, 2); err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// Test13RoomChat verifies proximity talk returns nearby recipients.
func Test13RoomChat(t *testing.T) {
	svc, _, chatSvc := setupE2E(t)
	ctx := context.Background()
	inst, err := svc.LoadRoom(ctx, domain.Room{ID: 3, ModelSlug: "model_a"})
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	sender := domain.NewPlayerEntity(0, 1, "c1", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	target := domain.NewPlayerEntity(1, 2, "c2", "Bob", "", "", "M",
		domain.Tile{X: 1, Y: 3, State: domain.TileOpen})
	if err := svc.EnterRoom(ctx, inst, &sender, 3, 1); err != nil {
		t.Fatalf("enter sender: %v", err)
	}
	if err := svc.EnterRoom(ctx, inst, &target, 3, 2); err != nil {
		t.Fatalf("enter target: %v", err)
	}
	recipients, err := chatSvc.Talk(ctx, inst, &sender, 3, "hello", 0)
	if err != nil {
		t.Fatalf("talk: %v", err)
	}
	if len(recipients) == 0 {
		t.Fatal("expected at least one recipient")
	}
}

// Test13RoomShout verifies shout returns all room entities.
func Test13RoomShout(t *testing.T) {
	svc, _, chatSvc := setupE2E(t)
	ctx := context.Background()
	inst, err := svc.LoadRoom(ctx, domain.Room{ID: 4, ModelSlug: "model_a"})
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	sender := domain.NewPlayerEntity(0, 1, "c3", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	neighbor := domain.NewPlayerEntity(1, 2, "c4", "Bob", "", "", "M",
		domain.Tile{X: 1, Y: 3, State: domain.TileOpen})
	if err := svc.EnterRoom(ctx, inst, &sender, 4, 1); err != nil {
		t.Fatalf("enter sender: %v", err)
	}
	if err := svc.EnterRoom(ctx, inst, &neighbor, 4, 2); err != nil {
		t.Fatalf("enter neighbor: %v", err)
	}
	recipients, err := chatSvc.Shout(ctx, inst, &sender, 4, "hey", 0)
	if err != nil {
		t.Fatalf("shout: %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("expected 2 shout recipients, got %d", len(recipients))
	}
}

// Test13RoomLeave verifies a player entity can leave a room.
func Test13RoomLeave(t *testing.T) {
	svc, _, _ := setupE2E(t)
	ctx := context.Background()
	inst, err := svc.LoadRoom(ctx, domain.Room{ID: 5, ModelSlug: "model_a"})
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	entity := domain.NewPlayerEntity(0, 1, "c5", "Alice", "", "", "M",
		domain.Tile{X: 1, Y: 2, State: domain.TileOpen})
	if err := svc.EnterRoom(ctx, inst, &entity, 5, 1); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	if err := svc.LeaveRoom(ctx, inst, &entity, 5, 1); err != nil {
		t.Fatalf("leave room: %v", err)
	}
	entities := inst.Entities()
	if len(entities) != 0 {
		t.Fatalf("expected 0 entities after leave, got %d", len(entities))
	}
}
