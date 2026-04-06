package tests

import (
	"context"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testLayout builds a small 5x5 open room layout.
func testLayout() domain.Layout {
	grid := make([][]domain.Tile, 5)
	for y := 0; y < 5; y++ {
		grid[y] = make([]domain.Tile, 5)
		for x := 0; x < 5; x++ {
			grid[y][x] = domain.Tile{X: x, Y: y, Z: 0, State: domain.TileOpen}
		}
	}
	return domain.Layout{Slug: "test", DoorX: 0, DoorY: 0, DoorDir: 2, Grid: grid}
}

// testEntity builds a player entity for testing.
func testEntity() *domain.RoomEntity {
	e := domain.NewPlayerEntity(0, 1, "conn-1", "TestUser", "hr-100", "hi", "M",
		domain.Tile{X: 0, Y: 0, Z: 0, State: domain.TileOpen})
	return &e
}

// noopBroadcaster is a no-op entity broadcaster for tests.
func noopBroadcaster(_ int, _ []domain.RoomEntity, _ []byte) {}

// TestNewInstance verifies initial instance state.
func TestNewInstance(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	assert.Equal(t, 1, inst.RoomID)
	assert.Equal(t, engine.StateCreated, inst.State())
	assert.Equal(t, 0, inst.EntityCount())
}

// TestInstance_StartStop verifies goroutine lifecycle.
func TestInstance_StartStop(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	ctx := context.Background()
	inst.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, engine.StateActive, inst.State())
	inst.Stop()
	select {
	case <-inst.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("instance did not stop in time")
	}
	assert.Equal(t, engine.StateStopped, inst.State())
}

// TestInstance_Enter verifies entity enter via message.
func TestInstance_Enter(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	ok := inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	require.True(t, ok)
	err := <-reply
	require.NoError(t, err)
	assert.Equal(t, 1, inst.EntityCount())
	assert.True(t, entity.VirtualID > 0)
}

// TestInstance_Leave verifies entity removal.
func TestInstance_Leave(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	reply2 := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLeave, Entity: entity, Reply: reply2})
	err := <-reply2
	require.NoError(t, err)
	assert.Equal(t, 0, inst.EntityCount())
}

// TestInstance_SendFull verifies channel backpressure.
func TestInstance_SendFull(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	dropped := 0
	for i := 0; i < 300; i++ {
		if !inst.Send(engine.Message{Type: engine.MsgStop}) {
			dropped++
		}
	}
	assert.True(t, dropped > 0, "some messages should be dropped when channel full")
}

// TestInstance_ContextCancel verifies context-driven shutdown.
func TestInstance_ContextCancel(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	ctx, cancel := context.WithCancel(context.Background())
	inst.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-inst.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("instance did not stop on context cancel")
	}
	assert.Equal(t, engine.StateStopped, inst.State())
}

// TestLookTo_Standing verifies body and head both rotate toward target when standing.
func TestLookTo_Standing(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	lookReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLookTo, Entity: entity, TargetX: 4, TargetY: 0, Reply: lookReply})
	require.NoError(t, <-lookReply)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, e.HeadRotation, e.BodyRotation, "body and head must match when standing")
}

// TestLookTo_Self verifies facing direction 2 when clicking own tile.
func TestLookTo_Self(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	selfReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLookTo, Entity: entity, TargetX: entity.Position.X, TargetY: entity.Position.Y, Reply: selfReply})
	require.NoError(t, <-selfReply)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 2, e.BodyRotation, "self-click must face frontward (dir 2)")
	assert.Equal(t, 2, e.HeadRotation, "self-click must face frontward (dir 2)")
}

// TestEjectSittingEntitiesAt_StandsInPlace verifies that ejecting an auto-sitting entity
// clears its sit state and leaves it standing in place without walking toward the door.
func TestEjectSittingEntitiesAt_StandsInPlace(t *testing.T) {
	layout := testLayout()
	inst := engine.NewInstance(1, layout, zap.NewNop(), noopBroadcaster)
	inst.SetTileSeatChecker(func(_ int, x, y int) (float64, int, bool, bool) {
		if x == 2 && y == 2 {
			return 1.0, 2, true, false
		}
		return 0, 0, false, false
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok, "entity must be in room after walk")
	require.True(t, e.IsSitting, "entity should be auto-sitting on arrival")
	require.True(t, e.IsSittingAuto, "auto-sit flag must be set")
	updated := inst.EjectSittingEntitiesAt(2, 2)
	require.Len(t, updated, 1, "one entity must be ejected")
	e, ok = inst.Entity(entity.VirtualID)
	require.True(t, ok, "entity must remain in room after ejection")
	assert.False(t, e.IsSitting, "sit flag must be cleared")
	assert.False(t, e.IsSittingAuto, "auto-sit flag must be cleared")
	assert.False(t, e.IsWalking, "entity must not walk to door after ejection")
	assert.Equal(t, 2, e.Position.X, "entity must stay at ejection tile X")
	assert.Equal(t, 2, e.Position.Y, "entity must stay at ejection tile Y")
	_, hasSit := e.Statuses["sit"]
	assert.False(t, hasSit, "sit status must be removed")
}

// TestEjectSittingEntitiesAt_ClearsManualSit verifies stale sit state is cleared even when the posture was not marked auto-sit.
func TestEjectSittingEntitiesAt_ClearsManualSit(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	sitReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSit, Entity: entity, Reply: sitReply})
	require.NoError(t, <-sitReply)
	updated := inst.EjectSittingEntitiesAt(2, 2)
	require.Len(t, updated, 1)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.False(t, e.IsSitting)
	assert.False(t, e.IsSittingAuto)
	_, hasSit := e.Statuses["sit"]
	assert.False(t, hasSit, "manual sit status must be removed")
	assert.False(t, e.IsWalking, "manual sit ejection must not force a walk")
}

