package http

import (
	"errors"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// isPublicPath reports whether a request path bypasses API key enforcement.
func (module *Module) isPublicPath(path string) bool {
	if isPublicDocsRoute(path) {
		return true
	}
	normalized := normalizePath(path)
	module.webSocketMutex.RLock()
	_, exists := module.webSocketPaths[normalized]
	module.webSocketMutex.RUnlock()
	return exists
}

// RegisterWebSocket registers websocket upgrade and endpoint handlers.
func (module *Module) RegisterWebSocket(path string, handler WebSocketHandler) error {
	if handler == nil {
		return errNilWebSocketHandler
	}
	normalizedPath := normalizePath(path)
	if normalizedPath == "" || normalizedPath == "/" {
		return errWebSocketPathRequired
	}
	module.app.Use(normalizedPath, func(ctx *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(ctx) {
			return ctx.Next()
		}
		return fiber.NewError(nethttp.StatusUpgradeRequired, "websocket upgrade required")
	})
	module.webSocketMutex.Lock()
	module.webSocketPaths[normalizedPath] = struct{}{}
	module.webSocketMutex.Unlock()
	module.app.Get(normalizedPath, websocket.New(func(connection *websocket.Conn) {
		module.trackWebSocketConnection(connection)
		defer func() {
			module.untrackWebSocketConnection(connection)
			_ = connection.Close()
		}()
		handler(connection)
	}))
	return nil
}

// Dispose shuts down the Fiber application and releases network resources.
func (module *Module) Dispose() error {
	module.disposeOnce.Do(func() {
		webSocketError := module.closeWebSocketConnections()
		shutdownError := module.app.ShutdownWithTimeout(DefaultShutdownTimeout)
		module.disposeError = errors.Join(webSocketError, shutdownError)
	})
	return module.disposeError
}

// trackWebSocketConnection registers one websocket connection as active.
func (module *Module) trackWebSocketConnection(connection *websocket.Conn) {
	module.webSocketMutex.Lock()
	module.webSocketConnections[connection] = struct{}{}
	module.webSocketMutex.Unlock()
}

// untrackWebSocketConnection removes one websocket connection from active set.
func (module *Module) untrackWebSocketConnection(connection *websocket.Conn) {
	module.webSocketMutex.Lock()
	delete(module.webSocketConnections, connection)
	module.webSocketMutex.Unlock()
}

// closeWebSocketConnections sends close packets and closes active websocket connections.
func (module *Module) closeWebSocketConnections() error {
	module.webSocketMutex.Lock()
	connections := make([]*websocket.Conn, 0, len(module.webSocketConnections))
	for connection := range module.webSocketConnections {
		connections = append(connections, connection)
	}
	module.webSocketConnections = map[*websocket.Conn]struct{}{}
	module.webSocketMutex.Unlock()
	var closeErrors []error
	for _, connection := range connections {
		if connection == nil {
			continue
		}
		if err := connection.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server shutdown"),
			time.Now().Add(DefaultWebSocketCloseTimeout),
		); err != nil {
			closeErrors = append(closeErrors, err)
		}
		if err := connection.Close(); err != nil {
			closeErrors = append(closeErrors, err)
		}
	}
	return errors.Join(closeErrors...)
}

// normalizePath canonicalizes route paths for middleware checks.
func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	normalized := strings.TrimSuffix(trimmed, "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}
