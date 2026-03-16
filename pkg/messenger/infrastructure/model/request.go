package model

import "time"

// Request stores one pending friend request row in PostgreSQL.
type Request struct {
	// ID stores the auto-incremented request identifier.
	ID int `gorm:"primaryKey;autoIncrement"`
	// FromUserID stores the requesting user identifier.
	FromUserID int `gorm:"not null;uniqueIndex:idx_request_pair,priority:1"`
	// ToUserID stores the target user identifier.
	ToUserID int `gorm:"not null;uniqueIndex:idx_request_pair,priority:2;index"`
	// CreatedAt stores the request creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Request.
func (Request) TableName() string { return "friend_requests" }
