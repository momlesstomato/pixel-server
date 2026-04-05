package model

import "time"

// RoomVote stores one room score vote entry in PostgreSQL.
type RoomVote struct {
	// ID stores the stable vote identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// RoomID stores the voted room reference.
	RoomID uint `gorm:"not null;uniqueIndex:idx_room_votes_pair,priority:1"`
	// UserID stores the voting user identifier.
	UserID uint `gorm:"not null;uniqueIndex:idx_room_votes_pair,priority:2"`
	// CreatedAt stores the vote timestamp.
	CreatedAt time.Time `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for RoomVote.
func (RoomVote) TableName() string { return "room_votes" }
