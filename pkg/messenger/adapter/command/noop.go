package command

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/connection"
)

// noopBroadcaster is a no-op broadcaster used for CLI-only service instantiation.
type noopBroadcaster struct{}

// Publish discards the payload silently.
func (n *noopBroadcaster) Publish(_ context.Context, _ string, _ []byte) error {
	return nil
}

// Subscribe returns a closed channel immediately.
func (n *noopBroadcaster) Subscribe(_ context.Context, _ string) (<-chan []byte, connection.Disposable, error) {
	ch := make(chan []byte)
	close(ch)
	return ch, connection.DisposeFunc(func() error { return nil }), nil
}

// noopSessionRegistry is a no-op session registry for CLI use.
type noopSessionRegistry struct{}

// Register discards the session record.
func (n *noopSessionRegistry) Register(_ connection.Session) error { return nil }

// FindByUserID always returns not found.
func (n *noopSessionRegistry) FindByUserID(_ int) (connection.Session, bool) {
	return connection.Session{}, false
}

// FindByConnID always returns not found.
func (n *noopSessionRegistry) FindByConnID(_ string) (connection.Session, bool) {
	return connection.Session{}, false
}

// Touch performs no operation.
func (n *noopSessionRegistry) Touch(_ string) error { return nil }

// Remove performs no operation.
func (n *noopSessionRegistry) Remove(_ string) {}

// ListAll always returns an empty slice.
func (n *noopSessionRegistry) ListAll() ([]connection.Session, error) { return nil, nil }
