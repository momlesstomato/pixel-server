package domain

import (
	"errors"
	"time"
)

// ErrInvalidUserID indicates ticket operations require a valid user id.
var ErrInvalidUserID = errors.New("invalid user id")

// ErrInvalidTicket indicates ticket value is invalid.
var ErrInvalidTicket = errors.New("invalid ticket")

// ErrTicketNotFound indicates ticket does not exist or is expired.
var ErrTicketNotFound = errors.New("ticket not found")

// TicketStore defines SSO ticket creation and validation behavior.
type TicketStore interface {
	// Create generates one ticket for a user id with an expiration TTL.
	Create(userID int32, ttl time.Duration) (string, error)
	// Consume validates and consumes one ticket, returning user id.
	Consume(ticket string) (int32, error)
	// Revoke removes one ticket if it exists.
	Revoke(ticket string) error
}
