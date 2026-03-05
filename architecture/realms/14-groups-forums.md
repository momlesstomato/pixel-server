# Realm: Groups & Forums

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 130 | **Phase:** 10 (Groups) | **Packets:** 64 (36 c2s, 28 s2c)
> **Services:** social (groups), game (badge display) | **Status:** Not yet implemented

---

## Overview

Groups & Forums is the fourth-largest realm at 64 packets. It covers guild creation, membership management, group badges, group home rooms, forum threads, forum posts, and forum moderation. Groups are a deeply cross-cutting feature: they affect profile display, room badges, navigator filtering, and social interactions.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 10

---

## Packet Inventory

### C2S -- 36 packets

#### Group Management

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 230 | `group.create` | name, description, roomId, colorA, colorB, badge parts | Create group |
| 1004 | `group.get_info` | `groupId:int32`, `newWindow:boolean` | Get group details |
| 2991 | `group.get_members` | `groupId`, `page`, `query`, `filter` | Get member list (paginated) |
| 1203 | `group.join` | `groupId:int32` | Request to join group |
| 641 | `group.leave` | `groupId:int32` | Leave group |
| 3032 | `group.accept_member` | `groupId`, `userId` | Accept join request |
| 1894 | `group.decline_member` | `groupId`, `userId` | Decline join request |
| 3620 | `group.remove_member` | `groupId`, `userId` | Remove member |
| 2528 | `group.promote_member` | `groupId`, `userId` | Promote to admin |
| 722 | `group.demote_member` | `groupId`, `userId` | Demote from admin |
| 926 | `group.update_settings` | `groupId`, name, description, state, rights | Update group settings |
| 1991 | `group.update_badge` | `groupId`, badge parts | Update group badge |
| 2066 | `group.update_colors` | `groupId`, `colorA`, `colorB` | Update group colors |
| 1236 | `group.set_favourite` | `groupId:int32` | Set as favourite group |
| 1820 | `group.remove_favourite` | _(none)_ | Remove favourite group |
| 2082 | `group.delete` | `groupId:int32` | Delete group |
| 2467 | `group.get_badge_parts` | _(none)_ | Get available badge parts |

#### Forums

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 873 | `forum.get_list` | `page:int32` | Get forum list |
| 3149 | `forum.get_threads` | `groupId`, `page` | Get threads in group forum |
| 232 | `forum.get_messages` | `groupId`, `threadId`, `page` | Get messages in thread |
| 3529 | `forum.post_thread` | `groupId`, `subject`, `body` | Create new thread |
| 3060 | `forum.post_message` | `groupId`, `threadId`, `body` | Reply to thread |
| 1397 | `forum.update_thread` | `groupId`, `threadId`, `action` | Pin/unpin/lock/hide thread |
| 286 | `forum.moderate_message` | `groupId`, `threadId`, `messageId`, `action` | Moderate (hide/delete) message |
| 3493 | `forum.update_settings` | `groupId`, `readPermission`, `postPermission`, `threadPermission`, `moderatePermission` | Update forum permissions |

#### Additional packets for group room management, member searching, group purchase, etc.

### S2C -- 28 packets

| ID | Name | Summary |
|----|------|---------|
| 1702 | `group.info` | Full group data (name, description, badge, members, state) |
| 1200 | `group.members` | Paginated member list |
| 3914 | `group.member_updated` | Member role changed |
| 2815 | `group.badge_parts` | Available badge editor parts |
| 1459 | `group.created` | Group creation result |
| 1180 | `group.join_failed` | Join request failed (full, closed, banned) |
| 3025 | `group.favourite_updated` | Favourite group changed |
| -- | `forum.list` | Forum list with unread counts |
| -- | `forum.threads` | Thread list for group |
| -- | `forum.messages` | Messages in thread |
| -- | `forum.thread_created` | New thread notification |
| -- | `forum.message_posted` | New reply notification |
| -- | `forum.settings_updated` | Forum settings changed |

---

## Implementation Analysis

### Group Creation

