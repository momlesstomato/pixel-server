# 04 - User & Profile Realm

## Overview

The User & Profile realm owns the player identity model: avatar appearance,
motto, settings, wardrobe, ignore list, profile viewing, name changes,
respects, and the new-user experience (NUX). It is the second realm a
client interacts with after the Handshake & Security phase.

**Permissions, clubs, perks, and security levels** have been extracted to a
dedicated Permission & Group realm (see `05-PERMISSION-SYSTEM.md`). The user
table stores a `group_id` FK that references the permission group, replacing
vendor-style `club_level`, `security_level`, and `is_ambassador` columns.

The pixel-protocol lists **56 packets** in this realm (29 C2S, 27 S2C).
Many of these packets depend on features outside this realm (rooms, groups,
email). Our approach: implement the **core identity model** and the packets
that operate on it, and defer packets that require other realms.

---

## Vendor Cross-Reference

Analysis of three reference implementations (Sodium C#, Gladiator Java,
Galaxy Java) and the pixel-protocol YAML spec.

### Post-Auth Burst (user packets)

Vendors send these user-related packets in the post-auth burst immediately
after `authentication.ok`:

| Order | Packet ID | Name | Notes |
|-------|-----------|------|-------|
| 1 | 2725 | `user.info` | Own user data: figure, motto, gender, respects |
| 2 | 411 | `user.permissions` | Club level, security level, ambassador |
| 3 | 513 | `user.settings` | Volume, chat style, room invites, flag bits |
| 4 | 2875 | `user.home_room` | Home room ID or -1 |
| 5 | 2586 | `user.perks` | Feature gating flags (USE_GUIDE, CAMERA, etc.) |
| 6 | 3738 | `user.noobness_level` | Account age tier (0, 1, 2) |
| 7 | 126 | `user.ignored_users` | Ignore list on login |

**Sodium** sends 5/7 of these. **Gladiator** sends all 7. **Galaxy** sends 6/7.
All vendors agree `user.info` must come before any other user packet.

### User Database Schema (Vendor Comparison)

