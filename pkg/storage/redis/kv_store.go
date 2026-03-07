package redis

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"pixelsv/pkg/storage/interfaces"
)

// KVStore implements interfaces.KeyValueStore over Redis.
type KVStore struct {
	client goredis.Cmdable
	prefix string
}

// NewKVStore creates a KVStore.
func NewKVStore(client goredis.Cmdable, cfg Config) *KVStore {
	return &KVStore{client: client, prefix: cfg.KeyPrefix}
}

// Set stores a value by key with TTL.
func (s *KVStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		log.Printf("level=debug component=redis op=set key=%q ttl=%s bytes=%d", NamespacedKey(s.prefix, key), ttl, len(value))
	}
	return s.client.Set(ctx, NamespacedKey(s.prefix, key), value, ttl).Err()
}

// Get loads a value by key.
func (s *KVStore) Get(ctx context.Context, key string) ([]byte, error) {
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		log.Printf("level=debug component=redis op=get key=%q", NamespacedKey(s.prefix, key))
	}
	body, err := s.client.Get(ctx, NamespacedKey(s.prefix, key)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, interfaces.ErrNotFound
	}
	return body, err
}

// Delete removes a key.
func (s *KVStore) Delete(ctx context.Context, key string) error {
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		log.Printf("level=debug component=redis op=del key=%q", NamespacedKey(s.prefix, key))
	}
	return s.client.Del(ctx, NamespacedKey(s.prefix, key)).Err()
}
