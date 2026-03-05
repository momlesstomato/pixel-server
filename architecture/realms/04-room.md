# Realm: Room

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 60 | **Phase:** 3 (Room Entry & Movement) | **Packets:** 90 (46 c2s, 44 s2c)
> **Services:** game (room worker) | **Status:** Not yet implemented

---

## Overview

The Room realm is the second-largest realm (90 packets) and the most architecturally significant. It governs room entry, model/heightmap loading, settings management, rights assignment, banning, doorbell flow, events, bot management, and room moderation. This realm brings the ECS world and 20 Hz tick loop online -- it is the foundation for all in-room gameplay.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 3

---

## Packet Inventory

### C2S (Client to Server) -- 46 packets

#### Room Entry & Lifecycle

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2312 | `room.enter` | `roomId:int32`, `password:string`, `unknown:int32` | Request to enter a room |
| 1644 | `room.doorbell` | `username:string` | Ring doorbell (locked rooms) |
| 105 | `room.desktop_view` | _(none)_ | Leave room to hotel view |
| 3093 | `room.change_queue` | `targetRoomId:int32` | Change room queue position |

#### Room Settings & Configuration

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 3129 | `room.get_settings` | `roomId:int32` | Request room settings |
| 1969 | `room.save_settings` | name, description, state, password, maxUsers, category, tags, tradeMode, allowPets, allowPetsEat, walkThrough, hideWall, wallThickness, floorThickness, chatMode, chatWeight, chatSpeed, chatDistance, floodProtection | Save room settings |
| 3385 | `room.get_rights_list` | `roomId:int32` | Request users with room rights |
| 2267 | `room.get_ban_list` | `roomId:int32` | Request banned users list |
| 3637 | `room.toggle_mute` | `roomId:int32` | Toggle room-wide mute |
| 2300 | `room.get_model` | _(none)_ | Request room model/heightmap |
| 3559 | `room.get_entry_tile` | _(none)_ | Request door tile position |
| 875 | `room.save_floor_plan` | `heightmap:string` | Save custom floor plan |
| 1687 | `room.get_occupied_tiles` | _(none)_ | Request tiles with items |
| 3582 | `room.like` | _(none)_ | Like/rate the room (+1 score) |

#### Rights Management

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 808 | `room.give_rights` | `userId:int32` | Grant room rights to user |
| 2064 | `room.take_rights` | `userId:int32` | Revoke room rights from user |
| 2683 | `room.remove_all_rights` | _(none)_ | Revoke all room rights |
| 3182 | `room.remove_own_rights` | _(none)_ | Remove own room rights |

#### Room Moderation

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1477 | `room.ban_user` | `userId:int32`, `roomId:int32`, `banType:string` | Ban user from room |
| 992 | `room.unban_user` | `userId:int32`, `roomId:int32` | Unban user from room |
| 1320 | `room.kick_user` | `userId:int32` | Kick user from room |
| 3485 | `room.mute_user` | `userId:int32`, `roomId:int32`, `minutes:int32` | Mute user in room |
| 2996 | `room.ambassador_alert` | `userId:int32` | Ambassador alert to user |

#### Bots

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1592 | `room.place_bot` | `botId:int32`, `x:int32`, `y:int32` | Place bot from inventory |
| 3323 | `room.pickup_bot` | `botId:int32` | Return bot to inventory |
| 2624 | `room.save_bot_skill` | `botId:int32`, `skillId:int32`, `data:string` | Configure bot behavior |
| 1986 | `room.get_bot_configuration` | `botId:int32` | Get bot settings |

#### Object Data

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 3608 | `room.set_object_data` | `itemId:int32`, `data:string` | Set room object state data |
| 3964 | `room.get_item_data` | `itemId:int32` | Get item data |

#### Additional (20+ packets for events, promotions, word filter, polls, room ads)

Remaining C2S packets cover room events, promotion management, word filter configuration, room polls, advertising, and various room-specific queries.

