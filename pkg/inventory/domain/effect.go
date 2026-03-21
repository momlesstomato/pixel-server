package domain

import "time"

// Effect defines one user-owned avatar effect entry.
type Effect struct {
	// ID stores stable effect row identifier.
	ID int
	// UserID stores the effect owner identifier.
	UserID int
	// EffectID stores the avatar effect type identifier.
	EffectID int
	// Duration stores total duration in seconds.
	Duration int
	// Quantity stores remaining activations.
	Quantity int
	// ActivatedAt stores first activation timestamp, nil when inactive.
	ActivatedAt *time.Time
	// IsPermanent stores whether the effect never expires.
	IsPermanent bool
	// CreatedAt stores effect award timestamp.
	CreatedAt time.Time
}

// IsExpired reports whether a finite-duration effect has elapsed.
func (e Effect) IsExpired() bool {
	if e.IsPermanent || e.ActivatedAt == nil {
		return false
	}
	return time.Since(*e.ActivatedAt).Seconds() > float64(e.Duration)
}

// ExpiredEffect pairs one expired effect with its owner for cleanup.
type ExpiredEffect struct {
	// UserID stores the effect owner identifier.
	UserID int
	// EffectID stores the expired effect type identifier.
	EffectID int
}
