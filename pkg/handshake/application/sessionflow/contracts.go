package sessionflow

import (
	"errors"

	"github.com/gofiber/contrib/websocket"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// PongTimeoutCloseCode defines websocket close code for heartbeat timeout.
const PongTimeoutCloseCode = websocket.CloseAbnormalClosure

// ErrPongTimeoutElapsed defines heartbeat timeout expiration behavior.
var ErrPongTimeoutElapsed = errors.New("pong timeout elapsed")

// Transport defines packet send and connection close behavior.
type Transport interface {
	// Send writes one encoded packet payload to one connection.
	Send(string, uint16, []byte) error
	// Close closes one connection with code and reason payload.
	Close(string, int, string) error
}

// SessionRegistry defines session lifecycle storage behavior.
type SessionRegistry interface {
	// Register stores one session record.
	Register(coreconnection.Session) error
	// FindByConnID resolves one session by connection identifier.
	FindByConnID(string) (coreconnection.Session, bool)
	// Remove deletes one session by connection identifier.
	Remove(string)
}
