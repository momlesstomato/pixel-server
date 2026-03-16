package model

import "time"

// OfflineMessage stores one offline message row in PostgreSQL.
type OfflineMessage struct {
	// ID stores the auto-incremented message identifier.
	ID int `gorm:"primaryKey;autoIncrement"`
	// FromUserID stores the sender user identifier.
	FromUserID int `gorm:"not null"`
	// ToUserID stores the recipient user identifier.
	ToUserID int `gorm:"not null;index"`
	// Message stores the message content up to 255 characters.
	Message string `gorm:"not null;type:varchar(255)"`
	// SentAt stores the original send timestamp.
	SentAt time.Time `gorm:"not null"`
}

// TableName returns the PostgreSQL table name for OfflineMessage.
func (OfflineMessage) TableName() string { return "offline_messages" }
