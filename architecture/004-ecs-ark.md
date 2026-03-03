# ECS with Ark

## Correction Notice

Previous drafts referenced `github.com/mlange-42/arche` (Arche). The correct library is  
**`github.com/mlange-42/ark`** (Ark) — a distinct, newer, and more feature-rich project by the same author.  
All code samples below use the Ark v0.7.x API.

---

## Should we use ECS at all?

### What ECS solves in a hotel server

A room contains four types of entities: **Avatars** (players), **Bots**, **Pets**, and **Items (furniture)**. Each type shares a subset of traits (position, walkability, status) but has exclusive traits (chat cooldown for avatars, AI state for bots/pets, item interaction state for furniture).

Legacy emulators model this with inheritance hierarchies:

```
RoomUnit → Habbo
         → Bot
         → Pet
HabboItem (flat, but state grows via interaction sub-types)
```

This leads to:
- Fat objects with fields irrelevant to most ticks.
- Iteration over heterogeneous collections requiring `instanceof` checks.
- Difficult composition of new entity types (e.g. "quest NPC" requires subclassing again).

ECS separates **data** (components) from **logic** (queries/systems) and stores components in contiguous slices by archetype. Iterating all entities with `Position + WalkPath` is a sequential memory scan, not a pointer chase.

### Verdict: Yes, use Ark

`github.com/mlange-42/ark` **v0.7.1** (latest as of March 2026) is the correct choice:

| Property | Details |
|---|---|
| Architecture | Archetype-based → contiguous component storage per archetype |
| API style | Typed generics (`NewMap2[T1,T2]`, `NewFilter3[T1,T2,T3]`, …) — fully type-safe, no reflection in hot paths |
| Entity relationships | First-class support via `ecs.Relation` — models owner→pet, group→member without manual join tables |
| Event system | Built-in filterable events (`EntityCreated`, `EntityRemoved`, `ComponentAdded/Removed`) |
| Batch operations | `batch.NewBatchQ[T]` for mass component mutation without iterating manually |
| World serialization | `ark-serde` library for JSON World snapshots — useful for room save/restore |
| Scheduler/systems | `ark-tools` provides a `Scheduler` and `System` interface if desired |
| Dependencies | **Zero** external dependencies in the core `ecs` package |
| Coverage | 100% test coverage; 222 GitHub stars; MIT + Apache 2.0 |
| Scope per room | One `*ecs.World` per room goroutine — no global state |

---

## Installation

```
go get github.com/mlange-42/ark@v0.7.1
go get github.com/mlange-42/ark-serde@latest   # optional: room serialization
go get github.com/mlange-42/ark-tools@latest   # optional: system scheduler
```

---

## Component definitions (`pkg/ecs/components.go`)

```go
package ecs

import "github.com/mlange-42/ark/ecs"

// Position in tile-space. Z is stack height (e.g. 1.0, 1.5, 2.0 …).
type Position struct {
    X, Y float32
    Z    float32
}

// TileRef is the grid-snapped tile index for collision and pathfinding lookups.
type TileRef struct {
    X, Y int16
}

// WalkPath holds the ordered walk steps assigned to this entity.
type WalkPath struct {
    Steps  []PathStep // from pkg/pathfinding
    Cursor int
}

// EntityKind distinguishes the simulation role of an entity.
type EntityKind struct {
    Kind uint8 // KindAvatar=1  KindBot=2  KindPet=3  KindItem=4
}

// AvatarID links an ECS entity to the database user record and room-scoped unit index.
type AvatarID struct {
    UserID   int64
    RoomUnit int32 // room-scoped unit index sent to clients
}

// Status encodes posture and visual effects as compact bit fields.
type Status struct {
    Posture uint8  // sit=1 stand=2 lay=3 wave=4 …
    Effects uint32 // bitmask of active effect IDs
}

// ChatCooldown tracks the rate-limiter counter (decremented every odd tick).
type ChatCooldown struct {
    Counter int32
}

// BotAI is present only on bot entities.
type BotAI struct {
    Behaviour uint8
    ChatLines []string
    ChatIndex int
}

// PetAI is present only on pet entities.
// OwnerID is also modelled as an ecs.Relation (see section below).
type PetAI struct {
    HappyLevel int32
    Energy     int32
}

// ItemInteraction is present only on interactive floor/wall items.
type ItemInteraction struct {
    FurniID    int64
    ExtraData  string
    CycleCount int
}

// Dirty flags an entity as having state that must be broadcast this tick.
type Dirty struct{}
```

