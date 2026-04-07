package cli

import (
	"bytes"
	"errors"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	fws "github.com/gofiber/contrib/websocket"
	gws "github.com/gorilla/websocket"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/initializer"
	corelogging "github.com/momlesstomato/pixel-server/core/logging"
)

// syncBuffer captures writes and tracks Sync calls from zap cleanup.
type syncBuffer struct {
	// Buffer stores encoded log output payload.
	bytes.Buffer
	// syncCalls stores total Sync invocation count.
	syncCalls int32
}

// Sync records one sync invocation for cleanup assertions.
func (buffer *syncBuffer) Sync() error {
	atomic.AddInt32(&buffer.syncCalls, 1)
	return nil
}

// SyncCount returns total sync calls performed on this writer.
func (buffer *syncBuffer) SyncCount() int32 {
	return atomic.LoadInt32(&buffer.syncCalls)
}

// TestRunServeLifecycleWithSignalsDisposesResources verifies graceful interrupt shutdown behavior.
func TestRunServeLifecycleWithSignalsDisposesResources(t *testing.T) {
	logBuffer := &syncBuffer{}
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "debug"}, logBuffer)
	if err != nil {
		t.Fatalf("expected logger creation success, got %v", err)
	}
	module := corehttp.New(corehttp.Options{Logger: logger})
	if err := module.RegisterWebSocket("/ws", NewEchoWebSocketHandler(logger)); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	runtime := &initializer.Runtime{Logger: logger}
	signals := make(chan os.Signal, 1)
	lifecycleResult := make(chan error, 1)
	go func() {
		lifecycleResult <- runServeLifecycleWithSignals(runtime, module, listener.Addr().String(), func(httpModule *corehttp.Module, _ string) error {
			return httpModule.App().Listener(listener)
		}, signals)
	}()
	connection := dialWebSocket(t, "ws://"+listener.Addr().String()+"/ws")
	defer connection.Close()
	signals <- os.Interrupt
	select {
	case runErr := <-lifecycleResult:
		if runErr != nil {
			t.Fatalf("expected lifecycle success on interrupt, got %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected lifecycle completion after interrupt")
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, _, readErr := connection.ReadMessage()
		if readErr == nil {
			continue
		}
		var closeErr *gws.CloseError
		if !errors.As(readErr, &closeErr) {
			t.Fatalf("expected websocket close after interrupt, got %v", readErr)
		}
		if closeErr.Code != corehttp.DefaultShutdownWebSocketCloseCode {
			t.Fatalf("expected shutdown close code %d, got %d", corehttp.DefaultShutdownWebSocketCloseCode, closeErr.Code)
		}
		break
	}
	if logBuffer.SyncCount() == 0 {
		t.Fatalf("expected logger sync to execute during cleanup")
	}
}

// TestRunServeLifecycleRejectsInvalidInputs verifies startup precondition checks.
func TestRunServeLifecycleRejectsInvalidInputs(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := runServeLifecycle(nil, module, "127.0.0.1:0", nil); err == nil {
		t.Fatalf("expected runtime logger precondition failure")
	}
	if err := runServeLifecycle(&initializer.Runtime{}, module, "127.0.0.1:0", nil); err == nil {
		t.Fatalf("expected runtime logger precondition failure")
	}
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "debug"}, bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("expected logger creation success, got %v", err)
	}
	if err := runServeLifecycle(&initializer.Runtime{Logger: logger}, nil, "127.0.0.1:0", nil); err == nil {
		t.Fatalf("expected http module precondition failure")
	}
}

// TestRunServeLifecycleWithSignalsFlushesRegisteredWebSocketCloser verifies shutdown waits long enough for registered closers to flush final frames.
func TestRunServeLifecycleWithSignalsFlushesRegisteredWebSocketCloser(t *testing.T) {
	logBuffer := &syncBuffer{}
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "debug"}, logBuffer)
	if err != nil {
		t.Fatalf("expected logger creation success, got %v", err)
	}
	module := corehttp.New(corehttp.Options{Logger: logger})
	if err := module.RegisterWebSocket("/ws", func(connection *fws.Conn) {
		module.RegisterWebSocketCloser(connection, func() {
			_ = connection.WriteMessage(gws.BinaryMessage, []byte("shutdown"))
			_ = connection.WriteControl(gws.CloseMessage, gws.FormatCloseMessage(corehttp.DefaultShutdownWebSocketCloseCode, "server shutdown"), time.Now().Add(time.Second))
			_ = connection.Close()
		})
		defer module.UnregisterWebSocketCloser(connection)
		for {
			if _, _, readErr := connection.ReadMessage(); readErr != nil {
				return
			}
		}
	}); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	runtime := &initializer.Runtime{Logger: logger}
	signals := make(chan os.Signal, 1)
	lifecycleResult := make(chan error, 1)
	go func() {
		lifecycleResult <- runServeLifecycleWithSignals(runtime, module, listener.Addr().String(), func(httpModule *corehttp.Module, _ string) error {
			return httpModule.App().Listener(listener)
		}, signals)
	}()
	connection := dialWebSocket(t, "ws://"+listener.Addr().String()+"/ws")
	defer connection.Close()
	signals <- os.Interrupt
	connection.SetReadDeadline(time.Now().Add(time.Second))
	messageType, payload, readErr := connection.ReadMessage()
	if readErr != nil {
		t.Fatalf("expected shutdown payload before close, got %v", readErr)
	}
	if messageType != gws.BinaryMessage || string(payload) != "shutdown" {
		t.Fatalf("expected shutdown payload, got type=%d payload=%q", messageType, string(payload))
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, _, readErr = connection.ReadMessage()
		if readErr == nil {
			continue
		}
		var closeErr *gws.CloseError
		if !errors.As(readErr, &closeErr) {
			t.Fatalf("expected websocket close after shutdown payload, got %v", readErr)
		}
		if closeErr.Code != corehttp.DefaultShutdownWebSocketCloseCode {
			t.Fatalf("expected shutdown close code %d, got %d", corehttp.DefaultShutdownWebSocketCloseCode, closeErr.Code)
		}
		break
	}
	select {
	case runErr := <-lifecycleResult:
		if runErr != nil {
			t.Fatalf("expected lifecycle success on interrupt, got %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected lifecycle completion after interrupt")
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
