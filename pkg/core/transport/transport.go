package transport

import (
	"context"
	"errors"
	"io"
)

// ErrClosed indicates the transport bus has been closed.
var ErrClosed = errors.New("transport is closed")

// ErrEmptyTopic indicates event topic is required.
var ErrEmptyTopic = errors.New("event topic is required")

// Message is the runtime event envelope delivered by transport adapters.
type Message struct {
	// Topic is the stable transport topic string.
	Topic string
	// Payload carries binary event data.
	Payload []byte
}

// Handler consumes one runtime transport message.
type Handler func(ctx context.Context, message Message) error

// EventPublisher publishes messages to runtime transport topics.
type EventPublisher interface {
	// Publish emits a message payload to a topic.
	Publish(ctx context.Context, topic string, payload []byte) error
}

// EventSubscriber subscribes handlers to runtime transport topics.
type EventSubscriber interface {
	// Subscribe registers a handler and returns a cancellable subscription.
	Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
}

// Subscription represents an active transport topic subscription.
type Subscription interface {
	// Unsubscribe removes the handler from its topic.
	Unsubscribe() error
}

// Bus combines publish/subscribe behavior with lifecycle control.
type Bus interface {
	EventPublisher
	EventSubscriber
	io.Closer
}

// ValidateTopic checks whether a topic can be used by transport adapters.
func ValidateTopic(topic string) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	return nil
}
