package redisstore

import (
	"context"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
	redislib "github.com/redis/go-redis/v9"
)

// TestNewStoreRejectsNilClient verifies constructor validation behavior.
func TestNewStoreRejectsNilClient(t *testing.T) {
	if _, err := NewStore(nil, ""); err == nil {
		t.Fatalf("expected constructor validation failure")
	}
}

// TestStoreLoadSaveAndCompareAndSwap verifies status persistence and CAS behavior.
func TestStoreLoadSaveAndCompareAndSwap(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	store, _ := NewStore(client, "hotel:status:test")
	status, found, err := store.Load(context.Background())
	if err != nil || found {
		t.Fatalf("expected missing status, got %#v found=%v err=%v", status, found, err)
	}
	open := statusdomain.HotelStatus{State: statusdomain.StateOpen}
	if err := store.Save(context.Background(), open); err != nil {
		t.Fatalf("expected save success, got %v", err)
	}
	loaded, found, err := store.Load(context.Background())
	if err != nil || !found || loaded.State != statusdomain.StateOpen {
		t.Fatalf("expected open status after save, got %#v found=%v err=%v", loaded, found, err)
	}
	closing := statusdomain.HotelStatus{State: statusdomain.StateClosing}
	swapped, err := store.CompareAndSwap(context.Background(), open, closing)
	if err != nil || !swapped {
		t.Fatalf("expected CAS success, got swapped=%v err=%v", swapped, err)
	}
	swapped, err = store.CompareAndSwap(context.Background(), open, open)
	if err != nil || swapped {
		t.Fatalf("expected CAS miss for stale snapshot, got swapped=%v err=%v", swapped, err)
	}
}
