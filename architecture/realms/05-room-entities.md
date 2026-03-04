# Realm: Room Entities

> **Position:** 70 | **Phase:** 3 (Room Entry & Movement) | **Packets:** 34 (14 c2s, 20 s2c)
> **Services:** game (ECS systems) | **Status:** Not yet implemented

---

## Overview

Room Entities is the real-time gameplay core. It covers avatar movement, chat (say/shout/whisper), posture/expression changes, typing indicators, hand items, dance, signs, and entity state broadcasting. This realm operates entirely within the ECS tick loop at 20 Hz and demands the lowest possible latency.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 3

---

## Packet Inventory

### C2S (Client to Server) -- 14 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 3320 | `room_entities.walk` | `x:int32`, `y:int32` | Walk to tile |
| 2080 | `room_entities.dance` | `danceId:int32` | Start/stop dancing |
| 2456 | `room_entities.action` | `actionId:int32` | Trigger expression (wave, idle, etc.) |
| 2235 | `room_entities.set_posture` | `posture:string` | Set posture (sit, lay, stand) |
| 1975 | `room_entities.hold_sign` | `signId:int32` | Hold up a numbered sign |
| 3301 | `room_entities.look_at` | `x:int32`, `y:int32` | Face a specific tile |
| 2941 | `room_entities.give_hand_item` | `userId:int32` | Give hand item to another user |
| 2814 | `room_entities.drop_hand_item` | _(none)_ | Drop current hand item |
| 1314 | `room_entities.say` | `message:string`, `color:int32` | Normal chat message |
| 2085 | `room_entities.shout` | `message:string`, `color:int32` | Shout (visible to entire room) |
| 1543 | `room_entities.whisper` | `message:string`, `targetName:string`, `color:int32` | Whisper to specific user |
| 1597 | `room_entities.typing_start` | _(none)_ | User started typing |
| 1474 | `room_entities.typing_stop` | _(none)_ | User stopped typing |
| 1030 | `room_entities.set_chat_style` | `styleId:int32` | Change chat bubble style |

### S2C (Server to Client) -- 20 packets

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 374 | `room_entities.units` | entities array (id, username, figure, motto, x, y, z, dir, gender, type, groupBadge) | Full entity list for room |
| 1640 | `room_entities.status` | statuses array (id, x, y, z, headDir, bodyDir, action string) | Entity position/state batch update |
| 3920 | `room_entities.info` | extended entity info | Detailed entity information |
| 1446 | `room_entities.chat` | `unitId`, `message`, `expression`, `color`, `links[]`, `messageCount` | Chat bubble broadcast |
| 1036 | `room_entities.shout` | same as chat | Shout bubble broadcast |
| 2704 | `room_entities.whisper` | same as chat | Whisper bubble (sender + target only) |
| 1747 | `room_entities.typing` | `unitId`, `typing:boolean` | Typing indicator state |
| 1191 | `room_entities.action` | `unitId`, `actionId` | Expression broadcast |
| 3582 | `room_entities.dance` | `unitId`, `danceId` | Dance state broadcast |
| 2016 | `room_entities.effect` | `unitId`, `effectId` | Avatar effect broadcast |
| 3831 | `room_entities.hand_item` | `unitId`, `itemId` | Hand item state broadcast |
| 1032 | `room_entities.idle` | `unitId`, `idle:boolean` | Idle state change |
| 1002 | `room_entities.remove` | `unitId:string` | Entity left room |
| 2401 | `room_entities.sleep` | `unitId`, `sleeping:boolean` | Sleep (zzz) state |
| 3785 | `room_entities.figure_change` | `unitId`, `figure`, `gender`, `motto`, `achievementScore` | Figure/motto update in-room |
| 1926 | `room_entities.carry_object` | `unitId`, `itemType` | Carrying object state |
| 2275 | `room_entities.slide_object` | slide data | Object sliding on roller |
| 2446 | `room_entities.expression` | `unitId`, `expressionId` | Expression change |
| 3189 | `room_entities.pet_respect` | `petId`, `respect:int32` | Pet respect given |
| 2700 | `room_entities.sign` | `unitId`, `signId` | Hold sign broadcast |

---

## Architecture Mapping

### ECS Integration

All room entity packets flow through the ECS world:

```
C2S Packet ──▶ Room Worker Command Channel ──▶ ECS Command
                                                    │
                                              20 Hz Tick Loop
                                                    │
                                            ┌───────┼───────┐
                                            ▼       ▼       ▼
                                      Movement  ChatCooldown  Broadcast
                                       System     System      System
                                                    │
                                               S2C Packets
                                                    │
                                            session.output via NATS
```

