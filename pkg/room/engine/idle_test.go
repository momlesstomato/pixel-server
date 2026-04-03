package engine

import (
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// newIdleInstance builds a minimal instance without starting it.
func newIdleInstance() *Instance {
	grid := make([][]domain.Tile, 3)
	for y := 0; y < 3; y++ {
		grid[y] = make([]domain.Tile, 3)
		for x := 0; x < 3; x++ {
			grid[y][x] = domain.Tile{X: x, Y: y, State: domain.TileOpen}
		}
	}
	layout := domain.Layout{Slug: "idle_test", DoorX: 0, DoorY: 0, DoorDir: 2, Grid: grid}
	return NewInstance(1, layout, zap.NewNop(), func(_ int, _ []domain.RoomEntity, _ []byte) {})
}

// seedEntity places an entity directly into the instance without a goroutine.
func seedEntity(inst *Instance, vid int) {
	e := &domain.RoomEntity{VirtualID: vid, Statuses: make(map[string]string)}
	inst.entities[vid] = e
	if inst.nextVID <= vid {
		inst.nextVID = vid + 1
	}
}

// TestProcessEntityIdle_SleepAtThreshold verifies IsIdle transitions at idleSleepTicks.
func TestProcessEntityIdle_SleepAtThreshold(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IdleTimer = idleSleepTicks - 1

	newlySlept, kicked := inst.processEntityIdle()

	assert.Len(t, newlySlept, 1)
	assert.Empty(t, kicked)
	assert.Equal(t, 1, newlySlept[0].VirtualID)
	assert.True(t, inst.entities[1].IsIdle)
}

// TestProcessEntityIdle_KickAtThreshold verifies entity removal at idleKickTicks.
func TestProcessEntityIdle_KickAtThreshold(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IdleTimer = idleKickTicks - 1

	_, kicked := inst.processEntityIdle()

	assert.Len(t, kicked, 1)
	assert.Equal(t, 1, kicked[0].VirtualID)
	assert.Equal(t, 0, len(inst.entities))
}

// TestProcessEntityIdle_NoDoubleSleep verifies already-idle entity is not re-reported.
func TestProcessEntityIdle_NoDoubleSleep(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IdleTimer = idleSleepTicks - 1
	inst.entities[1].IsIdle = true

	newlySlept, _ := inst.processEntityIdle()

	assert.Empty(t, newlySlept)
}

// TestProcessEntityIdle_WalkingSkipped verifies walking entities bypass idle increment.
func TestProcessEntityIdle_WalkingSkipped(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IsWalking = true

	inst.processEntityIdle()

	assert.Equal(t, 0, inst.entities[1].IdleTimer)
}

// TestSleepNotifier_FiredOnSleep verifies sleep callback invoked when entity transitions.
func TestSleepNotifier_FiredOnSleep(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IdleTimer = idleSleepTicks - 1
	var sleptVIDs []int
	inst.SetSleepNotifier(func(_ int, vid int, sleeping bool) {
		if sleeping {
			sleptVIDs = append(sleptVIDs, vid)
		}
	})

	newlySlept, _ := inst.processEntityIdle()
	for _, e := range newlySlept {
		inst.sleepNotifier(inst.RoomID, e.VirtualID, true)
	}

	assert.Equal(t, []int{1}, sleptVIDs)
}

// TestKickNotifier_FiredOnKick verifies kick callback invoked when entity times out.
func TestKickNotifier_FiredOnKick(t *testing.T) {
	inst := newIdleInstance()
	seedEntity(inst, 1)
	inst.entities[1].IdleTimer = idleKickTicks - 1
	var kickedVIDs []int
	inst.SetKickNotifier(func(_ int, e domain.RoomEntity) {
		kickedVIDs = append(kickedVIDs, e.VirtualID)
	})

	_, kicked := inst.processEntityIdle()
	for _, e := range kicked {
		inst.kickNotifier(inst.RoomID, e)
	}

	assert.Equal(t, []int{1}, kickedVIDs)
}

// TestResetEntityIdle_ClearsTimerAndFlag verifies both fields reset to zero.
func TestResetEntityIdle_ClearsTimerAndFlag(t *testing.T) {
	e := &domain.RoomEntity{IdleTimer: 100, IsIdle: true}
	resetEntityIdle(e)
	assert.Equal(t, 0, e.IdleTimer)
	assert.False(t, e.IsIdle)
}
