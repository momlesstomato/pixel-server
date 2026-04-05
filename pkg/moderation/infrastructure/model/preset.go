package model

import "time"

// ModerationPreset defines the GORM model for moderation_presets table.
type ModerationPreset struct {
	// ID stores the primary key.
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// Category stores the preset category name.
	Category string `gorm:"column:category;type:varchar(50);not null"`
	// Name stores the preset display name.
	Name string `gorm:"column:name;type:varchar(100);not null"`
	// ActionType stores the default action type.
	ActionType string `gorm:"column:action_type;type:varchar(20);not null"`
	// DefaultDuration stores the default duration in minutes.
	DefaultDuration int `gorm:"column:default_duration"`
	// DefaultReason stores the default reason text.
	DefaultReason string `gorm:"column:default_reason;type:text;not null;default:''"`
	// Active indicates whether the preset is enabled.
	Active bool `gorm:"column:active;not null;default:true"`
	// CreatedAt stores the creation timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName returns the database table name.
func (ModerationPreset) TableName() string {
	return "moderation_presets"
}
