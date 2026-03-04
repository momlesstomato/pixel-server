# User Profile — Data Models

All domain structs, in-memory implementations, and repository interfaces used
by the user-profile realm.

---

## User (`pkg/user.User`)

The core user record. Persisted in PostgreSQL (in-memory during development).

| Field | Type | Description |
|---|---|---|
| `ID` | `int32` | Unique user identifier (database primary key) |
| `Username` | `string` | Display name, shown to all players |
| `Figure` | `string` | Avatar figure string (Nitro format, e.g. `"hr-115-42.hd-195-19"`) |
| `Gender` | `string` | `"M"` (male) or `"F"` (female) |
| `Motto` | `string` | Freeform profile motto text |
| `Credits` | `int32` | Currency balance displayed in top bar |
| `Rank` | `int32` | Staff rank level; drives `SecurityLevel` in permissions |
| `HomeRoom` | `int32` | Default room ID to enter at login (0 = no home room) |
| `RespectPoints` | `int32` | Respect count received from other players |
| `AllowNameChange` | `bool` | Whether `change_username` (2977) is permitted |
| `SafetyLocked` | `bool` | Whether account is safety locked |
| `AccountCreated` | `time.Time` | Account creation timestamp; drives `noobness_level` |

---

## Settings (`pkg/user.Settings`)

Per-user UI and preference state. Persisted separately from the user record.

| Field | Type | Default | Description |
|---|---|---|---|
| `UserID` | `int32` | — | Foreign key to `User.ID` |
| `VolumeSystem` | `int32` | `0` | System audio volume level |
| `VolumeFurni` | `int32` | `0` | Furniture sound effect volume |
| `VolumeTrax` | `int32` | `0` | Trax music volume |
| `OldChat` | `bool` | `false` | Whether to use classic chat bubble style |
| `IgnoreRoomInvites` | `bool` | `false` | Whether to block room invite notifications |
| `BlockCameraFollow` | `bool` | `false` | Disable camera auto-follow |
| `FriendBarOpen` | `bool` | `false` | Friend bar expanded state |
| `UIFlags` | `int32` | `0` | Packed bitmask of miscellaneous UI state flags |

---

## Badge (`pkg/user.Badge`)

An awarded achievement badge on the user's profile.

| Field | Type | Description |
|---|---|---|
| `UserID` | `int32` | Owner user ID |
| `Code` | `string` | Badge code (e.g. `"EASTER23"`) |
| `Slot` | `int32` | Display slot 1–5 (0 = not displayed) |

---

## WardrobeOutfit (`pkg/user.WardrobeOutfit`)

A saved avatar configuration in a wardrobe slot.

| Field | Type | Description |
|---|---|---|
| `UserID` | `int32` | Owner user ID |
| `SlotID` | `int32` | Wardrobe slot index, 1–10 (1-indexed) |
| `Figure` | `string` | Figure string |
| `Gender` | `string` | `"M"` or `"F"` |

A user can have a maximum of 10 saved outfits. Slots are overwritten on save.

---

## Permission (`pkg/user.Permission`)

A named capability that can be granted to a role.

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Unique machine-readable identifier (e.g. `"club_hc"`, `"ambassador"`) |
| `Description` | `string` | Human-readable description |
| `Weight` | `int32` | Minimum role weight required to grant this permission (informational) |

---

## Role (`pkg/user.Role`)

A named group of permissions. Users are assigned one or more roles. Roles carry
a numeric weight that drives security level derivation.

| Field | Type | Description |
|---|---|---|
| `ID` | `int32` | Unique role identifier |
| `Name` | `string` | Display name (e.g. `"Moderator"`) |
| `Description` | `string` | Human-readable description |
| `Weight` | `int32` | Numeric authority level (see table below) |
| `Badge` | `string` | Optional badge code awarded when the role is assigned |
| `Permissions` | `[]string` | Slice of permission names included in this role |

### Predefined weight tiers

| Weight | Suggested role |
|---|---|
| 0 | Guest |
| 100 | User |
| 200 | Hotel Club (HC) |
| 300 | VIP |
| 400 | Helper |
| 500 | Moderator |
| 700 | Manager |
| 1000 | Administrator |

### `SecurityDivisor`

```go
const SecurityDivisor = 100
```

Divides `MaxRoleWeight` to produce a `SecurityLevel` integer between 1 and 7.

### `SecurityLevelFromWeight(weight int32) int32`

```go
func SecurityLevelFromWeight(weight int32) int32
```

Returns `max(1, min(7, weight / SecurityDivisor))`. Used by `BuildRoleProfile` to
derive the client-facing security level from the highest assigned role weight.

---

## UserRole (`pkg/user.UserRole`)

Links a user to a role. Stored separately from core user data.

| Field | Type | Description |
|---|---|---|
| `UserID` | `int32` | ID of the user |
| `RoleID` | `int32` | ID of the assigned role |

