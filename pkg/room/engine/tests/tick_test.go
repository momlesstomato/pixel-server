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

// TestTick_Walk verifies entity walks toward target.
func TestTick_Walk(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{
		Type: engine.MsgWalk, Entity: entity,
		TargetX: 4, TargetY: 4, Reply: walkReply,
	})
	err := <-walkReply
	require.NoError(t, err)
	time.Sleep(3 * time.Second)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 4, e.Position.X)
	assert.Equal(t, 4, e.Position.Y)
}

// TestTick_WalkBlocked verifies path blocked returns error.
func TestTick_WalkBlocked(t *testing.T) {
	grid := make([][]domain.Tile, 3)
	for y := 0; y < 3; y++ {
		grid[y] = make([]domain.Tile, 3)
		for x := 0; x < 3; x++ {
			grid[y][x] = domain.Tile{X: x, Y: y, Z: 0, State: domain.TileOpen}
		}
	}
	grid[0][1].State = domain.TileBlocked
	grid[1][0].State = domain.TileBlocked
	grid[1][1].State = domain.TileBlocked
	layout := domain.Layout{Slug: "test", DoorX: 0, DoorY: 0, DoorDir: 2, Grid: grid}
	inst := engine.NewInstance(1, layout, zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	walkReply := make(chan error, 1)
	inst.Send(engine.Message{
		Type: engine.MsgWalk, Entity: entity,
		TargetX: 2, TargetY: 2, Reply: walkReply,
	})
	err := <-walkReply
	assert.ErrorIs(t, err, domain.ErrPathBlocked)
}

// TestTick_EntityIdle verifies entity idle timer increments.
func TestTick_EntityIdle(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	time.Sleep(2 * time.Second)
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.True(t, e.IdleTimer > 0)
}

// TestTick_LeaveNilEntity verifies nil entity error handling.
func TestTick_LeaveNilEntity(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLeave, Entity: nil, Reply: reply})
	err := <-reply
	assert.ErrorIs(t, err, domain.ErrEntityNotFound)
}

// TestTick_WalkNilEntity verifies walk with nil entity returns error.
func TestTick_WalkNilEntity(t *testing.T) {
	inst := engine.NewInstance(1, testLayout(), zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: nil, Reply: reply})
	err := <-reply
	assert.ErrorIs(t, err, domain.ErrEntityNotFound)
}

// TestTick_EnterPositionsDoorTile verifies entity spawns at door.
func TestTick_EnterPositionsDoorTile(t *testing.T) {
	layout := testLayout()
	layout.DoorX = 2
	layout.DoorY = 3
	layout.DoorDir = 4
	inst := engine.NewInstance(1, layout, zap.NewNop(), noopBroadcaster)
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	e, ok := inst.Entity(entity.VirtualID)
	require.True(t, ok)
	assert.Equal(t, 2, e.Position.X)
	assert.Equal(t, 3, e.Position.Y)
	assert.Equal(t, 4, e.BodyRotation)
}

// TestTick_DoorExitNotifier verifies the notifier fires and entity is removed when walking to the door tile.
func TestTick_DoorExitNotifier(t *testing.T) {
	layout := testLayout()
	layout.DoorX = 2
	layout.DoorY = 2
	inst := engine.NewInstance(1, layout, zap.NewNop(), noopBroadcaster)
	exitCh := make(chan int, 1)
	inst.SetDoorExitNotifier(func(roomID int, _ domain.RoomEntity) {
		exitCh <- roomID
	})
	inst.Start(context.Background())
	defer inst.Stop()
	time.Sleep(50 * time.Millisecond)
	entity := testEntity()
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply})
	<-reply
	walkAway := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 0, TargetY: 0, Reply: walkAway})
	require.NoError(t, <-walkAway)
	time.Sleep(2 * time.Second)
	walkBack := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: 2, TargetY: 2, Reply: walkBack})
	require.NoError(t, <-walkBack)
	select {
	case roomID := <-exitCh:
		assert.Equal(t, 1, roomID)
	case <-time.After(3 * time.Second):
		t.Fatal("door exit notifier was not fired")
	}
	_, stillPresent := inst.Entity(entity.VirtualID)
	assert.False(t, stillPresent, "entity must be removed after door exit")
}
