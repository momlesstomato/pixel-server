package e2e_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"pixelsv/internal/auth"
	"pixelsv/internal/sessionconnection"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport/local"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/plugin/eventbus"
	"pixelsv/pkg/protocol"
)

// Test08SessionConnectionE2E validates phase-1 session-connection runtime behavior.
func Test08SessionConnectionE2E(t *testing.T) {
	bus := local.New()
	address := openLocalAddress(t)
	server, err := httpserver.New(httpserver.Config{Address: address, DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}, nil, bus)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authRuntime, err := auth.Register(ctx, server.App(), bus, eventbus.New(), nil, "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = sessionconnection.Register(ctx, bus, eventbus.New(), nil, sessionconnection.DefaultConfig())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ticket, _, err := authRuntime.Service.CreateTicket(21, 60)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe(ctx) }()
	waitHealth(t, address)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer conn.Close()
	sendPacket(t, conn, &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1})
	sendPacket(t, conn, &protocol.SecuritySsoTicketPacket{Ticket: ticket})
	assertContainsHeaders(t, conn, map[uint16]int{2491: 1, 2033: 1})
	sendPacket(t, conn, &protocol.ClientLatencyTestPacket{RequestId: 33})
	assertOneHeader(t, conn, 10)
	sendPacket(t, conn, &protocol.SessionDesktopViewPacket{})
	assertOneHeader(t, conn, 122)
	cancel()
	select {
	case shutdownErr := <-errCh:
		if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
			t.Fatalf("unexpected shutdown error: %v", shutdownErr)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected server shutdown")
	}
}

// assertContainsHeaders reads websocket output until all expected headers are found.
func assertContainsHeaders(t *testing.T, conn *websocket.Conn, expected map[uint16]int) {
	t.Helper()
	remaining := 0
	for _, count := range expected {
		remaining += count
	}
	deadline := time.Now().Add(time.Second)
	for remaining > 0 && time.Now().Before(deadline) {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		frames, err := codec.SplitFrames(payload)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		for _, frame := range frames {
			if expected[frame.Header] > 0 {
				expected[frame.Header]--
				remaining--
			}
		}
	}
	if remaining != 0 {
		t.Fatalf("expected headers not found: %+v", expected)
	}
}