### ECS Components (from `pkg/room/components.go`)

| Component | Fields | Used By |
|-----------|--------|---------|
| `Position` | X, Y, Z float32 | All entities |
| `TileRef` | X, Y int16 | Collision detection |
| `WalkPath` | Steps[]PathStep, Cursor int | MovementSystem |
| `EntityKind` | Kind uint8 (Avatar=1, Bot=2, Pet=3, Item=4) | All systems |
| `AvatarID` | UserID int64, RoomUnit int32 | Avatar entities |
| `Status` | Posture uint8, Effects uint32 | BroadcastSystem |
| `ChatCooldown` | LastChat time.Time, MuteUntil time.Time | ChatCooldownSystem |
| `Dirty` | flag component | BroadcastSystem (marks entity for update broadcast) |

### ECS Systems (20 Hz tick order)

| Order | System | Reads | Writes | Packets Emitted |
|-------|--------|-------|--------|-----------------|
| 1 | `MovementSystem` | WalkPath, Position | Position, TileRef, Dirty | _(none, dirty flag set)_ |
| 2 | `ArrivalSystem` | Position, WalkPath | WalkPath (clear on arrive) | _(none)_ |
| 3 | `RollerSystem` | Position, TileRef | Position, Dirty | `room_entities.slide_object` |
| 4 | `ChatCooldownSystem` | ChatCooldown | ChatCooldown (decrement) | _(none)_ |
| 5 | `IdleSystem` | AvatarID, last input time | Status, Dirty | `room_entities.idle` |
| 6 | `BroadcastSystem` | Dirty, Position, Status | Dirty (clear) | `room_entities.status` |

---

## Implementation Analysis

### Movement Pipeline

`room_entities.walk` (3320) is the highest-frequency C2S packet:

```
1. Client sends walk(x, y)
2. Room worker receives command
3. Validate target tile:
   a. Within room bounds
   b. Not blocked (wall, item collision)
   c. Reachable from current position
4. Compute 3D A* path (pkg/pathfinding)
5. Set WalkPath component on entity
6. Mark entity as Dirty
7. Each tick, MovementSystem advances one step:
   a. Update Position component
   b. Update TileRef component
   c. Set Dirty flag
8. BroadcastSystem collects all Dirty entities:
   a. Build room_entities.status (1640) with updated positions
   b. Publish to all session outputs in room
   c. Clear Dirty flags
```

**Performance targets:**
- Pathfinding: < 100 us for 64x64 rooms (see [005-pathfinding-3d.md](../005-pathfinding-3d.md))
- Status broadcast: batched per tick (one packet per 50ms, not per step)
- Maximum entities per room: 200+ (ECS handles ~30-80 us per tick)

### Chat Pipeline

Chat packets (`say`, `shout`, `whisper`) follow this flow:

```
1. Client sends room_entities.say (1314) with message + color
2. Room worker validates:
   a. User is not muted (check ChatCooldown.MuteUntil)
   b. Message passes word filter
   c. Flood protection check (ChatCooldown.LastChat)
3. Apply chat distance:
   - Normal say: visible within chatDistance tiles (room setting)
   - Shout: visible to entire room
   - Whisper: visible to sender + named target only
4. Build room_entities.chat (1446) with unit ID, message, color
5. Publish to applicable session outputs:
   - Say: users within distance
   - Shout: all users
   - Whisper: sender + target
6. Update ChatCooldown.LastChat timestamp
```

**Flood protection levels:**
- 0 (off): no rate limiting
- 1 (normal): 1 message per 1 second
- 2 (strict): 1 message per 3 seconds

After exceeding the limit, subsequent messages are silently dropped for the cooldown duration. After 3 violations, the user is muted for 30 seconds.

### Entity Status Format

The `room_entities.status` (1640) S2C packet uses a string-encoded action format:

```
"/<unitId> <x>,<y>,<z>/<headDir>/<bodyDir>/[action string]"

Action string components:
  mv x,y,z     -- moving to tile
  sit z        -- sitting at height z
  lay z        -- laying at height z
  sign n       -- holding sign number n
  flatctrl n   -- has room rights (level n)
  dance n      -- dancing (style n)
  carryd n     -- carrying drink/item
  gest sml|sad|srp|agr -- gesture expression
```

This format is legacy but must be preserved for Nitro client compatibility. The BroadcastSystem must serialize entity state into this exact format.

