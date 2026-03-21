package realtime

import (
	"context"

	packet "github.com/momlesstomato/pixel-server/pkg/catalog/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated catalog packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.GetIndexPacketID:
		return true, runtime.handleGetIndex(ctx, connID, userID)
	case packet.GetPagePacketID:
		return true, runtime.handleGetPage(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleGetIndex responds with catalog page index.
func (runtime *Runtime) handleGetIndex(ctx context.Context, connID string, userID int) error {
	pages, err := runtime.service.ListPages(ctx)
	if err != nil {
		runtime.logger.Error("get catalog index failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = pages
	return nil
}

// handleGetPage responds with catalog page details.
func (runtime *Runtime) handleGetPage(ctx context.Context, connID string, userID int, body []byte) error {
	_ = body
	return nil
}
