package model

import "time"

// ModerationRoomVisit defines the GORM model for moderation_room_visits table.
type ModerationRoomVisit struct {
	// ID stores the primary key.
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// UserID stores the visiting user.
	UserID int `gorm:"column:user_id;not null"`
	// RoomID stores the visited room.
	RoomID int `gorm:"column:room_id;not null"`
	// VisitedAt stores the visit timestamp.
	VisitedAt time.Time `gorm:"column:visited_at;autoCreateTime"`
}

// TableName returns the database table name.
func (ModerationRoomVisit) TableName() string {
	return "moderation_room_visits"
}
