# Realm: Messenger & Social

> **Position:** 40 | **Phase:** 5 (Social) | **Packets:** 30 (14 c2s, 16 s2c)
> **Services:** social | **Status:** Not yet implemented

---

## Overview

The Messenger & Social realm manages friend lists, friend requests, private messaging, room invitations, user search, and online status broadcasting. This realm is uniquely cross-cutting: messenger state must be synchronized across all connected users in real-time, which makes it the most NATS-intensive realm after room entities.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 5

---

## Packet Inventory

### C2S (Client to Server) -- 14 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2781 | `messenger.init` | _(none)_ | Initialize messenger, request friend list |
| 3567 | `messenger.chat` | `userId:int32`, `message:string` | Send private message to friend |
| 3157 | `messenger.request_friend` | `username:string` | Send friend request |
| 137 | `messenger.accept_friend` | `count:int32`, `userIds:int32[]` | Accept one or more friend requests |
| 2890 | `messenger.decline_friend` | `removeAll:boolean`, `count:int32`, `userIds:int32[]` | Decline friend requests |
| 1689 | `messenger.remove_friend` | `count:int32`, `userIds:int32[]` | Remove friends |
| 3997 | `messenger.follow_friend` | `userId:int32` | Follow friend to their room |
| 1210 | `messenger.search` | `query:string` | Search for users by name |
| 1276 | `messenger.room_invite` | `count:int32`, `userIds:int32[]`, `message:string` | Invite friends to current room |
| 2448 | `messenger.get_requests` | _(none)_ | Request pending friend request list |
| 516 | `messenger.find_new_friends` | _(none)_ | Find random rooms with users |
| 1419 | `messenger.refresh` | _(none)_ | Force-refresh friend list state |
| 1148 | `messenger.friend_request_quest_complete` | _(none)_ | Mark friend-request quest as done |
| 1523 | `messenger.messenger_friends` | _(none)_ | Alternative friend list request |

### S2C (Server to Client) -- 16 packets

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 1605 | `messenger.init` | `userFriendLimit`, `normalFriendLimit`, `extendedFriendLimit`, `categories[]` | Messenger initialization with limits |
| 3130 | `messenger.friends` | `friends[]` (id, username, figure, online, inRoom, motto, categoryId, lastAccess, realName, isMobileOnline) | Full friend list |
| 2800 | `messenger.update` | `categories[]`, `updates[]` | Friend status updates (online/offline/room change) |
| 2219 | `messenger.friend_request` | `userId`, `username`, `figure` | Incoming friend request notification |
| 280 | `messenger.friend_requests` | `requests[]` | Pending friend requests list |
| 1587 | `messenger.chat` | `userId`, `message`, `timestamp`, `extra` | Received private message |
| 896 | `messenger.accept_result` | `failures[]` | Accept request result (list of failed IDs) |
| 973 | `messenger.search_result` | `friends[]`, `strangers[]` | User search results split into friends/non-friends |
| 3870 | `messenger.room_invite` | `userId`, `message` | Room invitation received |
| 462 | `messenger.room_invite_error` | `errorCode:int32`, `userIds:int32[]` | Room invite failure |
| 892 | `messenger.message_error` | `errorCode:int32`, `userId:int32` | Message delivery failure |
| 3359 | `messenger.instant_message_error` | `errorCode:int32`, `userId:int32` | Instant message error (extended) |
| 1210 | `messenger.find_friends_result` | `results[]` | Find-new-friends room results |
| 3048 | `messenger.follow_failed` | `errorCode:int32` | Follow friend to room failed |
| 2803 | `messenger.minimail_count` | `count:int32` | Unread minimail count |
| 1911 | `messenger.minimail_new` | _(none)_ | New minimail notification |

---

## Architecture Mapping

### Service Ownership

```
Client ──packet──▶ Gateway ──NATS(social.input.<sid>)──▶ Social Service
                                                              │
                        ┌─────NATS(session.output.<sid>)──────┘
                        ▼
                   Gateway ──▶ Recipient Client
```

