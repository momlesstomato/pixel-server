package ws

import (
	"context"
	"errors"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/core/session"
	"pixelsv/pkg/core/transport"
)

// ErrNilBus indicates transport bus dependency is required.
var ErrNilBus = errors.New("transport bus is required")

// Gateway owns websocket session ingress and egress wiring.
type Gateway struct {
	// bus is the runtime transport adapter.
	bus transport.Bus
	// logger receives websocket lifecycle logs.
	logger *zap.Logger
	// sessions stores active websocket sessions.
	sessions *session.Manager
	// ids generates monotonic runtime session ids.
	ids *SessionIDGenerator
}

// NewGateway creates a websocket gateway from transport and logger dependencies.
func NewGateway(bus transport.Bus, logger *zap.Logger) (*Gateway, error) {
	if bus == nil {
		return nil, ErrNilBus
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Gateway{
		bus:      bus,
		logger:   logger,
		sessions: session.NewManager(),
		ids:      NewSessionIDGenerator(),
	}, nil
}

// Start subscribes to session output topics for websocket fan-out writes.
func (g *Gateway) Start(ctx context.Context) error {
	_, err := g.bus.Subscribe(ctx, sessionmessaging.OutputWildcardTopic(), g.handleSessionOutput)
	if err != nil {
		return err
	}
	_, err = g.bus.Subscribe(ctx, sessionmessaging.DisconnectWildcardTopic(), g.handleSessionDisconnect)
	return err
}

// Sessions returns active websocket session state.
func (g *Gateway) Sessions() *session.Manager {
	return g.sessions
}

// UpgradeMiddleware validates websocket upgrade requirements on the ws route.
func (g *Gateway) UpgradeMiddleware(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return c.SendStatus(fiber.StatusUpgradeRequired)
}

// HandleConnection serves one websocket connection lifecycle.
func (g *Gateway) HandleConnection(conn *websocket.Conn) {
	sessionID := g.ids.Next()
	if err := g.sessions.Register(sessionID, NewFiberConnectionAdapter(conn)); err != nil {
		g.logger.Error("failed to register websocket session", zap.Error(err))
		_ = conn.Close()
		return
	}
	g.logger.Info("websocket session connected", zap.String("session_id", sessionID))
	_ = g.bus.Publish(context.Background(), sessionmessaging.TopicConnected, []byte(sessionID))
	g.logger.Debug("session connected event published", zap.String("session_id", sessionID), zap.String("topic", sessionmessaging.TopicConnected))
	g.handleConnectionReadLoop(context.Background(), sessionID, conn)
	_ = g.bus.Publish(context.Background(), sessionmessaging.TopicDisconnected, []byte(sessionID))
	g.logger.Debug("session disconnected event published", zap.String("session_id", sessionID), zap.String("topic", sessionmessaging.TopicDisconnected))
	if err := g.sessions.Remove(sessionID); err != nil {
		g.logger.Debug("websocket session remove failed", zap.Error(err))
	}
	g.logger.Info("websocket session disconnected", zap.String("session_id", sessionID))
}

// handleConnectionReadLoop consumes websocket binary messages for one session.
func (g *Gateway) handleConnectionReadLoop(ctx context.Context, sessionID string, conn *websocket.Conn) {
	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		if err := g.handleBinary(ctx, sessionID, raw); err != nil {
			g.logger.Warn(
				"websocket packet handling failed",
				zap.String("session_id", sessionID),
				zap.Error(err),
			)
			return
		}
	}
}
