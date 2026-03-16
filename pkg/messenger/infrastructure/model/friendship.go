package model

import "time"

// Friendship stores one canonical friendship row in PostgreSQL.
type Friendship struct {
	// UserOneID stores the lower user identifier of the pair.
	UserOneID int `gorm:"primaryKey;not null"`
	// UserTwoID stores the higher user identifier of the pair.
	UserTwoID int `gorm:"primaryKey;not null"`
	// Relationship stores the relationship type from UserOneID perspective.
	Relationship int16 `gorm:"not null;default:0"`
	// CreatedAt stores the friendship creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Friendship.
func (Friendship) TableName() string { return "messenger_friendships" }
