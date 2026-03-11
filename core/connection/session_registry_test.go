package connection

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
)

// TestRedisSessionRegistryRegisterFindRemove verifies registry lifecycle behavior.
func TestRedisSessionRegistryRegisterFindRemove(t *testing.T) {
	registry, closeRegistry := createTestRegistry(t)
	defer closeRegistry()
	session := Session{
		ConnID: "conn-1", UserID: 15, State: StateConnected, CreatedAt: time.Unix(100, 0),
	}
	if err := registry.Register(session); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	byConnID, ok := registry.FindByConnID("conn-1")
	if !ok || byConnID.UserID != 15 {
		t.Fatalf("expected session by conn id, got %+v", byConnID)
	}
	byUserID, ok := registry.FindByUserID(15)
	if !ok || byUserID.ConnID != "conn-1" {
		t.Fatalf("expected session by user id, got %+v", byUserID)
	}
	registry.Remove("conn-1")
	if _, ok := registry.FindByConnID("conn-1"); ok {
		t.Fatalf("expected removed conn id lookup to fail")
	}
	if _, ok := registry.FindByUserID(15); ok {
		t.Fatalf("expected removed user id lookup to fail")
	}
}

// TestRedisSessionRegistryRejectsMissingConnID verifies validation behavior.
func TestRedisSessionRegistryRejectsMissingConnID(t *testing.T) {
	registry, closeRegistry := createTestRegistry(t)
	defer closeRegistry()
	if err := registry.Register(Session{}); err == nil {
		t.Fatalf("expected register failure for empty conn id")
	}
}

// TestRedisSessionRegistryOverwritesDuplicateUser verifies user index reassignment behavior.
func TestRedisSessionRegistryOverwritesDuplicateUser(t *testing.T) {
	registry, closeRegistry := createTestRegistry(t)
	defer closeRegistry()
	first := Session{ConnID: "conn-1", UserID: 8, State: StateConnected, CreatedAt: time.Unix(200, 0)}
	second := Session{ConnID: "conn-2", UserID: 8, State: StateAuthenticated, CreatedAt: time.Unix(201, 0)}
	if err := registry.Register(first); err != nil {
		t.Fatalf("expected first register success, got %v", err)
	}
	if err := registry.Register(second); err != nil {
		t.Fatalf("expected second register success, got %v", err)
	}
	byUserID, ok := registry.FindByUserID(8)
	if !ok || byUserID.ConnID != "conn-2" {
		t.Fatalf("expected reassigned user lookup, got %+v", byUserID)
	}
	if _, ok := registry.FindByConnID("conn-1"); ok {
		t.Fatalf("expected previous connection to be deleted")
	}
}

// TestNewRedisSessionRegistryRejectsNilClient verifies constructor validation.
func TestNewRedisSessionRegistryRejectsNilClient(t *testing.T) {
	if _, err := NewRedisSessionRegistry(nil); err == nil {
		t.Fatalf("expected constructor failure for nil client")
	}
}

// createTestRegistry builds a redis-backed session registry with isolated test resources.
func createTestRegistry(t *testing.T) (*RedisSessionRegistry, func()) {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	registry, err := NewRedisSessionRegistry(client)
	if err != nil {
		t.Fatalf("expected registry creation success, got %v", err)
	}
	cleanup := func() {
		_ = client.Close()
		server.Close()
	}
	return registry, cleanup
}
