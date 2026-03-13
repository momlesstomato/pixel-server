package httpapi

import (
	"time"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
)

// sessionResponse defines session API response payload.
type sessionResponse struct {
	// ConnID defines connection identifier.
	ConnID string `json:"conn_id"`
	// UserID defines authenticated user identifier.
	UserID int `json:"user_id"`
	// MachineID defines machine fingerprint.
	MachineID string `json:"machine_id,omitempty"`
	// State defines session lifecycle state name.
	State string `json:"state"`
	// InstanceID defines owning server instance identifier.
	InstanceID string `json:"instance_id"`
	// CreatedAt defines session creation timestamp.
	CreatedAt string `json:"created_at"`
}

// mapSession converts a domain session to an API response.
func mapSession(s coreconnection.Session) sessionResponse {
	state := "connected"
	switch s.State {
	case coreconnection.StateAuthenticated:
		state = "authenticated"
	case coreconnection.StateDisconnecting:
		state = "disconnecting"
	}
	return sessionResponse{
		ConnID: s.ConnID, UserID: s.UserID, MachineID: s.MachineID,
		State: state, InstanceID: s.InstanceID,
		CreatedAt: s.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// mapHotelStatus converts a domain hotel status to an API response.
func mapHotelStatus(s statusdomain.HotelStatus) hotelStatusResponse {
	resp := hotelStatusResponse{State: string(s.State), ThrowUsers: s.UserThrownOutAtClose}
	if s.CloseAt != nil {
		t := s.CloseAt.UTC().Format(time.RFC3339)
		resp.CloseAt = &t
	}
	if s.ReopenAt != nil {
		t := s.ReopenAt.UTC().Format(time.RFC3339)
		resp.ReopenAt = &t
	}
	return resp
}
