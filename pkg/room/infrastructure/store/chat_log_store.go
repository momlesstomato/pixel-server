package store

import (
	"context"
	"errors"
	"time"

	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"gorm.io/gorm"
)

// ChatLogStore implements domain.ChatLogRepository using PostgreSQL.
type ChatLogStore struct {
	// db stores the database connection.
	db *gorm.DB
}

// NewChatLogStore creates one chat log store instance.
func NewChatLogStore(db *gorm.DB) (*ChatLogStore, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &ChatLogStore{db: db}, nil
}

// Append persists one chat log entry.
func (s *ChatLogStore) Append(ctx context.Context, entry domain.ChatLogEntry) error {
	rec := roommodel.ChatLog{
		RoomID:    uint(entry.RoomID),
		UserID:    uint(entry.UserID),
		Username:  entry.Username,
		Message:   entry.Message,
		ChatType:  entry.ChatType,
		CreatedAt: entry.CreatedAt,
	}
	return s.db.WithContext(ctx).Create(&rec).Error
}

// ListByRoom returns chat entries for one room filtered by time range.
func (s *ChatLogStore) ListByRoom(ctx context.Context, roomID int, from time.Time, to time.Time) ([]domain.ChatLogEntry, error) {
	var records []roommodel.ChatLog
	err := s.db.WithContext(ctx).
		Where("room_id = ? AND created_at >= ? AND created_at <= ?", uint(roomID), from, to).
		Order("created_at ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	entries := make([]domain.ChatLogEntry, len(records))
	for i, rec := range records {
		entries[i] = domain.ChatLogEntry{
			RoomID: int(rec.RoomID), UserID: int(rec.UserID),
			Username: rec.Username, Message: rec.Message,
			ChatType: rec.ChatType, CreatedAt: rec.CreatedAt,
		}
	}
	return entries, nil
}
