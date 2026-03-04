package user

// Permission defines a named capability.
// Weight is purely informational — it does not auto-grant permissions.
// Higher weight conventionally indicates a more privileged capability and is
// used for ordering in admin UIs. The name is the canonical key used in
// RoleProfile.Perks and in Role.Permissions slices.
type Permission struct {
	// Name is the unique machine-readable identifier (e.g. "mod_tools", "club_vip").
	Name string

	// Description is a human-readable explanation of what the permission allows.
	Description string

	// Weight controls display ordering (higher = more privileged).
	Weight int32
}

// Role is a named collection of permissions that can be assigned to users.
// The role's Weight drives the effective SecurityLevel in the wire protocol:
// when a user has multiple roles, SecurityLevel equals the maximum Weight
// across all assigned roles, divided by the SecurityDivisor constant.
//
// Conventional weight scale:
//
//	Guest    =   0
//	User     = 100
//	HC       = 200
//	VIP      = 300
//	Helper   = 400
//	Moderator= 500
//	Manager  = 700
//	Admin    = 1000
type Role struct {
	// ID is the unique database identifier.
	ID int32

	// Name is the human-readable role label shown in admin UIs (e.g. "Moderator").
	Name string

	// Description describes the purpose of this role.
	Description string

	// Weight defines the role's authority level. See the conventional scale above.
	Weight int32

	// Badge is an optional visual badge code rendered next to the username.
	Badge string

	// Permissions is the explicit list of permission Names granted to users with
	// this role. Perms from multiple roles are unioned at login time.
	Permissions []string
}

// UserRole records the assignment of a Role to a User.
type UserRole struct {
	// UserID is the user this assignment belongs to.
	UserID int32

	// RoleID is the assigned role.
	RoleID int32
}

// SecurityDivisor scales the maximum role Weight to produce a SecurityLevel
// value in the range the wire protocol expects (typically 1–7).
// A role with Weight=700 produces SecurityLevel=7; Weight=1000 → capped at 7.
const SecurityDivisor int32 = 100

// SecurityLevelFromWeight converts a role weight to a protocol security level.
// The result is clamped to [1, 7].
func SecurityLevelFromWeight(weight int32) int32 {
	level := weight / SecurityDivisor
	if level < 1 {
		level = 1
	}
	if level > 7 {
		level = 7
	}
	return level
}
