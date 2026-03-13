package domain

import "time"

// State defines hotel lifecycle state.
type State string

const (
	// StateOpen defines normal open operation state.
	StateOpen State = "open"
	// StateClosing defines scheduled closing countdown state.
	StateClosing State = "closing"
	// StateClosed defines maintenance/closed state.
	StateClosed State = "closed"
)

// HotelStatus defines one persisted hotel lifecycle snapshot.
type HotelStatus struct {
	// State stores current hotel lifecycle state.
	State State `json:"state"`
	// CloseAt stores scheduled close timestamp in UTC when state is closing.
	CloseAt *time.Time `json:"close_at,omitempty"`
	// ReopenAt stores scheduled reopen timestamp in UTC when state is closed.
	ReopenAt *time.Time `json:"reopen_at,omitempty"`
	// UserThrownOutAtClose stores whether connected users are removed at close time.
	UserThrownOutAtClose bool `json:"user_thrown_out_at_close"`
}

// IsOpen reports whether current status allows hotel access.
func (status HotelStatus) IsOpen() bool {
	return status.State == StateOpen || status.State == StateClosing
}

// OnShutdown reports whether hotel is in shutdown countdown state.
func (status HotelStatus) OnShutdown() bool {
	return status.State == StateClosing
}
