package bus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Bus wraps a NATS connection and JetStream context for publish/subscribe.
type Bus struct {
	conn *nats.Conn
	js   jetstream.JetStream
	log  *slog.Logger
}

// Option configures Bus construction.
type Option func(*busConfig)

type busConfig struct {
	logger *slog.Logger
}

// WithLogger sets a custom logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *busConfig) { c.logger = l }
}

// Connect establishes a NATS connection and initialises JetStream.
func Connect(ctx context.Context, url string, opts ...Option) (*Bus, error) {
	cfg := &busConfig{logger: slog.Default()}
	for _, o := range opts {
		o(cfg)
	}

	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(60),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			cfg.logger.Warn("nats disconnected", "err", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			cfg.logger.Info("nats reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("bus: connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("bus: jetstream init: %w", err)
	}

	return &Bus{conn: nc, js: js, log: cfg.logger}, nil
}

// Publish sends data to a NATS subject (core NATS, not JetStream).
func (b *Bus) Publish(subject string, data []byte) error {
	return b.conn.Publish(subject, data)
}

// Request sends a request and waits for a response with timeout.
func (b *Bus) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return b.conn.RequestWithContext(ctx, subject, data)
}

// Subscribe creates a core NATS subscription with a message handler.
func (b *Bus) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return b.conn.Subscribe(subject, handler)
}

// QueueSubscribe creates a queue subscription for load-balanced consumption.
func (b *Bus) QueueSubscribe(subject, queue string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return b.conn.QueueSubscribe(subject, queue, handler)
}

// JetStream returns the underlying JetStream context for advanced operations.
func (b *Bus) JetStream() jetstream.JetStream {
	return b.js
}

// Conn returns the underlying NATS connection.
func (b *Bus) Conn() *nats.Conn {
	return b.conn
}

// Close drains the connection and shuts down.
func (b *Bus) Close() error {
	b.conn.Close()
	return nil
}
