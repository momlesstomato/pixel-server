# 10 — Room Basics: Comprehensive Implementation Plan

## Overview

The Room realm is the most complex subsystem in the Habbo protocol, encompassing
90 core packets (46 C2S + 44 S2C), 34 entity packets (14 C2S + 20 S2C), and
serving as a prerequisite for trading, furniture interactions, pets, moderation,
and full navigator integration. This plan defines an ephemeral, isolated,
goroutine-per-room architecture designed for future multi-node distribution.

**Design principles:**
- No god objects — each room is a self-contained ephemeral environment
- Goroutine-per-room isolation with channel-based message passing
- Tick-driven state machine (500ms cycle matching vendor consensus)
- Server-authoritative pathfinding with client prediction
- Future-proof for node delegation via room serialization

---

## 1. Vendor Cross-Reference Matrix

| Feature | PlusEMU (.NET) | Arcturus (Java) | Comet-v2 (Java) | pixels-emulator (Go) |
|---------|----------------|-----------------|-----------------|---------------------|
| **Tick interval** | 500ms (RoomManager.OnCycle) | ~500ms (ScheduledFuture) | ~500ms (ProcessComponent) | N/A (no tick impl) |
| **Room isolation** | Task per room (C# Task) | Thread pool (ScheduledExecutor) | Single process thread | N/A |
| **Heightmap delimiter** | Char(13) = CR | `\r` after removing `\n` | `\r` after removing `\n` | `\\r` literal (BUG) |
| **Height encoding** | `0-9` = 0-9, custom Parse() | `0-9` = 0-9, `A-Z` = 10-35 | `0-9` = 0-9, `a-z` = 10-35 | `0-9` then `A-Z` = 10-35 |
| **Blocked tile** | `'x'` | `'x'` | `'x'` | `'x'` |
| **Pathfinding** | A* with MinHeap, 8-dir | A* with PriorityQueue, 8-dir | A* with MinMaxPQ, 8-dir | N/A |
| **Height in path** | Not validated in step | `maximumStepHeight` check | Deferred to mapping | N/A |
| **Diagonal blocking** | None (free diagonal) | Modern: both cardinals open | Not implemented | N/A |
| **Idle unload** | 60 cycles (30s) | 240 cycles (~2min) | Configurable | N/A |
| **Lag detection** | 30 missed cycles = crash | None | `isProcessing` guard | N/A |
| **Chat proximity** | Server filters recipients | Server filters recipients | `broadcastChatMessage` | N/A |
| **Room access** | Open/Locked/Password | Open/Locked/Password/Invisible | Open/Locked/Password | N/A |
| **Entity limit** | ConcurrentDictionary | ConcurrentHashMap | ConcurrentHashMap | N/A |
| **Status updates** | Batch per cycle end | Per-entity on cycle | Batch broadcast | N/A |

---

## 2. Heightmap Parsing — Critical Bug Analysis

### The Bug (pixels-emulator)

The pixels-emulator has an **escape sequence mismatch** between storage and parsing:

1. **`NewLayout()`** (line 179): `strings.ReplaceAll(hMap.Heightmap, "\n", "")` — removes actual newlines
2. **`generateGrid()`** (line 51): `strings.ReplaceAll(l.hMap, "\\n", "")` — tries to remove *literal* `\n`
3. **`generateGrid()`** (line 52): `strings.Split(c, "\\r")` — splits by *literal* `\r`
4. **`SendHeightMapPackets()`** (protocol.go line 28): `strings.ReplaceAll(l.RawMap(), "\\r\\n", "\r")` — converts *literal* `\r\n` to actual CR

**Root Cause:** If the DB stores actual `\r\n` bytes, step 1 strips `\n` but step 2 can't find literal `\\n`. If the DB stores escaped `\r\n` as string literals, step 1 does nothing useful.

### Our Solution

Pixel-server must normalize heightmaps at the **ingestion boundary** (DB load or API input):

```
Step 1: Replace literal "\r\n" sequences with "\r"
Step 2: Replace actual \r\n bytes with \r
Step 3: Replace actual \n bytes with empty string
Step 4: Split by \r to get rows
Step 5: Validate each row has consistent length
```

Client expects heightmap rows separated by `\r` (char 13) only. Heights are base-36:
`0-9` = heights 0-9, `a-z` (case-insensitive) = heights 10-35, `x` = blocked tile.

### Client Protocol Expectations (from Nitro)

The client expects **three packets** in sequence for room geometry:

1. **FloorHeightMapEvent** — raw heightmap string (rows separated by `\r`), scale flag, wall height
2. **RoomHeightMapEvent** — width (int), totalTiles (int), then short[] of stacking heights
   - Each short: `(value & 16383) / 256` = tile height, `(value & 0x4000)` = stacking blocked
3. **RoomHeightMapUpdateEvent** — incremental updates when furniture changes: `{x, y, height}`

---

## 3. Architecture — Ephemeral Room Environments

### Package Structure

```
pkg/room/
├── domain/                     # Core contracts & entities
│   ├── room.go                 # Room aggregate (metadata, state, access)
│   ├── entity.go               # RoomEntity, RoomUnit (position, rotation, status)
│   ├── tile.go                 # Tile, TileState, Coordinate
│   ├── layout.go               # Layout contract (heightmap grid)
│   ├── repository.go           # RoomModel repository interface
│   └── errors.go               # Domain error definitions
│
├── application/                # Use cases
│   ├── room_service.go         # Room lifecycle (load, unload, enter, exit)
│   ├── entity_service.go       # Entity management (add, remove, move)
│   ├── chat_service.go         # Chat with proximity
│   └── tests/                  # Unit tests
│       ├── room_service_test.go
│       ├── entity_test.go
│       └── chat_test.go
│
├── engine/                     # Room tick engine (goroutine isolation)
│   ├── instance.go             # Single room instance (goroutine + channels)
│   ├── manager.go              # Room instance registry (load/unload)
│   ├── tick.go                 # Tick cycle logic (entities, items, idle)
│   └── tests/
│       ├── instance_test.go
│       ├── manager_test.go
│       └── tick_test.go
│
├── pathfinding/                # A* pathfinding with 3D support
│   ├── astar.go                # Core A* algorithm with height validation
│   ├── node.go                 # PathfinderNode with priority queue
│   ├── grid.go                 # Grid adapter for Layout
│   ├── options.go              # PathfinderOptions (diagonal, height limit, 3D)
│   └── tests/
│       ├── astar_test.go
│       ├── grid_test.go
│       └── height_test.go
│
├── heightmap/                  # Heightmap parsing & validation
│   ├── parser.go               # Parse heightmap string to grid
│   ├── encoder.go              # Encode grid to client packet format
│   └── tests/
│       ├── parser_test.go
│       └── encoder_test.go
│
├── infrastructure/             # Persistence
│   ├── model/
│   │   ├── room_model.go       # GORM RoomModel (heightmap templates)
│   │   └── room_ban.go         # GORM RoomBan
│   ├── store/
│   │   ├── model_store.go      # RoomModel repository implementation
│   │   └── ban_store.go        # Ban repository implementation
│   └── migration/
│       ├── migration.go        # Room model & ban migrations
│       └── step_models.go      # Room models table
│
├── packet/                     # Protocol codecs
│   ├── constants.go            # Packet ID constants
│   ├── room_packet.go          # Room loading packets (heightmap, ready, entry)
│   ├── entity_packet.go        # Entity position/status packets
│   └── chat_packet.go          # Chat message packets
│
└── adapter/                    # External interfaces
    ├── realtime/
    │   ├── runtime.go          # WebSocket handler registration
    │   ├── dispatch.go         # Packet dispatch
    │   └── dispatch_room.go    # Room-specific handlers
    ├── httpapi/
    │   ├── routes.go           # REST API routes
    │   └── openapi.go          # OpenAPI spec
    └── command/
        ├── command.go          # CLI commands
        └── actions.go          # CLI actions
```

### Goroutine-Per-Room Model

```
                    ┌─────────────────────┐
                    │   Room Manager      │
                    │  (instance registry)│
                    └──────┬──────────────┘
                           │ Load/Unload
              ┌────────────┼────────────────┐
              ▼            ▼                ▼
     ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
     │  Room 101    │ │  Room 102    │ │  Room 103    │
     │  goroutine   │ │  goroutine   │ │  goroutine   │
     │              │ │              │ │              │
     │  ← msgChan   │ │  ← msgChan   │ │  ← msgChan   │
     │  ← ticker    │ │  ← ticker    │ │  ← ticker    │
     │  → done      │ │  → done      │ │  → done      │
     └──────────────┘ └──────────────┘ └──────────────┘
```

Each room instance runs in its own goroutine with:
- **`msgChan chan RoomMessage`** — inbound commands (enter, leave, walk, chat, etc.)
- **`ticker *time.Ticker`** — 500ms tick cycle
- **`done chan struct{}`** — shutdown signal
- **`ctx context.Context`** — cancellation propagation

**Why this model:**
- **Isolation:** A panic in one room doesn't crash others (with recover)
- **No locks:** All room state mutation happens inside the owning goroutine
- **Scalability:** Future node delegation sends serialized room state + redirects connections
- **Backpressure:** Channel capacity controls message queueing; full channels = room overloaded

### Future Node Delegation Path

```
Phase 1 (current): All rooms in-process, goroutine-per-room
Phase 2 (future):  Room Manager becomes a router
                   - Rooms can be "remote" (on another node)
                   - RoomMessage gets serialized to protobuf/msgpack
                   - Manager routes to local goroutine OR remote node via gRPC/NATS
                   - Client is transparently redirected
```

No code changes needed for Phase 1 beyond the channel-based interface. Phase 2 requires
only implementing a `RemoteRoomInstance` that forwards messages over the network.

---

## 4. Room Lifecycle

### State Machine

```
                    ┌──────────┐
         ┌──────────│  CREATED │ (DB row exists, not loaded)
         │          └────┬─────┘
         │               │ First user enters
         │          ┌────▼─────┐
         │          │ LOADING  │ (heightmap parsed, items loaded)
         │          └────┬─────┘
         │               │ Load complete
         │          ┌────▼─────┐
         ├──────────│  ACTIVE  │ ◄── tick cycle running
         │          └────┬─────┘
         │               │ Last user leaves → idle timer starts
         │          ┌────▼─────┐
         │          │   IDLE   │ (idle timer counting, still ticking)
         │          └────┬─────┘
         │               │ User re-enters → back to ACTIVE
         │               │ Idle timeout (120 cycles = 60s) → UNLOADING
         │          ┌────▼──────┐
         │          │ UNLOADING │ (save state, flush items)
         │          └────┬──────┘
         │               │
         └───────────────┘ Back to CREATED
```

### Room Loading Sequence

1. User sends `GetGuestRoom` (2230) — gets room metadata (already implemented in navigator)
2. User sends `OpenFlatConnection` (C2S) — enter request
3. Server validates: access (open/locked/password), ban check, capacity
4. Server loads room instance if not loaded (heightmap, items, bots)
5. Server creates RoomUnit for user, assigns virtual ID
6. Server sends packet sequence:
   - `RoomReady` (S2C) — room model name + room ID
   - `FloorHeightMap` (S2C) — heightmap string
   - `RoomHeightMap` (S2C) — furniture stacking heights
   - `RoomEntryTile` (S2C) — door coordinates + direction
   - `ObjectsDataUpdate` (S2C) — users already in room
   - `RoomUnit` (S2C) — avatar positions/rotations of all entities

### Room Access Control

| State | Behavior |
|-------|----------|
| **Open** | Everyone enters directly |
| **Locked** | Doorbell — owner/rights holders accept or deny |
| **Password** | Must match stored password hash (bcrypt, not plaintext) |
| **Invisible** | Hidden from navigator, only direct URL/friend follow |

**Ban system:** Timed bans with expiry. Checked before any access logic. Room owner and
users with `ACC_ENTERANYROOM` permission bypass bans.

---

## 5. Room Tick Cycle

### Tick Processing Order (per 500ms cycle)

```
1. Check idle timeout → unload if exceeded
2. Process scheduled tasks (one-shot callbacks)
3. Process entity movement (advance one step per path)
4. Process entity status updates (sit, lay, carry items, effects)
5. Process entity idle timers (sleep animation at 300 cycles = 2.5min)
6. Process item cycles (rollers, wired triggers, timed furniture)
7. Batch broadcast status updates to all room entities
8. Decrement chat flood counters (every other cycle)
```

### Performance Characteristics

| Metric | Target | Rationale |
|--------|--------|-----------|
| Tick interval | 500ms | Vendor consensus, smooth movement |
| Max entities/room | 50 (configurable) | Memory + broadcast cap |
| Max rooms loaded | 1000 (configurable) | ~500KB/room × 1000 = ~500MB |
| Idle unload | 120 ticks (60s) | Balance memory vs reload cost |
| Lag detection | 30 missed ticks = panic recovery | Prevent stuck goroutines |
| Message channel | Buffered 256 | Handle burst without blocking senders |

### Bottleneck Analysis

| Bottleneck | Impact | Solution |
|------------|--------|----------|
| **Status broadcast per tick** | O(n²) — n entities × n recipients | Batch into single packet, only broadcast on change (dirty flag) |
| **Pathfinding per walk** | A* on large maps (50×50) | Cache paths, limit re-pathing to 1 per 2 ticks per entity |
| **Room loading (cold start)** | DB queries for model + items + bots | Preload popular room models in Redis; lazy-load items on first placement |
| **Chat flood** | High-frequency chat = excessive broadcasts | Server-side flood counter: 3 messages/3s, then mute 30s |
| **Goroutine overhead** | 1000 rooms × 1 goroutine | Minimal: ~4KB stack per goroutine, total ~4MB |
| **Channel contention** | Many users walking simultaneously | Buffered channel 256 + backpressure (drop old messages under load) |

---

## 6. Pathfinding — 2D A* with 3D Extension

### Core A* Implementation

Standard A* with MinHeap priority queue, matching the vendor consensus:

- **8-directional movement** (configurable to 4)
- **Cost function:** `f(n) = g(n) + h(n)` where g = actual cost, h = Manhattan distance
- **Diagonal cost:** 14 (vs 10 for cardinal), matching Arcturus pattern
- **Height validation:** `abs(nextTile.Z - currentTile.Z) <= maxStepHeight` (default 1.5)
- **Diagonal blocking:** Modern style — both adjacent cardinals must be passable

### 3D Pathfinding Extension (Novel Feature)

Standard Habbo A* treats Z-height as a walk validation only. Our extension adds:

1. **Height-aware cost function:** Climbing costs more than descending
   - `g_height = abs(nextZ - currentZ) * heightCostMultiplier`
   - Ascending multiplier: 2.0 (climbing is slow)
   - Descending multiplier: 0.5 (going down is fast)
   - Flat: 0.0 (no extra cost)

2. **Multi-level pathfinding:** For future stacked room support
   - Z-levels treated as additional graph layers
   - Stairs/ladders connect layers at specific tiles
   - Each layer has its own heightmap grid

3. **Configurable per room:** Room owner can toggle 3D pathfinding mode
   - `PathMode: "flat"` (standard 2D) | `"height_aware"` (cost-adjusted) | `"multi_level"` (future)

**Implementation phases:**
- Phase 1: Standard 2D A* matching vendor behavior
- Phase 2: Height-aware cost function (backward compatible)
- Phase 3: Multi-level support (requires furniture stairs/ladders)

### Walk Sequence

```
Client: RoomUnitWalk(x=5, y=8)
  │
  ▼ Server (inside room goroutine)
  1. Validate destination tile exists and is walkable
  2. Run A* from entity position to (5, 8)
  3. If path found: set entity.Path = path, entity.IsWalking = true
  4. Each tick: advance entity one step along path
     - Update entity position (X, Y, Z)
     - Update body rotation toward next tile
     - Set status "mv" with target coords
     - Broadcast AvatarUpdate to room
  5. If path blocked mid-walk: recalculate or stop
  6. On arrival: clear path, clear "mv" status

Client: receives AvatarUpdate, interpolates movement over 500ms
```

---

## 7. Entity System

### Entity Types

| Type | Description | Virtual ID Range |
|------|-------------|-----------------|
| **Player** | Human user with session | 1-9999 |
| **Bot** | Server-controlled NPC | 10000-19999 |
| **Pet** | Player-owned pet entity | 20000-29999 |

### RoomUnit State

```go
type RoomUnit struct {
    VirtualID    int        // Session-local ID
    EntityType   EntityType // Player, Bot, Pet
    Position     Tile       // Current X, Y, Z
    GoalPosition *Tile      // Walk target (nil if not walking)
    Path         []Tile     // Remaining path steps
    BodyRotation int        // 0-7 (N, NE, E, SE, S, SW, W, NW)
    HeadRotation int        // 0-7
    Statuses     map[string]string // "mv", "sit", "lay", "dance", etc.
    IsWalking    bool
    IsIdle       bool
    IdleTimer    int        // Ticks since last action
    CanWalk      bool       // Movement permission
    UpdateNeeded bool       // Dirty flag for broadcast
}
```

### Status Types

| Status Key | Value Format | Description |
|------------|-------------|-------------|
| `mv` | `"x,y,z"` | Moving to position |
| `sit` | `"z"` | Sitting at height z |
| `lay` | `"z"` | Laying at height z |
| `dance` | `"style"` | Dance animation (1-4) |
| `sign` | `"id"` | Holding sign (0-17) |
| `flatctrl` | `"level"` | Room rights level indicator |
| `carry` | `"itemId"` | Carrying hand item |

---

## 8. Chat System — Proximity & Types

### Chat Types

| Type | Packet C2S | Packet S2C | Range | Description |
|------|-----------|-----------|-------|-------------|
| **Talk** | RoomUnitChat | RoomUnitChatEvent | ~14 tiles | Normal chat bubble |
| **Shout** | RoomUnitShout | RoomUnitChatShout | Room-wide | Larger bubble, all hear |
| **Whisper** | RoomUnitWhisper | RoomUnitChatWhisper | Target only | Private, only sender+target see |

### Proximity Algorithm

```
For TALK messages:
  distance = Manhattan(sender, recipient)
  if distance > CHAT_RANGE (14):
    skip recipient (they don't see the message)
  if distance > CHAT_FADE_START (10):
    client renders smaller/faded bubble (client handles this)

For SHOUT messages:
  No distance filtering — all entities in room receive

For WHISPER messages:
  Only sender entity and target entity receive
```

### Chat Flood Control

- Counter incremented per message, decremented every other tick (~1s)
- Threshold: 3 messages before flood
- Penalty: mute for 30 seconds (configurable)
- Moderators and room owners exempt

### Chat Features

- **Bubble styles:** Integer bubble ID, rank-gated (permissions check)
- **Emotions/gestures:** Server detects `:)`, `:(`, etc., sends gesture int alongside message
- **Word filter:** Room-specific + global filter lists
- **Private chat (tents):** Furniture items that create isolated chat zones

---

## 9. Room Actions

### Sitting

Triggered when user walks onto a furniture item with `canSit` property:
1. Entity arrives at sit-tile
2. Server checks furniture at tile for `canSit` flag
3. Sets entity status `sit` with height = furniture height
4. Sets body rotation to furniture rotation
5. Broadcasts status update

### Actions List

| Action | Trigger | Implementation |
|--------|---------|---------------|
| **Walk** | C2S `RoomUnitWalk(x,y)` | A* pathfind, step-per-tick |
| **Sit** | Walk to seat furniture | Auto-sit on arrival + status |
| **Lay** | Walk to bed furniture | Auto-lay on arrival + status |
| **Dance** | C2S `RoomUnitDance(style)` | Set/clear dance status |
| **Wave** | C2S `RoomUnitAction(1)` | Brief gesture animation |
| **Sign** | C2S `RoomUnitSign(id)` | Hold sign for N ticks |
| **Carry** | C2S `RoomUnitCarry(itemId)` | Hold hand item with timer |
| **Type** | C2S `RoomUnitTyping(true/false)` | Typing indicator bubble |
| **Idle** | 300 ticks no activity | Sleep animation (Zzz) |
| **Look at** | C2S `RoomUnitLookTo(x,y)` | Rotate head toward point |

---

## 10. Room Ingress Flow — Complete Protocol

### Entry Packets (Minimum Viable)

| # | Direction | Packet | ID (est.) | Purpose |
|---|-----------|--------|-----------|---------|
| 1 | C2S | `OpenFlatConnection` | 2312 | Request to enter room |
| 2 | S2C | `OpenConnectionMessageEvent` | — | Acknowledge connection |
| 3 | S2C | `RoomReady` | 2031 | Room model name + ID |
| 4 | S2C | `FloorHeightMap` | 1301 | Heightmap string |
| 5 | S2C | `RoomHeightMap` | 2753 | Stacking heights array |
| 6 | S2C | `RoomEntryTile` | 1664 | Door X, Y, direction |
| 7 | S2C | `FurnitureAliases` | — | Sprite aliases (if needed) |
| 8 | S2C | `RoomVisualizationSettings` | — | Wall/floor settings |
| 9 | S2C | `ObjectsDataUpdate` | — | Floor items |
| 10 | S2C | `WallItems` | — | Wall items |
| 11 | S2C | `RoomEntitiesStatus` | — | All entity statuses |
| 12 | S2C | `RoomUnit` | — | All entity positions |
| 13 | C2S | `RoomUnitWalk` | — | Initial user movement |

### Doorbell Flow (Locked Rooms)

```
1. Client sends OpenFlatConnection with room ID
2. Server detects LOCKED state
3. IF owner/rights holder in room:
   a. Server sends DoorbellAddUser to rights holders
   b. Server sends DoorbellRinging to requester (wait state)
   c. Rights holder accepts → normal entry flow
   d. Rights holder denies → CloseConnection with reason
4. IF no rights holder in room:
   a. Server sends FlatAccessDenied to requester
   b. Server sends HotelView to redirect back
```

### Password Flow

```
1. Client sends OpenFlatConnection with room ID + password
2. Server compares password hash (bcrypt)
3. IF match: normal entry flow
4. IF mismatch: GenericError(WRONG_PASSWORD) + HotelView
```

---

## 11. Database Schema

### New Tables

#### `room_models` — Predefined Room Templates

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | SERIAL | PK | Auto-incrementing ID |
| `slug` | VARCHAR(50) | UNIQUE, NOT NULL | Model identifier (e.g., "model_a") |
| `heightmap` | TEXT | NOT NULL | Raw heightmap string (rows separated by `\r`) |
| `door_x` | INT | NOT NULL | Door tile X coordinate |
| `door_y` | INT | NOT NULL | Door tile Y coordinate |
| `door_z` | INT | NOT NULL, DEFAULT 0 | Door tile Z height |
| `door_dir` | INT | NOT NULL, DEFAULT 2 | Door facing direction (0-7) |
| `wall_height` | INT | NOT NULL, DEFAULT -1 | Custom wall height (-1 = auto) |

#### `room_bans` — Room Access Bans

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | SERIAL | PK | Auto-incrementing ID |
| `room_id` | INT | NOT NULL, FK rooms(id) | Room reference |
| `user_id` | INT | NOT NULL, INDEX | Banned user |
| `expires_at` | TIMESTAMP | NULL | NULL = permanent |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Ban creation time |

#### `room_rights` — Room Rights List

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | SERIAL | PK | Auto-incrementing ID |
| `room_id` | INT | NOT NULL, FK rooms(id) | Room reference |
| `user_id` | INT | NOT NULL | Rights holder |
| UNIQUE | | `(room_id, user_id)` | One entry per user per room |

### Room Table Extension (Migration)

Add to existing `rooms` table:
- `model_slug` VARCHAR(50) DEFAULT 'model_a' — links to room_models
- `custom_heightmap` TEXT NULL — user-created heightmap (overrides model)
- `wall_height` INT DEFAULT -1 — custom wall height
- `floor_thickness` INT DEFAULT 0 — floor rendering thickness
- `wall_thickness` INT DEFAULT 0 — wall rendering thickness
- `password_hash` VARCHAR(255) DEFAULT '' — bcrypt hash (NOT plaintext)
- `allow_pets` BOOL DEFAULT true — pet placement allowed
- `allow_trading` BOOL DEFAULT false — trading enabled

---

## 12. Seeding — Room Models

Essential bootstrap room models matching Habbo standard layouts:

| Slug | Size | Description |
|------|------|-------------|
| `model_a` | 10×10 | Small square room |
| `model_b` | 14×10 | Medium rectangular |
| `model_c` | 18×12 | Large open space |
| `model_d` | 12×12 | L-shaped room |
| `model_e` | 16×16 | Large square |
| `model_f` | 8×8 | Tiny room |
| `model_g` | 10×14 | Vertical rectangle |
| `model_h` | 12×8 | Hallway-style |
| `model_i` | 20×20 | Extra-large |

These will be defined in `pkg/room/infrastructure/seed/` following the established pattern.

---

## 13. SDK Events

Following the cancellable mutation event pattern:

| Event | Type | Cancellable | Domain File |
|-------|------|-------------|-------------|
| `RoomLoading` | Before load | Yes | `sdk/events/room/room_loading.go` |
| `RoomLoaded` | After load | No | `sdk/events/room/room_loaded.go` |
| `RoomUnloading` | Before unload | Yes | `sdk/events/room/room_unloading.go` |
| `RoomUnloaded` | After unload | No | `sdk/events/room/room_unloaded.go` |
| `RoomEntering` | Before user enters | Yes | `sdk/events/room/room_entering.go` |
| `RoomEntered` | After user enters | No | `sdk/events/room/room_entered.go` |
| `RoomLeaving` | Before user leaves | Yes | `sdk/events/room/room_leaving.go` |
| `RoomLeft` | After user leaves | No | `sdk/events/room/room_left.go` |
| `EntityMoving` | Before walk step | Yes | `sdk/events/room/entity_moving.go` |
| `EntityMoved` | After walk step | No | `sdk/events/room/entity_moved.go` |
| `ChatSending` | Before chat | Yes | `sdk/events/room/chat_sending.go` |
| `ChatSent` | After chat | No | `sdk/events/room/chat_sent.go` |

---

## 14. Testing Plan

### Unit Tests

| Package | Tests | Coverage Focus |
|---------|-------|---------------|
| `heightmap/` | parser_test, encoder_test | All height encodings (0-35), edge cases, malformed input, escape chars |
| `pathfinding/` | astar_test, grid_test, height_test | Straight path, diagonal, blocked, height delta, no-path, large grid |
| `engine/` | instance_test, manager_test, tick_test | Goroutine lifecycle, channel message delivery, tick ordering |
| `domain/` | N/A (pure structs) | Covered by application tests |
| `application/` | room_service_test, entity_test, chat_test | Enter/leave, walk validation, chat proximity |
| `packet/` | In packet/tests/ | Encode/decode round-trips for all room packets |

### E2E Tests

| Flow | Folder | Tests |
|------|--------|-------|
| Room loading | `e2e/12_room/` | Load room model, validate heightmap parse, verify layout grid |
| Room entry | `e2e/12_room/` | Enter open room, password check, doorbell flow, ban check |
| Room walking | `e2e/12_room/` | Walk to tile, pathfinding around obstacles, height validation |
| Room chat | `e2e/12_room/` | Talk proximity, shout broadcast, whisper target-only |
| Room lifecycle | `e2e/12_room/` | Load on first user, idle unload, re-enter after unload |

**Test naming:** `e2e/12_room/12_room_test.go` per established convention.

---

## 15. Implementation Milestones

### Milestone 10.1 — Room Domain & Heightmap

**Status:** NOT STARTED
**Packages:** `pkg/room/domain/`, `pkg/room/heightmap/`, `pkg/room/infrastructure/model/`,
`pkg/room/infrastructure/store/`, `pkg/room/infrastructure/migration/`, `pkg/room/infrastructure/seed/`

| Task | Details |
|------|---------|
| Domain entities | Room, RoomUnit, Tile, TileState, Coordinate, Layout interface |
| Heightmap parser | String → grid parser with escape normalization, base-36 heights |
| Heightmap encoder | Grid → client packet format (FloorHeightMap, RoomHeightMap) |
| Room model table | Migration for `room_models` with predefined layouts |
| Room extensions | Migration adding model_slug, custom_heightmap, password_hash to rooms |
| Room model seeds | Standard models (model_a through model_i) |
| Unit tests | Parser edge cases, encoder round-trip, all height values |

### Milestone 10.2 — Pathfinding Engine

**Status:** NOT STARTED
**Packages:** `pkg/room/pathfinding/`

| Task | Details |
|------|---------|
| A* core | MinHeap + 8-directional + path trace |
| Height validation | Step height limit, no-fall configuration |
| Diagonal blocking | Modern style (both cardinals open) |
| Grid adapter | Layout → pathfinder grid conversion |
| Options | Configurable diagonal, height cost, max iterations |
| Unit tests | Straight, diagonal, blocked, height, no-path, performance |

### Milestone 10.3 — Room Engine (Goroutine Isolation)

**Status:** NOT STARTED
**Packages:** `pkg/room/engine/`

| Task | Details |
|------|---------|
| Room instance | Goroutine lifecycle, channel message loop, ticker |
| Room manager | Instance registry, load/unload, lookup by ID |
| Tick cycle | Entity movement, status broadcast, idle/lag detection |
| Panic recovery | Graceful room crash handling with logging |
| Unit tests | Instance lifecycle, message delivery, tick ordering |

### Milestone 10.4 — Room Entry & Access Control

**Status:** NOT STARTED
**Packages:** `pkg/room/application/`, `pkg/room/packet/`, `pkg/room/adapter/`

| Task | Details |
|------|---------|
| Room service | Enter, leave, kick, ban, rights management |
| Access control | Open, locked (doorbell), password (bcrypt), invisible |
| Entry packets | OpenFlatConnection, RoomReady, FloorHeightMap, etc. |
| Ban system | Room bans table, timed expiry, permission bypass |
| Rights system | Room rights table, CRUD |
| SDK events | RoomEntering/Entered, RoomLeaving/Left, RoomLoading/Loaded |
| CLI wiring | Register room services in serve_services.go, routes in serve_routes.go |

### Milestone 10.5 — Entities, Walking & Chat

**Status:** NOT STARTED
**Packages:** `pkg/room/application/`, `pkg/room/packet/`, SDK events

| Task | Details |
|------|---------|
| Entity service | Add/remove entities, position tracking, status management |
| Walking | C2S walk → pathfind → step-per-tick → broadcast |
| Sitting/laying | Auto-sit on seat furniture, status management |
| Chat service | Talk (proximity), shout (room-wide), whisper (targeted) |
| Chat flood | Counter per entity, mute penalty |
| Actions | Dance, wave, sign, carry, typing indicator |
| Idle system | Timer → sleep animation → kick after extended idle |
| SDK events | EntityMoving/Moved, ChatSending/ChatSent |

### Milestone 10.6 — E2E Tests & Integration

**Status:** NOT STARTED
**Packages:** `e2e/12_room/`

| Task | Details |
|------|---------|
| Room loading E2E | Load model, parse heightmap, verify grid |
| Room entry E2E | Open room, password, doorbell, ban |
| Walking E2E | Walk to tile, obstacle avoidance |
| Chat E2E | Talk proximity, shout, whisper |
| Lifecycle E2E | Load, idle, unload, re-enter |

---

## 16. Packet Registry — Room Realm (Core)

### Room Loading & Management

| # | Packet | ID (est.) | Dir | Milestone |
|---|--------|-----------|-----|-----------|
| 1 | `OpenFlatConnection` | 2312 | C2S | 10.4 |
| 2 | `RoomReady` | 2031 | S2C | 10.4 |
| 3 | `FloorHeightMap` | 1301 | S2C | 10.1 |
| 4 | `RoomHeightMap` | 2753 | S2C | 10.1 |
| 5 | `RoomHeightMapUpdate` | — | S2C | 10.5 |
| 6 | `RoomEntryTile` | 1664 | S2C | 10.4 |
| 7 | `RoomVisualizationSettings` | — | S2C | 10.4 |
| 8 | `CloseConnection` | — | S2C | 10.4 |
| 9 | `GetRoomSettings` | — | C2S | 10.4 |
| 10 | `RoomSettings` | — | S2C | 10.4 |
| 11 | `SaveRoomSettings` | — | C2S | 10.4 |

### Room Access

| # | Packet | ID (est.) | Dir | Milestone |
|---|--------|-----------|-----|-----------|
| 12 | `DoorbellRinging` | — | S2C | 10.4 |
| 13 | `DoorbellAddUser` | — | S2C | 10.4 |
| 14 | `LetUserIn` | — | C2S | 10.4 |
| 15 | `FlatAccessDenied` | — | S2C | 10.4 |
| 16 | `RoomEnterError` | — | S2C | 10.4 |

### Room Entities

| # | Packet | ID (est.) | Dir | Milestone |
|---|--------|-----------|-----|-----------|
| 17 | `RoomUnit` | — | S2C | 10.5 |
| 18 | `AvatarUpdate` | — | S2C | 10.5 |
| 19 | `RoomUnitWalk` | — | C2S | 10.5 |
| 20 | `RoomUnitChat` | — | C2S | 10.5 |
| 21 | `RoomUnitChatEvent` | — | S2C | 10.5 |
| 22 | `RoomUnitShout` | — | C2S | 10.5 |
| 23 | `RoomUnitChatShout` | — | S2C | 10.5 |
| 24 | `RoomUnitWhisper` | — | C2S | 10.5 |
| 25 | `RoomUnitChatWhisper` | — | S2C | 10.5 |
| 26 | `RoomUnitDance` | — | C2S | 10.5 |
| 27 | `RoomUnitAction` | — | C2S | 10.5 |
| 28 | `RoomUnitSign` | — | C2S | 10.5 |
| 29 | `RoomUnitTyping` | — | C2S | 10.5 |
| 30 | `RoomUnitLookTo` | — | C2S | 10.5 |
| 31 | `RoomUnitRemove` | — | S2C | 10.5 |
| 32 | `RoomUnitIdle` | — | S2C | 10.5 |

---

## 17. Performance Optimization Strategy

### Memory Budget (per loaded room)

| Component | Estimate | Notes |
|-----------|----------|-------|
| Heightmap grid (50×50) | ~10KB | 2500 tiles × 4 bytes |
| Pathfinding scratch | ~20KB | Allocated per pathfind call, pooled |
| Entity list (50 max) | ~50KB | 50 entities × ~1KB state |
| Item list (est. 500) | ~250KB | 500 items × ~500 bytes |
| Channel buffers | ~8KB | 256 messages × 32 bytes |
| Goroutine stack | ~4KB | Go default, grows as needed |
| **Total per room** | **~342KB** | Conservative estimate |

### Optimization Techniques

1. **Object pooling:** Reuse pathfinder node arrays via `sync.Pool`
2. **Dirty flags:** Only broadcast entity statuses that changed since last tick
3. **Batch encoding:** Encode all entity updates into one packet per tick
4. **Height cache:** Pre-compute stacking heights, update only when furniture changes
5. **Room preloading:** Load popular rooms into Redis serialized state
6. **Goroutine recycling:** Don't destroy goroutine on unload if room is likely to reload soon

### Monitoring

- Per-room tick duration histogram (detect slow rooms)
- Message channel fill level (detect backpressure)
- Loaded rooms gauge
- Entities per room histogram
- Pathfinding calls per second

---

## 18. Comparison: Our Design vs Vendors

| Aspect | PlusEMU | Arcturus | Comet | **Pixel (ours)** |
|--------|---------|----------|-------|------------------|
| **Language** | C# | Java | Java | Go |
| **Concurrency** | Task per room | Thread pool | Single thread | **Goroutine per room** |
| **Communication** | Direct method calls | Direct + events | Direct calls | **Channel messages** |
| **Lock strategy** | ConcurrentDictionary | synchronized blocks | ConcurrentHashMap | **Lock-free (channel-only)** |
| **Crash isolation** | Task catch | try/catch | try/catch | **Goroutine recover** |
| **Memory model** | Shared state + locks | Shared state + locks | Shared state + locks | **Isolated state per goroutine** |
| **Scalability** | Single process | Single process | Single process | **Single process → multi-node** |
| **Pathfinding** | In-process A* | In-process A* | In-process A* | **In-process A* + 3D extension** |
| **God objects** | RoomManager (global) | Room (has everything) | Room (has everything) | **Separated engine/service/domain** |

**Key advantage:** Our channel-based architecture eliminates shared state between rooms,
making future node delegation a transport-only change rather than an architectural rewrite.

---

## 19. Caveats & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Heightmap format inconsistency** | Client rendering bugs | Normalize at DB boundary, comprehensive parser tests |
| **Goroutine leak** | Memory exhaustion over time | Idle timeout + panic recovery + context cancellation |
| **Pathfinding timeout** | Tick delay on large maps | Max iteration limit (10000), timeout (5ms), async pathfinding |
| **Chat proximity edge cases** | Messages not received/over-received | Unit test proximity at boundary distances |
| **Room model seeding** | Missing models = can't create rooms | Validate model existence in room creation service |
| **Password stored plaintext** | Security vulnerability | Enforce bcrypt hashing at service boundary |
| **Entity virtual ID exhaustion** | Can't add more entities | Recycle IDs when entities leave |
| **Channel deadlock** | Room goroutine hangs | Timeout on channel sends, detect stuck goroutines |
| **3D pathfinding complexity** | Implementation scope creep | Phase 2 — only add after 2D is proven |
| **Custom heightmap validation** | Malformed user input | Strict parser + size limits (max 64×64) |

---

## 20. Dependencies

This plan depends on:
- ✅ Navigator realm (rooms table, categories, CRUD) — implemented
- ✅ Permission system (room access, moderation perks) — implemented
- ✅ Connection/session system (WebSocket, user sessions) — implemented
- ✅ Codec system (packet encoding/decoding) — implemented
- ❌ Furniture/item system (for room items) — partially implemented
- ❌ Bot system (for room NPCs) — not started
- ❌ Pet system — not started

Milestones 10.1-10.3 have NO blockers.
Milestone 10.4 requires only navigator + permissions (both done).
Milestone 10.5 requires entities which are self-contained.
Full furniture interaction requires completing the furniture realm.
