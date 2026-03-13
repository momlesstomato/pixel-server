# 05 - Permission & Group System

## Overview

The permission system replaces vendor-style numeric `security_level`,
`club_level`, and `is_ambassador` columns with a **group-based model**
using **dotted-notation string permissions**. Each user belongs to exactly
one permission group. Groups define protocol attributes (club level,
security level, ambassador) and grant an extensible set of string
permissions checked at runtime.

Permission strings use dotted notation (`moderation.ban`, `perk.camera`,
`room.enter.locked`) following the Bukkit/Minecraft convention. There is
**no permission definition table** — the string itself is the identifier.
Only grants are stored in the database.

---

## Vendor Cross-Reference

### How Vendors Handle Permissions

| Aspect | Sodium (C#) | Gladiator (Java) | Galaxy (Java) |
|--------|------------|------------------|---------------|
| Storage | `rank` INT on user | `rank` INT on user | `rank` INT on user |
| Resolution | Hardcoded `if (rank >= N)` | `permissions` table per rank | Hardcoded checks |
| Granularity | Coarse (numeric) | Moderate (rank + permissions table) | Coarse (numeric) |
| Wildcards | None | None | None |
| Perks | Hardcoded per rank + club | Config-defined min-rank | Hardcoded per rank |
| Group names | None (numeric only) | None (numeric only) | None (numeric only) |

**All three vendors** store a single integer rank and use numeric
comparisons. None support string permissions, wildcards, or named groups.

### Gladiator Permission Table

Gladiator comes closest to our design with a `permissions` table:

```
permissions
├── id          INT PK
├── rank_id     INT FK
├── permission  VARCHAR(255)
```

But it uses integer rank IDs, not named groups, and does not support
wildcard resolution or dotted hierarchy.

### Our Improvements Over Vendors

1. **Named groups** instead of numeric ranks (human-readable)
2. **Dotted string permissions** instead of integer level checks
3. **Wildcard resolution** (`moderation.*` grants all moderation perms)
4. **Protocol attributes on groups** (club, security, ambassador live on
   the group, not the user)
5. **Plugin-extensible** — plugins define their own permission strings
6. **No magic numbers** — Go constants for all built-in permissions

---

## Design Decisions

### Single Group Per User

Each user belongs to **exactly one** permission group. This avoids the
complexity of multi-group inheritance and permission merging. If a server
needs a "VIP Moderator" role, the admin creates a group with both VIP
club level and moderation permissions. This is explicit and predictable.

**Rationale:** Habbo vendors use single-rank assignment. Multi-group
systems (like Bukkit) add complexity that is unnecessary for the Habbo
use case where role combinations are finite and admin-defined.

### No Permission Definition Table

The `group_permissions` table stores `(group_id, permission_string)`.
There is no separate `permission_definitions` table listing all possible
permissions. Permissions are self-describing strings defined as Go
constants in the owning realm package.

**Rationale:** The user explicitly requested this. A definition table
adds maintenance overhead without functional benefit. The Go constants
serve as documentation and prevent typos at compile time.

### Protocol Attributes on Groups

The `user.permissions` packet (ID 411) requires `clubLevel`,
`securityLevel`, and `isAmbassador`. Rather than storing these on the
user, they are defined on the permission group. When a user's group
changes, their protocol attributes change atomically.

### Dotted Notation with Wildcards

Permissions follow `<realm>.<action>` or `<realm>.<sub>.<action>` format.
Wildcard `*` at any segment grants all children:

- `*` → grants every permission
- `moderation.*` → grants `moderation.ban`, `moderation.kick`, etc.
- `perk.*` → grants all perks

Resolution checks exact match first, then walks up the hierarchy.

---

## Database Model

### Table 1: `permission_groups`

Core group definition. Each group is a named role with protocol
attributes and optional description.

```
permission_groups
├── id              BIGINT PK AUTO
├── name            VARCHAR(64) UNIQUE NOT NULL      -- "admin", "moderator", "vip"
├── display_name    VARCHAR(128) NOT NULL DEFAULT ''  -- "Administrator"
├── priority        INT NOT NULL DEFAULT 0            -- higher = more authority
├── club_level      INT NOT NULL DEFAULT 0            -- 0=none, 1=club, 2=vip
├── security_level  INT NOT NULL DEFAULT 0            -- maps to user.permissions
├── is_ambassador   BOOLEAN NOT NULL DEFAULT false
├── is_default      BOOLEAN NOT NULL DEFAULT false    -- assigned to new users
├── created_at      TIMESTAMP
├── updated_at      TIMESTAMP
```

**Constraints:**
- Exactly one group must have `is_default = true`
- `name` is lowercase alphanumeric + hyphens (validated in application layer)
- `priority` is used for display ordering and admin audit, not permission
  resolution (no inheritance)

### Table 2: `group_permissions`

Permission grants for a group. Each row is one permission string granted
to the group. The permission string IS the primary identifier — no
separate definition table.

```
group_permissions
├── group_id        BIGINT NOT NULL (FK → permission_groups.id ON DELETE CASCADE)
├── permission      VARCHAR(128) NOT NULL
├── PRIMARY KEY(group_id, permission)
├── INDEX(permission)
```

**No foreign key on `permission`** — it is a free-form dotted string.
Validation happens in the application layer using Go constants.

### Entity Relationship

```
permission_groups ──1:N──> group_permissions
users ──N:1──> permission_groups (via group_id FK)
```

---

## Permission Constants

Permissions are defined as Go `const` in the package that owns the
behavior. No magic strings in business logic.

### Core Permissions (`core/permission/constants.go`)

```go
package permission

const (
    Wildcard = "*"
)
```

### Moderation Permissions (`pkg/moderation/domain/permission.go`)

```go
const (
    ModerationWildcard = "moderation.*"
    ModerationBan      = "moderation.ban"
    ModerationKick     = "moderation.kick"
    ModerationMute     = "moderation.mute"
    ModerationAlert    = "moderation.alert"
    ModerationRoomCtrl = "moderation.room"
)
```

### Perk Permissions (`pkg/user/domain/permission.go`)

```go
const (
    PerkWildcard      = "perk.*"
    PerkCamera        = "perk.camera"
    PerkTrade         = "perk.trade"
    PerkGuide         = "perk.guide"
    PerkGuideTours    = "perk.guide.tours"
    PerkChatReviews   = "perk.chat_reviews"
    PerkCompetitions  = "perk.competitions"
    PerkHelpers       = "perk.helpers"
    PerkCitizen       = "perk.citizen"
    PerkHeightmap     = "perk.heightmap_editor"
    PerkBuilder       = "perk.builder"
    PerkRoomThumbnail = "perk.room_thumbnail"
    PerkMouseZoom     = "perk.mouse_zoom"
    PerkNavigatorV2   = "perk.navigator_v2"
    PerkSafeChat      = "perk.safe_chat"
    PerkClubOffer     = "perk.club_offer"
)
```

### Room Permissions (`pkg/room/domain/permission.go`)

```go
const (
    RoomWildcard   = "room.*"
    RoomEnter      = "room.enter"
    RoomEnterLocked = "room.enter.locked"
    RoomKick       = "room.kick"
    RoomBan        = "room.ban"
)
```

### Plugin-Defined Permissions

Plugins define their own constants in their module. The permission system
stores and resolves any valid dotted string:

```go
const MyPluginFeature = "myplugin.vip_lounge"
```

---

## Permission Resolution Algorithm

### Checking a Permission

```
HasPermission(group, "moderation.ban"):
  1. Check group_permissions for exact match "moderation.ban" → found? return true
  2. Check group_permissions for parent wildcard "moderation.*" → found? return true
  3. Check group_permissions for root wildcard "*" → found? return true
  4. Return false
```

For `a.b.c.d`, the check order is:
`a.b.c.d` → `a.b.c.*` → `a.b.*` → `a.*` → `*`

### Caching Strategy

Permission groups and their grants are **read-heavy, write-rare**.

- **Redis cache:** `perm:group:{groupId}` → JSON with group attributes
  and permission set
- **TTL:** 5 minutes
- **Invalidation:** On any admin write (group update, permission add/remove),
  delete the cache key
- **Warm on login:** First login loads from PostgreSQL and populates cache

### In-Memory Resolution

The resolver loads the full permission set into a `map[string]struct{}`
and performs O(1) exact lookups. Wildcard checks walk the dotted segments
(max 4 lookups for 3-segment permission). This is faster than SQL for
hot-path checks (e.g., room entry, chat filter).

---

## Perk Resolution from Permissions

The `user.perks` packet (ID 2586) sends a list of perks with
`(perkCode, errorMessage, isAllowed)`. Perk resolution maps known perk
codes to permission checks.

### Perk Code to Permission Mapping

| Client Perk Code | Permission String | Error Message |
|-----------------|-------------------|---------------|
| `USE_GUIDE_TOOL` | `perk.guide` | `Requires guide role` |
| `GIVE_GUIDE_TOURS` | `perk.guide.tours` | `Requires guide role` |
| `JUDGE_CHAT_REVIEWS` | `perk.chat_reviews` | `Requires moderator` |
| `VOTE_IN_COMPETITIONS` | `perk.competitions` | `` |
| `CALL_ON_HELPERS` | `perk.helpers` | `` |
| `CITIZEN` | `perk.citizen` | `` |
| `TRADE` | `perk.trade` | `Requires Club membership` |
| `HEIGHTMAP_EDITOR_BETA` | `perk.heightmap_editor` | `` |
| `BUILDER_AT_WORK` | `perk.builder` | `` |
| `NAVIGATOR_ROOM_THUMBNAIL_CAMERA` | `perk.room_thumbnail` | `` |
| `CAMERA` | `perk.camera` | `Requires Club membership` |
| `MOUSE_ZOOM` | `perk.mouse_zoom` | `` |
| `NAVIGATOR_PHASE_TWO` | `perk.navigator_v2` | `` |
| `SAFE_CHAT` | `perk.safe_chat` | `` |
| `HABBO_CLUB_OFFER_BETA` | `perk.club_offer` | `` |

### Resolution Logic

```
For each known perk code:
  1. Map code to permission string
  2. Check HasPermission(group, permission)
  3. If granted → isAllowed=true, errorMessage=""
  4. If denied → isAllowed=false, errorMessage=configured message
  5. Add to packet array
```

### Default Group Seed Perks

| Group | Granted Perks |
|-------|--------------|
| `default` | `perk.safe_chat`, `perk.helpers`, `perk.citizen` |
| `vip` | `perk.*` (all perks via wildcard) |
| `moderator` | `perk.*`, `moderation.kick`, `moderation.mute`, `moderation.alert` |
| `admin` | `*` (everything) |

---

## Packet Integration

### user.permissions (S2C 411)

Fields resolved from the user's permission group:

```
clubLevel      ← group.club_level
securityLevel  ← group.security_level
isAmbassador   ← group.is_ambassador
```

Sent during post-auth burst after `user.info`.

### user.perks (S2C 2586)

Fields resolved via perk-to-permission mapping described above. Sent
during post-auth burst after `user.permissions`.

### Impact on Wardrobe Slot Limits

Wardrobe slot limits are resolved from `group.club_level`:
- `club_level = 0` → 5 slots
- `club_level = 1` → 10 slots
- `club_level = 2` → 20 slots

No separate config — the group's club level drives this.

---

## API & CLI Endpoints

### REST API

All behind API key middleware.

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/groups` | List all permission groups | **M1** |
| `GET` | `/api/groups/{id}` | Get group with permissions | **M1** |
| `POST` | `/api/groups` | Create permission group | **M1** |
| `PATCH` | `/api/groups/{id}` | Update group attributes | **M1** |
| `DELETE` | `/api/groups/{id}` | Delete group (if not default, no users assigned) | **M1** |
| `GET` | `/api/groups/{id}/permissions` | List group permissions | **M1** |
| `POST` | `/api/groups/{id}/permissions` | Add permissions to group | **M1** |
| `DELETE` | `/api/groups/{id}/permissions/{perm}` | Remove permission from group | **M1** |
| `PATCH` | `/api/users/{id}/group` | Assign user to group | **M2** |

### API Request/Response Examples

**POST /api/groups**
```json
{
  "name": "moderator",
  "displayName": "Moderator",
  "priority": 50,
  "clubLevel": 0,
  "securityLevel": 1,
  "isAmbassador": false
}
```

**POST /api/groups/{id}/permissions**
```json
{
  "permissions": ["moderation.kick", "moderation.mute", "moderation.alert", "perk.*"]
}
```

**PATCH /api/users/{id}/group**
```json
{
  "groupId": 3
}
```

### CLI Commands

Mirror API 1:1 per AGENTS.md.

| Command | Description | Milestone |
|---------|-------------|-----------|
| `pixelsv group list` | List all groups | **M1** |
| `pixelsv group get <name>` | Get group details + permissions | **M1** |
| `pixelsv group create <name> --club 0 --security 1` | Create group | **M1** |
| `pixelsv group update <name> --display "Senior Mod"` | Update group | **M1** |
| `pixelsv group delete <name>` | Delete group | **M1** |
| `pixelsv group perm add <group> <perm> [<perm>...]` | Grant permissions | **M1** |
| `pixelsv group perm remove <group> <perm>` | Revoke permission | **M1** |
| `pixelsv group perm list <group>` | List permissions | **M1** |
| `pixelsv user group <userId> <groupName>` | Assign user to group | **M2** |

---

## Plugin Events & Usage

### SDK Events

New events added to `sdk/event.go`:

| Event | Cancellable | Fields | Milestone |
|-------|-------------|--------|-----------|
| `UserGroupChanged` | Yes | ConnID, UserID, OldGroupID, NewGroupID | **M2** |
| `PermissionChecked` | No | UserID, Permission, Granted | **M3** |

### Plugin Permission Check API

New method on `sdk.Server`:

```go
type PermissionAPI interface {
    HasPermission(userID int, permission string) bool
    GetGroup(userID int) (GroupInfo, bool)
}

type GroupInfo struct {
    ID            int
    Name          string
    ClubLevel     int
    SecurityLevel int
    IsAmbassador  bool
}
```

### Plugin Usage Example

```go
func (p *MyPlugin) Enable(srv sdk.Server) error {
    srv.Events().Subscribe(func(e *sdk.PacketReceived) {
        if e.PacketID == 123 {
            if !srv.Permissions().HasPermission(getUserID(e.ConnID), "myplugin.feature") {
                e.Cancel()
            }
        }
    })
    return nil
}
```

### Plugin-Defined Custom Permissions

Plugins can define their own permission strings. The admin adds them to
groups via API/CLI. The server does not need to know about custom
permissions in advance.

---

## Hexagonal Architecture Layout

```
core/permission/
├── constants.go      ← Wildcard constant, shared types
├── resolver.go       ← HasPermission algorithm (wildcard matching)
├── resolver_test.go  ← Resolver unit tests

pkg/permission/
├── domain/
│   ├── group.go      ← Group aggregate, Repository interface
│   └── grant.go      ← Permission grant value object
├── application/
│   ├── service.go    ← CRUD use cases for groups + grants
│   └── perk.go       ← Perk resolution use case
├── adapter/
│   └── httpapi/
│       ├── routes.go ← REST routes registration
│       └── openapi.go← OpenAPI spec
├── infrastructure/
│   ├── model/
│   │   ├── group.go  ← GORM model
│   │   └── grant.go  ← GORM model
│   └── store/
│       └── repository.go ← PostgreSQL repository
└── packet/
    ├── permissions.go ← user.permissions S2C (ID 411)
    └── perks.go       ← user.perks S2C (ID 2586)
```

---

## Default Seed Data

Migration creates these default groups:

| Name | Display | Priority | Club | Security | Ambassador | Default | Permissions |
|------|---------|----------|------|----------|------------|---------|-------------|
| `default` | Default | 0 | 0 | 0 | false | true | `perk.safe_chat`, `perk.helpers`, `perk.citizen` |
| `vip` | VIP | 10 | 2 | 0 | false | false | `perk.*` |
| `moderator` | Moderator | 50 | 0 | 1 | false | false | `perk.*`, `moderation.kick`, `moderation.mute`, `moderation.alert` |
| `admin` | Administrator | 100 | 2 | 3 | true | false | `*` |

---

## Edge Cases & Extreme Cases

### 1. Deleting a Group with Assigned Users

Deletion of a group with users assigned is **rejected** with a 409
Conflict error. Admin must reassign users first. The API returns the
count of affected users.

### 2. Deleting the Default Group

Rejected. Exactly one group must be `is_default = true` at all times.
To change the default, set `is_default` on another group first.

### 3. Multiple Default Groups

Application-layer validation ensures only one group has `is_default =
true`. Setting a new default atomically unsets the previous one in a
transaction.

### 4. Wildcard Explosion

A group with `*` permission grants everything, including permissions
defined by plugins after the group was created. This is intentional —
admin groups should have full access.

### 5. Permission String Validation

- Max length: 128 characters
- Allowed characters: lowercase `a-z`, digits `0-9`, dots `.`,
  underscores `_`, and `*` for wildcards
- Wildcard `*` is only valid as the last segment (`moderation.*` is
  valid, `*.ban` is not)
- Empty strings rejected
- Validated in application layer, not database constraint

### 6. Cache Invalidation Race

When an admin changes a group's permissions, the Redis cache is deleted.
Concurrent requests may see stale permissions for up to 5 minutes if the
delete fails. Mitigation: admin API returns a warning if cache delete
fails, suggesting server restart for immediate effect.

### 7. Group Assignment During Active Session

When a user's group changes via API while they are online:
- `UserGroupChanged` event fires on the instance with the session
- Server sends updated `user.permissions` (411) and `user.perks` (2586)
  packets to the client
- Client updates UI immediately without reconnection

### 8. Permission Check Performance

Hot-path permission checks (room entry, chat) use in-memory cached
permission sets. No database query per check. Redis cache miss triggers
a PostgreSQL load (< 1ms for group + permissions). Worst case: cold
cache on first request after 5-minute TTL expiry.

### 9. Empty Permission Groups

A group with zero permissions is valid. Users in that group have no
special access. The `user.perks` packet returns all perks as
`isAllowed = false`.

### 10. Large Permission Sets

A group with 500+ permissions is supported but unusual. The
`map[string]struct{}` in-memory set handles this efficiently. The
wildcard `*` should be used instead of enumerating all permissions.

### 11. Concurrent Group Modifications

Two admins adding permissions to the same group simultaneously: both
succeed independently. The composite PK `(group_id, permission)` prevents
duplicate entries. No transaction needed for idempotent inserts.

---

## Implementation Roadmap

### Milestone 1: Core Group Model + API

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 1 | Create `permission_groups` table + migration | - | PENDING |
| 2 | Create `group_permissions` table + migration | 1 | PENDING |
| 3 | Create `Group` domain aggregate + `Repository` interface | 1 | PENDING |
| 4 | Create `Grant` value object | 2 | PENDING |
| 5 | Implement `HasPermission` resolver with wildcard matching | 3 | PENDING |
| 6 | Implement GORM repository for groups + permissions | 3,4 | PENDING |
| 7 | Implement group CRUD application service | 6 | PENDING |
| 8 | Seed default groups (default, vip, moderator, admin) | 6 | PENDING |
| 9 | REST API: group CRUD + permission management | 7 | PENDING |
| 10 | CLI: group commands | 9 | PENDING |
| 11 | OpenAPI spec for group endpoints | 9 | PENDING |
| 12 | Redis cache for group + permissions | 6 | PENDING |
| 13 | Unit tests: resolver wildcard matching | 5 | PENDING |
| 14 | Unit tests: repository CRUD | 6 | PENDING |
| 15 | Integration tests: API endpoints | 9 | PENDING |

### Milestone 2: User Integration + Live Updates

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 16 | Add `group_id` FK to `users` table (migration) | M1 | PENDING |
| 17 | Wire group resolution into `user.permissions` packet | 16 | PENDING |
| 18 | Wire perk resolution into `user.perks` packet | 17 | PENDING |
| 19 | API: PATCH /api/users/{id}/group | 16 | PENDING |
| 20 | CLI: user group assignment | 19 | PENDING |
| 21 | Fire `UserGroupChanged` plugin event | 19 | PENDING |
| 22 | Live packet update on group change (send 411 + 2586) | 21 | PENDING |
| 23 | Unit + integration tests for M2 | all M2 | PENDING |

### Milestone 3: Plugin API + Permission Checks

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 24 | Add `PermissionAPI` to `sdk.Server` interface | M2 | PENDING |
| 25 | Implement `pluginPermissionAPI` wrapper | 24 | PENDING |
| 26 | Fire `PermissionChecked` event (opt-in, not default) | 25 | PENDING |
| 27 | E2E test: custom plugin permission check | 25 | PENDING |
| 28 | Documentation: permission system wiki page | all | PENDING |

---

## Caveats & Technical Notes

### Migration Order

The `permission_groups` and `group_permissions` tables must be created
BEFORE the `users` table migration that adds `group_id`. The seed script
for default groups must run between these two migrations.

### Perk Code Registry

The known perk codes are defined as a Go `var` slice in the perk
resolution package. This is the single source of truth for the
code-to-permission mapping. Adding a new perk requires adding one entry
to this slice and one Go constant.

### Relation to Subscription & Offers Realm

The `club_level` on a permission group is a **static assignment** — the
admin sets it when creating the group. Time-based subscription logic
(monthly VIP, payment integration) belongs to the Subscription & Offers
realm which is DEFERRED. When that realm is implemented, it may
dynamically assign users to VIP/Club groups based on subscription status.

### No Inheritance

Groups do not inherit from other groups. If a "senior moderator" needs
all moderator permissions plus extras, the admin creates a group with
both sets. This keeps resolution O(1) per permission check and avoids
graph traversal complexity.

### Permission Realm Ownership

Each realm defines its own permission constants. The permission system
(`core/permission/`) owns only the resolver algorithm and wildcard
matching. It does not know about specific permissions like `moderation.ban`.
This follows the distributed ownership principle from AGENTS.md.