The **social service** owns all messenger logic:
- Friend list CRUD
- Message routing
- Status broadcasting
- Search queries

### Database Tables

| Table | Columns (Key) | Usage |
|-------|---------------|-------|
| `messenger_friendships` | user_id, friend_id, created_at | Bidirectional friendship links |
| `messenger_requests` | sender_id, receiver_id, created_at | Pending friend requests |
| `messenger_messages` | sender_id, receiver_id, message, timestamp, read | Message history |
| `messenger_categories` | user_id, category_id, name | Friend list categories (custom folders) |

### Redis Keys

| Key Pattern | Usage |
|-------------|-------|
| `user:online:<userId>` | Online status flag (SET with TTL) |
| `user:room:<userId>` | Current room ID (0 = lobby) |
| `friends:<userId>` | Cached friend ID set for fast lookups |

### NATS Subjects

| Subject | Direction | Purpose |
|---------|-----------|---------|
| `social.input.<sessionID>` | gateway -> social | Incoming messenger packets |
| `session.output.<sessionID>` | social -> gateway | Outgoing responses |
| `social.status.<userID>` | social -> social | Online/offline/room-change broadcasts |
| `social.notification.<userID>` | social -> gateway | Push notifications to specific user |

---

## Implementation Analysis

### Friend List Initialization

On `messenger.init` (2781):
1. Load `messenger_friendships` for the user.
2. For each friend, check Redis `user:online:<friendId>` for online status.
3. For online friends, check Redis `user:room:<friendId>` for current room.
4. Build `messenger.init` (1605) with friend limits and categories.
5. Build `messenger.friends` (3130) with full friend list.

**Performance concern:** Users with 500+ friends cause expensive initialization. Strategy:
- Cache friend list in Redis (`friends:<userId>` as sorted set).
- Batch Redis MGET for online status checks.
- Load friend details from PostgreSQL in a single `WHERE id IN (...)` query.

### Real-Time Status Broadcasting

When a user comes online, goes offline, or changes rooms, the social service must notify all online friends:

```
User A goes online:
  1. Social service receives session.authenticated event
  2. Load friend list for User A from Redis
  3. For each friend who is online:
     - Publish messenger.update (2800) to their session.output
  4. Update Redis user:online:<A> = true
```

**Caveat from Comet v2:** Broadcasting to all friends on every status change is O(n) per event. For users with many friends, this creates NATS message fan-out. Mitigate by:
- Batching status updates (collect changes over 1-second windows, send once).
- Only broadcasting to friends whose messenger is initialized.

### Private Messaging

`messenger.chat` (3567) sends a message to a specific user:

1. Validate sender and receiver are friends.
2. Check if receiver has sender ignored.
3. Check if receiver is online (Redis lookup).
4. **If online:** Route `messenger.chat` (1587) to receiver's `session.output.<sid>`.
5. **If offline:** Store in `messenger_messages` table for delivery on next login.
6. **If blocked/not friends:** Send `messenger.message_error` (892) to sender.

**Offline delivery is critical.** Legacy emulators (Comet v2) often skip this -- messages to offline users are silently dropped. pixel-server must persist offline messages and deliver them during messenger initialization.

### Room Invitations

`messenger.room_invite` (1276) sends an invitation to multiple friends:

1. Validate sender is currently in a room.
2. For each target user:
   - Validate friendship exists.
   - Check if target has `blockInvites` setting enabled.
   - If target is online: send `messenger.room_invite` (3870).
   - If target is offline or blocking: add to error list.
3. If any failures: send `messenger.room_invite_error` (462) with failed user IDs.

**Error codes for room_invite_error:**
- 1: User offline
- 2: User blocking invites
- 3: Not friends

### Follow Friend to Room

`messenger.follow_friend` (3997):
1. Look up friend's current room from Redis `user:room:<friendId>`.
2. If friend is not in a room: send `messenger.follow_failed` (3048) with error code 1.
3. If room is locked/full: send follow_failed with appropriate code.
4. If valid: trigger room entry for the follower (delegate to game service).

