package model

import "time"

// RoomBan stores one room access ban entry in PostgreSQL.
type RoomBan struct {
	// ID stores the stable ban identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// RoomID stores the room reference.
	RoomID uint `gorm:"not null;index:idx_room_bans_room"`
	// UserID stores the banned user identifier.
	UserID uint `gorm:"not null;index:idx_room_bans_user"`
	// ExpiresAt stores the ban expiry timestamp (NULL = permanent).
	ExpiresAt *time.Time
	// CreatedAt stores the ban creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for RoomBan.
func (RoomBan) TableName() string { return "room_bans" }

// RoomRight stores one room rights grant entry in PostgreSQL.
type RoomRight struct {
	// ID stores the stable rights identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// RoomID stores the room reference.
	RoomID uint `gorm:"not null;uniqueIndex:idx_room_rights_pair,priority:1"`
	// UserID stores the rights holder identifier.
	UserID uint `gorm:"not null;uniqueIndex:idx_room_rights_pair,priority:2"`
}

// TableName returns the PostgreSQL table name for RoomRight.
func (RoomRight) TableName() string { return "room_rights" }
