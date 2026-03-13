package tests

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	redislib "github.com/redis/go-redis/v9"
)

// TestRedisSessionRegistryRegisterFindRemove verifies registry lifecycle behavior.
func TestRedisSessionRegistryRegisterFindRemove(t *testing.T) {
	registry, closeRegistry := createTestRegistry(t)
	defer closeRegistry()
	session := coreconnection.Session{ConnID: "conn-1", UserID: 15, State: coreconnection.StateConnected, CreatedAt: time.Unix(100, 0)}
	if err := registry.Register(session); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	byConnID, ok := registry.FindByConnID("conn-1")
	if !ok || byConnID.UserID != 15 {
		t.Fatalf("expected session by conn id, got %+v", byConnID)
	}
	if byConnID.InstanceID == "" {
		t.Fatalf("expected generated instance id for registered session")
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
	if err := registry.Register(coreconnection.Session{}); err == nil {
		t.Fatalf("expected register failure for empty conn id")
	}
	if err := registry.Touch(""); err == nil {
		t.Fatalf("expected touch failure for empty conn id")
	}
}

// TestRedisSessionRegistryOverwritesDuplicateUser verifies user index reassignment behavior.
func TestRedisSessionRegistryOverwritesDuplicateUser(t *testing.T) {
	registry, closeRegistry := createTestRegistry(t)
	defer closeRegistry()
	first := coreconnection.Session{ConnID: "conn-1", UserID: 8, State: coreconnection.StateConnected, CreatedAt: time.Unix(200, 0)}
	second := coreconnection.Session{ConnID: "conn-2", UserID: 8, State: coreconnection.StateAuthenticated, CreatedAt: time.Unix(201, 0)}
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
	if _, err := coreconnection.NewRedisSessionRegistry(nil); err == nil {
		t.Fatalf("expected constructor failure for nil client")
	}
}

// TestRedisSessionRegistrySessionTTLExpiry verifies TTL expiration behavior.
func TestRedisSessionRegistrySessionTTLExpiry(t *testing.T) {
	registry, server, closeRegistry := createConfiguredTestRegistry(t, coreconnection.RedisSessionRegistryOptions{
		TTL:             2 * time.Second,
		RefreshInterval: time.Second,
		InstanceID:      "instance-test",
	})
	defer closeRegistry()
	session := coreconnection.Session{ConnID: "conn-ttl", UserID: 33, State: coreconnection.StateAuthenticated, CreatedAt: time.Unix(300, 0)}
	if err := registry.Register(session); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	server.FastForward(3 * time.Second)
	if _, ok := registry.FindByConnID("conn-ttl"); ok {
		t.Fatalf("expected conn session to expire")
	}
	if _, ok := registry.FindByUserID(33); ok {
		t.Fatalf("expected user index to expire")
	}
}

// TestRedisSessionRegistryTouchRefreshesLease verifies touch lease refresh behavior.
func TestRedisSessionRegistryTouchRefreshesLease(t *testing.T) {
	registry, server, closeRegistry := createConfiguredTestRegistry(t, coreconnection.RedisSessionRegistryOptions{
		TTL:             2 * time.Second,
		RefreshInterval: time.Second,
		InstanceID:      "instance-touch",
	})
	defer closeRegistry()
	session := coreconnection.Session{ConnID: "conn-touch", UserID: 44, State: coreconnection.StateAuthenticated, CreatedAt: time.Unix(301, 0)}
	if err := registry.Register(session); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	server.FastForward(1500 * time.Millisecond)
	if err := registry.Touch("conn-touch"); err != nil {
		t.Fatalf("expected touch success, got %v", err)
	}
	server.FastForward(1500 * time.Millisecond)
	if _, ok := registry.FindByConnID("conn-touch"); !ok {
		t.Fatalf("expected conn session to remain after touch refresh")
	}
	server.FastForward(2500 * time.Millisecond)
	if _, ok := registry.FindByConnID("conn-touch"); ok {
		t.Fatalf("expected conn session to expire after refreshed ttl")
	}
}

// createTestRegistry builds a redis-backed session registry with isolated test resources.
func createTestRegistry(t *testing.T) (*coreconnection.RedisSessionRegistry, func()) {
	t.Helper()
	registry, _, cleanup := createConfiguredTestRegistry(t, coreconnection.RedisSessionRegistryOptions{})
	return registry, cleanup
}

// createConfiguredTestRegistry builds a redis-backed session registry with explicit options.
func createConfiguredTestRegistry(t *testing.T, options coreconnection.RedisSessionRegistryOptions) (*coreconnection.RedisSessionRegistry, *miniredis.Miniredis, func()) {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	registry, err := coreconnection.NewRedisSessionRegistryWithOptions(client, options)
	if err != nil {
		t.Fatalf("expected registry creation success, got %v", err)
	}
	cleanup := func() {
		_ = client.Close()
		server.Close()
	}
	return registry, server, cleanup
}