### Hand Items

Hand items (drinks, food, carry items) are visual-only props:
- `room_entities.give_hand_item` (2941): Transfer hand item to another user within interaction distance.
- `room_entities.drop_hand_item` (2814): Remove hand item.
- Broadcast `room_entities.hand_item` (3831) or `room_entities.carry_object` (1926) to all users.

Hand items have a 15-minute auto-drop timer.

### Posture & Expression

| Posture ID | Name | Notes |
|------------|------|-------|
| 0 | Stand | Default |
| 1 | Sit | Only valid on sittable tiles/furniture |
| 2 | Lay | Only valid on layable furniture |

| Expression ID | Name | Notes |
|---------------|------|-------|
| 1 | Wave | 3-second duration |
| 2 | Blow kiss | 3-second duration |
| 3 | Laugh | 3-second duration |
| 4 | Idle | Triggered after 10 minutes inactivity |
| 5 | Jump | 1-second duration |

---

## Caveats & Edge Cases

### 1. Walk Cancellation
If a user clicks a new destination while walking, the current path is replaced immediately. The new path starts from the entity's current position (the tile they're currently stepping onto), not their original destination.

### 2. Status Update Batching
Sending individual status updates per entity per tick creates packet storms. The BroadcastSystem must batch all dirty entities into a single `room_entities.status` packet per tick. With 50 entities moving, this means one packet with 50 entries every 50ms -- not 50 packets.

### 3. Chat Bubble Rendering
The `messageCount` field in chat packets is a counter that increments per message from each unit. The client uses this to stack chat bubbles vertically. The server must track per-unit message counters.

### 4. Whisper Privacy
Whispers must only be delivered to the sender and the named target. If the target is not in the room, send an error. Never broadcast whispers to other users or log them in room chat history (moderation caveat: some emulators still log whispers for mod review).

### 5. Typing Indicator Spam
Clients send `typing_start` on every keystroke and `typing_stop` on clear. Rate-limit typing indicator broadcasts to 1 per second per entity to prevent packet flooding.

### 6. Sit/Lay Validation
Setting posture to "sit" is only valid if the entity is on a sittable tile (chair, sofa). "Lay" is only valid on layable furniture (bed). The server must check the tile's item interaction type before allowing the posture change.

### 7. Dance While Walking
If a user is dancing and starts walking, the dance should stop. If a user starts dancing while sitting, they should stand first. Enforce state transition rules.

### 8. Entity Removal on Disconnect
When a session disconnects unexpectedly, the entity must be removed from the ECS world immediately. The BroadcastSystem should send `room_entities.remove` to all other users.

### 9. Direction Facing
`room_entities.look_at` (3301) sets the entity's head/body direction without moving. The direction is computed from the current position to the target tile using octagonal direction math (8 directions: N, NE, E, SE, S, SW, W, NW).

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **State updates** | Per-entity packets (N packets per tick) | Batched single status packet per tick |
| **Pathfinding** | 2D A* (Z ignored) | 3D A* with height costs |
| **Chat distance** | Euclidean distance check | Tile-based Manhattan distance (more accurate) |
| **Flood protection** | Global cooldown | Per-room configurable (3 levels) |
| **Posture validation** | Minimal (sit anywhere) | Tile-aware (requires sittable furniture) |
| **Entity model** | Inheritance hierarchy | ECS components (composable, cache-friendly) |
| **Tick determinism** | Timer-based (non-deterministic) | Fixed 20 Hz ticker (50ms, deterministic) |

---

## Dependencies

- **Phase 3 (Room)** -- room entry creates entities; room worker provides ECS world
- **pkg/room** -- ECS components (Position, WalkPath, Status, etc.)
- **pkg/pathfinding** -- 3D A* for movement
- **pkg/ecs** -- Ark v0.7.1 World, Mappers, Filters

---

## Testing Strategy

### Unit Tests
- Pathfinding produces correct paths for various room layouts
- Chat flood protection state machine
- Direction calculation (8 directions from tile to tile)
- Status format string serialization
- Walk cancellation replaces path correctly
- Posture validation against tile type

### Integration Tests
- Full movement cycle: walk command -> path computed -> status broadcast
- Chat message flows through word filter and distance check
- Entity spawn/remove lifecycle in ECS world
- Batch status updates with multiple moving entities

### E2E Tests
- Two clients in a room: A walks to tile, B sees movement
- A says a message, B receives chat bubble with correct styling
- A whispers to B, C (third client) does not see it
- A goes idle after inactivity, B sees idle animation
