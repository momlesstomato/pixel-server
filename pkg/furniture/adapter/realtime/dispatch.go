package realtime

import (
	"context"

	packet "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// Handle dispatches one authenticated furniture packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	_, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.PlacePacketID:
		return true, runtime.handlePlace(ctx, connID, body)
	case packet.PickupPacketID:
		return true, runtime.handlePickup(ctx, connID, body)
	case packet.ToggleMultistatePacketID:
		return true, runtime.handleToggleMultistate(ctx, connID, body)
	default:
		return false, nil
	}
}

// handlePlace processes a furniture placement request.
func (runtime *Runtime) handlePlace(ctx context.Context, connID string, body []byte) error {
	_ = body
	return nil
}

// handlePickup processes a furniture pickup request.
func (runtime *Runtime) handlePickup(ctx context.Context, connID string, body []byte) error {
	_ = body
	return nil
}

// handleToggleMultistate processes a furniture state toggle.
func (runtime *Runtime) handleToggleMultistate(ctx context.Context, connID string, body []byte) error {
	_ = body
	return nil
}
