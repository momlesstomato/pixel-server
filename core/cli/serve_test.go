package cli

import (
	"bytes"
	"errors"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeBuildsServer verifies initializer-driven startup wiring.
func TestExecuteServeBuildsServer(t *testing.T) {
	logBuffer := bytes.NewBuffer(nil)
	envFile := writeServeEnvFile(t)
	listenCalled := false
	err := ExecuteServe(ServeOptions{
		EnvFile: envFile, WebSocketPath: "/realtime", Output: logBuffer,
	}, func(module *corehttp.Module, _ string) error {
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
	if logBuffer.Len() == 0 {
		t.Fatalf("expected log output to be written")
	}
	if !strings.Contains(logBuffer.String(), "http server starting") {
		t.Fatalf("expected startup log entry, got %s", logBuffer.String())
	}
}

// TestExecuteServeFailsWithMissingConfig verifies startup error propagation.
func TestExecuteServeFailsWithMissingConfig(t *testing.T) {
	err := ExecuteServe(ServeOptions{
		EnvFile: filepath.Join(t.TempDir(), ".env"),
	}, nil)
	if err == nil {
		t.Fatalf("expected serve execution failure for missing mandatory config")
	}
}

// TestNewServeCommandAppliesFlags verifies command flag parsing behavior.
func TestNewServeCommandAppliesFlags(t *testing.T) {
	envFile := writeServeEnvFile(t)
	captured := ""
	command := NewServeCommand(ServeDependencies{
		Listen: func(_ *corehttp.Module, address string) error { captured = address; return nil },
	})
	command.SetArgs([]string{"--env-file", envFile, "--ws-path", "/events"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected command execution success, got %v", err)
	}
	if captured == "" {
		t.Fatalf("expected listen callback to receive bind address")
	}
}

// TestExecuteServePropagatesListenError verifies listener failures are returned.
func TestExecuteServePropagatesListenError(t *testing.T) {
	expected := errors.New("listen error")
	err := ExecuteServe(ServeOptions{
		EnvFile: writeServeEnvFile(t),
	}, func(_ *corehttp.Module, _ string) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped listen error, got %v", err)
	}
}

// TestDefaultListenRejectsInvalidAddress verifies network startup validation.
func TestDefaultListenRejectsInvalidAddress(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := defaultListen(module, "invalid-address"); err == nil {
		t.Fatalf("expected invalid address error")
	}
}

// TestEchoWebSocketHandlerMirrorsFrames verifies websocket echo behavior.
func TestEchoWebSocketHandlerMirrorsFrames(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.RegisterWebSocket("/ws", EchoWebSocketHandler); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	serverError := make(chan error, 1)
	go func() { serverError <- module.App().Listener(listener) }()
	defer func() {
		_ = module.Dispose()
		_ = <-serverError
	}()
	connection := dialWebSocket(t, "ws://"+listener.Addr().String()+"/ws")
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

// dialWebSocket creates a websocket client connection with retries.
func dialWebSocket(t *testing.T, url string) *gws.Conn {
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

// writeServeEnvFile writes a valid env file for serve startup tests.
func writeServeEnvFile(t *testing.T) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), ".env")
	content := []byte("APP_BIND_IP=127.0.0.1\nAPP_PORT=3987\nAPP_API_KEY=test-key\nREDIS_ADDRESS=localhost:6379\nPOSTGRES_DSN=dsn\nUSERS_JWT_SECRET=secret\nLOGGING_LEVEL=debug\n")
	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return filePath
}
