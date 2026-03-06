package session

import (
	"errors"
	"reflect"
	"testing"
)

// TestManagerLifecycle validates register, send, ids, and remove behavior.
func TestManagerLifecycle(t *testing.T) {
	manager := NewManager()
	connection := &stubConnection{}
	if err := manager.Register("s1", connection); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := manager.Send("s1", []byte("hello")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := manager.Count(); got != 1 {
		t.Fatalf("expected count 1, got %d", got)
	}
	if got := manager.IDs(); !reflect.DeepEqual(got, []string{"s1"}) {
		t.Fatalf("unexpected ids: %#v", got)
	}
	if err := manager.Remove("s1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !connection.closed {
		t.Fatalf("expected connection close")
	}
	if len(connection.writes) != 1 || string(connection.writes[0]) != "hello" {
		t.Fatalf("unexpected writes: %#v", connection.writes)
	}
}

// TestManagerErrors validates manager error behavior.
func TestManagerErrors(t *testing.T) {
	manager := NewManager()
	if err := manager.Register("", &stubConnection{}); !errors.Is(err, ErrEmptySessionID) {
		t.Fatalf("expected empty session id error, got %v", err)
	}
	if err := manager.Register("s1", nil); !errors.Is(err, ErrNilConnection) {
		t.Fatalf("expected nil connection error, got %v", err)
	}
	_ = manager.Register("s1", &stubConnection{})
	if err := manager.Register("s1", &stubConnection{}); !errors.Is(err, ErrSessionExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	if err := manager.Send("missing", []byte("x")); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected missing session send error, got %v", err)
	}
	if err := manager.Remove("missing"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected missing session remove error, got %v", err)
	}
}

// stubConnection captures writes in-memory for tests.
type stubConnection struct {
	// writes stores all binary payload writes.
	writes [][]byte
	// closed tracks close calls.
	closed bool
}

// WriteBinary stores one payload in the write log.
func (s *stubConnection) WriteBinary(payload []byte) error {
	value := append([]byte(nil), payload...)
	s.writes = append(s.writes, value)
	return nil
}

// Close marks the connection as closed.
func (s *stubConnection) Close() error {
	s.closed = true
	return nil
}