| Field | Sodium | Gladiator | Galaxy | pixelsv (proposed) |
|-------|--------|-----------|--------|--------------------|
| figure | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` | `VARCHAR(255)` |
| gender | `CHAR(1)` | `CHAR(1)` | `VARCHAR(1)` | `CHAR(1)` |
| motto | `VARCHAR(127)` | `VARCHAR(127)` | `VARCHAR(128)` | `VARCHAR(127)` |
| respects_received | `INT` | `INT` | `INT` | `INT DEFAULT 0` (materialized counter) |
| respects_remaining | `INT` | `INT` | `INT` | ~~removed~~ → computed from `user_respects` |
| respects_pet_remaining | `INT` | `INT` | `INT` | ~~removed~~ → computed from `user_respects` |
| home_room | `INT` | `INT` | `INT` | `INT DEFAULT -1` |
| can_change_name | `BOOL` | `BOOL` | `BOOL` | `BOOL DEFAULT false` |
| club_level | `INT` | `INT` | `INT` | ~~removed~~ → via `permission_groups` |
| security_level | `INT` | `INT` | `INT` | ~~removed~~ → via `permission_groups` |
| is_ambassador | `BOOL` | `BOOL` | `BOOL` | ~~removed~~ → via `permission_groups` |
| last_access | `TIMESTAMP` | `TIMESTAMP` | `TIMESTAMP` | `TIMESTAMP` |
| noobness_level | - | `INT` | - | `INT DEFAULT 2` |
| group_id | - | - | - | `BIGINT FK → permission_groups.id` |

**Respect tracking (pixelsv vs vendors):** All vendors use a simple counter
column (`respects_remaining`) on the user table, reset daily on first login.
We replace this with a **`user_respects` audit table** that stores individual
respect events. Remaining respects are computed at query time as
`3 - COUNT(today's records)`. This provides audit trail without bloating the
user table with mutable daily counters.

All vendors store settings in a **separate table** (1:1 with users).
Wardrobe slots use a **separate table** (1:N with users).
Ignore lists use a **separate table** (N:N between users).

**Permissions** are extracted to `permission_groups` + `group_permissions`
tables. See `05-PERMISSION-SYSTEM.md`.

---

## Packet Registry

### Server-to-Client (27 packets)

| ID | Name | Fields | Phase | Priority |
|----|------|--------|-------|----------|
| 2725 | `user.info` | userId, username, figure, gender, motto, realName, directMail, respectsReceived, respectsRemaining, respectsPetRemaining, streamPublishingAllowed, lastAccessDate, canChangeName, safetyLocked | post-auth | **M1** |
| 411 | `user.permissions` | clubLevel, securityLevel, isAmbassador | post-auth | **M1** |
| 513 | `user.settings` | volumeSystem, volumeFurni, volumeTrax, oldChat, roomInvites, cameraFollow, flags, chatType | post-auth | **M2** |
| 2875 | `user.home_room` | homeRoomId, roomIdToEnter | post-auth | **M2** |
| 2586 | `user.perks` | count, [perkCode, errorMessage, isAllowed]* | post-auth | **M1** |
| 3738 | `user.noobness_level` | noobnessLevel | post-auth | **M1** |
| 126 | `user.ignored_users` | count, [username]* | post-auth | **M3** |
| 2429 | `user.figure` | figure, gender | on-demand | **M2** |
| 2815 | `user.respect_received` | userId, respectsReceived | on-demand | **M2** |
| 3898 | `user.profile` | userId, username, figure, motto, registration, achievementPoints, friendsCount, isMyFriend, requestSent, isOnline, groupsCount, [groups]*, secondsSinceLastVisit, openProfileWindow | on-demand | **M3** |
| 3315 | `user.wardrobe_page` | pageIndex, count, [slotId, figure, gender]* | on-demand | **M3** |
| 118 | `user.change_name_result` | resultCode, name, suggestionCount, [suggestion]* | on-demand | **M4** |
| 2182 | `user.name_change` | webId, id, newName | on-demand | **M4** |
| 966 | `user.classification` | count, [userId, username, userClass]* | on-demand | **DEFER** |
| 2016 | `user.relationship_status` | userId, count, [type, friendCount, randomFriendId, randomFriendName, randomFriendFigure]* | on-demand | **DEFER** |
| 1255 | `user.tags` | roomUnitId, count, [tag]* | on-demand | **DEFER** |
| 1683 | `user.banned` | message | on-demand | **M1** |
| 563 | `user.check_name_result` | resultCode, name, suggestionCount, [suggestion]* | on-demand | **M4** |
| 2707 | `user.welcome_gift_status` | email, isVerified, allowChange, furniId, requestedByUser | on-demand | **DEFER** |
| 207 | `user.ignore_result` | result, name | on-demand | **M3** |
| (various) | `user.email_status`, `user.safety_lock_status`, `user.in_client_link`, `user.extended_profile_changed`, `user.approve_name_result`, `user.welcome_gift_change_email_result`, `user.change_email_result` | varies | on-demand | **DEFER** |

### Client-to-Server (29 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 357 | `user.get_info` | (empty) | **M1** |
| 2228 | `user.update_motto` | motto | **M2** |
| 2730 | `user.update_figure` | gender, figure | **M2** |
| 1367 | `user.settings_volume` | volumeSystem, volumeFurni, volumeTrax | **M2** |
| 2742 | `user.get_wardrobe` | pageId | **M3** |
| 800 | `user.save_wardrobe_outfit` | slotId, figure, gender | **M3** |
| 3265 | `user.get_profile` | userId, openProfile | **M3** |
| 3878 | `user.get_ignored` | username | **M3** |
| 1117 | `user.ignore` | username | **M3** |
| (varies) | `user.unignore`, `user.ignore_id` | username / userId | **M3** |
| 3950 | `user.check_name` | name | **M4** |
| 2977 | `user.change_name` | name | **M4** |
| (varies) | `user.approve_name` | name | **M4** |
| 1740 | `user.set_home_room` | roomId | **M2** |
| (varies) | `user.settings_room_invites`, `user.settings_old_chat` | boolean | **M2** |
| 2694 | `user.respect` | userId | **M2** |
| (varies) | `user.get_relationship_status`, `user.set_relationship_status` | userId, type | **DEFER** |
| (varies) | `user.set_classification`, `user.get_tags` | varies | **DEFER** |
| (varies) | `user.nux_proceed`, `user.nux_get_gifts` | varies | **DEFER** |
| (varies) | `user.effect_enable` | effectId | **DEFER** |
| (varies) | `user.get_email_status`, `user.change_email`, `user.welcome_gift_change_email`, `user.get_profile_by_name` | varies | **DEFER** |

---

## Database Model Design

### Normalized Schema

The User & Profile realm requires **5 PostgreSQL tables**. All ORM models
live in their realm infrastructure package per AGENTS.md rules.

#### Table 1: `users` (extend existing)

Currently stores `id`, `username`, `owner_id`, timestamps. Must be extended
with identity fields. This is the **core aggregate root**.

Permission-related columns (`club_level`, `security_level`, `is_ambassador`)
are NOT stored here — they live on the permission group resolved via
`group_id`. Daily respect limits are NOT stored here — they are computed
from the `user_respects` audit table.

```
users
├── id                      BIGINT PK AUTO
├── username                VARCHAR(64) UNIQUE NOT NULL
├── figure                  VARCHAR(255) NOT NULL DEFAULT 'hr-115-42.hd-180-1.ch-3030-82.lg-275-82.sh-295-62'
├── gender                  CHAR(1) NOT NULL DEFAULT 'M'
├── motto                   VARCHAR(127) NOT NULL DEFAULT ''
├── real_name               VARCHAR(64) NOT NULL DEFAULT ''
├── respects_received       INT NOT NULL DEFAULT 0           -- materialized total counter
├── home_room_id            INT NOT NULL DEFAULT -1
├── can_change_name         BOOLEAN NOT NULL DEFAULT false
├── noobness_level          INT NOT NULL DEFAULT 2           -- 0=veteran, 1=recent, 2=new
├── safety_locked           BOOLEAN NOT NULL DEFAULT false
├── last_access_at          TIMESTAMP
├── group_id                BIGINT NOT NULL (FK → permission_groups.id)
├── owner_id                BIGINT INDEX (FK → users.id)     -- admin creator (existing)
├── created_at              TIMESTAMP
├── updated_at              TIMESTAMP
└── deleted_at              TIMESTAMP INDEX (soft delete)
```

#### Table 2: `user_settings`

One-to-one with users. Holds client preference data that is NOT part of
identity (volumes, chat style, flags). Separate table because:
- settings are updated frequently (volume changes on every session)
- decouples identity writes from settings writes
- clear ownership: user realm, not session realm

```
user_settings
├── id          BIGINT PK AUTO
├── user_id     BIGINT UNIQUE NOT NULL (FK → users.id)
├── volume_system   INT NOT NULL DEFAULT 100
├── volume_furni    INT NOT NULL DEFAULT 100
├── volume_trax     INT NOT NULL DEFAULT 100
├── old_chat        BOOLEAN NOT NULL DEFAULT false
├── room_invites    BOOLEAN NOT NULL DEFAULT true
├── camera_follow   BOOLEAN NOT NULL DEFAULT true
├── flags           INT NOT NULL DEFAULT 0
├── chat_type       INT NOT NULL DEFAULT 0          -- 0=normal, 1=wide
├── created_at      TIMESTAMP
└── updated_at      TIMESTAMP
```

#### Table 3: `user_wardrobe_slots`

One-to-many with users. Each row is one saved outfit slot (max 20 per user
for VIP, 5 for non-VIP). Vendors use 10 slots base, 20 VIP.

```
user_wardrobe_slots
├── id          BIGINT PK AUTO
├── user_id     BIGINT NOT NULL (FK → users.id) INDEX
├── slot_id     INT NOT NULL
├── figure      VARCHAR(255) NOT NULL
├── gender      CHAR(1) NOT NULL DEFAULT 'M'
├── UNIQUE(user_id, slot_id)
```

#### Table 4: `user_ignores`

Many-to-many self-referential. Storing ignore relationships.

```
user_ignores
├── id              BIGINT PK AUTO
├── user_id         BIGINT NOT NULL (FK → users.id) INDEX
├── ignored_user_id BIGINT NOT NULL (FK → users.id) INDEX
├── created_at      TIMESTAMP
├── UNIQUE(user_id, ignored_user_id)
```

#### Table 5: `user_respects`

Audit table tracking individual respect events. Replaces vendor-style
`respects_remaining` counter columns. Each row represents one respect given
by one user to another (or to a pet) on a specific UTC date.

```
user_respects
├── id              BIGINT PK AUTO
├── actor_user_id   BIGINT NOT NULL (FK → users.id)
├── target_id       BIGINT NOT NULL                   -- user ID or pet ID
├── target_type     SMALLINT NOT NULL DEFAULT 0       -- 0=user, 1=pet
├── respected_at    DATE NOT NULL                     -- UTC date
├── INDEX(actor_user_id, respected_at, target_type)
```

**Why this design instead of a counter:**

| Approach | Pros | Cons |
|----------|------|------|
| Counter column (vendors) | Fast read, simple | No audit trail, mutable daily column, must reset on login |
| Full event log (every respect forever) | Complete audit | Unbounded growth, N×3×days rows |
| **Audit table (ours)** | Audit trail, computed remaining, no daily reset needed | One extra query for remaining count |

**Growth estimate:** Max 3 user + 3 pet respects per user per day. For a
1000-user server over 1 year: `1000 × 6 × 365 ≈ 2.2M rows`. Manageable
with index on `(actor_user_id, respected_at, target_type)`.

**Daily remaining computation:**
```sql
SELECT 3 - COUNT(*) FROM user_respects
WHERE actor_user_id = ? AND respected_at = CURRENT_DATE AND target_type = 0
```

**No daily reset needed.** The count resets naturally at UTC midnight because
`respected_at = CURRENT_DATE` returns 0 rows for the new day. This eliminates
the vendor pattern of resetting counters on first login.

**Materialized counter:** `users.respects_received` is incremented atomically
on each respect event (not recomputed). This counter is the fast-read path
for `user.info` packet and profile viewing. The audit table provides the
daily limit enforcement.

**Pet respects:** Use `target_type = 1` with `target_id` pointing to a pet
ID. Pet respects are deferred to the Room Entities realm but the schema
supports them now. Until pets are implemented, `respectsPetRemaining` in
`user.info` returns `3` (hardcoded, no records exist with type=1).

### Entity Relationship Diagram

```
users ──1:1──> user_settings
users ──1:N──> user_wardrobe_slots
users ──N:N──> user_ignores (self-referential)
users ──1:N──> user_respects (as actor)
users ──N:1──> permission_groups (via group_id FK, see 05-PERMISSION-SYSTEM.md)
users ──1:N──> login_events (existing)
```

---

## Hexagonal Architecture Layout

```
pkg/user/
├── domain/
│   ├── user.go             ← User aggregate (extend with new fields)
│   ├── settings.go         ← Settings value object
│   └── permission.go       ← Perk permission constants (perk.camera, etc.)
├── application/
│   ├── service.go          ← Identity use cases (get info, update motto/figure)
│   ├── settings.go         ← Settings use cases (volume, chat, invites)
│   └── wardrobe.go         ← Wardrobe use cases (get/save slots)
├── adapter/
│   └── realtime/
│       ├── handler.go      ← Packet dispatch for user realm C2S
│       └── transport.go    ← Send S2C packets
├── infrastructure/
│   ├── model/
│   │   ├── record.go       ← Extend existing user GORM model
│   │   ├── settings.go     ← Settings GORM model
│   │   ├── wardrobe.go     ← Wardrobe GORM model
│   │   ├── ignore.go       ← Ignore GORM model
│   │   └── login_event.go  ← Existing login event model
│   └── store/
│       ├── repository.go   ← Extend existing user repository
│       └── settings.go     ← Settings repository
└── packet/
    ├── identity/           ← user.info, user.figure, user.noobness_level, etc.
    ├── settings/           ← user.settings, user.settings_volume, etc.
    ├── wardrobe/           ← user.wardrobe_page, user.save_wardrobe_outfit
    ├── ignore/             ← user.ignored_users, user.ignore, user.unignore
    └── profile/            ← user.profile, user.get_profile
```

**Note:** `user.permissions` (ID 411) and `user.perks` (ID 2586) packets
live in `pkg/permission/packet/` under the permission realm, not the user
realm. They are resolved from the permission group, not user data. See
`05-PERMISSION-SYSTEM.md`.

**Note:** `user_respects` model lives in `pkg/user/infrastructure/model/`
alongside the other user models. This brings the model package to 6 files
(at the limit per AGENTS.md). The `respect.go` model file includes both
the GORM model and the daily count query helper.

---

## API & CLI Endpoints

### REST API Endpoints

All behind API key middleware (matching existing pattern).

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/users/{id}` | Get user by ID (includes group info) | **M1** |
| `PATCH` | `/api/users/{id}` | Update user fields (figure, motto, home room) | **M2** |
| `GET` | `/api/users/{id}/settings` | Get user client settings | **M2** |
| `PATCH` | `/api/users/{id}/settings` | Update user settings | **M2** |
| `GET` | `/api/users/{id}/wardrobe` | Get wardrobe slots | **M3** |
| `POST` | `/api/users/{id}/respect` | Admin-grant respect to user | **M2** |
| `GET` | `/api/users/{id}/respects` | Get respect history (paginated) | **M2** |
| `POST` | `/api/users/{id}/name-change` | Admin-force name change | **M4** |

**Permissions API** (group assignment, group CRUD, permission management)
lives under `/api/groups/` — see `05-PERMISSION-SYSTEM.md`.

### CLI Commands

Mirror API 1:1 per AGENTS.md:

| Command | Description | Milestone |
|---------|-------------|-----------|
| `pixelsv user get <id>` | Get user details + group | **M1** |
| `pixelsv user update <id> --motto "x" --figure "x"` | Update user fields | **M2** |
| `pixelsv user respect <id>` | Admin-grant respect | **M2** |
| `pixelsv user rename <id> <name>` | Force name change | **M4** |

**Group commands** (`pixelsv group ...`) live under the permission realm
CLI — see `05-PERMISSION-SYSTEM.md`.

---

## Plugin Events

New events to add to `sdk/event.go`:

| Event | Cancellable | Fields | Milestone |
|-------|-------------|--------|-----------|
| `UserInfoRequested` | No | ConnID, UserID | **M1** |
| `UserMottoChanged` | Yes | ConnID, UserID, OldMotto, NewMotto | **M2** |
| `UserFigureChanged` | Yes | ConnID, UserID, OldFigure, NewFigure, Gender | **M2** |
| `UserRespected` | Yes | ActorConnID, ActorUserID, TargetUserID | **M2** |
| `UserNameChanged` | Yes | UserID, OldName, NewName | **M4** |
| `UserIgnored` | Yes | UserID, IgnoredUserID | **M3** |
| `UserUnignored` | Yes | UserID, IgnoredUserID | **M3** |

**Permission-related events** (`UserGroupChanged`, `PermissionChecked`)
live in the permission realm SDK — see `05-PERMISSION-SYSTEM.md`.

---

## Edge Cases & Caveats

### Respect System

**Daily limit (3 per type):** Server validates every `user.respect` C2S
by counting today's `user_respects` rows for the actor. If count >= 3, the
packet is silently dropped. The client also tracks remaining locally for UX,
but the server is authoritative.

**Self-respect prevention:** Server rejects `user.respect` where
`actor_user_id == target_id`. Client UI also prevents this.

**Offline target:** Respecting an offline user is valid. The
`respects_received` counter is incremented and the target sees the updated
total on next login. No live notification needed (user is not in a room).

**Race condition (concurrent respect):** Two connections sending
`user.respect` simultaneously for the same actor: the count check and insert
must be atomic. Use a PostgreSQL transaction with `SELECT FOR UPDATE` or a
Redis-based rate limiter for the hot path.

**Rate limiting:** Beyond the 3/day limit, rate-limit `user.respect` to
1 per second per connection to prevent protocol flooding.

**Pet respects:** Same 3/day limit with `target_type = 1`. Server validates
count independently from user respects. Until Room Entities realm is
implemented, pet respect C2S packets are silently dropped (no pets exist).

**Archival:** Records older than 90 days can be periodically deleted. The
`respects_received` materialized counter retains the total permanently.
Archival is optional and not required for correctness.

### Figure Validation

The `figure` string has a specific format: `hr-XXX-YY.hd-XXX-YY.ch-...`.
Vendors do NOT validate figure codes server-side — the client enforces valid
combinations. Server stores whatever the client sends. If future validation
is desired, a plugin can cancel `UserFigureChanged` and reject invalid
figures.

### Motto Length

Vendors cap motto at 127 characters. The server must truncate before
persisting. Some vendors also strip HTML tags.

### Name Change Flow

1. Client sends `user.check_name` → server returns `user.check_name_result`
   with result code (0=available, 1=taken, 2=invalid)
2. Client sends `user.change_name` → server applies change, returns
   `user.change_name_result`, broadcasts `user.name_change` to room
3. `can_change_name` is set to `false` after successful change
4. Admin API can re-enable `can_change_name`

### Wardrobe Slot Limits

Slot limits are resolved from the user's permission group `club_level`:
- `club_level = 0` → 5 slots
- `club_level = 1` → 10 slots
- `club_level = 2` → 20 slots

Server enforces this on `user.save_wardrobe_outfit`. See
`05-PERMISSION-SYSTEM.md` for how `club_level` is resolved from groups.

### Permissions & Perks

**Removed from this realm.** The `user.permissions` (ID 411) and
`user.perks` (ID 2586) packets are owned by the permission realm. They
are resolved from the user's permission group via `group_id` FK. See
`05-PERMISSION-SYSTEM.md` for full details including perk-to-permission
mapping, wildcard resolution, and default group seeds.

### Ignore List

Maximum size varies by vendor (50-100 entries). We use 100. The ignore list
is loaded once at login and kept in-memory per-session by the client.
Server-side we only need to persist and query. Ignore checks for chat/friend
requests are done by social/room realms (DEFERRED).

---

## What Gets Deferred

| Feature | Reason | Depends On |
|---------|--------|------------|
| Relationships (`user.relationship_status`) | Requires Messenger & Social realm | Messenger |
| Tags (`user.tags`, `user.get_tags`) | Room-entity feature | Room Entities |
| Classification (`user.classification`, `user.set_classification`) | Friends list feature | Messenger |
| NUX (`user.nux_proceed`, `user.nux_get_gifts`) | Requires reward/furniture system | Catalog |
| Welcome Gift (`user.welcome_gift_*`) | Requires email + furniture system | Catalog |
| Email (`user.get_email_status`, `user.change_email`) | External email integration | Infrastructure |
| Effects (`user.effect_enable`) | Requires subscription + inventory | Subscription |
| Safety Lock (`user.safety_lock_status`) | Parental controls | Infrastructure |
| Profile Groups (inside `user.profile`) | Requires Groups realm | Groups |
| Profile Friends Count (inside `user.profile`) | Requires Messenger realm | Messenger |
| Profile Badge Display | Requires Achievement realm | Achievements |
| In-Client Link (`user.in_client_link`) | Marketing feature | Infrastructure |

---

## Optimizations

### Redis Caching for Hot User Data

User info is read on every `/navigate` and every room entry. Cache the core
identity (figure, motto, group info) in Redis with 5-minute TTL. Invalidate
on write. Pattern: `user:info:{userID}` → JSON blob.

### Respect Remaining Fast Path

Computing `3 - COUNT(*)` on every `user.info` request adds one SQL query.
For hot-path performance, cache today's respect count per actor in Redis:
`user:respects:{userID}:{date}` → INT, TTL 24h. Increment on respect,
auto-expires at midnight.

### Settings Write Batching

Client sends `user.settings_volume` on every slider movement. Vendors
debounce server-side with a 2-second coalesce window before writing to
PostgreSQL. We achieve this by storing pending writes in-memory on the
session and flushing on interval or disconnect.

### Lazy Profile Loading

`user.profile` is expensive (groups, friends count, achievements). Load only
the fields we have. For deferred fields (groups, friends), return 0 counts
and empty arrays until those realms are implemented.

---

## Implementation Roadmap

### Milestone 1: Core Identity & Post-Auth

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 1 | Extend `users` table with identity fields + `group_id` FK (migration) | 05-M1 | DONE (group_id column added; FK deferred until permission schema finalization) |
| 2 | Extend `User` domain aggregate with new fields | 1 | DONE |
| 3 | Extend `Record` GORM model with identity columns | 1 | DONE |
| 4 | Extend `Repository` with FindByID returning full user + group | 2 | DONE |
| 5 | Create `user.info` S2C packet (ID 2725) | 2 | DONE |
| 6 | Create `user.noobness_level` S2C packet (ID 3738) | - | DONE |
| 7 | Create `user.banned` S2C packet (ID 1683) | - | DONE |
| 8 | Create `user.get_info` C2S packet (ID 357) | - | DONE |
| 9 | Wire user.info + permissions + perks + noobness into post-auth burst | 5,6 | DONE |
| 10 | API: GET /api/users/{id} | 4 | DONE |
| 11 | CLI: user get | 10 | DONE |
| 12 | Unit + integration tests for M1 | all M1 | DONE |

### Milestone 2: Appearance, Settings & Respects

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 13 | Create `user_settings` table + model + migration | - | DONE |
| 14 | Create `Settings` domain value object | 13 | DONE |
| 15 | Create settings repository | 14 | DONE |
| 16 | Create `user.settings` S2C packet (ID 513) | 14 | DONE |
| 17 | Create `user.home_room` S2C packet (ID 2875) | - | DONE |
| 18 | Create `user.figure` S2C packet (ID 2429) | - | DONE |
| 19 | Create C2S packets: update_motto, update_figure, settings_volume, settings_room_invites, settings_old_chat, set_home_room | - | PARTIAL (room_invites/old_chat IDs still pending protocol-final mapping) |
| 20 | Create `user_respects` table + model + migration | - | DONE |
| 21 | Create `user.respect_received` S2C packet (ID 2815) | - | DONE |
| 22 | Create C2S packet: user.respect (ID 2694) | - | DONE |
| 23 | Implement motto/figure update use case with plugin events | 19 | PARTIAL (use case done, plugin events pending) |
| 24 | Implement settings persistence use case (with write debounce) | 15 | PARTIAL (persistence done, debounce pending) |
| 25 | Implement respect use case: daily limit check, atomic increment | 20,22 | DONE |
| 26 | Compute respectsRemaining/petRemaining from user_respects in user.info | 20 | DONE |
| 27 | Wire settings + home_room into post-auth burst | 16,17 | DONE |
| 28 | API: PATCH /api/users/{id}, GET/PATCH settings, POST respect | 23,24,25 | DONE |
| 29 | CLI: user update, user respect | 28 | DONE |
| 30 | Plugin events: UserMottoChanged, UserFigureChanged, UserRespected | 23,25 | PENDING |
| 31 | Unit + integration tests for M2 | all M2 | DONE |

### Milestone 3: Wardrobe, Ignore List & Profile

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 32 | Create `user_wardrobe_slots` table + model + migration | - | PENDING |
| 33 | Create `user_ignores` table + model + migration | - | PENDING |
| 34 | Create wardrobe repository | 32 | PENDING |
| 35 | Create ignore repository | 33 | PENDING |
| 36 | Create `user.wardrobe_page` S2C + C2S get/save packets | 34 | PENDING |
| 37 | Create `user.ignored_users` S2C + C2S ignore/unignore packets | 35 | PENDING |
| 38 | Wire ignored_users into post-auth burst | 37 | PENDING |
| 39 | Create `user.profile` S2C packet (ID 3898, partial) | - | PENDING |
| 40 | Create `user.get_profile` C2S packet handler (ID 3265) | 39 | PENDING |
| 41 | API: GET /api/users/{id}/wardrobe, GET /api/users/{id}/respects | 34,20 | PENDING |
| 42 | Plugin events: UserIgnored, UserUnignored | 37 | PENDING |
| 43 | Unit + integration tests for M3 | all M3 | PENDING |

### Milestone 4: Name Changes

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 44 | Create `user.check_name` C2S (ID 3950) + `user.check_name_result` S2C (ID 563) | - | PENDING |
| 45 | Create `user.change_name` C2S (ID 2977) + `user.change_name_result` S2C (ID 118) | - | PENDING |
| 46 | Create `user.name_change` S2C broadcast packet (ID 2182) | - | PENDING |
| 47 | Implement name validation use case (length, chars, uniqueness) | 44 | PENDING |
| 48 | Implement name change use case with `can_change_name` guard | 45,47 | PENDING |
| 49 | Create `user.approve_name` C2S + result for admin approval flow | - | PENDING |
| 50 | API: POST /api/users/{id}/name-change | 48 | PENDING |
| 51 | CLI: user rename | 50 | PENDING |
| 52 | Plugin event: UserNameChanged | 48 | PENDING |
| 53 | Unit + integration tests for M4 | all M4 | PENDING |

---

## Caveats & Technical Notes

### Migration Strategy

Extending the existing `users` table is a non-destructive `ALTER TABLE
ADD COLUMN` with defaults. The `group_id` FK requires the
`permission_groups` table to exist first — this means **05-PERMISSION-SYSTEM
Milestone 1 must complete before 04-USER-PROFILE Milestone 1**.

New tables (`user_settings`, `user_wardrobe_slots`, `user_ignores`,
`user_respects`) are created fresh with no data migration.

### Domain Identity vs Session Identity

The `User` domain aggregate owns persistent identity (figure, motto, etc.).
The `Session` struct in `core/connection` owns transient identity (connID,
state, instanceID). These MUST NOT merge. The session references user by
`UserID` only.

### Post-Auth Burst Integration

Currently the post-auth burst is wired in
`pkg/session/application/postauth/usecase.go`. The user realm packets
(`user.info`, `user.settings`, etc.) must be added to this burst in the
correct order. Permission realm packets (`user.permissions`, `user.perks`)
are also added to the burst but resolved from the permission group.

### GORM Model File Count

The `pkg/user/infrastructure/model/` package currently has 2 files. Adding
`settings.go`, `wardrobe.go`, `ignore.go`, and `respect.go` brings it to 6
(at the limit per AGENTS.md).

### Figure Default

The default figure `hr-115-42.hd-180-1.ch-3030-82.lg-275-82.sh-295-62` is
the Habbo "Frank" starter avatar used by all vendors. Gender defaults to `M`.
