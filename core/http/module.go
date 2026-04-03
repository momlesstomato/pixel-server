package http

import (
	"crypto/subtle"
	"errors"
	nethttp "net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var errNilWebSocketHandler = errors.New("websocket handler is required")
var errAPIKeyRequired = errors.New("api key is required")
var errWebSocketPathRequired = errors.New("websocket path is required")

// WebSocketHandler defines the websocket endpoint handler signature.
type WebSocketHandler func(*websocket.Conn)

// DefaultAPIKeyHeader defines the default header used for API key auth.
const DefaultAPIKeyHeader = "X-API-Key"

// DefaultReadBufferSize defines the default request read buffer size in bytes.
const DefaultReadBufferSize = 16 * 1024

// DefaultOpenAPISpecPath defines the raw OpenAPI document route.
const DefaultOpenAPISpecPath = "/openapi.json"

// DefaultSwaggerUIPath defines the Swagger UI route.
const DefaultSwaggerUIPath = "/swagger"

// DefaultWebSocketCloseTimeout defines close control write timeout.
const DefaultWebSocketCloseTimeout = 2 * time.Second

// DefaultShutdownWebSocketCloseCode defines close code used during server shutdown.
const DefaultShutdownWebSocketCloseCode = websocket.CloseGoingAway

// DefaultDisconnectReasonPacketID defines disconnect reason packet identifier.
const DefaultDisconnectReasonPacketID uint16 = 4000

// DefaultShutdownDisconnectReasonCode defines protocol disconnect reason for shutdown.
const DefaultShutdownDisconnectReasonCode int32 = 19

// DefaultShutdownTimeout defines graceful application shutdown timeout.
const DefaultShutdownTimeout = 5 * time.Second

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
	// webSocketPaths stores websocket routes that bypass API key checks.
	webSocketPaths map[string]struct{}
	// webSocketConnections stores active websocket connections.
	webSocketConnections map[*websocket.Conn]struct{}
	// webSocketClosers stores per-connection graceful close functions.
	webSocketClosers map[*websocket.Conn]func()
	// webSocketMutex guards websocket routes and connection tracking.
	webSocketMutex sync.RWMutex
	// disposeOnce guarantees disposal executes exactly once.
	disposeOnce sync.Once
	// disposeError stores the first disposal error.
	disposeError error
}

// New creates a Fiber module with zapfiber and ray-id trace middlewares pre-configured.
func New(options Options) *Module {
	logger := options.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	fiberConfig := options.FiberConfig
	fiberConfig.DisableStartupMessage = true
	if fiberConfig.ReadBufferSize <= 0 {
		fiberConfig.ReadBufferSize = DefaultReadBufferSize
	}
	if fiberConfig.ErrorHandler == nil {
		fiberConfig.ErrorHandler = buildErrorHandler(logger)
	}
	app := fiber.New(fiberConfig)
	app.Use(TraceMiddleware())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
		Next: func(_ *fiber.Ctx) bool {
			return !logger.Core().Enabled(zap.DebugLevel)
		},
	}))
	return &Module{
		app:                  app,
		webSocketPaths:       map[string]struct{}{},
		webSocketConnections: map[*websocket.Conn]struct{}{},
		webSocketClosers:     map[*websocket.Conn]func(){},
	}
}

// buildErrorHandler returns a Fiber error handler that logs errors with ray_id context.
func buildErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := nethttp.StatusInternalServerError
		message := "internal server error"
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			message = fiberErr.Message
		}
		rayID := RayID(ctx)
		if code >= nethttp.StatusInternalServerError {
			logger.Error("http error", zap.String("ray_id", rayID), zap.Int("status", code), zap.Error(err))
		}
		ctx.Set(HeaderRayID, rayID)
		ctx.Status(code)
		return ctx.JSON(fiber.Map{"error": message})
	}
}

// App returns the underlying Fiber application.
func (module *Module) App() *fiber.App {
	return module.app
}

// RegisterWebSocketCloser stores a graceful close function for one active connection.
// The function is called instead of a raw write when the server shuts down, ensuring
// the disconnect reason packet is sent through the encrypted transport.
func (module *Module) RegisterWebSocketCloser(conn *websocket.Conn, fn func()) {
	module.webSocketMutex.Lock()
	module.webSocketClosers[conn] = fn
	module.webSocketMutex.Unlock()
}

// UnregisterWebSocketCloser removes the graceful close function for one connection.
func (module *Module) UnregisterWebSocketCloser(conn *websocket.Conn) {
	module.webSocketMutex.Lock()
	delete(module.webSocketClosers, conn)
	module.webSocketMutex.Unlock()
}

// RegisterGET registers an HTTP GET endpoint on the Fiber application.
func (module *Module) RegisterGET(path string, handler fiber.Handler) {
	module.app.Get(path, handler)
}

// RegisterPOST registers an HTTP POST endpoint on the Fiber application.
func (module *Module) RegisterPOST(path string, handler fiber.Handler) {
	module.app.Post(path, handler)
}

// RegisterPATCH registers an HTTP PATCH endpoint on the Fiber application.
func (module *Module) RegisterPATCH(path string, handler fiber.Handler) {
	module.app.Patch(path, handler)
}

// RegisterDELETE registers an HTTP DELETE endpoint on the Fiber application.
func (module *Module) RegisterDELETE(path string, handler fiber.Handler) {
	module.app.Delete(path, handler)
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
		if module.isPublicPath(ctx.Path()) {
			return ctx.Next()
		}
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

// isPublicDocsRoute reports whether a route must bypass API key enforcement.
func isPublicDocsRoute(path string) bool {
	if path == DefaultOpenAPISpecPath || path == DefaultSwaggerUIPath {
		return true
	}
	return strings.HasPrefix(path, DefaultSwaggerUIPath+"/")
}
