package model

import "time"

// Assignment defines one persisted user_permission_groups row.
type Assignment struct {
	// UserID stores assigned user identifier.
	UserID uint `gorm:"primaryKey;not null;index"`
	// GroupID stores assigned group identifier.
	GroupID uint `gorm:"primaryKey;not null;index"`
	// CreatedAt stores assignment creation timestamp.
	CreatedAt time.Time
}