### S2C (Server to Client) -- 44 packets

#### Room Entry Sequence

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 758 | `room.open` | _(none)_ | Room opened, prepare for data |
| 2031 | `room.model` | `model:string`, `roomId:int32` | Room model name for layout |
| 1301 | `room.heightmap` | `heightmap:string` | Full room heightmap data |
| 2753 | `room.relative_map` | `relativeMap:string` | Relative heightmap |
| 1186 | `room.score` | `score:int32` | Room rating score |
| 2454 | `room.paint` | `type:string`, `value:string` | Wall/floor decoration |
| 687 | `room.rights` | `level:int32` | User's rights level in room |
| 339 | `room.rights_list` | `roomId`, `users[]` | All users with rights |
| 1200 | `room.info` | full room object | Complete room metadata |

#### Room State Changes

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 3828 | `room.settings` | full settings object | Room settings data |
| 1500 | `room.settings_saved` | `roomId:int32` | Settings save confirmation |
| 2208 | `room.settings_error` | `roomId`, `errorCode`, `message` | Settings save error |
| 1245 | `room.muted` | `muted:boolean` | Room mute state toggled |
| 997 | `room.doorbell_ringing` | `username:string` | Someone is at the doorbell |
| 3783 | `room.doorbell_accepted` | `username:string` | Doorbell accepted |
| 878 | `room.doorbell_rejected` | _(none)_ | Doorbell rejected |
| 3736 | `room.flat_access_denied` | `username:string` | Access denied |
| 3963 | `room.no_rights` | _(none)_ | User has no rights notification |

#### Bans, Kicks, Rights Updates

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 1869 | `room.banned_users` | `users[]` | Banned users list |
| 2999 | `room.user_banned` | `userId:int32` | User banned notification |
| 1890 | `room.user_kicked` | _(none)_ | User kicked notification |
| 3785 | `room.rights_updated` | `roomId`, `userId`, `level` | Rights changed |

#### Additional S2C packets for heightmap updates, floor plan editor, entry tile, occupied tiles, room events, promotions, and more.

---

## Architecture Mapping

### Service Ownership

The **game service** owns room lifecycle through room workers:

```
Gateway ──NATS(room.input.<sid>)──▶ Game Service
                                         │
                                    Room Router
                                         │
                                    Room Worker Pool
                                         │
                                    ┌─────┴─────┐
                                    ▼            ▼
                              Room Worker 1  Room Worker 2 ...
                              (ECS World)    (ECS World)
                              (20 Hz tick)   (20 Hz tick)
```

Each room worker:
- Owns an isolated `*ecs.World` (Ark v0.7.1).
- Runs on its own goroutine with a 50ms ticker.
- Receives commands via `chan Envelope` (never direct method calls).
- Broadcasts state updates to all session outputs in the room.

### Room Worker Lifecycle

```
1. Room Enter Request
   └─ Room Router checks if worker exists
      ├─ YES: Forward command to existing worker
      └─ NO: Create new worker
              ├─ Load room from PostgreSQL
              ├─ Initialize ECS World
              ├─ Load room model + heightmap
              ├─ Load placed items
              ├─ Load bots + pets
              ├─ Start 20 Hz tick loop
              └─ Forward enter command

2. Room Idle (no users for 60 seconds)
   └─ Worker saves state to PostgreSQL
   └─ Worker disposes ECS World
   └─ Worker goroutine exits

3. Room Unload
   └─ Save all dirty state
   └─ Remove from worker pool
   └─ Release memory
```

### Database Tables

| Table | Usage |
|-------|-------|
| `rooms` | Room metadata, settings, owner |
| `room_models` | Predefined room layouts |
| `room_rights` | user_id, room_id pairs |
| `room_bans` | user_id, room_id, ban_type, expires_at |
| `room_mutes` | user_id, room_id, expires_at |
| `room_bots` | bot definitions, positions |
| `room_word_filter` | Per-room word filter entries |
| `room_promotions` | Active room promotions |

