package connection

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	redislib "github.com/redis/go-redis/v9"
)

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

// Touch refreshes session key TTL for an active connection.
func (registry *RedisSessionRegistry) Touch(connID string) error {
	if strings.TrimSpace(connID) == "" {
		return fmt.Errorf("connection id is required")
	}
	ctx := context.Background()
	session, found, err := registry.fetchByConnID(ctx, connID)
	if err != nil || !found {
		return err
	}
	pipeline := registry.client.TxPipeline()
	pipeline.Expire(ctx, registry.connKey(connID), registry.ttl)
	if session.UserID > 0 {
		pipeline.Expire(ctx, registry.userKey(session.UserID), registry.ttl)
	}
	_, err = pipeline.Exec(ctx)
	return err
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

// ListAll returns all sessions stored in the registry using SCAN.
func (registry *RedisSessionRegistry) ListAll() ([]Session, error) {
	ctx, pattern := context.Background(), registry.prefix+":conn:*"
	var sessions []Session
	iter := registry.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		payload, err := registry.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var s Session
		if json.Unmarshal(payload, &s) == nil {
			sessions = append(sessions, s)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// fetchByConnID loads one session record by connection identifier.
func (registry *RedisSessionRegistry) fetchByConnID(ctx context.Context, connID string) (Session, bool, error) {
	payload, err := registry.client.Get(ctx, registry.connKey(connID)).Bytes()
	if err == redislib.Nil {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	var session Session
	if err = json.Unmarshal(payload, &session); err != nil {
		return Session{}, false, err
	}
	return session, true, nil
}

// connKey returns the namespaced Redis key for one connection session record.
func (registry *RedisSessionRegistry) connKey(connID string) string {
	return registry.prefix + ":conn:" + connID
}

// userKey returns the namespaced Redis key for one user-to-connection index.
func (registry *RedisSessionRegistry) userKey(userID int) string {
	return registry.prefix + ":user:" + strconv.Itoa(userID)
}

// resolveRedisSessionRegistryOptions resolves options with defaults.
func resolveRedisSessionRegistryOptions(options RedisSessionRegistryOptions) (RedisSessionRegistryOptions, error) {
	resolved := options
	if strings.TrimSpace(resolved.Prefix) == "" {
		resolved.Prefix = defaultSessionRegistryPrefix
	}
	if resolved.TTL <= 0 {
		resolved.TTL = defaultSessionRegistryTTL
	}
	if resolved.RefreshInterval <= 0 {
		resolved.RefreshInterval = defaultSessionRefreshInterval
	}
	if resolved.RefreshInterval >= resolved.TTL {
		resolved.RefreshInterval = resolved.TTL / 2
	}
	if strings.TrimSpace(resolved.InstanceID) == "" {
		instanceID, err := generateRegistryInstanceID()
		if err != nil {
			return RedisSessionRegistryOptions{}, err
		}
		resolved.InstanceID = instanceID
	}
	return resolved, nil
}

// generateRegistryInstanceID builds one default server instance identifier.
func generateRegistryInstanceID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "instance"
	}
	buffer := make([]byte, 6)
	if _, err = rand.Read(buffer); err != nil {
		return "", err
	}
	return hostname + ":" + hex.EncodeToString(buffer), nil
}
