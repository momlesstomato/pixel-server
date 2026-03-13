package model

import "time"

// RespectTargetTypeUser identifies user-to-user respect records.
const RespectTargetTypeUser int16 = 0

// RespectTargetTypePet identifies user-to-pet respect records.
const RespectTargetTypePet int16 = 1

// Respect stores one respect audit event row.
type Respect struct {
	// ID stores stable respect row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// ActorUserID stores respecting user identifier.
	ActorUserID uint `gorm:"not null;index:idx_user_respects_actor_date_type,priority:1"`
	// TargetID stores target user or pet identifier.
	TargetID uint `gorm:"not null"`
	// TargetType stores respect target type marker.
	TargetType int16 `gorm:"not null;default:0;index:idx_user_respects_actor_date_type,priority:3"`
	// RespectedAt stores the UTC day when respect happened.
	RespectedAt time.Time `gorm:"type:date;not null;index:idx_user_respects_actor_date_type,priority:2"`
}