---

## Implementation Analysis

### Room Entry Flow (Critical Path)

The room entry is the most complex packet flow in the system:

```
1. Client sends room.enter (2312) with roomId and optional password
2. Game service routes to room worker (create if needed)
3. Room worker validates:
   a. Room exists
   b. User is not banned
   c. Room state access check:
      - "open": allow
      - "locked": check rights → if no rights, trigger doorbell flow
      - "password": validate password
      - "invisible": check rights or owner
   d. Room is not full (users_now < users_max)
4. On success:
   a. Send room.open (758)
   b. Send room.model (2031)
   c. Send room.heightmap (1301) + room.relative_map (2753)
   d. Send room.paint (2454) for wall/floor decorations
   e. Send room.info (1200) with full room metadata
   f. Send room.rights (687) with user's permission level
   g. Send room.score (1186)
   h. Spawn ECS entity for the user at door tile
   i. Broadcast room_entities.units to all users in room
   j. Send existing entities list to the entering user
   k. Send furniture lists (Phase 6)
5. On failure:
   a. Send room.flat_access_denied (3736) or appropriate error
```

### Doorbell Flow (Locked Rooms)

```
1. User A enters locked room → no rights
2. Server sends room.doorbell_ringing (997) to room owner/users with rights
3. Owner/rights holder sends room.doorbell (1644) with accept/reject
4. If accepted: room.doorbell_accepted (3783) → proceed with entry
5. If rejected: room.doorbell_rejected (878) → user denied
6. If timeout (30 seconds): auto-reject
```

**Caveat:** If no rights holders are online, the doorbell rings indefinitely. Implement a 30-second timeout with auto-reject.

### Rights System

Room rights have three tiers:
- **Level 0:** No rights (normal user)
- **Level 1:** Has rights (can place/move items, use room tools)
- **Level 2:** Owner (full control including settings, rights management, bans)

The `room_rights` table maps user_id to room_id. Rights are checked on every room command. Reference emulators (Comet v2) cache rights in the room component.

pixel-server should cache rights in the room worker's memory, invalidated on rights changes. This avoids database round-trips on every furniture interaction.

### Room Settings Save

`room.save_settings` (1969) is one of the most field-heavy packets:
- name (3-25 chars), description (0-128), state, password, maxUsers, category
- tags (max 2, max 15 chars each), tradeMode (0=disabled, 1=rights, 2=all)
- allowPets, allowPetsEat, walkThrough, hideWall
- wallThickness (0-2), floorThickness (0-2)
- chatMode, chatWeight, chatSpeed, chatDistance, floodProtection

**Validation is critical.** Every field must be bounds-checked. Reference emulators often skip validation, allowing:
- Room names with HTML/script injection
- MaxUsers exceeding server capacity
- Invalid category IDs
- Chat speed below 0

### Custom Floor Plans

`room.save_floor_plan` (875) allows room owners to edit the heightmap:

```
Heightmap format:
  x = wall (blocked)
  0-9, a-z = tile height (0=ground, 1=0.5 units, ... z=highest)
  \r = row separator

Example 4x3 room:
  xxxx\r
  x000\r
  x000
```

**Validation requirements:**
- At least one walkable tile adjacent to the door.
- Maximum dimensions: 64x64 (configurable).
- No disconnected sections.
- Door tile must be valid.

### Bot Management

Bots are NPC entities placed in rooms:
- `room.place_bot` (1592): Creates ECS entity with `BotAI` component at specified tile.
- `room.pickup_bot` (3323): Removes ECS entity, returns to inventory.
- `room.save_bot_skill` (2624): Configures bot behavior (greetings, chat lines, movement patterns).
- `room.get_bot_configuration` (1986): Returns current bot settings.

Bot AI runs within the ECS `BotAISystem` each tick:
- Walk to random tiles on a timer.
- Say configured chat lines at intervals.
- Respond to configured triggers.

