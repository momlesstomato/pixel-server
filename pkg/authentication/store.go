package authentication

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

var errTicketNotFound = errors.New("sso ticket not found")

// Store defines ticket persistence and single-use validation behavior.
type Store interface {
	// Store persists one ticket mapped to one user ID for a bounded lifetime.
	Store(context.Context, string, int, time.Duration) error
	// Validate consumes one ticket and returns the associated user ID.
	Validate(context.Context, string) (int, error)
}

// RedisStore implements Store using Redis SET and GETDEL operations.
type RedisStore struct {
	// client stores Redis connectivity used by ticket operations.
	client *redislib.Client
	// prefix stores ticket key namespace prefix.
	prefix string
}

// NewRedisStore creates a Redis-backed ticket store.
func NewRedisStore(client *redislib.Client, prefix string) (*RedisStore, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		trimmed = "sso"
	}
	return &RedisStore{client: client, prefix: trimmed}, nil
}

// Store persists one ticket mapped to one user ID for a bounded lifetime.
func (store *RedisStore) Store(ctx context.Context, ticket string, userID int, ttl time.Duration) error {
	return store.client.Set(ctx, store.key(ticket), strconv.Itoa(userID), ttl).Err()
}

// Validate consumes one ticket and returns the associated user ID.
func (store *RedisStore) Validate(ctx context.Context, ticket string) (int, error) {
	value, err := store.client.GetDel(ctx, store.key(ticket)).Result()
	if err == redislib.Nil {
		return 0, errTicketNotFound
	}
	if err != nil {
		return 0, err
	}
	userID, parseErr := strconv.Atoi(value)
	if parseErr != nil {
		return 0, fmt.Errorf("parse user id for ticket: %w", parseErr)
	}
	return userID, nil
}

// key returns the namespaced Redis key for one ticket.
func (store *RedisStore) key(ticket string) string {
	return store.prefix + ":" + ticket
}
