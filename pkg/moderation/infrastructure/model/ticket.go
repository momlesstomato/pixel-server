package model

import "time"

// ModerationTicket defines the GORM model for moderation_tickets table.
type ModerationTicket struct {
	// ID stores the primary key.
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// ReporterID stores the user who submitted the ticket.
	ReporterID int `gorm:"column:reporter_id;not null"`
	// ReportedID stores the user being reported.
	ReportedID int `gorm:"column:reported_id"`
	// RoomID stores the room where the incident occurred.
	RoomID int `gorm:"column:room_id"`
	// Category stores the ticket category.
	Category string `gorm:"column:category;type:varchar(50);not null"`
	// Message stores the reporter description.
	Message string `gorm:"column:message;type:text;not null;default:''"`
	// Status stores the ticket lifecycle state.
	Status string `gorm:"column:status;type:varchar(20);not null;default:'open'"`
	// AssignedTo stores the moderator handling the ticket.
	AssignedTo int `gorm:"column:assigned_to"`
	// CreatedAt stores the submission timestamp.
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt stores the last modification timestamp.
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// ClosedAt stores the resolution timestamp.
	ClosedAt *time.Time `gorm:"column:closed_at"`
}

// TableName returns the database table name.
func (ModerationTicket) TableName() string {
	return "moderation_tickets"
}
