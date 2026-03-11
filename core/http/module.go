package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var errNilWebSocketHandler = errors.New("websocket handler is required")

// WebSocketHandler defines the websocket endpoint handler signature.
type WebSocketHandler func(*websocket.Conn)

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
