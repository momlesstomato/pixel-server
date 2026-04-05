package model

import "time"

// ModerationWordFilter defines the GORM model for moderation_word_filters table.
type ModerationWordFilter struct {
	// ID stores the primary key.
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// Pattern stores the text pattern to match.
	Pattern string `gorm:"column:pattern;type:varchar(255);not null"`
	// Replacement stores the substitution text.
	Replacement string `gorm:"column:replacement;type:varchar(255);not null;default:'***'"`
	// IsRegex indicates whether the pattern uses regex.
	IsRegex bool `gorm:"column:is_regex;not null;default:false"`
	// Scope stores "global" or "room".
	Scope string `gorm:"column:scope;type:varchar(10);not null;default:'global'"`
	// RoomID stores the room identifier when scope is room.
	RoomID int `gorm:"column:room_id"`
	// Active indicates whether the filter is enabled.
	Active bool `gorm:"column:active;not null;default:true"`
	// CreatedAt stores the creation timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName returns the database table name.
func (ModerationWordFilter) TableName() string {
	return "moderation_word_filters"
}
