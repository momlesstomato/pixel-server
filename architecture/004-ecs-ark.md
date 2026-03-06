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
RoomUnit -> Habbo
          -> Bot
          -> Pet
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
| Architecture | Archetype-based: contiguous component storage per archetype |
| API style | Typed generics (`NewMap2[T1,T2]`, `NewFilter3[T1,T2,T3]`, ...) — fully type-safe, no reflection in hot paths |
| Entity relationships | First-class support via `ecs.Relation` — models owner/pet, group/member without manual join tables |
| Event system | Built-in filterable events (`EntityCreated`, `EntityRemoved`, `ComponentAdded/Removed`) |
| Batch operations | `batch.NewBatchQ[T]` for mass component mutation without iterating manually |
| World serialization | `ark-serde` library for JSON World snapshots — useful for room save/restore |
| Scheduler/systems | `ark-tools` provides a `Scheduler` and `System` interface if desired |
| Dependencies | **Zero** external dependencies in the core `ecs` package |
| Coverage | 100% test coverage; MIT + Apache 2.0 |
| Scope per room | One `*ecs.World` per room goroutine — no global state |

---

## Installation

```
go get github.com/mlange-42/ark@v0.7.1
go get github.com/mlange-42/ark-serde@latest   # optional: room serialization
go get github.com/mlange-42/ark-tools@latest   # optional: system scheduler
```

---

## Component definitions (`internal/game/domain/components.go`)

```go
package domain

import "github.com/mlange-42/ark/ecs"

// Position in tile-space. Z is stack height (e.g. 1.0, 1.5, 2.0 ...).
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
    Posture uint8  // sit=1 stand=2 lay=3 wave=4 ...
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

ECS components are **domain types** and live in `internal/game/domain/`. They have no framework imports except `github.com/mlange-42/ark/ecs` which is the ECS framework itself — this is an accepted domain-level dependency for the game realm specifically.

---

## World initialization

Ark's `World` returns a pointer from `ecs.NewWorld()` (changed in v0.7.0). Each room worker goroutine creates its own World:

```go
// internal/game/domain/world.go

type RoomWorld struct {
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

func NewRoomWorld() *RoomWorld {
    w := ecs.NewWorld()
    return &RoomWorld{
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

## Room Worker Integration

The room worker is the primary consumer of ECS. It lives in `internal/game/domain/` and owns one `RoomWorld`:

```go
// internal/game/domain/worker.go

type RoomWorker struct {
    roomID    int64
    rw        *RoomWorld
    inbox     chan Envelope
    sessions  map[string]SessionWriter // active sessions in this room
    tickCount uint64
}

func (w *RoomWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(50 * time.Millisecond) // 20 Hz
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            w.shutdown()
            return
        case env := <-w.inbox:
            w.handleCommand(env)
        case <-ticker.C:
            w.tick()
            w.tickCount++
        }
    }
}
```

`SessionWriter` is a port interface — in all-in-one mode it writes directly to the WebSocket connection; in distributed mode it publishes via NATS.

---

## ECS Systems (at 20 Hz)

Systems are plain Go functions called sequentially inside the tick loop:

```go
func (w *RoomWorker) tick() {
    MovementSystem(w.rw)
    ArrivalSystem(w.rw)
    RollerSystem(w.rw)
    ItemInteractionSystem(w.rw)
    PetAISystem(w.rw)
    BotAISystem(w.rw)
    ChatCooldownSystem(w.rw, w.tickCount)
    WiredSystem(w.rw)
    BroadcastSystem(w.rw, w.sessions)
}
```

### MovementSystem (full example)

```go
func MovementSystem(rw *RoomWorld) {
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
}
```

### BroadcastSystem

In the single-binary model, `BroadcastSystem` writes directly to session WebSocket connections via the `SessionWriter` port. No NATS round-trip for state broadcasts:

```go
func BroadcastSystem(rw *RoomWorld, sessions map[string]SessionWriter) {
    query := rw.dirtyFilter.Query()
    for query.Next() {
        // Encode entity state update packet
        // Write to every session in the room
        for _, sess := range sessions {
            sess.Send(encoded)
        }
    }
}
```

---

## Entity Relationships (Ark first-class feature)

Ark supports entity relationships natively:

- Pet to Owner: `ecs.Relation` from pet entity to its owner entity.
- Item to Room: linking items to their room world.

```go
petOwnerRel := ecs.NewRelation[PetAI]()

petEntity := rw.world.NewEntityRel(petOwnerRel.Make(ownerEntity),
    &Position{...}, &TileRef{...}, &PetAI{HappyLevel: 100, Energy: 100})

petQuery := rw.world.Query(petOwnerRel.Filter(ownerEntity))
for petQuery.Next() {
    // found a pet owned by this avatar
}
```

---

## World Serialization (`ark-serde`)

For room state persistence (crash recovery, room templates):

```go
import arkserde "github.com/mlange-42/ark-serde"

// Save room state to Redis
jsonBytes, err := arkserde.Serialize(rw.world)
redisClient.Set(ctx, fmt.Sprintf("room:snapshot:%d", roomID), jsonBytes, 0)

// Restore room state from Redis
data, _ := redisClient.Get(ctx, fmt.Sprintf("room:snapshot:%d", roomID)).Bytes()
err = arkserde.Deserialize(rw.world, data)
```

This enables crash recovery in distributed mode: when a game worker restarts, it loads room state from Redis snapshots instead of rebuilding from PostgreSQL.

---

## Why NOT to use ECS everywhere

ECS is appropriate inside the `game` realm for room simulation only.

| Realm | Use ECS? | Reason |
|---|---|---|
| `auth` | No | Stateless request/response |
| `catalog` | No | DB read dominated |
| `navigator` | No | Filtered SQL queries |
| `social` | No | Event-driven, no simulation loop |
| `game` | **Yes** | Tight simulation loop, 4 entity types, >200 entities per room |

---

## Performance estimate (Ark v0.7.x)

| System | Entities | Estimate per tick |
|---|---|---|
| MovementSystem | 200 walking entities | ~1-2 us |
| ItemInteractionSystem | 100 interactive items | ~3-10 us |
| BroadcastSystem (dirty scan) | 200 entities | ~5-20 us (session write dominates) |
| All systems combined | 200+100 | ~30-80 us per 50 ms tick |

At 100 active rooms per process, total ECS CPU is approximately 8 ms/s — well under 1% of one core. I/O (Redis, storage adapters) dominates at scale, not ECS iteration.
