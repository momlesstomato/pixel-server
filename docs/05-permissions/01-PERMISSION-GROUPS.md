# Permission Groups

## Overview

The permission system organizes users into groups. Each group holds a set of
dotted-notation permissions. Users can belong to multiple groups; their
effective permissions are the union of all group grants. One group is always
marked as the default and is automatically assigned to users with no explicit
assignment.

## Domain Model

### Group

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | Primary key |
| `Name` | string | Unique slug, `^[a-z0-9-]{2,64}$` |
| `DisplayName` | string | Human-readable label (max 128 chars) |
| `Priority` | int | Conflict resolution; higher wins |
| `ClubLevel` | int | Habbo Club tier (0 = none, 1 = HC, 2 = VIP) |
| `SecurityLevel` | int | Staff tier (0 = normal, 1 = mod, 2 = senior, 3 = admin) |
| `IsAmbassador` | bool | Ambassador role flag |
| `IsDefault` | bool | Assigned automatically to new users |

### Grant

| Field | Type | Description |
|-------|------|-------------|
| `GroupID` | int | Owning group |
| `Permission` | string | Dotted notation, max 128 chars |

Permission format: `^[a-z0-9_]+(\.[a-z0-9_]+|\.\*)*$` or `*` (wildcard).

### Access

Resolved at runtime per user:

| Field | Type | Description |
|-------|------|-------------|
| `UserID` | int | Target user |
| `PrimaryGroup` | Group | Highest-priority assigned group |
| `GroupIDs` | []int | All assigned group IDs |
| `Permissions` | map | Merged grants from all groups |

## Permission Resolution

The resolver operates on a flat set of grants with hierarchical matching:

1. **Exact match** — `"perk.trade"` matches `"perk.trade"`
2. **Full wildcard** — `"*"` matches everything
3. **Hierarchical wildcard** — `"perk.*"` matches `"perk.trade"`, `"perk.camera"`, etc.
4. **Prefix walk** — For `"a.b.c"`, the resolver tries `"a.b.*"` then `"a.*"`

```go
Resolve([]string{"perk.*", "moderation.kick"}, "perk.trade")     // true
Resolve([]string{"perk.*", "moderation.kick"}, "moderation.ban") // false
Resolve([]string{"*"}, "anything.at.all")                        // true
```

## Default Groups (Seeds)

| Name | Priority | Club | Security | Ambassador | Default | Grants |
|------|----------|------|----------|------------|---------|--------|
| `default` | 0 | 0 | 0 | No | **Yes** | `perk.safe_chat`, `perk.helpers`, `perk.citizen` |
| `vip` | 10 | 2 | 0 | No | No | `perk.*` |
| `moderator` | 50 | 0 | 1 | No | No | `perk.*`, `moderation.kick`, `moderation.mute`, `moderation.alert` |
| `admin` | 100 | 2 | 3 | **Yes** | No | `*` |

## REST API

### Group Management

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/groups` | List all groups with permissions |
| `GET` | `/api/v1/groups/{id}` | Get group details |
| `POST` | `/api/v1/groups` | Create group |
| `PATCH` | `/api/v1/groups/{id}` | Update group |
| `DELETE` | `/api/v1/groups/{id}` | Delete group |

**POST body:**

| Field | Type | Constraint |
|-------|------|------------|
| `name` | string | Required, `[a-z0-9-]{2,64}` |
| `display_name` | string | Required, max 128 |
| `priority` | int | Required |
| `club_level` | int | Optional (0) |
| `security_level` | int | Optional (0) |
| `is_ambassador` | bool | Optional (false) |
| `is_default` | bool | Optional (false) |

Deleting a group fails if it is the default group or if any users are still
assigned to it.

### Permission Management

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/groups/{id}/permissions` | List group permissions |
| `POST` | `/api/v1/groups/{id}/permissions` | Add permissions |
| `DELETE` | `/api/v1/groups/{id}/permissions/{permission}` | Remove permission |

**POST body:**

```json
{"permissions": ["perk.trade", "moderation.kick"]}
```

### User Assignment

| Method | Path | Description |
|--------|------|-------------|
| `PATCH` | `/api/v1/users/{id}/group` | Set single group |
| `PATCH` | `/api/v1/users/{id}/groups` | Replace all groups |

**Single group body:**

```json
{"group_id": 2}
```

**Multi-group body:**

```json
{"group_ids": [1, 2, 3]}
```

## CLI

```bash
pixelsv group list                              # List all groups
pixelsv group get 1                             # Get group details
pixelsv group create vip \
  --display VIP --priority 10 --club 2          # Create group
pixelsv group update 1 --priority 50            # Update group
pixelsv group delete 3                          # Delete group

pixelsv group perm list 1                       # List group permissions
pixelsv group perm add 1 perk.trade perk.camera # Add permissions
pixelsv group perm remove 1 perk.trade          # Remove permission

pixelsv group assign-user 5 1 2                 # Assign user 5 to groups 1, 2
```

## Database

### permission_groups

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `name` | varchar(64) | Unique slug |
| `display_name` | varchar(128) | Label |
| `priority` | int | Ordering |
| `club_level` | int | HC tier |
| `security_level` | int | Staff tier |
| `is_ambassador` | bool | Ambassador flag |
| `is_default` | bool | Auto-assign (indexed) |

### group_permissions

Composite primary key: `(group_id, permission)`.

| Column | Type | Description |
|--------|------|-------------|
| `group_id` | int | FK to permission_groups |
| `permission` | varchar(128) | Dotted permission string (indexed) |

### user_permission_groups

Composite primary key: `(user_id, group_id)`.

| Column | Type | Description |
|--------|------|-------------|
| `user_id` | int | FK to users (indexed) |
| `group_id` | int | FK to permission_groups (indexed) |
| `created_at` | timestamp | Assignment time |

A migration backfills the legacy `users.group_id` column into this table
during initial setup.

## Caching

Group snapshots are cached in Redis as JSON with a configurable TTL:

| Key Pattern | Default TTL | Description |
|-------------|-------------|-------------|
| `perm:group:{id}` | 300s | Serialized group + permissions |

Cache is invalidated when permissions are added, removed, or when group
details change. The prefix and TTL are configurable via `PERMISSIONS_CACHE_PREFIX`
and `PERMISSIONS_CACHE_TTL_SECONDS`.

## Plugin Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `PermissionChecked` | No | UserID, Permission, Granted |
| `UserGroupChanged` | **Yes** | UserID, OldGroupID, NewGroupID, OldGroupIDs, NewGroupIDs |

`PermissionChecked` fires only when `PERMISSIONS_EMIT_CHECKED` is true. This
is primarily for audit/observability plugins.

`UserGroupChanged` fires before the new assignment is persisted. If cancelled,
the assignment is rolled back.

## Live Updates

When a user's groups change, the server pushes updated packets to the
connected client in real time:

1. `user.permissions` (411) — Updated club level, security level, ambassador
2. `user.perks` (2586) — Recalculated client perk grants

The `LiveUpdater` publishes these through the broadcaster, reaching the user's
notification channel even across instances.
