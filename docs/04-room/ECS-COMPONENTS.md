# Room — ECS Components

All component structs defined in `pkg/room/components.go`. Components are pure
data — no methods aside from small convenience helpers on `WalkPath`. Logic
lives in systems.

---

## Spatial Components

### `Position`

Continuous floating-point position in tile-space. Z is stack height (0.0 = floor).

| Field | Type | Description |
|---|---|---|
| `X` | `float32` | World X coordinate (tile fraction allowed during movement) |
| `Y` | `float32` | World Y coordinate |
| `Z` | `float32` | Stack height (0.0 = floor, 1.0 = one standard furniture height) |

### `TileRef`

Grid-snapped tile index. Used for collision detection and pathfinding lookups.
Always updated to match the tile an entity fully occupies.

| Field | Type | Description |
|---|---|---|
| `X` | `int16` | Tile column (integer) |
| `Y` | `int16` | Tile row (integer) |

---

## Movement Components

### `PathStep`

A single waypoint on a walk path.

| Field | Type | Description |
|---|---|---|
| `X` | `int16` | Target tile X |
| `Y` | `int16` | Target tile Y |
| `Z` | `float32` | Target stack height at this tile |

### `WalkPath`

Ordered list of `PathStep` waypoints assigned to an entity. Consumed one step
per tick by `MovementSystem`.

| Field | Type | Description |
|---|---|---|
| `Steps` | `[]PathStep` | Ordered waypoint slice |
| `Cursor` | `int` | Index of the next step to execute |

**Helper methods:**

| Method | Returns | Description |
|---|---|---|
| `HasSteps()` | `bool` | True if `Cursor < len(Steps)` |
| `Current()` | `PathStep` | Returns `Steps[Cursor]`, or zero if exhausted |
| `Advance()` | — | Increments `Cursor` by one |

---

## Identity Components

### `EntityKind`

Distinguishes the simulation role of an entity.

| Field | Type | Description |
|---|---|---|
| `Kind` | `uint8` | Entity type constant (see constants below) |

**Kind constants:**

| Constant | Value | Entity type |
|---|---|---|
| `KindAvatar` | `1` | Human-controlled player avatar |
| `KindBot` | `2` | Automated NPC bot |
| `KindPet` | `3` | Player-owned pet |
| `KindItem` | `4` | Placed furniture or interactive item |

### `AvatarID`

Links an ECS entity to the database user record and the room-scoped unit index
sent to clients.

| Field | Type | Description |
|---|---|---|
| `UserID` | `int64` | Database user ID |
| `RoomUnit` | `int32` | Room-scoped sequential index (sent to client as unit identifier) |

---

## Visual / Status Components

### `Status`

Compact bit-packed representation of posture and active visual effects.

| Field | Type | Description |
|---|---|---|
| `Posture` | `uint8` | Current posture constant (see below) |
| `Effects` | `uint32` | Bitmask of active visual effect IDs |

**Posture constants:**

| Constant | Value | Description |
|---|---|---|
| `PostureSit` | `1` | Entity is sitting on furniture |
| `PostureStand` | `2` | Entity is standing (default) |
| `PostureLay` | `3` | Entity is lying down |
| `PostureWave` | `4` | Entity is waving |

### `Dirty`

Marker component. Presence flags that this entity has state changes that must
be broadcast in the current tick. No fields — presence is the signal.

```go
type Dirty struct{}
```

---

## Rate-Limit Components

### `ChatCooldown`

Rate-limiter counter for the chat spam throttle. `ChatCooldownSystem` decrements
the counter on every odd tick; chat handlers check this before accepting input.

| Field | Type | Description |
|---|---|---|
| `Counter` | `int32` | Remaining cooldown ticks; 0 = allowed to chat |

---

## AI / Behaviour Components

### `BotAI`

Present only on bot entities (`KindBot`). Drives automated speech scheduling.

| Field | Type | Description |
|---|---|---|
| `Behaviour` | `uint8` | Behaviour preset ID (effect on NPC logic) |
| `ChatLines` | `[]string` | Ordered list of lines to cycle through |
| `ChatIndex` | `int` | Next line index to speak |

### `PetAI`

Present only on pet entities (`KindPet`). Tracks emotional state metrics.

| Field | Type | Description |
|---|---|---|
| `HappyLevel` | `int32` | Happiness percentage (0–100) |
| `Energy` | `int32` | Energy level percentage (0–100) |

---

## Item Component

### `ItemInteraction`

Present only on interactive floor/wall items (`KindItem`).

| Field | Type | Description |
|---|---|---|
| `FurniID` | `int64` | Furniture catalogue item ID |
| `ExtraData` | `string` | State string (colour, text, toggle state, etc.) |
| `CycleCount` | `int` | Current interaction cycle counter (driven by roller/item systems) |
