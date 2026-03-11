package http

import (
	"bytes"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	coreconfig "github.com/momlesstomato/pixel-server/core/config"
	corelogging "github.com/momlesstomato/pixel-server/core/logging"
)

// TestNewRegistersZapMiddleware verifies middleware emits request logs.
func TestNewRegistersZapMiddleware(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := corelogging.New(coreconfig.LoggingConfig{Format: "json", Level: "info"}, buffer)
	if err != nil {
		t.Fatalf("expected logger creation to succeed, got %v", err)
	}
	module := New(Options{Logger: logger})
	module.RegisterGET("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	request := httptest.NewRequest(nethttp.MethodGet, "/health", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected HTTP test request to succeed, got %v", err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	output := buffer.String()
	if output == "" || !strings.Contains(output, "/health") {
		t.Fatalf("expected zapfiber log output with request path, got %s", output)
	}
}

// TestRegisterWebSocketReturnsUpgradeRequired verifies non-upgrade requests are rejected.
func TestRegisterWebSocketReturnsUpgradeRequired(t *testing.T) {
	module := New(Options{})
	err := module.RegisterWebSocket("/ws", func(_ *websocket.Conn) {})
	if err != nil {
		t.Fatalf("expected websocket registration to succeed, got %v", err)
	}
	request := httptest.NewRequest(nethttp.MethodGet, "/ws", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected HTTP test request to succeed, got %v", err)
	}
	if response.StatusCode != nethttp.StatusUpgradeRequired {
		t.Fatalf("expected status 426, got %d", response.StatusCode)
	}
}

// TestRegisterWebSocketRejectsNilHandler verifies nil handlers are not accepted.
func TestRegisterWebSocketRejectsNilHandler(t *testing.T) {
	module := New(Options{})
	if err := module.RegisterWebSocket("/ws", nil); err == nil {
		t.Fatalf("expected nil websocket handler to fail")
	}
}

// TestDisposeShutsDownModule verifies module disposal path.
func TestDisposeShutsDownModule(t *testing.T) {
	module := New(Options{})
	if err := module.Dispose(); err != nil {
		t.Fatalf("expected dispose to succeed, got %v", err)
	}
}
