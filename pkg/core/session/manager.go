package session

import (
	"sort"
	"sync"
	"sync/atomic"
)

// Connection writes outbound binary payloads and supports close semantics.
type Connection interface {
	// WriteBinary sends one binary payload through the connection.
	WriteBinary(payload []byte) error
	// Close releases connection resources.
	Close() error
}

// Writer sends one payload to a session id.
type Writer interface {
	// Send writes payload bytes to one active session.
	Send(sessionID string, payload []byte) error
}

// Manager stores runtime session connections and provides thread-safe writes.
type Manager struct {
	// count tracks active sessions.
	count atomic.Int64
	// entries stores session id to session entry mappings.
	entries sync.Map
}

// NewManager creates a new empty session manager.
func NewManager() *Manager {
	return &Manager{}
}

// Register stores a connection under one session id.
func (m *Manager) Register(sessionID string, connection Connection) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	if connection == nil {
		return ErrNilConnection
	}
	entry := &sessionEntry{connection: connection}
	if _, loaded := m.entries.LoadOrStore(sessionID, entry); loaded {
		return ErrSessionExists
	}
	m.count.Add(1)
	return nil
}

// Remove closes and removes one registered session.
func (m *Manager) Remove(sessionID string) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	value, ok := m.entries.LoadAndDelete(sessionID)
	if !ok {
		return ErrSessionNotFound
	}
	m.count.Add(-1)
	return value.(*sessionEntry).connection.Close()
}

// Send writes one payload to an active session.
func (m *Manager) Send(sessionID string, payload []byte) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	value, ok := m.entries.Load(sessionID)
	if !ok {
		return ErrSessionNotFound
	}
	entry := value.(*sessionEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	return entry.connection.WriteBinary(payload)
}

// IDs returns currently registered session ids in sorted order.
func (m *Manager) IDs() []string {
	ids := make([]string, 0, 16)
	m.entries.Range(func(key any, _ any) bool {
		ids = append(ids, key.(string))
		return true
	})
	sort.Strings(ids)
	return ids
}

// Count returns current active session count.
func (m *Manager) Count() int {
	return int(m.count.Load())
}

// sessionEntry stores one session connection and write lock.
type sessionEntry struct {
	// connection is the outbound binary sink.
	connection Connection
	// mu serializes writes on one connection.
	mu sync.Mutex
}
