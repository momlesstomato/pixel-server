package http

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/momlesstomato/pixel-server/core/app"
	"go.uber.org/zap"
)

// Stage defines HTTP startup behavior.
type Stage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializeHTTP creates an HTTP module from the configured logger.
	InitializeHTTP(app.Config, *zap.Logger) (*Module, error)
}

// WebSocketStage defines websocket startup behavior.
type WebSocketStage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializeWebSocket registers websocket endpoints in the HTTP module.
	InitializeWebSocket(*Module) error
}

// Initializer provides default HTTP startup behavior.
type Initializer struct {
	// FiberConfig defines the Fiber app configuration.
	FiberConfig fiber.Config
	// APIKeyHeader defines the request header used for key transport.
	APIKeyHeader string
}

// Name returns the stable initializer name.
func (initializer Initializer) Name() string {
	return "http"
}

// InitializeHTTP builds and returns an HTTP module.
func (initializer Initializer) InitializeHTTP(loaded app.Config, logger *zap.Logger) (*Module, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	module := New(Options{Logger: logger, FiberConfig: initializer.FiberConfig})
	if err := module.ProtectWithAPIKey(loaded.APIKey, initializer.APIKeyHeader); err != nil {
		return nil, err
	}
	return module, nil
}

// WebSocketInitializer registers the websocket endpoint in the HTTP module.
type WebSocketInitializer struct {
	// Path defines the websocket endpoint route.
	Path string
	// Handler defines websocket connection behavior.
	Handler WebSocketHandler
}

// Name returns the stable initializer name.
func (initializer WebSocketInitializer) Name() string {
	return "websocket"
}

// InitializeWebSocket registers websocket behavior in the HTTP module.
func (initializer WebSocketInitializer) InitializeWebSocket(module *Module) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	path := initializer.Path
	if path == "" {
		path = "/ws"
	}
	if initializer.Handler == nil {
		return fmt.Errorf("websocket handler is required")
	}
	return module.RegisterWebSocket(path, func(connection *websocket.Conn) {
		initializer.Handler(connection)
	})
}