// TestRotateSittingEntitiesAt_RotatesManualSit verifies occupied chairs rotate users even when the posture is not auto-sit.
func TestRotateSittingEntitiesAt_RotatesManualSit(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	sitReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSit, Entity: entity, Reply: sitReply})
	require.NoError(t, <-sitReply)
	updated := inst.RotateSittingEntitiesAt(2, 2, 5)
	require.Len(t, updated, 1)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 4, e.BodyRotation, "odd directions must normalize to even chair rotation")
	assert.Equal(t, 4, e.HeadRotation, "head must rotate with the chair")
}

// TestAutoLayOnArrival verifies lay-capable furniture applies lay posture automatically on arrival.
func TestAutoLayOnArrival(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.SetTileSeatChecker(func(_ int, x, y int) (float64, int, bool, bool) {
		if x == 2 && y == 2 {
			return 0.75, 4, false, true
		}
		return 0, 0, false, false
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.True(t, e.IsSitting)
	assert.True(t, e.IsSittingAuto)
	assert.Equal(t, "0.75", e.Statuses["lay"])
	_, hasSit := e.Statuses["sit"]
	assert.False(t, hasSit)
	assert.Equal(t, 4, e.BodyRotation)
	assert.Equal(t, 4, e.HeadRotation)
}

// TestHandleSitUsesLayPosture verifies manual posture toggling uses lay on lay-capable furniture.
func TestHandleSitUsesLayPosture(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.SetTileSeatChecker(func(_ int, x, y int) (float64, int, bool, bool) {
		if x == 2 && y == 2 {
			return 1.25, 6, false, true
		}
		return 0, 0, false, false
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	inst.EjectSittingEntitiesAt(2, 2)
	sitReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSit, Entity: entity, Reply: sitReply})
	require.NoError(t, <-sitReply)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.True(t, e.IsSitting)
	assert.False(t, e.IsSittingAuto)
	assert.Equal(t, "1.25", e.Statuses["lay"])
	_, hasSit := e.Statuses["sit"]
	assert.False(t, hasSit)
	assert.Equal(t, 6, e.BodyRotation)
	assert.Equal(t, 6, e.HeadRotation)
	toggleReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSit, Entity: entity, Reply: toggleReply})
	require.NoError(t, <-toggleReply)
	e, ok = inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.False(t, e.IsSitting)
	_, hasLay := e.Statuses["lay"]
	assert.False(t, hasLay)
}

// TestWalkToLayTileUsesCanonicalAnchor verifies bed clicks reroute to a single lay anchor tile.
func TestWalkToLayTileUsesCanonicalAnchor(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.SetTileSeatChecker(func(_ int, x, y int) (float64, int, bool, bool) {
		if x == 2 && y == 2 {
			return 0.75, 4, false, true
		}
		return 0, 0, false, false
	})
	inst.SetSeatTargetResolver(func(_ int, x, y int) (int, int, bool) {
		if x >= 2 && x <= 3 && y >= 2 && y <= 4 {
			return 2, 2, true
		}
		return 0, 0, false
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	enterReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: enterReply})
	<-enterReply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 3, TargetY: 4, Reply: walkReply})
	require.NoError(t, <-walkReply)
	time.Sleep(3 * time.Second)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 2, e.Position.X)
	assert.Equal(t, 2, e.Position.Y)
	assert.True(t, e.IsSitting)
	assert.True(t, e.IsSittingAuto)
	assert.Equal(t, "0.75", e.Statuses["lay"])
}

// TestWalkToOccupiedBedAnchorBlocked verifies a second user cannot use another row of the same occupied bed.
func TestWalkToOccupiedBedAnchorBlocked(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.SetTileSeatChecker(func(_ int, x, y int) (float64, int, bool, bool) {
		if x == 2 && y == 2 {
			return 0.75, 4, false, true
		}
		return 0, 0, false, false
	})
	inst.SetSeatTargetResolver(func(_ int, x, y int) (int, int, bool) {
		if x >= 2 && x <= 3 && y >= 2 && y <= 4 {
			return 2, 2, true
		}
		return 0, 0, false
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	first := testEntity()
	enterFirstReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: first, Reply: enterFirstReply})
	<-enterFirstReply
	firstWalkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: first, TargetX: 3, TargetY: 4, Reply: firstWalkReply})
	require.NoError(t, <-firstWalkReply)
	time.Sleep(3 * time.Second)
	secondEntity := domain.NewPlayerEntity(0, 2, "conn-2", "SecondUser", "hr-200", "hi", "M", domain.Tile{X: 0, Y: 0, Z: 0, State: domain.TileOpen})
	second := &secondEntity
	enterSecondReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: second, Reply: enterSecondReply})
	<-enterSecondReply
	secondWalkReply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: second, TargetX: 3, TargetY: 3, Reply: secondWalkReply})
	assert.ErrorIs(t, <-secondWalkReply, domain.ErrPathBlocked)
	secondState, ok := inst.Entity(second.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 0, secondState.Position.X)
	assert.Equal(t, 0, secondState.Position.Y)
	assert.False(t, secondState.IsSitting)
}
