package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	"go.uber.org/zap"
)

// Transport defines websocket transport behavior for one active connection.
type Transport struct {
	// connID stores current websocket connection identifier.
	connID string
	// connection stores underlying websocket transport endpoint.
	connection *websocket.Conn
	// bus stores distributed close-signal behavior for foreign connections.
	bus CloseSignalBus
	// logger stores packet transport telemetry behavior.
	logger *zap.Logger
	// mutex serializes websocket writes.
	mutex sync.Mutex
}

// NewTransport creates websocket transport behavior.
func NewTransport(connID string, connection *websocket.Conn, bus CloseSignalBus, logger *zap.Logger) (*Transport, error) {
	if connID == "" {
		return nil, fmt.Errorf("connection id is required")
	}
	if connection == nil {
		return nil, fmt.Errorf("websocket connection is required")
	}
	if bus == nil {
		return nil, fmt.Errorf("close signal bus is required")
	}
	output := logger
	if output == nil {
		output = zap.NewNop()
	}
	return &Transport{connID: connID, connection: connection, bus: bus, logger: output}, nil
}

// Send writes one encoded packet to one connection.
func (transport *Transport) Send(connID string, packetID uint16, body []byte) error {
	if connID != transport.connID {
		return fmt.Errorf("target connection %s is not local", connID)
	}
	frame := codec.EncodeFrame(packetID, body)
	transport.mutex.Lock()
	err := transport.connection.WriteMessage(websocket.BinaryMessage, frame)
	transport.mutex.Unlock()
	if err == nil {
		transport.logger.Debug("websocket packet sent", zap.String("conn_id", connID), zap.Uint16("packet_id", packetID), zap.Int("size", len(body)))
	}
	return err
}

// Close closes one connection locally or sends a distributed close instruction.
func (transport *Transport) Close(connID string, code int, reason string) error {
	if connID == transport.connID {
		return transport.closeLocal(code, reason)
	}
	transport.logger.Debug("websocket close signal published", zap.String("conn_id", connID), zap.Int("code", code), zap.String("reason", reason))
	return transport.bus.Publish(context.Background(), connID, CloseSignal{Code: code, Reason: reason})
}

// closeLocal closes the local websocket connection with one close control frame.
func (transport *Transport) closeLocal(code int, reason string) error {
	transport.mutex.Lock()
	err := transport.connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
	transport.mutex.Unlock()
	if err != nil {
		return err
	}
	transport.logger.Debug("websocket connection closed", zap.String("conn_id", transport.connID), zap.Int("code", code), zap.String("reason", reason))
	return transport.connection.Close()
}
