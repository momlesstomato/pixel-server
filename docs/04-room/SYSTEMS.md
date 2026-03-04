# Room — Systems and World

Detailed description of the ECS systems, the `RoomWorld` wrapper, and how the
game loop drives them.

---

## RoomWorld (`pkg/room.RoomWorld`)

`RoomWorld` is a thin wrapper around an Ark `ecs.World`. It creates all
mappers and filters once at startup and reuses them every tick, which avoids
per-tick heap allocations.

### Mappers

Mappers allow creating and reading/writing specific component combinations on
a single entity.

| Field | Ark type | Components |
|---|---|---|
| `AvatarMapper` | `Map4` | `Position`, `TileRef`, `WalkPath`, `AvatarID` |
| `ItemMapper` | `Map3` | `Position`, `TileRef`, `ItemInteraction` |
| `PetMapper` | `Map3` | `Position`, `TileRef`, `PetAI` |
| `BotMapper` | `Map3` | `Position`, `TileRef`, `BotAI` |
| `StatusMapper` | `Map1` | `Status` |
| `CooldownMapper` | `Map1` | `ChatCooldown` |
| `DirtyMapper` | `Map1` | `Dirty` |
| `KindMapper` | `Map1` | `EntityKind` |

### Filters

Filters iterate all entities that carry a given component set.

| Field | Ark type | Iterated components | Used by |
|---|---|---|---|
| `WalkFilter` | `Filter3` | `Position`, `TileRef`, `WalkPath` | `MovementSystem` |
| `ChatFilter` | `Filter2` | `AvatarID`, `ChatCooldown` | `ChatCooldownSystem` |
| `ItemFilter` | `Filter2` | `TileRef`, `ItemInteraction` | Item interaction systems |
| `PetFilter` | `Filter2` | `TileRef`, `PetAI` | Pet AI system |
| `BotFilter` | `Filter2` | `TileRef`, `BotAI` | Bot AI system |
| `DirtyFilter` | `Filter1` | `Dirty` | Broadcast / clear dirty |

### Constructor

```go
rw := room.NewRoomWorld()
```

All mappers and filters are initialised before the room goroutine's first tick.

---

## Spawn Helpers

These helpers create an entity with its core component set already attached.
They accept only the data strictly required at spawn time; optional components
(e.g. `Status`, `ChatCooldown`) are added automatically where appropriate.

### `SpawnAvatar`

```go
func (rw *RoomWorld) SpawnAvatar(userID int64, roomUnit int32, x, y int16, z float32) ecs.Entity
```

Creates a player avatar. Automatically attaches:
- `Position{X, Y, Z}` and `TileRef{X, Y}`
- `WalkPath{}` (empty path)
- `AvatarID{UserID, RoomUnit}`
- `EntityKind{Kind: KindAvatar}`
- `Status{Posture: PostureStand}`
- `ChatCooldown{}`

### `SpawnBot`

```go
func (rw *RoomWorld) SpawnBot(behaviour uint8, chatLines []string, x, y int16, z float32) ecs.Entity
```

Creates a bot NPC. Attaches `Position`, `TileRef`, `BotAI`, `EntityKind{KindBot}`.

### `SpawnPet`

```go
func (rw *RoomWorld) SpawnPet(happy, energy int32, x, y int16, z float32) ecs.Entity
```

Creates a pet. Attaches `Position`, `TileRef`, `PetAI{HappyLevel, Energy}`,
`EntityKind{KindPet}`.

### `SpawnItem`

```go
func (rw *RoomWorld) SpawnItem(furniID int64, extraData string, x, y int16, z float32) ecs.Entity
```

Creates a furniture item. Attaches `Position`, `TileRef`,
`ItemInteraction{FurniID, ExtraData}`, `EntityKind{KindItem}`.

### `RemoveEntity`

```go
func (rw *RoomWorld) RemoveEntity(entity ecs.Entity)
```

Removes the entity and all its components from the World. Safe to call from
the room goroutine between ticks.

---

## Systems (20 Hz tick)

Systems are pure functions — they accept a `*RoomWorld` and produce no side
effects outside in-memory component mutations. All systems run sequentially
inside the room goroutine.

### `MovementSystem(rw *RoomWorld)`

Advances every entity that has a non-empty `WalkPath` by one step per call.

**Algorithm:**
1. Open `WalkFilter.Query()` to iterate all entities with `Position`,
   `TileRef`, `WalkPath`.
2. For each entity: if `path.HasSteps()` is false, skip.
3. Otherwise, read `path.Current()` → `step`.
4. Set `pos.X = float32(step.X)`, `pos.Y = float32(step.Y)`, `pos.Z = step.Z`.
5. Set `tile.X = step.X`, `tile.Y = step.Y`.
6. Call `path.Advance()` to move the cursor forward.

**Invariant:** After `MovementSystem`, `Position` and `TileRef` always match the
most recently consumed `PathStep`.

### `ChatCooldownSystem(rw *RoomWorld, tick uint64)`

Decrements chat rate-limit counters to prevent chat flood.

**Algorithm:**
1. If `tick % 2 == 0`, return early (runs only on odd ticks → 10 Hz).
2. Open `ChatFilter.Query()` to iterate entities with `AvatarID` + `ChatCooldown`.
3. If `cooldown.Counter > 0`, decrement by 1.

A chat handler sets `Counter` to a configured burst value when the user speaks.
The system drains it over time, re-enabling chat once `Counter == 0`.

### `MarkDirty(rw *RoomWorld, entity)`

Adds the `Dirty` marker component to an entity so the broadcast system picks
it up this tick. If `Dirty` is already present it must not be added again
(Ark panics on duplicate component addition).

### `ClearDirty(rw *RoomWorld)`

Removes `Dirty` from every entity that has it. Called at the end of each tick
after the broadcast pass.

**Constraint:** Components cannot be removed during an active Ark query
iteration; `ClearDirty` closes the query first, collects entities, then
removes the component.

---

## Game Loop Wiring

The game service drives room goroutines at a fixed 20 Hz. Per tick:

```
1. Drain inbound Envelope channel (feed external input into ECS).
2. MovementSystem(rw)
3. ChatCooldownSystem(rw, tick)
4. [future] BotAI system
5. [future] Broadcast dirty entities to clients
6. ClearDirty(rw)
7. tick++
```

Systems 3–6 are executed unconditionally every 50 ms regardless of player count.
