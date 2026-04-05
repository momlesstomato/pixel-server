package model

import "time"

// ChatLog stores one room chat message entry in PostgreSQL.
type ChatLog struct {
	// ID stores the stable chat log identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// RoomID stores the room where the message was sent.
	RoomID uint `gorm:"not null;index:idx_chat_logs_room_created,priority:1"`
	// UserID stores the sender user identifier.
	UserID uint `gorm:"not null"`
	// Username stores the sender display name at message time.
	Username string `gorm:"size:50;not null"`
	// Message stores the chat text payload.
	Message string `gorm:"size:512;not null"`
	// ChatType stores the message kind (talk, shout, whisper).
	ChatType string `gorm:"size:10;not null;default:talk"`
	// CreatedAt stores the message timestamp.
	CreatedAt time.Time `gorm:"not null;index:idx_chat_logs_created;index:idx_chat_logs_room_created,priority:2"`
}

// TableName returns the PostgreSQL table name for ChatLog.
func (ChatLog) TableName() string { return "room_chat_logs" }
