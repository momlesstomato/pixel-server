# 11 - Room Completion & Chat History

## Overview

This plan covers all remaining room-realm features required for a
production-ready room experience. It spans **chat history persistence**,
**entity SDK events**, **room voting**, **room soft-delete**, **room
promotions/ads**, **staff picks**, **room forward**, **ban list
management**, **room score display**, and several minor quality-of-life
packets.

The pixel-protocol currently handles **21 C2S** and **27 S2C** room
packets. After this plan, the server will add **10 new C2S** and **5 new
S2C** packets, bringing the total to **31 C2S** and **32 S2C**.

---

## Vendor Cross-Reference

### Room Feature Matrix

| Feature | PlusEMU (C#) | Arcturus (Java) | comet-v2 (Java) | pixelsv (proposed) |
|---------|-------------|-----------------|-----------------|-------------------|
| Chat history storage | DB `chatlogs` table | DB `chatlogs` table | DB `chatlogs` table | **Append-only `room_chat_logs`** |
| Chat log API | Mod tool query | Mod tool query | Mod tool query | **REST GET + CLI .log export** |
| Room voting | `rooms.score` column, 1 vote/user | `rooms.score`, via DB | `rooms.score`, in-memory | **`room_votes` table, 1 per user per room** |
| Room soft-delete | Room marked deleted | `Room.dispose()` | `RoomManager.removeRoom()` | **`deleted_at` column, furniture returns to inventory** |
| Room promotions/ads | `room_promotions` table | `room_promotions` | Navigator promoted list | **DEFER** (requires navigator overhaul) |
| Staff picks | `navigator_publics` table | `navigator_publics` | Navigator featured | **DEFER** (requires navigator categories) |
| Room forward | Redirect on room load | Redirect on room load | Redirect on room load | **`RoomForward` composer on entry** |
| Ban list management | `GetBannedUsers` packet | `GetBannedUsers` packet | `GetBannedUsers` packet | **`GetRoomBannedUsers` + `UnbanUser` packets** |
| Room mute toggle | `MuteAllInRoom` | `MuteAllInRoom` | `MuteAllInRoom` | **DEFER** (moderation plan) |
| Word filter | Per-room word list | Per-room word list | Global filter | **DEFER** (moderation plan) |
| Ambassador alerts | Staff-only alert | Staff-only alert | — | **DEFER** (moderation plan) |
| Entity SDK events | N/A | N/A | N/A | **7 missing event pairs (Dance, Action, Sign, Typing, LookTo, Sit)** |

### Chat Log Schema Comparison

| Aspect | PlusEMU | Arcturus | comet-v2 | pixelsv |
|--------|---------|----------|----------|---------|
| Table | `chatlogs` | `chatlogs` | `chatlogs` | `room_chat_logs` |
| Fields | room_id, user_id, message, timestamp | room_id, user_id, message, timestamp, type | room_id, user_id, message, timestamp | room_id, user_id, username, message, chat_type, created_at |
| Chat type stored | No (all same) | Yes (talk/shout/whisper) | No | **Yes** |
| Username denormalized | No | No | No | **Yes** (avoids JOIN on read) |
| Indexing | room_id + timestamp | room_id + timestamp | room_id | **Composite (room_id, created_at) + (created_at)** |

### Room Voting Schema Comparison

| Aspect | PlusEMU | Arcturus | comet-v2 | pixelsv |
|--------|---------|----------|----------|---------|
| Score storage | `rooms.score` INT | `rooms.score` INT | `rooms.score` INT | `rooms.score` INT |
| Vote tracking | `user_room_votes` (user_id, room_id) | `room_votes` | In-memory set | `room_votes` (user_id, room_id) unique |
| Max score | Unlimited | Unlimited | Unlimited | **Unlimited** |
| Vote value | +1 per user | Configurable | +1 | **+1 per user** |

### Our Improvements Over Vendors

1. **Structured chat history** — stores chat type (talk/shout/whisper)
   and denormalized username to avoid expensive JOINs on large datasets.
2. **REST API for chat logs** — vendors only expose chat via
   in-client moderation tools. We add a full HTTP API with date range
   filtering and a CLI `.log` download command.
3. **Room votes persisted** — comet-v2 tracks votes in memory only,
   losing them on restart. We persist in `room_votes`.
4. **Soft-delete with inventory return** — all vendors mark rooms
   deleted but do not automatically return placed furniture. We cascade
   furniture items back to owner inventory atomically.
5. **Full SDK entity events** — enables plugins to intercept/cancel
   Dance, Action, Sign, Typing, LookTo, and Sit mutations.

---

## Packet Registry

### Client-to-Server (new packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3616 | `room.give_room_score` | score (int32) | **M1** |
| 532 | `room.delete_room` | (empty, uses current room) | **M1** |
| 2652 | `room.get_banned_users` | roomId (int32) | **M1** |
| 3842 | `room.unban_user` | userId (int32), roomId (int32) | **M1** |

### Client-to-Server (deferred packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 2937 | `room.purchase_room_ad` | categoryId, name, description, roomId, expiry, costCredits | **DEFER** |
| 22 | `room.edit_room_ad` | promotionId, name, description | **DEFER** |
| 1920 | `room.staff_pick` | roomId (int32), pick (bool) | **DEFER** |
| 2973 | `room.modify_word_filter` | roomId, add (bool), word (string) | **DEFER** |
| 1973 | `room.get_word_filter` | roomId (int32) | **DEFER** |
| 0 | `room.toggle_mute_tool` | — | **DEFER** |

### Server-to-Client (new packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3271 | `room.room_score` | score (int32), canVote (bool) | **M1** |
| 511 | `room.room_forward` | roomId (int32) | **M1** |
| 1869 | `room.banned_users` | roomId (int32), count (int32), [userId, username]… | **M1** |

---

## Database Model Design

### `room_chat_logs` (new table)

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PK |
| room_id | INTEGER | NOT NULL, INDEX |
| user_id | INTEGER | NOT NULL |
| username | VARCHAR(50) | NOT NULL |
| message | VARCHAR(512) | NOT NULL |
| chat_type | VARCHAR(10) | NOT NULL DEFAULT 'talk' |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

**Indexes:**
- `idx_room_chat_logs_room_created` ON (room_id, created_at)
- `idx_room_chat_logs_created` ON (created_at)

### `room_votes` (new table)

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PK |
| room_id | INTEGER | NOT NULL |
| user_id | INTEGER | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

**Constraints:**
- UNIQUE (room_id, user_id)

### `rooms` table changes

| Column | Change |
|--------|--------|
| deleted_at | ADD TIMESTAMPTZ DEFAULT NULL |

---

## Event Completeness Audit

### Missing Entity Events (7 pairs needed)

Each mutation in `entity_service.go` that modifies entity state must fire
a cancellable "before" event and a non-cancellable "after" event.

| Mutation | Before Event | After Event |
|----------|-------------|-------------|
| Dance | `EntityDancing` (cancellable) | `EntityDanced` |
| Action | `EntityActing` (cancellable) | `EntityActed` |
| Sign | `EntitySigning` (cancellable) | `EntitySigned` |
| StartTyping | `EntityTyping` (cancellable) | `EntityTyped` |
| StopTyping | `EntityTypingStopped` (cancellable) | `EntityTypingCleared` |
| LookTo | `EntityLooking` (cancellable) | `EntityLooked` |
| Sit | `EntitySitting` (cancellable) | `EntitySat` |

All event types will live under `sdk/events/room/entity/`.

---

## Design Decisions

1. **Chat log username denormalization** — storing username alongside
   user_id avoids a JOIN on the users table when reading large chat
   histories. Username changes are rare and historical accuracy per
   message is more valuable than normalization.

2. **Append-only chat logs** — no UPDATE or DELETE needed. The table
   grows linearly with chat activity. A future partition-by-month
   strategy can be added without schema changes.

3. **CLI `.log` export** — the `room chat-export` command writes plain
   text lines in the format `[HH:MM:SS] [TYPE] username: message`.
   This matches common game server log conventions.

4. **Soft-delete cascade** — when a room is deleted, all placed
   furniture items in that room have their `room_id` set to 0 (inventory)
   and placement fields cleared. This follows the PlusEMU pattern.

5. **Room forward** — when a room has a forward target, the server
   sends a `RoomForward` composer instead of loading the room. The
   client handles the redirect. Forward targets are stored as a new
   `forward_room_id` column on the rooms table (deferred until
   navigator support).

---

## Implementation Scope

### Implemented (M1)

- [x] SDK entity events (7 before/after pairs)
- [x] Entity service event wiring (Dance, Action, Sign, Typing, LookTo, Sit)
- [x] Chat history: migration, GORM model, store, domain repository
- [x] Chat history: HTTP API (GET by room + date range)
- [x] Chat history: CLI `db chat-export` command
- [x] Chat history: integration with ChatService (persist on send)
- [x] Room voting: migration, store, packet, handler
- [x] Room soft-delete: migration, store method, packet, handler
- [x] Ban list: GetBannedUsers packet + handler
- [x] Ban list: UnbanUser packet + handler
- [x] Room score: send score on room entry

### Implemented (Phase 2)

- [x] Room rights ownership: AssignRights, RemoveRights, RemoveMyRights, RemoveAllRights, GetRoomRights C2S packets + dispatch handlers
- [x] Room rights S2C: YouAreControllerComposer (sent on room entry for owner/rights holder), RoomRightsListComposer
- [x] Room mute toggle: engine `muted` flag, ToggleMuteToolPacket handler, chat/shout suppression for non-owners
- [x] Room forward: `ForwardRoomID` domain field, model column, Step08 migration, forward check in handleOpenFlat
- [x] Word filter per-room: WordFilterService + ChatService integration (implemented in moderation Phase 2)
- [x] Ambassador alerts: AlertAmbassadors on kick/ban (implemented in moderation Phase 2)
- [x] Room promotions: navigator model `promoted_until` + `promotion_name` columns, Step05 migration, domain Room fields, RoomFilter.PromotedOnly, store filter (new_ads search tab)
- [x] Staff picks: navigator model `staff_pick` column, Step06 migration, domain Room.StaffPick, RoomFilter.StaffPickOnly, store filter (official search tab), StaffPick packet handler with PermStaffPick permission check
- [x] Navigator permission.go: PermissionChecker interface + PermStaffPick constant in navigator domain

### Deferred

- Bot configuration (requires bot entity type)

### What to Test

1. **Entity events** — verify each of the 7 event pairs fires correctly,
   cancellation aborts the mutation, and after-event receives correct data.
2. **Chat history write** — verify Talk, Shout, Whisper all persist to
   `room_chat_logs` with correct chat_type and username.
3. **Chat history API** — test GET `/api/v1/rooms/:roomId/chat-logs`
   with date range query params, verify pagination and filtering.
4. **Chat export CLI** — test `room chat-export --room-id=1 --date=2025-01-01`
   produces valid `.log` format output.
5. **Room voting** — verify one vote per user, score increments, duplicate
   vote rejected, score sent on room entry with `canVote` flag.
6. **Room delete** — verify soft-delete sets `deleted_at`, furniture items
   return to inventory (room_id=0), room no longer loadable.
7. **Ban list** — verify `GetBannedUsers` returns current bans,
   `UnbanUser` removes ban and allows re-entry.
8. **Migration rollback** — verify all new migrations can be rolled back
   cleanly without data loss in other tables.
