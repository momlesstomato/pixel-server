package model

import "time"

// MessageLog stores one message log row in PostgreSQL for auditing and security.
type MessageLog struct {
	// ID stores the auto-incremented record identifier.
	ID int `gorm:"primaryKey;autoIncrement"`
	// FromUserID stores the sender user identifier.
	FromUserID int `gorm:"not null;index"`
	// ToUserID stores the recipient user identifier.
	ToUserID int `gorm:"not null;index"`
	// Message stores the message content up to 255 characters.
	Message string `gorm:"not null;type:varchar(255)"`
	// SentAt stores the original send timestamp.
	SentAt time.Time `gorm:"not null;index"`
}

// TableName returns the PostgreSQL table name for MessageLog.
func (MessageLog) TableName() string { return "messenger_message_log" }