```
1. Validate:
   a. Name: 3-29 chars, no HTML/special chars
   b. Description: 0-255 chars
   c. Room: must be owned by creator, not already a group room
   d. Badge parts: valid part IDs and positions
   e. User not at max groups (default: 5, HC: 10)
   f. Creation cost: 25 credits (configurable)
2. Create:
   a. INSERT group (name, description, state, room_id, creator_id, badge_code)
   b. INSERT group_member (creator as OWNER)
   c. UPDATE rooms SET group_id = ? WHERE id = ?
   d. Deduct credits
3. Send group.created with new group ID
```

### Group Badge System

Group badges are composed of layered parts:

```
Badge format: b{baseId}{colorId}s{symbol1Id}{colorId}{posX}{posY}s{symbol2Id}...

Parts:
  - 1 base shape (required)
  - Up to 4 symbol layers
  - Each layer has: part ID, color, position
  - Colors from predefined palette
```

The badge editor data (`group.get_badge_parts`) returns all available bases, symbols, and colors. The server must validate that all part IDs exist.

### Membership Levels

| Level | Name | Permissions |
|-------|------|-------------|
| 0 | Owner | Full control, cannot be removed |
| 1 | Admin | Accept/decline members, forum moderation |
| 2 | Member | Post in forums, display badge |
| 3 | Pending | Awaiting acceptance (closed groups only) |

### Group States

| State | Join Behavior |
|-------|--------------|
| OPEN | Anyone can join immediately |
| REQUEST | Join creates pending request, admin must accept |
| CLOSED | No new members (invite-only via admin) |

### Forum System

Forums are per-group discussion boards with threads and replies:

```
Permissions matrix (configurable per group):
  readPermission:     0=everyone, 1=members, 2=admins, 3=owner
  postPermission:     0=everyone, 1=members, 2=admins, 3=owner
  threadPermission:   0=everyone, 1=members, 2=admins, 3=owner
  moderatePermission: 0=everyone, 1=members, 2=admins, 3=owner

Thread states: normal, pinned, locked, hidden
Message states: visible, hidden, deleted
```

**Pagination:** Forums use server-side pagination (20 threads/page, 20 messages/page) to handle groups with thousands of posts.

---

## Caveats & Edge Cases

### 1. Group Room Ownership Transfer
If the group room owner changes (room transfer), the group association should be preserved. If the room is deleted, the group loses its home room but continues to exist.

### 2. Member Count Queries
Arcturus uses two separate COUNT queries for members and pending requests. pixel-server should use a single query with GROUP BY: `SELECT level_id, COUNT(*) FROM group_members WHERE group_id = ? GROUP BY level_id`.

### 3. Badge Regeneration
When group colors or badge parts change, all members' badges must be updated. This is a visual-only update (no data change) -- the badge is rendered client-side from parts. But the cached badge code must be updated in the group record.

### 4. Favourite Group Display
The favourite group's badge appears on the user's avatar in rooms. Changing the favourite group must trigger a `room_entities.figure_change` broadcast to all users in the current room.

### 5. Forum Abuse Prevention
- Rate limit: max 3 threads per hour, max 10 replies per 5 minutes per user.
- Content length: subject max 64 chars, body max 4096 chars.
- Word filter applies to forum content.
- Deleted messages should be soft-deleted (hidden, not purged) for moderation review.

### 6. Group Deletion Cascade
Deleting a group must: remove all members, remove the room association, remove the favourite group from all users, and archive forum content.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Member counts** | Two queries (N+1) | Single grouped query |
| **Badge validation** | None (client-trusted) | Server-side part validation |
| **Forum** | Basic, no pagination | Full pagination + permissions matrix |
| **Group deletion** | Incomplete cleanup | Full cascade in transaction |

---

## Dependencies

- **Phase 2 (Identity)** -- user profiles for member display
- **Phase 3 (Room)** -- room association for group home
- **Phase 5 (Social)** -- group invites via messenger
- **PostgreSQL** -- groups, members, forums tables

---

## Testing Strategy

### Unit Tests
- Badge part composition and validation
- Membership level permission checks
- Forum permission matrix evaluation
- Group state transition rules

### Integration Tests
- Full group CRUD lifecycle
- Forum thread/reply creation and pagination
- Member accept/decline/remove/promote
- Group deletion cascade verification

### E2E Tests
- Client creates group, sets room, invites member
- Members post in forum, see threaded replies
- Group badge appears on member avatars in rooms
