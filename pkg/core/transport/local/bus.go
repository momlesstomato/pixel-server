package local

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"pixelsv/pkg/core/transport"
)

// Bus is an in-process publish/subscribe transport adapter.
type Bus struct {
	mu     sync.RWMutex
	closed bool
	nextID atomic.Uint64
	subs   map[string]map[uint64]transport.Handler
}

// New creates an in-process local transport bus.
func New() *Bus {
	return &Bus{subs: map[string]map[uint64]transport.Handler{}}
}

// Publish emits a message to matching topic subscriptions.
func (b *Bus) Publish(ctx context.Context, topic string, payload []byte) error {
	if err := transport.ValidateTopic(topic); err != nil {
		return err
	}
	handlers, err := b.handlersForTopic(topic)
	if err != nil {
		return err
	}
	message := transport.Message{Topic: topic, Payload: payload}
	var joined error
	for _, handler := range handlers {
		if err := handler(ctx, message); err != nil {
			joined = errors.Join(joined, err)
		}
	}
	return joined
}

// Subscribe registers a handler for a topic pattern.
func (b *Bus) Subscribe(ctx context.Context, topic string, handler transport.Handler) (transport.Subscription, error) {
	if err := transport.ValidateTopic(topic); err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	id := b.nextID.Add(1)
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, transport.ErrClosed
	}
	if _, ok := b.subs[topic]; !ok {
		b.subs[topic] = map[uint64]transport.Handler{}
	}
	b.subs[topic][id] = handler
	b.mu.Unlock()
	sub := newSubscription(func() error { return b.remove(topic, id) })
	if done := ctx.Done(); done != nil {
		go func() {
			<-done
			_ = sub.Unsubscribe()
		}()
	}
	return sub, nil
}

// Close releases subscriptions and prevents further publish/subscribe operations.
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	b.subs = map[string]map[uint64]transport.Handler{}
	return nil
}

// handlersForTopic resolves all handlers matching the provided topic.
func (b *Bus) handlersForTopic(topic string) ([]transport.Handler, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return nil, transport.ErrClosed
	}
	handlers := make([]transport.Handler, 0, 4)
	for pattern, subscriptions := range b.subs {
		if !matchTopic(pattern, topic) {
			continue
		}
		for _, handler := range subscriptions {
			handlers = append(handlers, handler)
		}
	}
	return handlers, nil
}

// remove removes one subscription handler from the local topic map.
func (b *Bus) remove(topic string, id uint64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	subscriptions, ok := b.subs[topic]
	if !ok {
		return nil
	}
	delete(subscriptions, id)
	if len(subscriptions) == 0 {
		delete(b.subs, topic)
	}
	return nil
}
