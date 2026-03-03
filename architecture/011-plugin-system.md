# Plugin System (ECS-Aligned)

## Purpose

The plugin framework extends service behavior without violating core architecture:

- One ECS World per room goroutine remains authoritative.
- Cross-service communication remains NATS-only.
- Plugins never get direct mutable access to room ECS internals.

This document defines the required model for `pkg/plugin` and `services/game` integration.

---

## Design position

Minecraft-like plugin UX (drop-in binary, lifecycle callbacks, event listeners) is a useful **operator experience** reference, not the architecture source of truth.

The source of truth is pixel-server architecture:

- `003-service-topology.md` for service and NATS boundaries.
- `004-ecs-ark.md` for room/world ownership and tick model.
- `008-patterns.md` for hexagonal dependency direction.

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
- Cross-service direct calls that bypass NATS.

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

`ServiceScope` identifies runtime location (`Name`, `NodeID`, `Version`) to support distributed observability.

### Room facade

`RoomService` is command-oriented and read-oriented. It must preserve room goroutine ownership:

- `Snapshot(roomID)` returns immutable state view.
- `BroadcastPacket(roomID, headerID, payload)` delegates to the room owner path.
- `EmitEvent(event)` publishes through the in-process bus.

---

## Integration path for game service

`services/game` must wire plugin components in this order:

1. Build shared `EventBus` and `PacketInterceptor`.
2. Build a `RoomService` adapter backed by room worker channels.
3. Build `SimpleAPIProvider` (or custom provider) with `ServiceScope`.
4. Create `Registry` and call `LoadAll()` on startup.
5. Call `DisableAll()` during graceful shutdown.

Packet flow integration:

```
gateway -> NATS room.input -> game router
    -> interceptor.RunBefore
    -> default domain handler
    -> interceptor.RunAfter
    -> NATS session.output
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
- Integration tests when game service wiring is introduced.
- Documentation updates in this file and `AGENTS.md`.
