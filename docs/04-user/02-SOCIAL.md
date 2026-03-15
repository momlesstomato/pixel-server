# Social Features

## Overview

Social features include the wardrobe system, user respects, ignore list, and
profile viewing. Each feature follows the same hexagonal pattern: binary
packets for real-time interaction, REST endpoints for administration, and CLI
commands for operator tooling.

## Wardrobe

The wardrobe stores saved outfit slots per user.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `user.get_wardrobe` | 2742 | C2S | PageID |
| `user.wardrobe_page` | 3315 | S2C | PageID, Slots [{SlotID, Figure, Gender}] |
| `user.save_wardrobe_outfit` | 800 | C2S | SlotID (1–50), Figure, Gender |

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/{id}/wardrobe` | Get wardrobe slots |

### Database

Table `user_wardrobe_slots` with composite unique index on `(user_id, slot_id)`.

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `user_id` | int | Owner |
| `slot_id` | int | Slot number (1–50) |
| `figure` | string | Appearance string |
| `gender` | string | M or F |

## Respects

Players can send one respect per day per target type (user or pet). The
system enforces a daily limit of 3 respects per type.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `user.respect` | 2694 | C2S | UserID (int32) |
| `user.respect_received` | 2815 | S2C | UserID, RespectsReceived |

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/users/{id}/respect` | Send respect |
| `GET` | `/api/v1/users/{id}/respects` | List respect history |

**POST body:**

| Field | Type | Constraint |
|-------|------|------------|
| `actor_user_id` | int | Required, > 0 |

**GET query parameters:**

| Param | Type | Default |
|-------|------|---------|
| `limit` | int | — |
| `offset` | int | 0 |

**POST response:**

```json
{"respects_received": 5, "remaining": 2}
```

### CLI

```bash
pixelsv user respect 2 --actor-user-id 1
```

### Plugin Event

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `Respected` | Yes | ActorConnID, ActorUserID, TargetUserID |

### Database

Table `user_respects` with composite index on `(actor_user_id, respected_at, target_type)`.

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `actor_user_id` | int | Who sent the respect |
| `target_id` | int | Target entity ID |
| `target_type` | int | 0 = user, 1 = pet |
| `respected_at` | date | Day of respect (for limit enforcement) |

## Ignore System

Players can ignore other users via real-time packets. Administrators can
manage ignore lists through REST and CLI.

### Real-Time Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `user.get_ignored` | 3878 | C2S | (empty) |
| `user.ignored_users` | 126 | S2C | Usernames (string array) |
| `user.ignore` | 1117 | C2S | Username |
| `user.ignore_result` | 207 | S2C | Result (int32), Name |
| `user.ignore_by_id` | configurable | C2S | UserID (int32) |
| `user.unignore` | configurable | C2S | Username |

### Admin REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/{id}/ignores` | List ignored users |
| `POST` | `/api/v1/users/{id}/ignores` | Add ignore entry |
| `DELETE` | `/api/v1/users/{id}/ignores/{targetId}` | Remove ignore entry |

**POST body:**

| Field | Type | Constraint |
|-------|------|------------|
| `target_user_id` | int | Required, > 0 |

**GET response:**

```json
{"entries": [{"user_id": 5, "username": "bob"}]}
```

### CLI

```bash
pixelsv user ignore list 1          # List ignored users
pixelsv user ignore add 1 5         # User 1 ignores user 5
pixelsv user ignore remove 1 5      # User 1 unignores user 5
```

### Plugin Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `Ignored` | Yes | ConnID, UserID, IgnoredUserID |
| `Unignored` | Yes | ConnID, UserID, IgnoredUserID |

Admin operations (`AdminIgnoreUser`, `AdminUnignoreUser`) do not fire events.

### Database

Table `user_ignores` with composite unique index on `(user_id, ignored_user_id)`.

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `user_id` | int | Owner |
| `ignored_user_id` | int | Ignored target |
| `created_at` | timestamp | When the ignore was created |

## Profile View

Other players can view a user's public profile.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `user.get_profile` | 3265 | C2S | UserID, OpenProfileWindow |
| `user.profile` | 3898 | S2C | UserID, Username, Figure, Motto, Registration, AchievementPoints, FriendsCount, IsMyFriend, RequestSent, IsOnline, SecondsSinceLastVisit, OpenProfileWindow |

### REST API

Profile data is available through `GET /api/v1/users/{id}`.

## Name Changes

The name change system validates proposed names, generates suggestions for
taken names, and persists the change with an optional `CanChangeName` guard.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `user.check_name` | 3950 | C2S | Name |
| `user.change_name` | 2977 | C2S | Name |
| `user.name_result` (check) | 563 | S2C | ResultCode, Name, Suggestions |
| `user.name_result` (change) | 118 | S2C | ResultCode, Name, Suggestions |
| `user.name_change` | 2182 | S2C | WebID, UserID, NewName |
| `user.approve_name` | configurable | C2S | Name |

**Result codes:**

| Code | Meaning |
|------|---------|
| 0 | Available |
| 1 | Name taken |
| 2 | Invalid format |
| 3 | Name change not allowed |

**Validation:** Regex `^[A-Za-z0-9._-]{3,24}$`.

When a name is taken, the server generates deterministic suggestions:
`["name1", "name_2"]`.

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/users/{id}/name-change` | Force admin rename |

**POST body:**

| Field | Type | Constraint |
|-------|------|------------|
| `name` | string | Required, 3–24 chars |

### CLI

```bash
pixelsv user rename 1 newname
```

### Plugin Event

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `NameChanged` | Yes | ConnID, UserID, OldName, NewName |

Admin renames via CLI and REST use `force=true`, bypassing the `CanChangeName`
guard.
