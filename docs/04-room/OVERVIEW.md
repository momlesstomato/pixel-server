# Room — Overview

The room realm is the heart of multiplayer interaction in pixel-server. A room
is a virtual space where avatars, bots, pets, and items exist together. The
game service maintains one goroutine per active room. Inside that goroutine
lives a single Ark ECS World.

---

## What Is Built

| Capability | Status |
|---|---|
| Room domain model (`pkg/room`) | ✅ Complete |
| In-memory repository (`pkg/room/memory`) | ✅ Complete |
| ECS component definitions | ✅ Complete |
| `RoomWorld` — mapper + filter wiring | ✅ Complete |
| `MovementSystem` | ✅ Complete |
| `ChatCooldownSystem` | ✅ Complete |
| Room packet handlers (enter, exit, chat, walk) | 🔄 In progress |
| PostgreSQL repository | 🔄 Planned |

---

## Package Layout

```
pkg/room/
├── models.go        — Room, Model structs + ErrNotFound sentinel
├── components.go    — All ECS component types + kind constants + posture constants
├── world.go         — RoomWorld, NewRoomWorld, spawn/remove helpers
├── systems.go       — MovementSystem, ChatCooldownSystem, MarkDirty, ClearDirty
├── repository.go    — Repository interface
└── memory/
    └── room.go      — In-memory RoomRepo (tests + local dev)
```

---

## Design Constraints

- **One `*ecs.World` per room goroutine.** No external goroutine reads or
  writes the World directly.
- **No I/O inside systems.** Systems are pure computation; they never block,
  publish to NATS, or touch PostgreSQL.
- **Message passing only.** External input (packet dispatches, admin commands)
  arrives through the room goroutine's `chan Envelope`.
- **Fixed 20 Hz tick.** The room goroutine drives ECS at 50 ms intervals.

---

## Key Entry Points

| Symbol | File | Purpose |
|---|---|---|
| `Room` struct | `models.go` | Persistent room state |
| `Model` struct | `models.go` | Static layout (heightmap, door position) |
| `room.Repository` | `repository.go` | CRUD interface for Room + Model |
| `RoomWorld` | `world.go` | ECS wrapper: mappers, filters, spawn methods |
| `MovementSystem()` | `systems.go` | Advances walk paths each tick |
| `ChatCooldownSystem()` | `systems.go` | Rate-limits chat every odd tick |

---

## Further Reading

| Page | Contents |
|---|---|
| [DATA-MODELS.MD](DATA-MODELS.MD) | Room, Model struct fields; Repository interface; in-memory impl |
| [ECS-COMPONENTS.MD](ECS-COMPONENTS.MD) | All ECS component structs + constants |
| [SYSTEMS.MD](SYSTEMS.MD) | Tick-driven systems; RoomWorld spawn helpers; game loop wiring |
| [PLUGIN-HOOKS.MD](PLUGIN-HOOKS.MD) | Planned plugin events, interceptors, realm relations |