### User Search

`messenger.search` (1210) searches by username prefix:
1. Query PostgreSQL: `SELECT id, username, figure, motto FROM users WHERE LOWER(username) LIKE LOWER($1 || '%') LIMIT 50`.
2. Split results into two lists: users who are friends, and users who are not.
3. Send `messenger.search_result` (973) with both lists.

**Performance:** Add a GIN trigram index on `users.username` for fast prefix search. Consider a separate search index if user count exceeds 100K.

---

## Caveats & Edge Cases

### 1. Friend Request Spam
No rate limiting on `messenger.request_friend` allows spam. Implement:
- Maximum 10 outgoing pending requests per user.
- Cooldown of 30 seconds between requests to the same user.
- Maximum 50 total pending requests (incoming + outgoing).

### 2. Mutual Friendship Consistency
Friendships are bidirectional: if A is friends with B, B must be friends with A. The `messenger_friendships` table should store both rows (A->B and B->A) in a single transaction to prevent half-friendships.

### 3. Batch Accept/Decline
`messenger.accept_friend` (137) and `messenger.decline_friend` (2890) accept arrays of user IDs. The handler must process each atomically -- partial failures are reported via `messenger.accept_result` (896). Use a database transaction that commits all or rolls back.

### 4. Message Ordering
Private messages must maintain causal ordering. Use PostgreSQL `SERIAL` or timestamp ordering. The client displays messages in order of `timestamp` from `messenger.chat` (1587).

### 5. Offline Message Limits
Without a cap, a user could accumulate thousands of offline messages. Limit to 100 unread messages per sender. After that, new messages from the same sender are rejected with `messenger.message_error`.

### 6. Status Update Storms
When a gateway instance with 500 users crashes, all 500 users go offline simultaneously. Each triggers status broadcasts to all their friends. This can create a thundering herd of NATS messages. Mitigate by:
- Batching offline events per gateway instance.
- Using a short delay (1-2 seconds) before broadcasting offline status (handles reconnection).

### 7. Follow to Private Rooms
`messenger.follow_friend` should respect room access settings:
- Open rooms: always allowed.
- Locked rooms: only if follower has rights.
- Password rooms: not allowed via follow (send error).
- Invisible rooms: not allowed via follow.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Offline messaging** | Messages dropped silently | Persistent storage with delivery on login |
| **Status broadcasting** | O(n) per event, no batching | Batched 1-second windows, NATS fan-out |
| **Friend list loading** | Sequential DB queries | Redis cache + batch MGET for online status |
| **Search** | LIKE query, no index | GIN trigram index, result capping |
| **Rate limiting** | None | Request cooldowns, pending limits |
| **Cross-service** | In-process calls (single server) | NATS-based; social service scales independently |
| **Consistency** | Single-row friendship | Bidirectional rows in transaction |

---

## Dependencies

- **Phase 2 (Identity)** -- user data (username, figure) required for friend list display
- **Phase 3 (Room)** -- room entry required for follow-to-room and invitations
- **pkg/social** -- domain models (Friendship, FriendRequest, Message)
- **PostgreSQL** -- friendship, request, message tables
- **Redis** -- online status, room presence, friend cache

---

## Testing Strategy

### Unit Tests
- Friend request validation (self-request, duplicate, limit)
- Message routing logic (online/offline paths)
- Search result splitting (friends vs strangers)
- Room invite error code generation
- Batch accept/decline partial failure handling

### Integration Tests
- Friendship CRUD against real PostgreSQL
- Offline message persistence and retrieval
- Redis online status set/get with TTL
- Status broadcasting to multiple friends
- Search with trigram index

### E2E Tests
- Two clients: A sends friend request to B, B accepts, both see updated friend lists
- A sends private message to B, B receives it in real-time
- A goes offline, B sends message, A comes online and receives it
- A invites B to room, B receives invitation and can enter
