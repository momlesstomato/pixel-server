package realtime

import (
	"sync"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/engine"
)

var sharedRuntimeStates sync.Map

// sharedRuntimeState stores mutable room runtime state shared by runtimes using the same manager.
type sharedRuntimeState struct {
	// mu protects shared room runtime state.
	mu sync.RWMutex
	// connRooms stores the active room identifier by connection.
	connRooms map[string]int
	// roomUserMutes stores per-room per-user mute expiries.
	roomUserMutes map[int]map[int]time.Time
	// access stores shared room access state.
	access *accessState
}

// loadSharedRuntimeState returns the shared room runtime state for one room manager.
func loadSharedRuntimeState(manager *engine.Manager) *sharedRuntimeState {
	if state, ok := sharedRuntimeStates.Load(manager); ok {
		return state.(*sharedRuntimeState)
	}
	created := &sharedRuntimeState{
		connRooms:     make(map[string]int),
		roomUserMutes: make(map[int]map[int]time.Time),
		access:        newAccessState(),
	}
	state, _ := sharedRuntimeStates.LoadOrStore(manager, created)
	return state.(*sharedRuntimeState)
}
