package httpapi

import (
	"context"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// SessionLister defines session query behavior for management routes.
type SessionLister interface {
	// FindByConnID retrieves one session by connection identifier.
	FindByConnID(string) (coreconnection.Session, bool)
	// ListAll returns all sessions currently stored in the registry.
	ListAll() ([]coreconnection.Session, error)
	// Remove deletes one session by connection identifier.
	Remove(string)
}

// HotelManager defines hotel status management behavior.
type HotelManager interface {
	// Current returns active hotel status snapshot.
	Current(context.Context) (statusdomain.HotelStatus, error)
	// ScheduleClose transitions hotel into closing state.
	ScheduleClose(context.Context, int32, int32, bool) (statusdomain.HotelStatus, error)
	// Reopen transitions hotel into open state.
	Reopen(context.Context) (statusdomain.HotelStatus, error)
}

// SessionCloser defines cross-instance session close signal behavior.
type SessionCloser interface {
	// Close publishes a close signal for one connection identifier.
	Close(ctx context.Context, connID string, code int, reason string) error
}
