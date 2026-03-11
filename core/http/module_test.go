package http

import (
	"bytes"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	coreapp "github.com/momlesstomato/pixel-server/core/app"
	corelogging "github.com/momlesstomato/pixel-server/core/logging"
	"go.uber.org/zap"
)

// TestNewRegistersZapMiddleware verifies middleware emits request logs.
func TestNewRegistersZapMiddleware(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "info"}, buffer)
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

// TestInitializerBuildsHTTPModule verifies package-owned initializer behavior.
func TestInitializerBuildsHTTPModule(t *testing.T) {
	module, err := (Initializer{}).InitializeHTTP(coreapp.Config{APIKey: "secret"}, zap.NewNop())
	if err != nil {
		t.Fatalf("expected http initializer success, got %v", err)
	}
	if module == nil {
		t.Fatalf("expected non-nil module")
	}
}

// TestInitializerRejectsNilLogger verifies logger precondition checks.
func TestInitializerRejectsNilLogger(t *testing.T) {
	if _, err := (Initializer{}).InitializeHTTP(coreapp.Config{APIKey: "secret"}, nil); err == nil {
		t.Fatalf("expected http initializer error for nil logger")
	}
}

// TestInitializerRejectsEmptyAPIKey verifies API key precondition checks.
func TestInitializerRejectsEmptyAPIKey(t *testing.T) {
	if _, err := (Initializer{}).InitializeHTTP(coreapp.Config{}, zap.NewNop()); err == nil {
		t.Fatalf("expected http initializer error for empty api key")
	}
}

// TestWebSocketInitializerRegistersRoute verifies websocket stage behavior.
func TestWebSocketInitializerRegistersRoute(t *testing.T) {
	module := New(Options{})
	stage := WebSocketInitializer{Path: "/events", Handler: func(_ *websocket.Conn) {}}
	if err := stage.InitializeWebSocket(module); err != nil {
		t.Fatalf("expected websocket initializer success, got %v", err)
	}
	request := httptest.NewRequest(nethttp.MethodGet, "/events", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusUpgradeRequired {
		t.Fatalf("expected status 426, got %d", response.StatusCode)
	}
}

// TestWebSocketInitializerRejectsInvalidInputs verifies stage precondition checks.
func TestWebSocketInitializerRejectsInvalidInputs(t *testing.T) {
	if err := (WebSocketInitializer{}).InitializeWebSocket(nil); err == nil {
		t.Fatalf("expected websocket initializer error for nil module")
	}
	module := New(Options{})
	if err := (WebSocketInitializer{}).InitializeWebSocket(module); err == nil {
		t.Fatalf("expected websocket initializer error for nil handler")
	}
}
