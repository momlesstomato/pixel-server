# Plugin System (ECS-Aligned)

## Purpose

The plugin framework extends realm module behavior without violating core architecture:

- One ECS World per room goroutine remains authoritative.
- Cross-module communication remains contract-topic based.
- Plugins never get direct mutable access to room ECS internals.

This document defines the required model for `pkg/plugin` and `internal/game` integration.

---

## Design position

Minecraft-like plugin UX (drop-in binary, lifecycle callbacks, event listeners) is a useful **operator experience** reference, not the architecture source of truth.

The source of truth is pixelsv architecture:

- [003-service-topology.md](003-service-topology.md) for module and contract boundaries.
- [004-ecs-ark.md](004-ecs-ark.md) for room/world ownership and tick model.
- [008-patterns.md](008-patterns.md) for hexagonal dependency direction.

---

## Runtime model

### Loading

- Plugins are compiled as `.so` (Linux/macOS) and discovered from `plugins/`.
- Each plugin exports `func NewPlugin() plugin.Plugin`.
- Registry loads all binaries, sorts by dependency metadata, and enables them.

### Lifecycle

1. `LoadAll()` discovers plugin binaries.
2. `EnableAll()` calls `OnEnable(api)` in dependency order.
3. `DisableAll()` calls `OnDisable()` in reverse order.

The registry wraps API adapters so every event/packet subscription is tracked and automatically deregistered on disable.

---

## ECS safety requirements

### Allowed plugin operations

- Subscribe to in-process events (`EventBus`).
- Register packet interceptors (`PacketInterceptor`).
- Read room snapshots through `RoomService.Snapshot`.
- Issue room-safe commands through `RoomService` methods.

### Forbidden plugin operations

- Direct access to `*ecs.World`.
- Blocking I/O inside event handlers executed on room goroutines.
- Cross-module direct calls that bypass contracts.

### Event context contract

Each event carries at minimum:

- `Name` (`EventName`)
- `RoomID`
- `Tick`
- `Entity` (`EntityRef`)
- `Payload` (`any`)

Handlers may call `Cancel()` only for cancellable event types.

---

## API surface

```go
type API interface {
    Scope() ServiceScope
    Events() EventBus
    Packets() PacketInterceptor
    Rooms() RoomService
    Logger() *slog.Logger
    Config() []byte
}
```

### Scope

`ServiceScope` identifies runtime location (`Name`, `NodeID`, `Version`) to support observability. In distributed mode, `NodeID` corresponds to `PIXELSV_INSTANCE_ID`.

### Room facade

`RoomService` is command-oriented and read-oriented. It must preserve room goroutine ownership:

- `Snapshot(roomID)` returns immutable state view.
- `BroadcastPacket(roomID, headerID, payload)` delegates to the room worker inbox channel.
- `EmitEvent(event)` publishes through the in-process event bus.

---

## Integration path for game realm

`internal/game/` must wire plugin components in this order:

1. Build shared `EventBus` and `PacketInterceptor`.
2. Build a `RoomService` adapter backed by room worker channels.
3. Build `SimpleAPIProvider` (or custom provider) with `ServiceScope`.
4. Create `Registry` and call `LoadAll()` on startup.
5. Call `DisableAll()` during graceful shutdown.

Packet flow integration:

```
WebSocket frame -> game adapter -> packet router
    -> interceptor.RunBefore
    -> default domain handler
    -> interceptor.RunAfter
    -> session.Send (direct or via NATS)
```

---

## Operational limits

- Go plugin runtime is Linux/macOS only.
- Plugin and host must use the same Go toolchain version.
- Plugins cannot be unloaded by Go runtime once loaded.

These constraints are accepted for Phase 0/1. Future adapter option: WASM runtime preserving the same API.

---

## Quality gates

Any change to plugin contracts must include:

- Unit tests for event, interceptor, registry behavior.
- Integration tests when game realm wiring is introduced.
- Documentation updates in this file and `AGENTS.md`.