---

## World initialization

Ark's `World` returns a pointer from `ecs.NewWorld()` (changed in v0.7.0). Each room goroutine creates its own World:

```go
import "github.com/mlange-42/ark/ecs"

type roomWorld struct {
    world *ecs.World

    // Mappers (constructed once, reused every tick)
    avatarMapper *ecs.Map4[Position, TileRef, WalkPath, AvatarID]
    itemMapper   *ecs.Map3[Position, TileRef, ItemInteraction]
    petMapper    *ecs.Map3[Position, TileRef, PetAI]
    botMapper    *ecs.Map3[Position, TileRef, BotAI]

    // Filters (constructed once, reused every tick)
    walkFilter   *ecs.Filter3[Position, TileRef, WalkPath]
    chatFilter   *ecs.Filter2[AvatarID, ChatCooldown]
    itemFilter   *ecs.Filter2[TileRef, ItemInteraction]
    petFilter    *ecs.Filter2[TileRef, PetAI]
    dirtyFilter  *ecs.Filter1[Dirty]
}

func newRoomWorld() *roomWorld {
    w := ecs.NewWorld()
    return &roomWorld{
        world:        w,
        avatarMapper: ecs.NewMap4[Position, TileRef, WalkPath, AvatarID](w),
        itemMapper:   ecs.NewMap3[Position, TileRef, ItemInteraction](w),
        petMapper:    ecs.NewMap3[Position, TileRef, PetAI](w),
        botMapper:    ecs.NewMap3[Position, TileRef, BotAI](w),
        walkFilter:   ecs.NewFilter3[Position, TileRef, WalkPath](w),
        chatFilter:   ecs.NewFilter2[AvatarID, ChatCooldown](w),
        itemFilter:   ecs.NewFilter2[TileRef, ItemInteraction](w),
        petFilter:    ecs.NewFilter2[TileRef, PetAI](w),
        dirtyFilter:  ecs.NewFilter1[Dirty](w),
    }
}
```

**Mappers and Filters must be saved and reused** — creating them inline in a hot loop defeats the purpose (they cache archetype lookups internally).

---

## Entity lifecycle

### Player enters room

```go
entity := rw.avatarMapper.NewEntity(
    &Position{X: float32(spawnX), Y: float32(spawnY), Z: 0},
    &TileRef{X: int16(spawnX), Y: int16(spawnY)},
    &WalkPath{},
    &AvatarID{UserID: userID, RoomUnit: nextUnitID()},
)
```

### Player leaves room

```go
rw.world.RemoveEntity(entity)
```

### Item placed

```go
entity := rw.itemMapper.NewEntity(
    &Position{X: float32(x), Y: float32(y), Z: z},
    &TileRef{X: int16(x), Y: int16(y)},
    &ItemInteraction{FurniID: furniID, ExtraData: extra},
)
```

### Entity starts walking

Ark's component update mutates in-place through the query pointer — no special "set" call needed:

```go
// During handle of walk command:
query := rw.walkFilter.Query()
for query.Next() {
    _, _, path := query.Get()
    if /* this entity */ {
        path.Steps = computedPath
        path.Cursor = 0
        break
    }
}
query.Close()
```

For single-entity lookup, use a stored `ecs.Entity` handle + component mapper's `Get` method:

```go
// rw.avatarMapper.Get(entity) → (*Position, *TileRef, *WalkPath, *AvatarID)
pos, _, path, _ := rw.avatarMapper.Get(entity)
path.Steps = computedPath
path.Cursor = 0
pos.Z = targetZ
```

---

## ECS Systems (at 20 Hz)

Ark explicitly has **no built-in system runner** — systems are plain Go functions. The `ark-tools` scheduler is available for optional use. For pixel-server, plain sequential function calls inside the tick loop are preferred for simplicity and predictable execution order.

```go
func (w *roomWorker) tick() {
    // 1. Advance walk paths, update Position + TileRef
    MovementSystem(rw)
    // 2. Detect arrival, trigger sit/stand posture
    ArrivalSystem(rw)
    // 3. Tick roller items, move entities on conveyor tiles
    RollerSystem(rw)
    // 4. Cycle interactive items (gate timers, crackables, etc.)
    ItemInteractionSystem(rw)
    // 5. Update pet happiness, energy, follow logic
    PetAISystem(rw)
    // 6. Bot chat timers, movement schedules
    BotAISystem(rw)
    // 7. Decrement chat rate-limit counters on odd ticks
    ChatCooldownSystem(rw)
    // 8. Evaluate WIRED condition → effect chains
    WiredSystem(rw)
    // 9. Collect Dirty entities, compose s2c packets, publish to NATS
    BroadcastSystem(rw)
}
```