---

## Repository Interfaces (`pkg/user`)

These interfaces are defined in the domain package. All implementations are
injected at startup — the identity logic never imports concrete packages.

### `user.Repository`

```go
GetByID(ctx, id int32) (*User, error)
GetByUsername(ctx, username string) (*User, error)
Create(ctx, u *User) error
Update(ctx, u *User) error
GetSettings(ctx, userID int32) (*Settings, error)
UpdateSettings(ctx, s *Settings) error
```

Sentinel errors:
- `user.ErrNotFound` — no user row matches the query.
- `user.ErrRoleNotFound` — role ID does not exist.
- `user.ErrPermissionNotFound` — permission name does not exist.

### `user.BadgeRepository`

```go
GetBadges(ctx, userID int32) ([]*Badge, error)
SetBadge(ctx, b *Badge) error
RemoveBadge(ctx, userID int32, code string) error
```

### `user.WardrobeRepository`

```go
GetOutfits(ctx, userID int32) ([]*WardrobeOutfit, error)
SaveOutfit(ctx, o *WardrobeOutfit) error
```

`SaveOutfit` upserts by `(UserID, SlotID)`.

### `user.IgnoreRepository`

```go
GetIgnored(ctx, userID int32) ([]string, error)  // returns usernames
AddIgnore(ctx, userID, targetID int32) error
RemoveIgnore(ctx, userID int32, username string) error
```

### `user.RoleRepository`

Full CRUD for roles plus user→role assignment management.

```go
GetByID(ctx, id int32) (*Role, error)
GetAll(ctx) ([]*Role, error)
GetForUser(ctx, userID int32) ([]*Role, error)
Create(ctx, r *Role) error
Update(ctx, r *Role) error
Delete(ctx, id int32) error
AssignRole(ctx, userID, roleID int32) error
RevokeRole(ctx, userID, roleID int32) error
HasRole(ctx, userID, roleID int32) (bool, error)
```

`AssignRole` is idempotent — assigning an already-held role is a no-op.
`GetForUser` returns nil (not an error) when the user has no assigned roles.

### `user.PermissionRepository`

```go
GetAll(ctx) ([]*Permission, error)
GetByName(ctx, name string) (*Permission, error)
Create(ctx, p *Permission) error
Delete(ctx, name string) error
```

---

## In-Memory Implementations (`pkg/user/memory`)

Used in unit tests and local development. All four repositories are backed by
`sync.RWMutex`-protected Go maps. There is no persistence across restarts.

| Type | Backing store |
|---|---|
| `memory.UserRepo` | `map[int32]*User` + `map[string]int32` (username index) |
| `memory.BadgeRepo` | `map[int32][]*Badge` |
| `memory.WardrobeRepo` | `map[int32][]*WardrobeOutfit` |
| `memory.IgnoreRepo` | `map[int32][]int32` (ignored user IDs) |
| `memory.RoleRepo` | `map[int32]*Role` (roles) + `map[int32][]int32` (user→roleIDs) |
| `memory.PermissionRepo` | `map[string]*Permission` |

All maps are protected by `sync.RWMutex`. `RoleRepo` uses `atomic.Int32` for
auto-incrementing IDs. `copyRole()` deep-copies `Permissions` slices to prevent
external mutation.

Constructors: `memory.NewUserRepo()`, `memory.NewBadgeRepo()`,
`memory.NewWardrobeRepo()`, `memory.NewIgnoreRepo()`, `memory.NewRoleRepo()`,
`memory.NewPermissionRepo()`.

---

## RoleProfile (`services/game/internal/identity`)

Derived at login time. Never persisted — rebuilt on every login from user rank
and perk data.

| Field | Type | Description |
|---|---|---|
| Field | Type | Description |
|---|---|---|
| `ClubLevel` | `int32` | 0 = none, 1 = HC, 2 = VIP |
| `SecurityLevel` | `int32` | Staff access level, 1–7 |
| `IsAmbassador` | `bool` | Whether the user is a community ambassador |
| `Perks` | `map[string]bool` | Union of all permission names from assigned roles; O(1) lookup |

Built by `BuildRoleProfile(rank int32, roles []*user.Role)` in
`services/game/internal/identity`.

- `ResolvePermissions(roles)` unions all `Role.Permissions` slices into the
  `Perks` map.
- `MaxRoleWeight(roles)` returns the highest `Role.Weight` across all assigned
  roles.
- `SecurityLevel` is derived via `user.SecurityLevelFromWeight(maxWeight)`,
  falling back to `max(1, rank)` when no roles are assigned.
- `ClubLevel` is determined by presence of `"club_vip"` (→ 2) or `"club_hc"`
  (→ 1) in the resolved perks.
- `IsAmbassador` is set when `"ambassador"` is present in resolved perks.

`RoleProfile` is cached in the session for the duration of the connection and
never persisted to the database.
