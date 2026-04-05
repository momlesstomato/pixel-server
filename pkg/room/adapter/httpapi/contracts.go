package httpapi

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// ChatLogService defines chat log operations required by the room HTTP adapter.
type ChatLogService interface {
	// ListByRoom returns chat entries for one room filtered by time range.
	ListByRoom(ctx context.Context, roomID int, from time.Time, to time.Time) ([]domain.ChatLogEntry, error)
}
