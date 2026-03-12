package authentication

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
)

// TestRedisStorePersistsAndConsumesTickets verifies store and single-use validation behavior.
func TestRedisStorePersistsAndConsumesTickets(t *testing.T) {
	server := startMiniRedis(t)
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	store, err := NewRedisStore(client, "sso")
	if err != nil {
		t.Fatalf("expected store creation success, got %v", err)
	}
	if err := store.Store(context.Background(), "ticket-1", 15, time.Minute); err != nil {
		t.Fatalf("expected store success, got %v", err)
	}
	userID, err := store.Validate(context.Background(), "ticket-1")
	if err != nil || userID != 15 {
		t.Fatalf("expected validate success with user 15, got user=%d err=%v", userID, err)
	}
	if _, err := store.Validate(context.Background(), "ticket-1"); err == nil {
		t.Fatalf("expected consumed ticket to fail validation")
	}
}

// TestRedisStoreRejectsNilClient verifies constructor precondition checks.
func TestRedisStoreRejectsNilClient(t *testing.T) {
	if _, err := NewRedisStore(nil, "sso"); err == nil {
		t.Fatalf("expected store creation failure for nil client")
	}
}

// TestRedisStoreRejectsInvalidUserIDPayload verifies validation parse checks.
func TestRedisStoreRejectsInvalidUserIDPayload(t *testing.T) {
	server := startMiniRedis(t)
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	store, err := NewRedisStore(client, "sso")
	if err != nil {
		t.Fatalf("expected store creation success, got %v", err)
	}
	if setErr := client.Set(context.Background(), "sso:broken", "abc", time.Minute).Err(); setErr != nil {
		t.Fatalf("expected test setup success, got %v", setErr)
	}
	if _, err := store.Validate(context.Background(), "broken"); err == nil {
		t.Fatalf("expected invalid user id payload failure")
	}
}

// startMiniRedis creates an isolated miniredis instance for tests.
func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	t.Cleanup(server.Close)
	return server
}