---

## Caveats & Edge Cases

### 1. Room Worker Memory Leaks
If rooms are never unloaded, memory grows unbounded. Implement aggressive idle timeout:
- No users for 60 seconds → begin unload.
- Save all dirty state before disposing ECS world.
- Log memory freed per unload for monitoring.

### 2. Concurrent Room Entry Race
Two users entering simultaneously may both trigger room creation. The room router must use a sync primitive (mutex or channel) to ensure only one worker is created per room ID.

### 3. Ban Types
Room bans have three types:
- `HOUR`: 1-hour ban
- `DAY`: 24-hour ban
- `PERMANENT`: indefinite (only removable by owner)

Ban expiry must be checked on every entry attempt. Use PostgreSQL `expires_at` column with `NULL` for permanent bans.

### 4. Rights Cascade on Room Delete
When a room is deleted (`room.delete`, 532), all associated data must be cleaned up:
- `room_rights` entries
- `room_bans` entries
- `room_mutes` entries
- Placed items returned to owner's inventory
- Bots returned to owner's inventory
- Navigator favourites referencing this room removed

### 5. Room State Transitions
State changes (open -> locked -> password -> invisible) must notify all users in the room. When changing to "password", existing users without rights should be kicked.

### 6. Chat Settings Validation
Chat settings (mode, weight, speed, distance, flood protection) have specific valid ranges:
- chatMode: 0 (free-flow), 1 (line-by-line)
- chatWeight: 0-3 (bubble size)
- chatSpeed: 0-2 (scroll speed)
- chatDistance: 1-99 (tiles)
- floodProtection: 0 (off), 1 (normal), 2 (strict)

### 7. Heightmap Parsing Robustness
Custom heightmaps from clients can be malformed. The parser must handle:
- Inconsistent row lengths
- Invalid characters
- Extremely large dimensions
- Empty heightmaps

### 8. Bot Limits
Limit bots per room (default: 10) and per user (default: 25 total across all rooms). Enforce at placement time.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Room lifecycle** | Never unloaded (memory leak) | Idle timeout with graceful unload |
| **ECS isolation** | Global state shared across rooms | One `*ecs.World` per room goroutine |
| **Rights caching** | DB query per action | In-memory cache invalidated on change |
| **Floor plan validation** | Minimal/none | Full connectivity + dimension checks |
| **Bot AI** | Hardcoded behavior per type | Configurable ECS component with skills |
| **Room entry** | Monolithic handler | Step-by-step pipeline with fail-fast |
| **Settings validation** | Missing/partial | Full bounds-checking on every field |
| **Concurrent access** | Shared mutable state + locks | Channel-based command dispatch |

---

## Dependencies

- **Phase 1 (Connection)** -- authenticated session
- **Phase 2 (Identity)** -- user data for room entry, rights validation
- **pkg/room** -- models (Room, Model), ECS components (Position, TileRef, etc.)
- **pkg/pathfinding** -- 3D A* for movement (Phase 3 room-entities)
- **pkg/ecs** -- Ark v0.7.1 World and component mappers
- **PostgreSQL** -- room data, rights, bans, models, bots
- **Redis** -- room user count, room presence

---

## Testing Strategy

### Unit Tests
- Room entry validation (all state types: open/locked/password/invisible)
- Rights level checks (owner, rights holder, normal user)
- Heightmap parsing and validation
- Room settings bounds checking
- Ban type expiry calculation
- Bot placement limit enforcement

### Integration Tests
- Full room entry flow against real PostgreSQL
- Rights CRUD operations
- Ban add/remove/expiry
- Custom floor plan save and reload
- Room creation and deletion with cascade cleanup

### E2E Tests
- Client enters open room, sees heightmap and entities
- Client enters locked room, doorbell rings, owner accepts
- Client enters password room with correct/incorrect password
- Owner changes room settings, other users see updates
- Bot placed and exhibits AI behavior
