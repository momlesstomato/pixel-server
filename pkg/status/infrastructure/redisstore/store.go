package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
	redislib "github.com/redis/go-redis/v9"
)

// Store defines Redis hotel status persistence behavior.
type Store struct {
	// client stores Redis client connectivity.
	client *redislib.Client
	// key stores Redis key name for hotel status payload.
	key string
}

// NewStore creates a Redis-backed hotel status store.
func NewStore(client *redislib.Client, key string) (*Store, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	resolved := strings.TrimSpace(key)
	if resolved == "" {
		resolved = "hotel:status"
	}
	return &Store{client: client, key: resolved}, nil
}

// Load retrieves persisted hotel status and reports whether a record exists.
func (store *Store) Load(ctx context.Context) (statusdomain.HotelStatus, bool, error) {
	payload, err := store.client.Get(ctx, store.key).Bytes()
	if err == redislib.Nil {
		return statusdomain.HotelStatus{}, false, nil
	}
	if err != nil {
		return statusdomain.HotelStatus{}, false, err
	}
	var status statusdomain.HotelStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		return statusdomain.HotelStatus{}, false, err
	}
	return status, true, nil
}

// Save persists one hotel status snapshot.
func (store *Store) Save(ctx context.Context, status statusdomain.HotelStatus) error {
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return store.client.Set(ctx, store.key, payload, 0).Err()
}

// CompareAndSwap updates stored status only when current value equals expected value.
func (store *Store) CompareAndSwap(ctx context.Context, expected statusdomain.HotelStatus, next statusdomain.HotelStatus) (bool, error) {
	swapped := false
	err := store.client.Watch(ctx, func(transaction *redislib.Tx) error {
		current, found, loadErr := store.loadWithClient(ctx, transaction)
		if loadErr != nil {
			return loadErr
		}
		if !found {
			current = statusdomain.HotelStatus{}
		}
		if !reflect.DeepEqual(current, expected) {
			return redislib.TxFailedErr
		}
		payload, marshalErr := json.Marshal(next)
		if marshalErr != nil {
			return marshalErr
		}
		_, pipelineErr := transaction.TxPipelined(ctx, func(pipe redislib.Pipeliner) error {
			pipe.Set(ctx, store.key, payload, 0)
			return nil
		})
		if pipelineErr == nil {
			swapped = true
		}
		return pipelineErr
	}, store.key)
	if err == redislib.TxFailedErr {
		return false, nil
	}
	return swapped, err
}

// loadWithClient retrieves persisted hotel status using one specific command target.
func (store *Store) loadWithClient(ctx context.Context, client redislib.Cmdable) (statusdomain.HotelStatus, bool, error) {
	payload, err := client.Get(ctx, store.key).Bytes()
	if err == redislib.Nil {
		return statusdomain.HotelStatus{}, false, nil
	}
	if err != nil {
		return statusdomain.HotelStatus{}, false, err
	}
	var status statusdomain.HotelStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		return statusdomain.HotelStatus{}, false, err
	}
	return status, true, nil
}
