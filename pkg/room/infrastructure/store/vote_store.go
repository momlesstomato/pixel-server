package store

import (
	"context"
	"errors"
	"time"

	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"

	"gorm.io/gorm"
)

// VoteStore implements domain.VoteRepository using PostgreSQL.
type VoteStore struct {
	// db stores the database connection.
	db *gorm.DB
}

// NewVoteStore creates one vote store instance.
func NewVoteStore(db *gorm.DB) (*VoteStore, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &VoteStore{db: db}, nil
}

// HasVoted reports whether one user has voted for one room.
func (s *VoteStore) HasVoted(ctx context.Context, roomID int, userID int) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&roommodel.RoomVote{}).
		Where("room_id = ? AND user_id = ?", uint(roomID), uint(userID)).
		Count(&count).Error
	return count > 0, err
}

// CastVote records one vote and increments the room score atomically.
func (s *VoteStore) CastVote(ctx context.Context, roomID int, userID int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		vote := roommodel.RoomVote{
			RoomID: uint(roomID), UserID: uint(userID), CreatedAt: time.Now(),
		}
		if err := tx.Create(&vote).Error; err != nil {
			return err
		}
		return tx.Table("rooms").Where("id = ?", roomID).
			UpdateColumn("score", gorm.Expr("score + 1")).Error
	})
}
