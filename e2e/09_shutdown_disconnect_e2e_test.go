package e2e_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"pixelsv/internal/auth"
	"pixelsv/internal/sessionconnection"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/plugin/eventbus"
)

// Test09ShutdownDisconnectE2E validates graceful shutdown publishes session.disconnected.
func Test09ShutdownDisconnectE2E(t *testing.T) {
	bus := local.New()
	address := openLocalAddress(t)
	server, err := httpserver.New(httpserver.Config{Address: address, DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}, nil, bus)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := auth.Register(ctx, server.App(), bus, eventbus.New(), nil, "secret"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := sessionconnection.Register(ctx, bus, eventbus.New(), nil, sessionconnection.DefaultConfig()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	disconnected := make(chan transport.Message, 1)
	_, _ = bus.Subscribe(context.Background(), sessionmessaging.TopicDisconnected, func(_ context.Context, message transport.Message) error {
		disconnected <- message
		return nil
	})
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe(ctx) }()
	waitHealth(t, address)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer conn.Close()
	cancel()
	select {
	case message := <-disconnected:
		if string(message.Payload) != "1" {
			t.Fatalf("unexpected session id: %s", string(message.Payload))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected session.disconnected publish on shutdown")
	}
	select {
	case shutdownErr := <-errCh:
		if shutdownErr != nil && shutdownErr != http.ErrServerClosed {
			t.Fatalf("unexpected shutdown error: %v", shutdownErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected server shutdown")
	}
}
