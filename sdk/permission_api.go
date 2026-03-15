package sdk

// GroupInfo defines plugin-facing group attributes.
type GroupInfo struct {
	// ID stores stable group identifier.
	ID int
	// Name stores unique group name.
	Name string
	// ClubLevel stores protocol club-level attribute.
	ClubLevel int
	// SecurityLevel stores protocol security-level attribute.
	SecurityLevel int
	// IsAmbassador stores protocol ambassador attribute.
	IsAmbassador bool
}

// PermissionAPI provides plugin-facing permission checks.
type PermissionAPI interface {
	// HasPermission reports whether one user has a permission.
	HasPermission(userID int, permission string) bool
	// GetGroup resolves the effective group for one user.
	GetGroup(userID int) (GroupInfo, bool)
}
