package broadcast

import (
	"context"
	"fmt"
	"sync"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// LocalBroadcaster defines in-process broadcast behavior for tests and single-instance flows.
type LocalBroadcaster struct {
	// mutex guards channel subscription maps.
	mutex sync.RWMutex
	// subscribers stores subscribers by channel name.
	subscribers map[string]map[chan []byte]struct{}
}

// NewLocalBroadcaster creates one local broadcaster instance.
func NewLocalBroadcaster() *LocalBroadcaster {
	return &LocalBroadcaster{subscribers: map[string]map[chan []byte]struct{}{}}
}

// Publish sends one payload to all active subscribers on one channel.
func (broadcaster *LocalBroadcaster) Publish(ctx context.Context, channel string, payload []byte) error {
	if channel == "" {
		return fmt.Errorf("channel is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	broadcaster.mutex.RLock()
	rawListeners := broadcaster.subscribers[channel]
	listeners := make([]chan []byte, 0, len(rawListeners))
	for listener := range rawListeners {
		listeners = append(listeners, listener)
	}
	broadcaster.mutex.RUnlock()
	for _, listener := range listeners {
		copyPayload := append([]byte(nil), payload...)
		select {
		case listener <- copyPayload:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Subscribe registers one subscriber channel and returns a disposer.
func (broadcaster *LocalBroadcaster) Subscribe(ctx context.Context, channel string) (<-chan []byte, coreconnection.Disposable, error) {
	if channel == "" {
		return nil, nil, fmt.Errorf("channel is required")
	}
	stream := make(chan []byte, 8)
	broadcaster.mutex.Lock()
	if _, found := broadcaster.subscribers[channel]; !found {
		broadcaster.subscribers[channel] = map[chan []byte]struct{}{}
	}
	broadcaster.subscribers[channel][stream] = struct{}{}
	broadcaster.mutex.Unlock()
	dispose := func() error {
		broadcaster.mutex.Lock()
		listeners := broadcaster.subscribers[channel]
		if _, found := listeners[stream]; found {
			delete(listeners, stream)
			close(stream)
		}
		if len(listeners) == 0 {
			delete(broadcaster.subscribers, channel)
		}
		broadcaster.mutex.Unlock()
		return nil
	}
	go func() {
		<-ctx.Done()
		_ = dispose()
	}()
	return stream, coreconnection.DisposeFunc(dispose), nil
}
