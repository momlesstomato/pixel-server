package engine

import (
	"context"
	"sync"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"go.uber.org/zap"
)

const tickInterval = 500 * time.Millisecond
const channelBuffer = 256
const idleUnloadTicks = 120
const lagThreshold = 30

// idleSleepTicks is the number of ticks before an entity enters sleep state (300s).
const idleSleepTicks = 600

// idleKickTicks is the number of ticks before an idle entity is auto-kicked (1800s).
const idleKickTicks = 1800

// SleepNotifier is called when an entity enters or exits sleep state.
type SleepNotifier func(roomID int, virtualID int, sleeping bool)

// KickNotifier is called when an entity is auto-kicked due to idle timeout.
type KickNotifier func(roomID int, entity domain.RoomEntity)

// DoorExitNotifier is called when an entity walks out through the room door tile.
type DoorExitNotifier func(roomID int, entity domain.RoomEntity)

// TileSeatChecker checks whether a tile has sittable or layable furniture.
// Returns the seat height, furniture direction, whether sitting and whether laying are possible.
type TileSeatChecker func(roomID, x, y int) (height float64, dir int, canSit, canLay bool)

// RoomState identifies the lifecycle phase of a room instance.
type RoomState int

const (
	// StateCreated indicates the room is loaded but not yet ticking.
	StateCreated RoomState = iota
	// StateActive indicates the room is running with entities present.
	StateActive
	// StateIdle indicates the room has no entities and is counting idle ticks.
	StateIdle
	// StateStopped indicates the room goroutine has terminated.
	StateStopped
)

// EntityBroadcaster sends encoded packets to room entities.
type EntityBroadcaster func(roomID int, entities []domain.RoomEntity, data []byte)

// Instance represents a single running room environment.
type Instance struct {
	// RoomID stores the stable room identifier.
	RoomID int
	// Layout stores the parsed room spatial grid.
	Layout domain.Layout
	// entities stores all entities present in the room.
	entities map[int]*domain.RoomEntity
	// nextVID stores the next virtual ID to assign.
	nextVID int
	// state stores the current lifecycle phase.
	state RoomState
	// idleTicks counts consecutive idle ticks.
	idleTicks int
	// msgChan receives room commands.
	msgChan chan Message
	// cancel stops the goroutine context.
	cancel context.CancelFunc
	// done signals goroutine completion.
	done chan struct{}
	// mu protects state reads from outside the goroutine.
	mu sync.RWMutex
	// logger stores the structured logger.
	logger *zap.Logger
	// broadcaster sends updates to connected clients.
	broadcaster EntityBroadcaster
	// sleepNotifier is called when an entity transitions to or from sleep.
	sleepNotifier SleepNotifier
	// kickNotifier is called when an entity is auto-kicked due to idle timeout.
	kickNotifier KickNotifier
	// doorExitNotifier is called when an entity exits through the room door tile.
	doorExitNotifier DoorExitNotifier
	// seatChecker checks whether a tile has sittable or layable furniture.
	seatChecker TileSeatChecker
	// muted reports whether the room has chat globally muted.
	muted bool
}

// NewInstance creates one room instance ready to start.
func NewInstance(roomID int, layout domain.Layout, logger *zap.Logger, broadcaster EntityBroadcaster) *Instance {
	return &Instance{
		RoomID:      roomID,
		Layout:      layout,
		entities:    make(map[int]*domain.RoomEntity),
		nextVID:     1,
		state:       StateCreated,
		msgChan:     make(chan Message, channelBuffer),
		done:        make(chan struct{}),
		logger:      logger.With(zap.Int("room_id", roomID)),
		broadcaster: broadcaster,
	}
}

// Start launches the room goroutine and tick cycle.
func (inst *Instance) Start(ctx context.Context) {
	ctx, inst.cancel = context.WithCancel(ctx)
	go inst.run(ctx)
}

// Send delivers a message into the room channel without blocking.
func (inst *Instance) Send(msg Message) bool {
	select {
	case inst.msgChan <- msg:
		return true
	default:
		return false
	}
}

// Stop requests a graceful room shutdown.
func (inst *Instance) Stop() {
	inst.Send(Message{Type: MsgStop})
}

// SetSleepNotifier configures the callback invoked when an entity sleeps or wakes.
func (inst *Instance) SetSleepNotifier(n SleepNotifier) {
	inst.sleepNotifier = n
}

// SetKickNotifier configures the callback invoked when an entity is auto-kicked.
func (inst *Instance) SetKickNotifier(n KickNotifier) {
	inst.kickNotifier = n
}

// SetDoorExitNotifier configures the callback invoked when an entity exits through the door tile.
func (inst *Instance) SetDoorExitNotifier(n DoorExitNotifier) {
	inst.doorExitNotifier = n
}

// SetTileSeatChecker configures the furniture seat lookup callback for this instance.
func (inst *Instance) SetTileSeatChecker(fn TileSeatChecker) {
	inst.seatChecker = fn
}

// Muted reports whether the room has chat globally muted.
func (inst *Instance) Muted() bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.muted
}

// SetMuted updates the room global chat mute state.
func (inst *Instance) SetMuted(v bool) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	inst.muted = v
}

// run is the main goroutine loop processing ticks and messages.
func (inst *Instance) run(ctx context.Context) {
	defer close(inst.done)
	defer func() {
		if r := recover(); r != nil {
			inst.logger.Error("room panic recovered", zap.Any("panic", r))
		}
		inst.mu.Lock()
		inst.state = StateStopped
		inst.mu.Unlock()
	}()
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	inst.mu.Lock()
	inst.state = StateActive
	inst.mu.Unlock()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-inst.msgChan:
			inst.handleMessage(msg)
		case <-ticker.C:
			inst.processTick()
			if inst.state == StateStopped {
				return
			}
		}
	}
}
