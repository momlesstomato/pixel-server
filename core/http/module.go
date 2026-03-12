package http

import (
	"crypto/subtle"
	"errors"
	nethttp "net/http"
	"strings"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var errNilWebSocketHandler = errors.New("websocket handler is required")
var errAPIKeyRequired = errors.New("api key is required")

// WebSocketHandler defines the websocket endpoint handler signature.
type WebSocketHandler func(*websocket.Conn)

// DefaultAPIKeyHeader defines the default header used for API key auth.
const DefaultAPIKeyHeader = "X-API-Key"

// Options defines configurable dependencies for the HTTP module.
type Options struct {
	// Logger defines the zap logger used by fiberzap middleware.
	Logger *zap.Logger
	// FiberConfig defines Fiber server settings.
	FiberConfig fiber.Config
}

// Module defines the HTTP delivery module for API and websocket traffic.
type Module struct {
	// app stores the configured Fiber application instance.
	app *fiber.App
}

// New creates a Fiber module with zapfiber middleware pre-configured.
func New(options Options) *Module {
	logger := options.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	app := fiber.New(options.FiberConfig)
	app.Use(fiberzap.New(fiberzap.Config{Logger: logger}))
	return &Module{app: app}
}

// App returns the underlying Fiber application.
func (module *Module) App() *fiber.App {
	return module.app
}

// RegisterGET registers an HTTP GET endpoint on the Fiber application.
func (module *Module) RegisterGET(path string, handler fiber.Handler) {
	module.app.Get(path, handler)
}

// RegisterPOST registers an HTTP POST endpoint on the Fiber application.
func (module *Module) RegisterPOST(path string, handler fiber.Handler) {
	module.app.Post(path, handler)
}

// ProtectWithAPIKey protects all registered routes using API key middleware.
func (module *Module) ProtectWithAPIKey(apiKey string, header string) error {
	trimmed := strings.TrimSpace(apiKey)
	if trimmed == "" {
		return errAPIKeyRequired
	}
	keyHeader := header
	if keyHeader == "" {
		keyHeader = DefaultAPIKeyHeader
	}
	module.app.Use(func(ctx *fiber.Ctx) error {
		provided := ctx.Get(keyHeader)
		if provided == "" {
			provided = ctx.Query("api_key")
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(trimmed)) != 1 {
			return fiber.NewError(nethttp.StatusUnauthorized, "invalid api key")
		}
		return ctx.Next()
	})
	return nil
}

// RegisterWebSocket registers websocket upgrade and endpoint handlers.
func (module *Module) RegisterWebSocket(path string, handler WebSocketHandler) error {
	if handler == nil {
		return errNilWebSocketHandler
	}
	module.app.Use(path, func(ctx *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(ctx) {
			return ctx.Next()
		}
		return fiber.NewError(nethttp.StatusUpgradeRequired, "websocket upgrade required")
	})
	module.app.Get(path, websocket.New(func(connection *websocket.Conn) {
		handler(connection)
	}))
	return nil
}

// Dispose shuts down the Fiber application and releases network resources.
func (module *Module) Dispose() error {
	return module.app.Shutdown()
}
