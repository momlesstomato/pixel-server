package model

import "time"

// Subscription stores one user subscription row in PostgreSQL.
type Subscription struct {
	// ID stores stable subscription identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the subscriber identifier.
	UserID uint `gorm:"not null;index"`
	// SubscriptionType stores the membership tier key.
	SubscriptionType string `gorm:"size:50;not null;default:habbo_club"`
	// StartedAt stores the subscription start timestamp.
	StartedAt time.Time `gorm:"not null"`
	// DurationDays stores total subscription length in days.
	DurationDays int `gorm:"not null"`
	// Active stores whether the subscription is currently active.
	Active bool `gorm:"not null;default:true"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Subscription.
func (Subscription) TableName() string { return "user_subscriptions" }

// ClubOffer stores one club membership offer row in PostgreSQL.
type ClubOffer struct {
	// ID stores stable club offer identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Name stores the offer display name.
	Name string `gorm:"size:100;not null"`
	// Days stores the membership duration in days.
	Days int `gorm:"not null"`
	// Credits stores the credit price.
	Credits int `gorm:"not null;default:0"`
	// Points stores the activity-point price.
	Points int `gorm:"not null;default:0"`
	// PointsType stores the activity-point currency type.
	PointsType int `gorm:"not null;default:0"`
	// OfferType stores the membership tier key.
	OfferType string `gorm:"size:10;not null;default:HC"`
	// Giftable stores whether the offer can be gifted.
	Giftable bool `gorm:"not null;default:false"`
	// Enabled stores whether the offer is currently purchasable.
	Enabled bool `gorm:"not null;default:true"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for ClubOffer.
func (ClubOffer) TableName() string { return "catalog_club_offers" }
