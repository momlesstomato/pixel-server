package tests

import (
	"bytes"
	"errors"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/cli"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeBuildsServer verifies initializer-driven startup wiring.
func TestExecuteServeBuildsServer(t *testing.T) {
	logBuffer := bytes.NewBuffer(nil)
	listenCalled := false
	err := cli.ExecuteServe(cli.ServeOptions{EnvFile: writeServeEnvFile(t), WebSocketPath: "/realtime", Output: logBuffer}, func(module *corehttp.Module, _ string) error {
		listenCalled = true
		request := httptest.NewRequest(nethttp.MethodGet, "/realtime", nil)
		request.Header.Set(corehttp.DefaultAPIKeyHeader, "test-key")
		response, testErr := module.App().Test(request)
		if testErr != nil {
			return testErr
		}
		if response.StatusCode != nethttp.StatusUpgradeRequired {
			t.Fatalf("expected status 426, got %d", response.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve execution success, got %v", err)
	}
	if !listenCalled {
		t.Fatalf("expected listen callback execution")
	}
	if !strings.Contains(logBuffer.String(), "http server starting") {
		t.Fatalf("expected startup log entry, got %s", logBuffer.String())
	}
}

// TestExecuteServeFailsWithMissingConfig verifies startup error propagation.
func TestExecuteServeFailsWithMissingConfig(t *testing.T) {
	err := cli.ExecuteServe(cli.ServeOptions{EnvFile: filepath.Join(t.TempDir(), ".env")}, nil)
	if err == nil {
		t.Fatalf("expected serve execution failure for missing mandatory config")
	}
}

// TestExecuteServePropagatesListenError verifies listener failures are returned.
func TestExecuteServePropagatesListenError(t *testing.T) {
	expected := errors.New("listen error")
	err := cli.ExecuteServe(cli.ServeOptions{EnvFile: writeServeEnvFile(t)}, func(_ *corehttp.Module, _ string) error { return expected })
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped listen error, got %v", err)
	}
}

// TestEchoWebSocketHandlerMirrorsFrames verifies websocket echo behavior.
func TestEchoWebSocketHandlerMirrorsFrames(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.RegisterWebSocket("/ws", cli.EchoWebSocketHandler); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	serverError := make(chan error, 1)
	go func() { serverError <- module.App().Listener(listener) }()
	defer func() { _ = module.Dispose(); _ = <-serverError }()
	connection := dialServeWebSocket(t, "ws://"+listener.Addr().String()+"/ws")
	defer connection.Close()
	if err := connection.WriteMessage(gws.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("expected websocket read success, got %v", err)
	}
	if string(payload) != "ping" {
		t.Fatalf("expected echoed payload, got %q", string(payload))
	}
}

// dialServeWebSocket creates a websocket client connection with retries.
func dialServeWebSocket(t *testing.T, url string) *gws.Conn {
	t.Helper()
	dialer := gws.Dialer{HandshakeTimeout: time.Second}
	for attempt := 0; attempt < 10; attempt++ {
		connection, _, err := dialer.Dial(url, nil)
		if err == nil {
			return connection
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected websocket dial success")
	return nil
}
