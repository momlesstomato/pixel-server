package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

const defaultSessionRegistryPrefix = "session"
const defaultSessionRegistryTTL = 120 * time.Second
const defaultSessionRefreshInterval = 60 * time.Second

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
	// InstanceID stores the server instance identifier that owns this session.
	InstanceID string
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
	// Touch refreshes session key TTL for an active connection.
	Touch(string) error
	// Remove deletes session indexes by connection ID.
	Remove(string)
	// ListAll returns all sessions currently stored in the registry.
	ListAll() ([]Session, error)
}

// RedisSessionRegistryOptions defines Redis session registry configuration.
type RedisSessionRegistryOptions struct {
	// Prefix stores Redis key namespace prefix.
	Prefix string
	// TTL stores Redis key expiration duration.
	TTL time.Duration
	// RefreshInterval stores recommended key lease refresh interval.
	RefreshInterval time.Duration
	// InstanceID stores default session instance identifier.
	InstanceID string
}

// RedisSessionRegistry implements SessionRegistry backed by Redis keys.
type RedisSessionRegistry struct {
	// client stores Redis connectivity used for session state operations.
	client *redislib.Client
	// prefix stores Redis key namespace prefix for session records.
	prefix string
	// ttl stores Redis lease duration for session records.
	ttl time.Duration
	// refreshInterval stores recommended lease refresh interval.
	refreshInterval time.Duration
	// instanceID stores default instance identifier for registered sessions.
	instanceID string
}

// NewRedisSessionRegistry creates a Redis-backed session registry.
func NewRedisSessionRegistry(client *redislib.Client) (*RedisSessionRegistry, error) {
	return NewRedisSessionRegistryWithOptions(client, RedisSessionRegistryOptions{})
}

// NewRedisSessionRegistryWithOptions creates a Redis-backed session registry with options.
func NewRedisSessionRegistryWithOptions(client *redislib.Client, options RedisSessionRegistryOptions) (*RedisSessionRegistry, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	resolved, err := resolveRedisSessionRegistryOptions(options)
	if err != nil {
		return nil, err
	}
	return &RedisSessionRegistry{
		client: client, prefix: resolved.Prefix, ttl: resolved.TTL,
		refreshInterval: resolved.RefreshInterval, instanceID: resolved.InstanceID,
	}, nil
}

// Register stores or updates a session.
func (registry *RedisSessionRegistry) Register(session Session) error {
	if session.ConnID == "" {
		return fmt.Errorf("session connection id is required")
	}
	effectiveSession := session
	if effectiveSession.InstanceID == "" {
		effectiveSession.InstanceID = registry.instanceID
	}
	ctx := context.Background()
	existing, found, err := registry.fetchByConnID(ctx, effectiveSession.ConnID)
	if err != nil {
		return err
	}
	previousConnID := ""
	if effectiveSession.UserID > 0 {
		previousConnID, err = registry.client.Get(ctx, registry.userKey(effectiveSession.UserID)).Result()
		if err != nil && err != redislib.Nil {
			return err
		}
	}
	payload, err := json.Marshal(effectiveSession)
	if err != nil {
		return err
	}
	pipeline := registry.client.TxPipeline()
	if found && existing.UserID > 0 && existing.UserID != effectiveSession.UserID {
		pipeline.Del(ctx, registry.userKey(existing.UserID))
	}
	if effectiveSession.UserID > 0 {
		if previousConnID != "" && previousConnID != effectiveSession.ConnID {
			pipeline.Del(ctx, registry.connKey(previousConnID))
		}
		pipeline.Set(ctx, registry.userKey(effectiveSession.UserID), effectiveSession.ConnID, registry.ttl)
	}
	pipeline.Set(ctx, registry.connKey(effectiveSession.ConnID), payload, registry.ttl)
	_, err = pipeline.Exec(ctx)
	return err
}
