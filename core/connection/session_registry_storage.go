package connection

import (
	"context"
	"encoding/json"
	"strconv"

	redislib "github.com/redis/go-redis/v9"
)

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
