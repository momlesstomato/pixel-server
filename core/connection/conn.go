package connection

import (
	"context"
	"fmt"
	"sync"
)

// Connection defines the core transport abstraction for packet IO.
type Connection interface {
	Disposable
	// ID returns the stable connection identifier.
	ID() string
	// Read receives a raw frame payload from the transport.
	Read(context.Context) ([]byte, error)
	// Write sends a raw frame payload through the transport.
	Write(context.Context, []byte) error
}

// MemoryConnection provides an in-memory transport for tests and local wiring.
type MemoryConnection struct {
	// id stores the connection identifier.
	id string
	// inbound stores frames available for reads.
	inbound chan []byte
	// outbound stores frames written by the connection.
	outbound chan []byte
	// mutex guards close-state transitions.
	mutex sync.RWMutex
	// disposed tracks whether the connection has been disposed.
	disposed bool
}

// NewMemoryConnection creates a memory-backed connection transport.
func NewMemoryConnection(id string, queueSize int) *MemoryConnection {
	return &MemoryConnection{
		id: id, inbound: make(chan []byte, queueSize), outbound: make(chan []byte, queueSize),
	}
}

// ID returns the stable connection identifier.
func (connection *MemoryConnection) ID() string {
	return connection.id
}

// Read receives a raw payload from the inbound queue.
func (connection *MemoryConnection) Read(ctx context.Context) ([]byte, error) {
	payload, ok := <-connection.inbound
	if !ok {
		return nil, fmt.Errorf("connection disposed")
	}
	return payload, nil
}

// Write sends a raw payload into the outbound queue.
func (connection *MemoryConnection) Write(ctx context.Context, payload []byte) error {
	connection.mutex.RLock()
	disposed := connection.disposed
	connection.mutex.RUnlock()
	if disposed {
		return fmt.Errorf("connection disposed")
	}
	connection.outbound <- payload
	return nil
}

// Dispose closes transport queues and marks the connection as disposed.
func (connection *MemoryConnection) Dispose() error {
	connection.mutex.Lock()
	if connection.disposed {
		connection.mutex.Unlock()
		return nil
	}
	connection.disposed = true
	close(connection.inbound)
	close(connection.outbound)
	connection.mutex.Unlock()
	return nil
}

// PushInbound enqueues payload for subsequent read operations.
func (connection *MemoryConnection) PushInbound(payload []byte) error {
	connection.mutex.RLock()
	disposed := connection.disposed
	connection.mutex.RUnlock()
	if disposed {
		return fmt.Errorf("connection disposed")
	}
	connection.inbound <- payload
	return nil
}

// ReadOutbound dequeues payload written by the connection.
func (connection *MemoryConnection) ReadOutbound(ctx context.Context) ([]byte, error) {
	payload, ok := <-connection.outbound
	if !ok {
		return nil, fmt.Errorf("connection disposed")
	}
	return payload, nil
}