### MovementSystem (full example)

```go
func MovementSystem(rw *roomWorld) {
    query := rw.walkFilter.Query()
    for query.Next() {
        pos, tile, path := query.Get()
        if path.Cursor >= len(path.Steps) {
            continue
        }
        step := path.Steps[path.Cursor]
        pos.X = float32(step.X)
        pos.Y = float32(step.Y)
        pos.Z = step.Z
        tile.X = int16(step.X)
        tile.Y = int16(step.Y)
        path.Cursor++
    }
    // query.Close() is implicit when the loop exhausts all entities;
    // call explicitly if breaking early.
}
```

### ChatCooldownSystem (odd-tick only)

```go
func ChatCooldownSystem(rw *roomWorld, tick uint64) {
    if tick%2 == 0 {
        return // only every other tick
    }
    query := rw.chatFilter.Query()
    for query.Next() {
        _, cooldown := query.Get()
        if cooldown.Counter > 0 {
            cooldown.Counter--
        }
    }
}
```

---

## Entity Relationships (Ark first-class feature)

Ark supports entity relationships natively, which is directly useful for:

- Pet → Owner: `ecs.Relation` from pet entity to its owner entity.
- Item → Room: linking items to their room world (less useful when each room has its own World, but useful for cross-room lookups in game-svc's supervisor).
- Group → Member: managed in the social service's own world.

```go
// Example: pet owns a relation to its owner avatar
petOwnerRel := ecs.NewRelation[PetAI]()  // PetAI is the relation component type

// When placing a pet:
petEntity := rw.world.NewEntityRel(petOwnerRel.Make(ownerEntity),
    &Position{...}, &TileRef{...}, &PetAI{HappyLevel: 100, Energy: 100})

// Query all pets owned by a specific avatar:
petQuery := rw.world.Query(petOwnerRel.Filter(ownerEntity))
for petQuery.Next() {
    // found a pet owned by this avatar
}
```

---

## Event System (Ark built-in)

Use Ark's event system to trigger side effects without coupling systems:

```go
// Listen for entity removal to clean up Redis presence
ecs.Subscribe(rw.world, func(world *ecs.World, evt ecs.EntityEvent) {
    if !evt.Contains(ecs.ComponentID[AvatarID](world)) {
        return
    }
    _, _, _, avatarID := rw.avatarMapper.Get(evt.Entity)
    // Remove from room presence in Redis
    redisClient.SRem(ctx, fmt.Sprintf("room:presence:%d", roomID), avatarID.UserID)
})
```

---

## World Serialization (`ark-serde`)

For room state persistence (crash recovery, room templates):

```go
import arkserde "github.com/mlange-42/ark-serde"

// Save room state
jsonBytes, err := arkserde.Serialize(rw.world)

// Restore room state
err = arkserde.Deserialize(rw.world, jsonBytes)
```

This replaces the manual per-item flush loops in legacy emulators. The full ECS world snapshot is written to Redis on clean shutdown and reloaded on startup, minimising DB load.

---

## Why NOT to use ECS everywhere

ECS is appropriate inside `game-svc` for room simulation only.

| Service | Use ECS? | Reason |
|---|---|---|
| `auth` | No | Stateless request/response |
| `catalog` | No | DB read dominated |
| `navigator` | No | Filtered SQL queries |
| `social` | No | Event-driven, no simulation loop |
| `game` | **Yes** | Tight simulation loop, 4 entity types, >200 entities per room |

---

## Performance estimate (Ark v0.7.x)

Ark's benchmarks (from its published benchmark suite) show ~2–5 ns per entity per query iteration on hot archetypes. For pixel-server:

| System | Entities | Estimate per tick |
|---|---|---|
| MovementSystem | 200 walking entities | ~1–2 µs |
| ItemInteractionSystem | 100 interactive items | ~3–10 µs |
| BroadcastSystem (dirty scan) | 200 entities | ~5–20 µs (NATS publish dominates) |
| All systems combined | 200+100 | ~30–80 µs per 50 ms tick |

At 100 active rooms per `game-svc` pod, total ECS CPU ≈ 8 ms/s = well under 1% of one core. I/O (NATS, Redis) dominates at scale, not ECS iteration.
