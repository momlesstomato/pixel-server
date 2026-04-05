package model

import "time"

// ModerationAction defines the GORM model for moderation_actions table.
type ModerationAction struct {
	// ID stores the primary key.
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// Scope stores room or hotel discriminator.
	Scope string `gorm:"column:scope;type:varchar(10);not null"`
	// ActionType stores the kind of action (kick, ban, mute, warn).
	ActionType string `gorm:"column:action_type;type:varchar(20);not null"`
	// TargetUserID stores the moderated user identifier.
	TargetUserID int `gorm:"column:target_user_id;not null"`
	// IssuerID stores the staff member who created the action.
	IssuerID int `gorm:"column:issuer_id;not null"`
	// RoomID stores the room identifier for room-scoped actions.
	RoomID int `gorm:"column:room_id"`
	// Reason stores the human-readable justification.
	Reason string `gorm:"column:reason;type:text;not null;default:''"`
	// DurationMinutes stores the intended duration.
	DurationMinutes int `gorm:"column:duration_minutes"`
	// ExpiresAt stores the computed expiry.
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	// Active stores whether the action is currently in effect.
	Active bool `gorm:"column:active;not null;default:true"`
	// DeactivatedBy stores which staff member lifted the action.
	DeactivatedBy int `gorm:"column:deactivated_by"`
	// DeactivatedAt stores when the action was deactivated.
	DeactivatedAt *time.Time `gorm:"column:deactivated_at"`
	// IPAddress stores IP for hotel IP bans.
	IPAddress string `gorm:"column:ip_address;type:varchar(45)"`
	// MachineID stores machine fingerprint for hotel machine bans.
	MachineID string `gorm:"column:machine_id;type:varchar(64)"`
	// CreatedAt stores the creation timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName returns the database table name.
func (ModerationAction) TableName() string {
	return "moderation_actions"
}
