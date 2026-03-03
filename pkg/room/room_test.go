package room_test

import (
	"testing"

	"pixel-server/pkg/room"
)

func TestNewRoomWorld(t *testing.T) {
	rw := room.NewRoomWorld()
	if rw.World == nil {
		t.Fatal("expected non-nil World")
	}
	if rw.AvatarMapper == nil {
		t.Fatal("expected non-nil AvatarMapper")
	}
	if rw.WalkFilter == nil {
		t.Fatal("expected non-nil WalkFilter")
	}
}

func TestSpawnAvatar(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnAvatar(42, 1, 5, 10, 0.0)
	if e.IsZero() {
		t.Fatal("expected non-zero entity")
	}
	pos, tile, _, avatar := rw.AvatarMapper.Get(e)
	if pos.X != 5 || pos.Y != 10 || pos.Z != 0 {
		t.Fatalf("position mismatch: got (%v,%v,%v)", pos.X, pos.Y, pos.Z)
	}
	if tile.X != 5 || tile.Y != 10 {
		t.Fatalf("tile mismatch: got (%v,%v)", tile.X, tile.Y)
	}
	if avatar.UserID != 42 || avatar.RoomUnit != 1 {
		t.Fatalf("avatar mismatch: got userID=%v roomUnit=%v", avatar.UserID, avatar.RoomUnit)
	}
}

func TestSpawnBot(t *testing.T) {
	rw := room.NewRoomWorld()
	lines := []string{"Hello!", "Bye!"}
	e := rw.SpawnBot(1, lines, 3, 4, 1.5)
	if e.IsZero() {
		t.Fatal("expected non-zero entity")
	}
	pos, tile, bot := rw.BotMapper.Get(e)
	if pos.X != 3 || pos.Y != 4 || pos.Z != 1.5 {
		t.Fatalf("position mismatch: %v", pos)
	}
	if tile.X != 3 || tile.Y != 4 {
		t.Fatalf("tile mismatch: %v", tile)
	}
	if bot.Behaviour != 1 || len(bot.ChatLines) != 2 {
		t.Fatalf("bot mismatch: %v", bot)
	}
}

func TestSpawnPet(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnPet(100, 80, 7, 8, 0)
	if e.IsZero() {
		t.Fatal("expected non-zero entity")
	}
	_, _, pet := rw.PetMapper.Get(e)
	if pet.HappyLevel != 100 || pet.Energy != 80 {
		t.Fatalf("pet mismatch: %v", pet)
	}
}

func TestSpawnItem(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnItem(999, "state:1", 2, 3, 0.5)
	if e.IsZero() {
		t.Fatal("expected non-zero entity")
	}
	_, _, item := rw.ItemMapper.Get(e)
	if item.FurniID != 999 || item.ExtraData != "state:1" {
		t.Fatalf("item mismatch: %v", item)
	}
}

func TestRemoveEntity(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnAvatar(1, 1, 0, 0, 0)
	rw.RemoveEntity(e)
	if rw.World.Alive(e) {
		t.Fatal("entity should be dead after removal")
	}
}

func TestMovementSystem(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnAvatar(1, 1, 0, 0, 0)
	_, _, path, _ := rw.AvatarMapper.Get(e)
	path.Steps = []room.PathStep{
		{X: 1, Y: 0, Z: 0},
		{X: 2, Y: 0, Z: 0},
		{X: 3, Y: 0, Z: 0.5},
	}
	path.Cursor = 0

	room.MovementSystem(rw)
	pos, tile, _, _ := rw.AvatarMapper.Get(e)
	if pos.X != 1 || pos.Y != 0 {
		t.Fatalf("tick 1: expected (1,0), got (%v,%v)", pos.X, pos.Y)
	}
	if tile.X != 1 || tile.Y != 0 {
		t.Fatalf("tick 1 tile: expected (1,0), got (%v,%v)", tile.X, tile.Y)
	}

	room.MovementSystem(rw)
	pos, _, _, _ = rw.AvatarMapper.Get(e)
	if pos.X != 2 {
		t.Fatalf("tick 2: expected X=2, got %v", pos.X)
	}

	room.MovementSystem(rw)
	pos, _, _, _ = rw.AvatarMapper.Get(e)
	if pos.X != 3 || pos.Z != 0.5 {
		t.Fatalf("tick 3: expected (3,_,0.5), got (%v,%v,%v)", pos.X, pos.Y, pos.Z)
	}

	room.MovementSystem(rw)
	pos2, _, _, _ := rw.AvatarMapper.Get(e)
	if pos2.X != 3 || pos2.Z != 0.5 {
		t.Fatalf("tick 4: position should not change")
	}
}

func TestChatCooldownSystem(t *testing.T) {
	rw := room.NewRoomWorld()
	e := rw.SpawnAvatar(1, 1, 0, 0, 0)
	cd := rw.CooldownMapper.Get(e)
	cd.Counter = 5

	room.ChatCooldownSystem(rw, 0)
	cd = rw.CooldownMapper.Get(e)
	if cd.Counter != 5 {
		t.Fatalf("even tick: expected 5, got %d", cd.Counter)
	}

	room.ChatCooldownSystem(rw, 1)
	cd = rw.CooldownMapper.Get(e)
	if cd.Counter != 4 {
		t.Fatalf("odd tick: expected 4, got %d", cd.Counter)
	}

	room.ChatCooldownSystem(rw, 3)
	cd = rw.CooldownMapper.Get(e)
	if cd.Counter != 3 {
		t.Fatalf("odd tick 2: expected 3, got %d", cd.Counter)
	}
}

func TestWalkPathHelpers(t *testing.T) {
	wp := room.WalkPath{
		Steps:  []room.PathStep{{X: 1, Y: 2, Z: 0}, {X: 3, Y: 4, Z: 0.5}},
		Cursor: 0,
	}
	if !wp.HasSteps() {
		t.Fatal("expected HasSteps=true")
	}
	s := wp.Current()
	if s.X != 1 || s.Y != 2 {
		t.Fatalf("expected first step (1,2), got (%d,%d)", s.X, s.Y)
	}
	wp.Advance()
	s = wp.Current()
	if s.X != 3 {
		t.Fatalf("expected second step X=3, got %d", s.X)
	}
	wp.Advance()
	if wp.HasSteps() {
		t.Fatal("expected HasSteps=false after exhaustion")
	}
	zero := wp.Current()
	if zero.X != 0 || zero.Y != 0 || zero.Z != 0 {
		t.Fatal("expected zero PathStep after exhaustion")
	}
}
