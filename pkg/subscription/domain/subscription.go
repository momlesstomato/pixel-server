package domain

import "time"

// SubscriptionType identifies the membership tier.
type SubscriptionType string

const (
	// SubscriptionHabboClub identifies standard HC membership.
	SubscriptionHabboClub SubscriptionType = "habbo_club"
	// SubscriptionBuildersClub identifies builders club membership.
	SubscriptionBuildersClub SubscriptionType = "builders_club"
)

// Subscription defines one active user membership entry.
type Subscription struct {
	// ID stores stable subscription identifier.
	ID int
	// UserID stores the subscriber identifier.
	UserID int
	// SubscriptionType stores the membership tier key.
	SubscriptionType SubscriptionType
	// StartedAt stores the subscription start timestamp.
	StartedAt time.Time
	// DurationDays stores total subscription length in days.
	DurationDays int
	// Active stores whether the subscription is currently active.
	Active bool
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}
