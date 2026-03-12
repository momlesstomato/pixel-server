package cli

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	corelogging "github.com/momlesstomato/pixel-server/core/logging"
)

// TestNewEchoWebSocketHandlerLogsPacketFlow verifies debug packet telemetry behavior.
func TestNewEchoWebSocketHandlerLogsPacketFlow(t *testing.T) {
	logBuffer := bytes.NewBuffer(nil)
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "debug"}, logBuffer)
	if err != nil {
		t.Fatalf("expected logger creation success, got %v", err)
	}
	module := corehttp.New(corehttp.Options{})
	if err := module.RegisterWebSocket("/ws", NewEchoWebSocketHandler(logger)); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- module.App().Listener(listener) }()
	connection, _, err := gws.DefaultDialer.Dial("ws://"+listener.Addr().String()+"/ws", nil)
	if err != nil {
		t.Fatalf("expected websocket dial success, got %v", err)
	}
	if err := connection.WriteMessage(gws.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
	if _, _, err := connection.ReadMessage(); err != nil {
		t.Fatalf("expected websocket read success, got %v", err)
	}
	if disposeErr := module.Dispose(); disposeErr != nil {
		t.Fatalf("expected module dispose success, got %v", disposeErr)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, readErr := connection.ReadMessage()
	if readErr == nil {
		t.Fatalf("expected close frame after module disposal")
	}
	var closeErr *gws.CloseError
	if !errors.As(readErr, &closeErr) || closeErr.Code != gws.CloseNormalClosure {
		t.Fatalf("expected normal close frame, got %v", readErr)
	}
	_ = <-serverErrors
	_ = connection.Close()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if strings.Contains(logBuffer.String(), "websocket connection disposed") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	output := logBuffer.String()
	if !strings.Contains(output, "websocket packet received") || !strings.Contains(output, "websocket packet sent") || !strings.Contains(output, "websocket connection disposed") {
		t.Fatalf("expected websocket telemetry logs, got %s", output)
	}
}

// TestIsIgnorableSyncErrorMatchesExpectedCases verifies logger sync error filtering behavior.
func TestIsIgnorableSyncErrorMatchesExpectedCases(t *testing.T) {
	if !isIgnorableSyncError(syscall.EBADF) {
		t.Fatalf("expected EBADF to be ignorable")
	}
	if !isIgnorableSyncError(syscall.EINVAL) {
		t.Fatalf("expected EINVAL to be ignorable")
	}
	if isIgnorableSyncError(nil) {
		t.Fatalf("expected nil error to be non-ignorable")
	}
}
