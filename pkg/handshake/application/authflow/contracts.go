package authflow

import (
	"context"

	"github.com/gofiber/contrib/websocket"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// UnauthorizedCloseCode defines the websocket close code for auth failures.
const UnauthorizedCloseCode = websocket.CloseAbnormalClosure

// DuplicateLoginCloseCode defines the websocket close code for duplicate sessions.
const DuplicateLoginCloseCode = websocket.ClosePolicyViolation

// AuthTimeoutCloseCode defines the websocket close code for auth timeout.
const AuthTimeoutCloseCode = websocket.ClosePolicyViolation

// TicketValidator defines ticket validation behavior for authentication flow.
type TicketValidator interface {
	// Validate consumes one ticket and returns associated user identifier.
	Validate(context.Context, string) (int, error)
}

// UserFinder defines user existence verification behavior for post-ticket validation.
type UserFinder interface {
	// FindByID verifies a user exists by identifier, returning error when not found.
	FindByID(context.Context, int) error
}

// SessionRegistry defines session lifecycle persistence behavior.
type SessionRegistry interface {
	// Register stores or updates one connection session.
	Register(coreconnection.Session) error
	// FindByUserID resolves one active session by user identifier.
	FindByUserID(int) (coreconnection.Session, bool)
	// Remove deletes one session by connection identifier.
	Remove(string)
}

// Transport defines packet and connection transport behavior.
type Transport interface {
	// Send writes one encoded packet to one connection.
	Send(string, uint16, []byte) error
	// Close closes one connection with code and reason.
	Close(string, int, string) error
}

// AuthenticateRequest defines one authentication attempt payload.
type AuthenticateRequest struct {
	// ConnID stores target connection identifier.
	ConnID string
	// Ticket stores raw SSO ticket value.
	Ticket string
	// MachineID stores normalized machine identifier.
	MachineID string
}

// AuthenticateResult defines one successful authentication output.
type AuthenticateResult struct {
	// UserID stores authenticated user identifier.
	UserID int
	// KickedConnID stores duplicate connection identifier removed by auth flow.
	KickedConnID string
}
