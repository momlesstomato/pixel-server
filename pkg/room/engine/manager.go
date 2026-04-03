package engine

import (
	"context"
	"sync"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"go.uber.org/zap"
)

// Manager maintains a registry of active room instances.
type Manager struct {
	// rooms stores active room instances by room ID.
	rooms map[int]*Instance
	// mu protects the rooms map.
	mu sync.RWMutex
	// ctx stores the parent context for all room goroutines.
	ctx context.Context
	// logger stores the structured logger.
	logger *zap.Logger
	// broadcaster stores the entity broadcast callback.
	broadcaster EntityBroadcaster
}

// NewManager creates a room instance manager.
func NewManager(ctx context.Context, logger *zap.Logger, broadcaster EntityBroadcaster) *Manager {
	return &Manager{
		rooms:       make(map[int]*Instance),
		ctx:         ctx,
		logger:      logger.Named("room_manager"),
		broadcaster: broadcaster,
	}
}

// Load creates and starts a room instance if not already loaded.
func (m *Manager) Load(roomID int, layout domain.Layout) *Instance {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.rooms[roomID]; ok {
		if inst.State() != StateStopped {
			return inst
		}
	}
	inst := NewInstance(roomID, layout, m.logger, m.broadcaster)
	inst.Start(m.ctx)
	m.rooms[roomID] = inst
	m.logger.Info("room loaded", zap.Int("room_id", roomID))
	return inst
}

// Unload stops and removes a room instance.
func (m *Manager) Unload(roomID int) {
	m.mu.Lock()
	inst, ok := m.rooms[roomID]
	if !ok {
		m.mu.Unlock()
		return
	}
	delete(m.rooms, roomID)
	m.mu.Unlock()
	inst.Stop()
	<-inst.Done()
	m.logger.Info("room unloaded", zap.Int("room_id", roomID))
}

// Get returns a loaded room instance by ID.
func (m *Manager) Get(roomID int) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inst, ok := m.rooms[roomID]
	if !ok || inst.State() == StateStopped {
		return nil, false
	}
	return inst, true
}

// Count returns the number of loaded rooms.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// StopAll shuts down every running room instance.
func (m *Manager) StopAll() {
	m.mu.Lock()
	rooms := make([]*Instance, 0, len(m.rooms))
	for _, inst := range m.rooms {
		rooms = append(rooms, inst)
	}
	m.rooms = make(map[int]*Instance)
	m.mu.Unlock()
	for _, inst := range rooms {
		inst.Stop()
		<-inst.Done()
	}
	m.logger.Info("all rooms stopped", zap.Int("count", len(rooms)))
}

// SetBroadcaster updates the entity broadcaster used by new room instances.
func (m *Manager) SetBroadcaster(bc EntityBroadcaster) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcaster = bc
}

// Cleanup removes stopped instances from the registry.
func (m *Manager) Cleanup() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	for id, inst := range m.rooms {
		if inst.State() == StateStopped {
			delete(m.rooms, id)
			removed++
		}
	}
	return removed
}
