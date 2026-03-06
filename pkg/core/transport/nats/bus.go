package natsbus

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
	"pixelsv/pkg/core/transport"
)

// Bus is a NATS-backed runtime transport adapter.
type Bus struct {
	conn *nats.Conn
}

// New creates a NATS transport bus from a NATS server URL.
func New(url string, options ...nats.Option) (*Bus, error) {
	if url == "" {
		return nil, errors.New("nats url is required")
	}
	conn, err := nats.Connect(url, options...)
	if err != nil {
		return nil, err
	}
	return &Bus{conn: conn}, nil
}

// NewFromConn creates a NATS transport bus from an existing connection.
func NewFromConn(conn *nats.Conn) (*Bus, error) {
	if conn == nil {
		return nil, errors.New("nats connection is required")
	}
	return &Bus{conn: conn}, nil
}

// Publish emits a payload on the provided NATS subject topic.
func (b *Bus) Publish(ctx context.Context, topic string, payload []byte) error {
	if err := transport.ValidateTopic(topic); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.conn.Publish(topic, payload)
}

// Subscribe registers a transport handler on a NATS subject pattern.
func (b *Bus) Subscribe(ctx context.Context, topic string, handler transport.Handler) (transport.Subscription, error) {
	if err := transport.ValidateTopic(topic); err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	sub, err := b.conn.Subscribe(topic, func(msg *nats.Msg) {
		_ = handler(ctx, transport.Message{Topic: msg.Subject, Payload: msg.Data})
	})
	if err != nil {
		return nil, err
	}
	if err := b.conn.Flush(); err != nil {
		_ = sub.Unsubscribe()
		return nil, err
	}
	handle := &subscription{subscription: sub}
	if done := ctx.Done(); done != nil {
		go func() {
			<-done
			_ = handle.Unsubscribe()
		}()
	}
	return handle, nil
}

// Close drains and closes the NATS connection.
func (b *Bus) Close() error {
	if b.conn.IsClosed() {
		return nil
	}
	if err := b.conn.Drain(); err != nil {
		b.conn.Close()
		return err
	}
	return nil
}

// subscription is a NATS subscription handle.
type subscription struct {
	subscription *nats.Subscription
}

// Unsubscribe removes the underlying NATS subscription.
func (s *subscription) Unsubscribe() error {
	return s.subscription.Unsubscribe()
}
