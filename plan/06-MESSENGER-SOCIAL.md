# 06 - Messenger & Social Realm

## Overview

The Messenger & Social realm owns the friend list, friend requests,
private messaging, room invites, friend relationships, search, follow-to-room,
and friend notifications. It is the primary social backbone between players.

The pixel-protocol references **14 C2S packets** and **16 S2C packets** for
this realm. All four vendor emulators implement the core messenger flow.
PlusEMU is the only vendor that persists offline messages.

---

## Vendor Cross-Reference

### Messenger Feature Matrix

| Feature | PlusEMU (C#) | Gladiator/Arcturus (Java) | comet-v2 (Java) | pixelsv (proposed) |
|---------|-------------|--------------------------|-----------------|-------------------|
| Offline messages | Persistent DB | Dropped | Dropped | **Persistent DB** |
| Relationships location | friendship table (migrated) | `relation` column | separate `player_relationships` | **friendship table** |
| Relationship values | 0–3 int | 0–3 short | HEART/SMILE/BOBBA/POOP enum | **Extendable registry** |
| Friend categories | Hardcoded 0 | DB-driven | 1 if group chat enabled | **Hardcoded 0** (defer categories) |
| Flood control | 12/5s window, 1m mute | 750ms cooldown | 750ms + progressive mute | **750ms + progressive mute** |
| Batch accept cap | 50 | Uncapped | Uncapped | **50** |
| Batch remove cap | 100 | Uncapped | Uncapped | **100** |
| Friend list fragmentation | 750/page | 750/page | 1 page (broken) | **750/page** |
| Auto-accept cross-request | Yes | No | No | **Yes** |
| Unfriend DB cleanup | Both directions | Both directions (OR) | One direction only (bug) | **Both directions** |
| Friend status propagation | On connect/disconnect | On connect/disconnect | On connect/disconnect | **Via Redis Pub/Sub** |
| Word filtering in messages | Yes | Configurable | Yes | **Yes** (via permission) |
| Staff/group chat | No | Group chat | Staff/Log/Alfa chat | **Defer** |

### Database Schema Comparison

| Aspect | PlusEMU | Arcturus | comet-v2 | pixelsv |
|--------|---------|----------|----------|---------|
| Friendship table | `messenger_friendships` | `messenger_friendships` | `messenger_friendships` | `messenger_friendships` |
| Friendship PK | `(user_one_id, user_two_id)` | `(user_one_id, user_two_id)` | `(user_one_id, user_two_id)` | `(user_one_id, user_two_id)` |
| Request table | `messenger_requests` | `messenger_friendrequests` | `messenger_requests` | `friend_requests` |
| Offline messages | `messenger_offline_messages` | None | None | `offline_messages` |
| Relationships | `relationship` col in friendship | `relation` col in friendship | `player_relationships` table | `relationship` col in friendship |

### Our Improvements Over Vendors

1. **Redis Pub/Sub for status propagation** — vendors iterate all online friends
   in-process. We fan out across instances via targeted channels.
2. **Offline message persistence** — only PlusEMU does this. All other vendors
   silently drop messages to offline users.
3. **Normalized relationships** — stored in the friendship table (like PlusEMU
   and Arcturus), not a separate table (avoids comet-v2's orphan relationship
   problem).
4. **Proper both-direction unfriend** — comet-v2 only deletes one row. We
   delete both atomically.
5. **Configurable flood control** — progressive mute like comet-v2, but with
   configurable thresholds via permission-based bypass.
6. **Auto-accept cross-requests** — if A requests B and B already requested A,
   accept immediately (PlusEMU behavior, cleaner UX).

---

## Packet Registry

### Client-to-Server (14 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 2781 | `messenger.init` | (empty) | **M1** |
| 1523 | `messenger.get_friends` | (empty) | **M1** |
| 2448 | `messenger.get_requests` | (empty) | **M1** |
| 3157 | `messenger.send_request` | username (string) | **M1** |
| 137 | `messenger.accept_friend` | count (int32), requestId (int32) × count | **M1** |
| 2890 | `messenger.decline_friend` | declineAll (bool), count (int32), requestId (int32) × count | **M1** |
| 1689 | `messenger.remove_friend` | count (int32), userId (int32) × count | **M1** |
| 3567 | `messenger.send_msg` | userId (int32), message (string) | **M1** |
| 1276 | `messenger.send_invite` | count (int32), userId (int32) × count, message (string) | **M2** |
| 3997 | `messenger.follow_friend` | friendId (int32) | **M2** |
| 1210 | `messenger.search` | query (string) | **M2** |
| 3768 | `messenger.set_relationship` | userId (int32), type (int32) | **M2** |
| 2138 | `messenger.get_relationships` | userId (int32) | **M2** |
| 516 | `messenger.find_new_friends` | (empty) | **DEFER** |

### Server-to-Client (16 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 1605 | `messenger.init` | userFriendLimit, normalLimit, extendedLimit, categoryCount | **M1** |
| 3130 | `messenger.friends` | totalFragments, fragmentNumber, friendCount, friendRecords… | **M1** |
| 280 | `messenger.requests` | totalRequests, count, [reqId, username, figure]… | **M1** |
| 2219 | `messenger.new_request` | requestId, username, figure | **M1** |
| 2800 | `messenger.friend_update` | categoryCount, categories…, updateCount, [action, record]… | **M1** |
| 1587 | `messenger.new_message` | senderId, message, secondsSinceSent | **M1** |
| 3359 | `messenger.message_error` | errorCode, userId, message | **M1** |
| 892 | `messenger.request_error` | clientMsgId, errorCode | **M1** |
| 896 | `messenger.accept_error` | count, [senderId, errorCode]… | **M1** |
| 973 | `messenger.search_result` | friendCount, friends…, otherCount, others… | **M2** |
| 3870 | `messenger.room_invite` | senderId, message | **M2** |
| 462 | `messenger.invite_error` | errorCode, failedCount, userId × count | **M2** |
| 3082 | `messenger.friend_notification` | friendId (string), typeCode, data | **M2** |
| 3048 | `messenger.follow_failed` | errorCode | **M2** |
| 2016 | `messenger.relationships` | userId, count, [type, count, friendId, username, figure]… | **M2** |
| 1210 | `messenger.find_friends_result` | success (bool) | **DEFER** |

### Friend Record Wire Format

Used in `messenger.friends` (3130) and `messenger.friend_update` (2800):

| Field | Type | Description |
|-------|------|-------------|
| `id` | int32 | User ID |
| `username` | string | Display name |
| `gender` | int32 | 0 = male, 1 = female |
| `online` | bool | Currently connected |
| `followingAllowed` | bool | Is in a room (can follow) |
| `figure` | string | Avatar appearance (empty if offline) |
| `categoryId` | int32 | Friend category (0 default) |
| `motto` | string | Player motto |
| `realName` | string | Real name (empty) |
| `lastAccess` | string | Last access info (empty) |
| `persistedMessageUser` | bool | Offline messaging capable |
| `vipMember` | bool | VIP status |
| `pocketHabboUser` | bool | Mobile user |
| `relationshipStatus` | int16 | 0=none, 1=heart, 2=smile, 3=bobba |

### Friend Update Action Codes

| Code | Meaning | Payload |
|------|---------|---------|
| -1 | Removed | friendId only |
| 0 | Updated | Full friend record |
| 1 | Added | Full friend record |

### Search Result Wire Format

| Field | Type | Description |
|-------|------|-------------|
| `avatarId` | int32 | User ID |
| `avatarName` | string | Username |
| `avatarMotto` | string | Motto |
| `isOnline` | bool | Currently online |
| `canFollow` | bool | In a room |
| `lastOnlineData` | string | Last access info |
| `avatarGender` | int32 | Gender |
| `avatarFigure` | string | Figure |
| `realName` | string | Real name |

Results are split into two lists: friends first, then non-friends.

---

## Database Model Design

### Table 1: `messenger_friendships`

Bidirectional friendship storage. Each friendship creates TWO rows (A→B
and B→A) so that all queries filter by `user_one_id` only, enabling
single-column index scans.

```
messenger_friendships
├── user_one_id     INT NOT NULL (FK → users.id)
├── user_two_id     INT NOT NULL (FK → users.id)
├── relationship    SMALLINT NOT NULL DEFAULT 0        -- 0=none, 1=heart, 2=smile, 3=bobba
├── created_at      TIMESTAMP NOT NULL DEFAULT NOW()
├── PRIMARY KEY(user_one_id, user_two_id)
├── INDEX(user_two_id)
```

**Constraint:** relationship values are validated against the `KnownRelationships` registry
in the domain layer. Plugins extend the registry at startup via `RegisterRelationship`.
Relationships are asymmetric — A can mark B as heart while B marks A as smile.

### Table 2: `friend_requests`

Pending friend requests. Deleted on accept or decline.

```
friend_requests
├── id              BIGINT PK AUTO
├── from_user_id    INT NOT NULL (FK → users.id)
├── to_user_id      INT NOT NULL (FK → users.id)
├── created_at      TIMESTAMP NOT NULL DEFAULT NOW()
├── UNIQUE(from_user_id, to_user_id)
├── INDEX(to_user_id)
```

### Table 3: `offline_messages`

Messages sent to offline users. Delivered and deleted atomically on next
login (PlusEMU pattern).

```
offline_messages
├── id              BIGINT PK AUTO
├── from_user_id    INT NOT NULL (FK → users.id)
├── to_user_id      INT NOT NULL (FK → users.id, INDEX)
├── message         VARCHAR(255) NOT NULL
├── sent_at         TIMESTAMP NOT NULL DEFAULT NOW()
```

**Cleanup:** Messages older than 30 days are purged by a periodic job
(configurable via `MESSENGER_OFFLINE_MESSAGE_TTL_DAYS`).

---

## Hexagonal Architecture

### Package Layout

```
pkg/messenger/
├── domain/
│   ├── friendship.go          -- Friendship, FriendRequest, OfflineMessage entities
│   ├── relationship.go        -- RelationshipType enum, validation
│   ├── repository.go          -- Repository interface
│   └── errors.go              -- Domain errors
├── application/
│   ├── service.go             -- Service struct, constructor
│   ├── friends.go             -- AddFriend, RemoveFriend, ListFriends
│   ├── requests.go            -- SendRequest, AcceptRequest, DeclineRequest
│   ├── messaging.go           -- SendMessage, DeliverOffline
│   └── social.go              -- Search, Follow, Relationships, Invites
├── adapter/
│   ├── realtime/
│   │   └── runtime.go         -- Packet handler dispatch
│   ├── httpapi/
│   │   ├── contracts.go       -- Service interface for HTTP
│   │   ├── friend_routes.go   -- Admin friend management
│   │   └── openapi.go         -- OpenAPI specs
│   └── command/
│       ├── command.go          -- CLI root
│       └── actions.go          -- CLI subcommands
├── infrastructure/
│   ├── model/
│   │   └── models.go          -- GORM models
│   ├── store/
│   │   └── repository.go      -- PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go      -- Schema migrations
│   └── seed/
│       └── seeds.go           -- Test data
└── stage.go                    -- Module bootstrap
```

### Domain Entities

```go
type Friendship struct {
    UserOneID    int
    UserTwoID    int
    Relationship RelationshipType
    CreatedAt    time.Time
}

type FriendRequest struct {
    ID         int
    FromUserID int
    ToUserID   int
    CreatedAt  time.Time
}

type OfflineMessage struct {
    ID         int
    FromUserID int
    ToUserID   int
    Message    string
    SentAt     time.Time
}

type RelationshipType int
const (
    RelationshipNone  RelationshipType = 0
    RelationshipHeart RelationshipType = 1
    RelationshipSmile RelationshipType = 2
    RelationshipBobba RelationshipType = 3
)

// KnownRelationships maps all registered relationship types to their labels.
// Plugins call RegisterRelationship to extend the set at startup.
var KnownRelationships = map[RelationshipType]string{
    RelationshipNone:  "none",
    RelationshipHeart: "heart",
    RelationshipSmile: "smile",
    RelationshipBobba: "bobba",
}

func RegisterRelationship(t RelationshipType, label string) { KnownRelationships[t] = label }
func IsValidRelationship(t RelationshipType) bool { _, ok := KnownRelationships[t]; return ok }
```

### Domain Errors

```go
var (
    ErrFriendListFull         = errors.New("friend list is full")
    ErrTargetFriendListFull   = errors.New("target friend list is full")
    ErrTargetNotAccepting     = errors.New("target not accepting requests")
    ErrTargetNotFound         = errors.New("target user not found")
    ErrAlreadyFriends         = errors.New("already friends")
    ErrNotFriends             = errors.New("not friends")
    ErrRequestNotFound        = errors.New("friend request not found")
    ErrSelfRequest            = errors.New("cannot send friend request to self")
    ErrSenderMuted            = errors.New("sender is muted")
    ErrRecipientOffline       = errors.New("recipient is offline")
    ErrInvalidRelationship    = errors.New("invalid relationship type")
    ErrFriendNotInRoom        = errors.New("friend not in a room")
    ErrFollowBlocked          = errors.New("friend blocked following")
)
```

---

## API Endpoints

### Admin REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/{id}/friends` | List user's friends |
| `POST` | `/api/v1/users/{id}/friends` | Force-add friendship |
| `DELETE` | `/api/v1/users/{id}/friends/{friendId}` | Force-remove friendship |
| `GET` | `/api/v1/users/{id}/friends/requests` | List pending requests |
| `GET` | `/api/v1/users/{id}/friends/{friendId}/relationship` | Get relationship |
| `PATCH` | `/api/v1/users/{id}/friends/{friendId}/relationship` | Set relationship |

### CLI Commands

```bash
pixelsv messenger friends list 1              # List user 1's friends
pixelsv messenger friends add 1 2             # Force friendship between 1 and 2
pixelsv messenger friends remove 1 2          # Remove friendship
pixelsv messenger requests list 1             # List pending requests for user 1
pixelsv messenger relationship set 1 2 heart  # Set relationship
```

---

## Plugin Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `FriendRequestSent` | **Yes** | FromUserID, ToUserID, ToUsername |
| `FriendRequestAccepted` | **Yes** | UserID, FriendUserID |
| `FriendRemoved` | **Yes** | UserID, FriendUserID |
| `PrivateMessageSent` | **Yes** | FromConnID, FromUserID, ToUserID, Message |
| `RoomInviteSent` | **Yes** | FromConnID, FromUserID, ToUserIDs, Message |
| `RelationshipChanged` | **Yes** | UserID, FriendUserID, OldType, NewType |
| `FriendFollowed` | No | UserID, FriendUserID, RoomID |

All cancellable events roll back the operation if cancelled. Plugins can
use these for logging, custom restrictions, or cross-system integrations.

---

## Cross-Instance Communication

### Friend Status Propagation

When a user connects or disconnects, all their friends must be notified.
Vendors iterate friends in-process. Our approach uses Redis Pub/Sub:

```
User A connects (instance 1)
  │
  ├── Load friend list from DB
  │
  ├── For each online friend:
  │   └── Publish to friend's notification channel:
  │       chan:user:{friendId} → FriendListUpdateComposer (action=0, updated record)
  │
  └── Friend's instance receives message → sends to friend's connection
```

Each user has a notification channel: `chan:user:{userId}`. When a
`FriendListUpdateComposer` arrives on this channel, the instance holding
that user's connection forwards it to the WebSocket.

### Private Message Routing

Messages to online friends are routed via the notification channel:

```
User A sends message to User B
  │
  ├── Check if B is online (session registry)
  │
  ├── If online:
  │   └── Publish to chan:user:{B.userId} → NewConsoleMessageComposer
  │
  └── If offline:
      └── Store in offline_messages table
```

### Room Invite Routing

Same pattern as private messages but sent to multiple recipients.

---

## Flood Control

### Message Rate Limiting

Progressive flood control (adapted from comet-v2):

| Threshold | Action |
|-----------|--------|
| < 750ms between messages | Silently dropped |
| 4 violations within window | Mute for `MESSENGER_FLOOD_MUTE_SECONDS` (default 20s) |
| During mute | All messages return error code 4 (sender muted) |

Staff with `messenger.flood_bypass` permission skip flood checks.

### Batch Operation Caps

| Operation | Max per packet |
|-----------|---------------|
| Accept friend requests | 50 |
| Remove friends | 100 |
| Room invite recipients | 50 |

### Message Length

Messages are truncated at 255 characters (matching all vendor
implementations).

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MESSENGER_MAX_FRIENDS` | 200 | Normal friend limit |
| `MESSENGER_MAX_FRIENDS_VIP` | 500 | VIP friend limit |
| `MESSENGER_FLOOD_COOLDOWN_MS` | 750 | Min interval between messages |
| `MESSENGER_FLOOD_VIOLATIONS` | 4 | Violations before mute |
| `MESSENGER_FLOOD_MUTE_SECONDS` | 20 | Mute duration after flood |
| `MESSENGER_OFFLINE_MESSAGE_TTL_DAYS` | 30 | Offline message retention |
| `MESSENGER_FRAGMENT_SIZE` | 750 | Friends per list fragment |

---

## Edge Cases

### Auto-Accept Cross-Request

If User A sends a request to User B, and B already has a pending request
from A, the system auto-accepts both directions instead of creating a
duplicate request. This follows PlusEMU behavior.

### Decline All

The `messenger.decline_friend` packet has a `declineAll` boolean. When
true, all pending requests are declined without reading individual IDs.

### Relationship Asymmetry

Relationships are per-direction. User A can mark B as heart while B marks
A as smile. The `messenger.relationships` packet (2016) for profile view
returns grouped counts per type with a random friend sample.

### Offline Message Delivery

On login, offline messages are loaded and deleted atomically. Each message
is delivered as `messenger.new_message` (1587) with `secondsSinceSent`
computed from `NOW() - sent_at`, not stored as a static value.

### Friend Notification String ID

The `messenger.friend_notification` (3082) packet serializes the friend ID
as a **string**, not an int. This is consistent across all vendors.

### Follow-to-Room Error Codes

| Code | Meaning |
|------|---------|
| 0 | Not in friend list |
| 1 | Friend is offline |
| 2 | Friend not in a room |
| 3 | Friend blocked following |

### Message Error Codes

| Code | Meaning |
|------|---------|
| 3 | Recipient muted messages |
| 4 | Sender is muted (flood) |
| 5 | Recipient offline (no offline storage) |
| 6 | Not friends |
| 7 | Recipient busy (DND) |
| 10 | Failed to store offline message |

### Request Error Codes

| Code | Meaning |
|------|---------|
| 1 | Own friend list full |
| 2 | Target friend list full |
| 3 | Target not accepting requests |
| 4 | Target not found |

---

## Deferred Items

| Feature | Reason | Dependency |
|---------|--------|------------|
| Friend categories | Low priority, only Arcturus supports DB categories | None |
| Staff/group chat | Complex, requires room-like group channels | Room realm |
| `messenger.find_new_friends` (516) | Requires room navigator/population data | Room realm |
| MiniMail (1911, 2803) | Legacy feature, low usage | None |
| Lovelock furni (382, 3775) | Furniture realm | Room + Furniture realms |
| Friend quest integration | Quest system | Quest realm |
| Word filtering | Relies on content filter module | Moderation realm |

---

## Optimizations

### Connection-Level Friend Cache

On `messenger.init`, load the full friend list into an in-memory map on
the connection. Use this for O(1) friendship checks (e.g., follow
permission, message routing) without hitting the database.

### Batch Redis Publishes

When a user with many friends connects, batch the status update publishes
using Redis pipelines instead of individual PUBLISH commands.

### Offline Message Cleanup Job

A periodic goroutine deletes messages older than the configured TTL.
Runs once per hour with a configurable interval. Uses `DELETE ... WHERE
sent_at < NOW() - INTERVAL ? DAYS LIMIT 1000` to avoid long-running
transactions.

---

## Implementation Roadmap

### Milestone 1: Core Messenger

| # | Task | Depends On | Status |
|---|------|-----------|--------|
| 1 | Domain model: entities, repository interface, errors | - | ✅ DONE |
| 2 | Database tables + migrations | 1 | ✅ DONE |
| 3 | Repository implementation (PostgreSQL) | 1, 2 | ✅ DONE |
| 4 | Application service: friend CRUD, requests, messaging | 1, 3 | ✅ DONE |
| 5 | `messenger.init` → `messenger.friends` + `messenger.requests` flow | 4 | ✅ DONE |
| 6 | `messenger.send_request` → `messenger.new_request` flow | 4 | ✅ DONE |
| 7 | `messenger.accept_friend` / `messenger.decline_friend` flows | 4 | ✅ DONE |
| 8 | `messenger.remove_friend` flow | 4 | ✅ DONE |
| 9 | `messenger.send_msg` → `messenger.new_message` + offline storage | 4 | ✅ DONE |
| 10 | `messenger.friend_update` on connect/disconnect via Redis Pub/Sub | 4, 5 | ✅ DONE |
| 11 | Plugin events for all operations | 4 | ✅ DONE |
| 12 | Flood control implementation | 4, 9 | ✅ DONE |

### Milestone 2: Social Features

| # | Task | Depends On | Status |
|---|------|-----------|--------|
| 1 | `messenger.search` → `messenger.search_result` flow | M1 | ✅ DONE |
| 2 | `messenger.follow_friend` → room navigation or `messenger.follow_failed` | M1 | ✅ DONE |
| 3 | `messenger.send_invite` → `messenger.room_invite` routing | M1 | ✅ DONE |
| 4 | `messenger.set_relationship` / `messenger.get_relationships` flows | M1 | ✅ DONE |
| 5 | `messenger.friend_notification` for various notification types | M1 | ✅ DONE |
| 6 | Friend list fragmentation (750 per fragment) | M1.5 | ✅ DONE |
| 7 | Admin REST API endpoints | M1 | ✅ DONE |
| 8 | Admin CLI commands | M1 | ✅ DONE |
| 9 | OpenAPI specifications | 7 | ✅ DONE |
| 10 | Offline message delivery on login | M1.9 | ✅ DONE |
| 11 | E2E tests for all flows | M1, M2.1–10 | ✅ DONE |
