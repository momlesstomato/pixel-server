package domain

// Group defines one permission group aggregate.
type Group struct {
	// ID stores stable group identifier.
	ID int
	// Name stores unique machine-friendly group name.
	Name string
	// DisplayName stores human-friendly group label.
	DisplayName string
	// Priority stores group priority for effective-group resolution.
	Priority int
	// ClubLevel stores protocol club level attribute.
	ClubLevel int
	// SecurityLevel stores protocol security level attribute.
	SecurityLevel int
	// IsAmbassador stores protocol ambassador attribute.
	IsAmbassador bool
	// IsDefault stores default-assignment marker.
	IsDefault bool
}

// Access defines one resolved user access snapshot.
type Access struct {
	// UserID stores resolved user identifier.
	UserID int
	// PrimaryGroup stores highest-priority resolved group.
	PrimaryGroup Group
	// GroupIDs stores all assigned group identifiers.
	GroupIDs []int
	// Permissions stores resolved permission grants.
	Permissions map[string]struct{}
}

// PerkGrant defines one perk-resolution output.
type PerkGrant struct {
	// Code stores client perk code.
	Code string
	// ErrorMessage stores denied reason shown to client.
	ErrorMessage string
	// IsAllowed stores final allowance marker.
	IsAllowed bool
}
