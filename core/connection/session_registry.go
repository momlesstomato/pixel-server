package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

const defaultSessionRegistryPrefix = "session"

// SessionState identifies connection authentication lifecycle state.
type SessionState int

const (
	// StateConnected indicates transport is connected and unauthenticated.
	StateConnected SessionState = iota
	// StateAuthenticated indicates SSO authentication has succeeded.
	StateAuthenticated
	// StateDisconnecting indicates graceful closure is in progress.
	StateDisconnecting
)

// Session represents a connected client lifecycle in handshake realm.
type Session struct {
	// ConnID stores the transport identifier.
	ConnID string
	// UserID stores authenticated user ID; zero when unauthenticated.
	UserID int
	// MachineID stores the machine fingerprint identifier.
	MachineID string
	// State stores current session lifecycle state.
	State SessionState
	// CreatedAt stores the session creation timestamp.
	CreatedAt time.Time
}

// SessionRegistry defines session lookup and lifecycle operations.
type SessionRegistry interface {
	// Register stores or updates a session.
	Register(Session) error
	// FindByUserID retrieves an active session by user ID.
	FindByUserID(int) (Session, bool)
	// FindByConnID retrieves an active session by connection ID.
	FindByConnID(string) (Session, bool)
	// Remove deletes session indexes by connection ID.
	Remove(string)
}

// RedisSessionRegistry implements SessionRegistry backed by Redis keys.
type RedisSessionRegistry struct {
	// client stores Redis connectivity used for session state operations.
	client *redislib.Client
	// prefix stores Redis key namespace prefix for session records.
	prefix string
}

// NewRedisSessionRegistry creates a Redis-backed session registry.
func NewRedisSessionRegistry(client *redislib.Client) (*RedisSessionRegistry, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	return &RedisSessionRegistry{client: client, prefix: defaultSessionRegistryPrefix}, nil
}

// Register stores or updates a session.
func (registry *RedisSessionRegistry) Register(session Session) error {
	if session.ConnID == "" {
		return fmt.Errorf("session connection id is required")
	}
	ctx := context.Background()
	existing, found, err := registry.fetchByConnID(ctx, session.ConnID)
	if err != nil {
		return err
	}
	previousConnID := ""
	if session.UserID > 0 {
		previousConnID, err = registry.client.Get(ctx, registry.userKey(session.UserID)).Result()
		if err != nil && err != redislib.Nil {
			return err
		}
	}
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}
	pipeline := registry.client.TxPipeline()
	if found && existing.UserID > 0 && existing.UserID != session.UserID {
		pipeline.Del(ctx, registry.userKey(existing.UserID))
	}
	if session.UserID > 0 {
		if previousConnID != "" && previousConnID != session.ConnID {
			pipeline.Del(ctx, registry.connKey(previousConnID))
		}
		pipeline.Set(ctx, registry.userKey(session.UserID), session.ConnID, 0)
	}
	pipeline.Set(ctx, registry.connKey(session.ConnID), payload, 0)
	_, err = pipeline.Exec(ctx)
	return err
}

// FindByUserID retrieves an active session by user ID.
func (registry *RedisSessionRegistry) FindByUserID(userID int) (Session, bool) {
	connID, err := registry.client.Get(context.Background(), registry.userKey(userID)).Result()
	if err != nil {
		return Session{}, false
	}
	session, found, fetchErr := registry.fetchByConnID(context.Background(), connID)
	if fetchErr != nil || !found {
		return Session{}, false
	}
	return session, true
}

// FindByConnID retrieves an active session by connection ID.
func (registry *RedisSessionRegistry) FindByConnID(connID string) (Session, bool) {
	session, found, err := registry.fetchByConnID(context.Background(), connID)
	if err != nil {
		return Session{}, false
	}
	return session, found
}

// Remove deletes session indexes by connection ID.
func (registry *RedisSessionRegistry) Remove(connID string) {
	ctx := context.Background()
	session, found, err := registry.fetchByConnID(ctx, connID)
	if err != nil {
		return
	}
	pipeline := registry.client.TxPipeline()
	pipeline.Del(ctx, registry.connKey(connID))
	if found && session.UserID > 0 {
		pipeline.Del(ctx, registry.userKey(session.UserID))
	}
	_, _ = pipeline.Exec(ctx)
}
