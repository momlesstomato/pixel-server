package interfaces

import (
	"context"
	"time"
)

// KeyValueStore defines generic byte-oriented key/value operations.
type KeyValueStore interface {
	// Set stores a value by key with a TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Get loads a value by key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Delete removes a key.
	Delete(ctx context.Context, key string) error
}
