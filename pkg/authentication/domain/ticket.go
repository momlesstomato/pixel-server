package domain

import (
	"context"
	"errors"
	"time"
)

// ErrTicketNotFound defines ticket validation miss behavior.
var ErrTicketNotFound = errors.New("sso ticket not found")

// IssueRequest defines one ticket issuance input.
type IssueRequest struct {
	// UserID defines the user identifier bound to the issued ticket.
	UserID int
	// TTL defines requested lifetime; zero uses configured default.
	TTL time.Duration
}

// IssueResult defines one issued ticket payload.
type IssueResult struct {
	// Ticket stores generated single-use authentication token.
	Ticket string
	// ExpiresAt stores token expiration timestamp.
	ExpiresAt time.Time
	// TTL stores the accepted lifetime used for storage.
	TTL time.Duration
}

// Store defines ticket persistence and single-use validation behavior.
type Store interface {
	// Store persists one ticket mapped to one user ID for a bounded lifetime.
	Store(context.Context, string, int, time.Duration) error
	// Validate consumes one ticket and returns the associated user ID.
	Validate(context.Context, string) (int, error)
}
