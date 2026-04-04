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
