package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"pixelsv/pkg/storage/interfaces"
)

// TestKVStoreRoundTrip validates set/get/delete behavior.
func TestKVStoreRoundTrip(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	cfg := Config{URL: "redis://" + server.Addr(), KeyPrefix: "pixelsv", SessionTTLSeconds: 60}
	store := NewKVStore(client, cfg)
	key := "k1"
	value := []byte(`{"ok":true}`)
	if err := store.Set(context.Background(), key, value, time.Minute); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	loaded, err := store.Get(context.Background(), key)
	if err != nil || string(loaded) != string(value) {
		t.Fatalf("unexpected loaded value: %s %v", string(loaded), err)
	}
	if err := store.Delete(context.Background(), key); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = store.Get(context.Background(), key)
	if !errors.Is(err, interfaces.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
