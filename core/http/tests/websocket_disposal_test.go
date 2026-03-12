package tests

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/gofiber/contrib/websocket"
	gws "github.com/gorilla/websocket"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestDisposeSendsCloseFrameToWebSocketClients verifies graceful websocket shutdown behavior.
func TestDisposeSendsCloseFrameToWebSocketClients(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.RegisterWebSocket("/ws", func(connection *websocket.Conn) {
		for {
			if _, _, err := connection.ReadMessage(); err != nil {
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
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- module.App().Listener(listener) }()
	dialer := gws.Dialer{HandshakeTimeout: time.Second}
	connection, _, err := dialer.Dial("ws://"+listener.Addr().String()+"/ws", nil)
	if err != nil {
		t.Fatalf("expected websocket dial success, got %v", err)
	}
	defer connection.Close()
	if disposeErr := module.Dispose(); disposeErr != nil {
		t.Fatalf("expected dispose success, got %v", disposeErr)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, _, readErr := connection.ReadMessage()
		if readErr == nil {
			continue
		}
		var closeErr *gws.CloseError
		if !errors.As(readErr, &closeErr) {
			t.Fatalf("expected close frame after dispose, got %v", readErr)
		}
		if closeErr.Code != corehttp.DefaultShutdownWebSocketCloseCode {
			t.Fatalf("expected shutdown close code %d, got %d", corehttp.DefaultShutdownWebSocketCloseCode, closeErr.Code)
		}
		break
	}
	_ = <-serverErrors
}
