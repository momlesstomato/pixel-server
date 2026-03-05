package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"pixelsv/pkg/storage/interfaces"
	"pixelsv/pkg/storage/postgres"
	"pixelsv/pkg/storage/redis"
)

// Test02RedisKVStoreRoundTrip validates redis key/value persistence.
func Test02RedisKVStoreRoundTrip(t *testing.T) {
	server := miniredis.RunT(t)
	cfg := redis.Config{
		URL:               "redis://" + server.Addr(),
		KeyPrefix:         "pixelsv-e2e",
		SessionTTLSeconds: 120,
	}
	service, err := redis.New(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer service.Close()
	store := redis.NewKVStore(service.Client(), cfg)
	type payload struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
	}
	value, err := json.Marshal(payload{UserID: 11, Name: "ian"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	key := "user:11"
	if err := store.Set(context.Background(), key, value, time.Minute); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	loaded, err := store.Get(context.Background(), key)
	if err != nil || string(loaded) != string(value) {
		t.Fatalf("unexpected value: %s %v", string(loaded), err)
	}
	if err := store.Delete(context.Background(), key); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = store.Get(context.Background(), key)
	if !errors.Is(err, interfaces.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// Test02PostgresPing validates postgres ping when env is provided.
func Test02PostgresPing(t *testing.T) {
	dsn := os.Getenv("POSTGRES_URL_E2E")
	if dsn == "" {
		t.Skip("set POSTGRES_URL_E2E to run postgres e2e test")
	}
	cfg := postgres.Config{URL: dsn, MinConns: 1, MaxConns: 2}
	service, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer service.Close()
	if err := service.Ping(context.Background()); err != nil {
		t.Fatalf("expected ping success, got %v", err)
	}
}
