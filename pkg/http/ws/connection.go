package ws

import "github.com/gofiber/contrib/websocket"

// FiberConnectionAdapter adapts fiber websocket connections into session ports.
type FiberConnectionAdapter struct {
	// conn is the active fiber websocket connection.
	conn *websocket.Conn
}

// NewFiberConnectionAdapter creates a new FiberConnectionAdapter.
func NewFiberConnectionAdapter(conn *websocket.Conn) *FiberConnectionAdapter {
	return &FiberConnectionAdapter{conn: conn}
}

// WriteBinary writes one payload as websocket binary message.
func (a *FiberConnectionAdapter) WriteBinary(payload []byte) error {
	return a.conn.WriteMessage(websocket.BinaryMessage, payload)
}

// Close closes the underlying websocket connection.
func (a *FiberConnectionAdapter) Close() error {
	return a.conn.Close()
}
