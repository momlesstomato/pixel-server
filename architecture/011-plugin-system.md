# Plugin System and Extensible Realm API

## Purpose

The plugin framework lets operators extend pixelsv without forking the codebase. Plugins hook into realm lifecycle events, intercept packets, register HTTP routes, and interact with room state — all through stable, version-gated APIs that enforce the same safety guarantees as internal realm code.

This document is the single source of truth for plugin contracts, loading semantics, realm extension points, and the boundary between "what plugins can do" and "what only internal realm code can do."

## References

- [003-service-topology.md](003-service-topology.md) — role system, transport adapter pattern
- [004-ecs-ark.md](004-ecs-ark.md) — ECS ownership, room worker model
- [007-messaging.md](007-messaging.md) — transport bus contracts
- [008-patterns.md](008-patterns.md) — hexagonal architecture rules

---

## Design Principles

1. **Realm-first, plugin-second.** All core behavior ships as internal realm code under `internal/`. Plugins extend or override behavior; they never implement primary realm logic.

2. **Contract-bound.** Plugins interact with pixelsv exclusively through the `plugin.API` interface. No exported internal types leak to plugin code.

3. **ECS-safe.** Plugins never receive `*ecs.World`. Room state is accessible only through command envelopes (writes) and snapshots (reads).

4. **Transport-agnostic.** Plugin event hooks are always in-process. Plugins don't know or care whether the host runs all-in-one or distributed.

5. **Fail-safe lifecycle.** A crashing plugin is isolated and disabled. Core realm behavior continues.

6. **Realm-scoped.** Each plugin declares which realm(s) it extends. The plugin loader only activates plugins for realms running on this process instance.

---

## Who Loads Plugins — Role Responsibility

### The `all` and individual role processes

Plugins are loaded **per-process**, not per-role. The startup orchestrator (`pkg/core/cli/startup.go`) is the sole plugin loading authority.

```
startup.go flow:
  1. Parse --role
  2. Build startup plan (transport, storage, HTTP)
  3. Load plugin registry (discover .so files)
  4. Register realm modules (filtered by active roles)
  5. Enable plugins (filtered by declared realm scope + active roles)
  6. Start HTTP listener / wait for context
  7. On shutdown: DisableAll plugins → Close realms → Close transport
```

### Why startup.go and not individual realms

- A plugin may target multiple realms (e.g., a "logging" plugin for auth + game).
- Plugin dependency resolution requires a global view of all loaded plugins.
- Auto-deregistration on shutdown must happen in one place.
- Realm `Register()` functions receive a `plugin.HookRegistry` they can pull hooks from, not the raw plugin loader.

### Role filtering

A plugin declaring `Realms: ["game", "social"]` loaded on a `--role=gateway` process is **skipped** — its `OnEnable` is never called. This prevents plugins from consuming resources on processes that don't run their target realms.

```go
// Pseudo-code in startup.go
for _, p := range registry.Plugins() {
    if !roles.intersects(p.Metadata().Realms) {
        logger.Debug("skipping plugin (realm not active)", zap.String("plugin", p.Metadata().Name))
        continue
    }
    if err := p.OnEnable(api); err != nil {
        logger.Error("plugin enable failed", zap.String("plugin", p.Metadata().Name), zap.Error(err))
        continue  // skip, don't crash the server
    }
}
```

---

## Plugin Contract

### Metadata

Every plugin provides immutable metadata used for discovery, dependency ordering, and role filtering.

```go
// pkg/plugin/metadata.go

type Metadata struct {
    // Name is the unique plugin identifier (e.g., "custom-commands").
    Name string
    // Version follows semver (e.g., "1.2.0").
    Version string
    // Realms lists realm names this plugin targets.
    // Empty means "all realms" (rare; prefer explicit).
    Realms []string
    // DependsOn lists plugin names that must load before this one.
    DependsOn []string
}
```

### Plugin interface

```go
// pkg/plugin/plugin.go

type Plugin interface {
    // Metadata returns immutable plugin identity.
    Metadata() Metadata
    // OnEnable is called with the API handle once the target realms are active.
    OnEnable(api API) error
    // OnDisable is called during shutdown in reverse dependency order.
    OnDisable() error
}
```

### Plugin API

The `API` is the plugin's window into pixelsv. Every method returns a narrow, purpose-built interface — not raw framework types.

```go
// pkg/plugin/api.go

type API interface {
    // Scope returns runtime identity (instance ID, version, environment).
    Scope() Scope

    // Events returns the in-process event bus for subscribing/emitting domain events.
    Events() EventBus

    // Packets returns the packet interceptor for pre/post processing hooks.
    Packets() PacketInterceptor

    // Rooms returns the room service facade for ECS-safe state access.
    Rooms() RoomService

    // HTTP returns the route registrar for adding plugin HTTP endpoints.
    HTTP() RouteRegistrar

    // Storage returns scoped key-value storage for plugin-owned data.
    Storage() PluginStore

    // Logger returns a named logger scoped to this plugin.
    Logger() *zap.Logger

    // Config returns raw plugin configuration bytes (from YAML/JSON in plugins/<name>/config.yml).
    Config() []byte
}
```

---

## API Sub-Contracts

### EventBus

In-process publish/subscribe for domain events. Always local — never bridges to NATS. Events fire synchronously within the originating goroutine's tick or handler.

```go
type EventBus interface {
    // On registers a listener for a named event type.
    // Returns a registration handle for auto-cleanup.
    On(event string, handler EventHandler) Registration

    // Emit fires an event to all registered listeners.
    Emit(event Event) error
}

type EventHandler func(event Event) error

type Event struct {
    // Name identifies the event type (e.g., "room.user.join", "auth.ticket.validated").
    Name string
    // RoomID is set for room-scoped events (-1 for non-room events).
    RoomID int64
    // SessionID is set for session-scoped events.
    SessionID string
    // Tick is set for ECS-originated events (room tick number).
    Tick uint64
    // Data carries typed event payload.
    Data any
    // cancelled tracks Cancel() state for cancellable events.
    cancelled bool
}

// Cancel marks a cancellable event as cancelled. Non-cancellable events ignore this.
func (e *Event) Cancel() { e.cancelled = true }

// Cancelled reports whether Cancel() was called.
func (e *Event) Cancelled() bool { return e.cancelled }
```

**Important distinction:** The `EventBus` is NOT the transport `Bus` from `pkg/core/transport/`. The transport bus carries inter-process messages over topics (`packet.c2s.*`, `session.output.*`). The EventBus is a local observer pattern for same-process domain hooks. They never cross:

```
┌──────────────────────────────────────────────────────────────────┐
│                       pixelsv process                            │
│                                                                  │
│  ┌─────────────┐    transport.Bus     ┌─────────────────────┐   │
│  │  gateway     │◄──(local/NATS)─────►│  auth / game / ...  │   │
│  │  ws ingress  │                     │  realm adapters      │   │
│  └─────────────┘                      └──────────┬──────────┘   │
│                                                   │              │
│                                        plugin.EventBus           │
│                                        (always local)            │
│                                                   │              │
│                                       ┌───────────▼──────────┐  │
│                                       │     plugins          │  │
│                                       │  (event listeners)   │  │
│                                       └──────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

#### Standard event names

Events follow the pattern `<realm>.<entity>.<action>`:

| Event | Realm | Cancellable | Fired When |
|---|---|---|---|
| `auth.ticket.validated` | auth | No | SSO ticket accepted, session authenticated |
| `session.connected` | gateway | No | WebSocket session established |
| `session.disconnected` | gateway | No | WebSocket session closed |
| `room.user.join` | game | Yes | User enters room (cancel = deny entry) |
| `room.user.leave` | game | No | User leaves room |
| `room.user.chat` | game | Yes | Chat message sent (cancel = mute) |
| `room.user.move` | game | Yes | Movement request (cancel = block) |
| `room.item.interact` | game | Yes | Furniture interaction (cancel = deny) |
| `room.tick.pre` | game | No | Before ECS systems run this tick |
| `room.tick.post` | game | No | After ECS systems run this tick |
| `navigator.search` | navigator | No | Room search query executed |
| `catalog.purchase` | catalog | Yes | Purchase attempt (cancel = deny) |
| `moderation.ban.issued` | moderation | No | Ban applied |

### PacketInterceptor

Plugins can wrap packet handling with before/after hooks. These run on the same goroutine as packet processing.

```go
type PacketInterceptor interface {
    // Before registers a pre-handler for a packet header ID.
    // Returning false from the hook cancels default handling.
    Before(headerID uint16, hook PacketHook) Registration

    // After registers a post-handler for a packet header ID.
    After(headerID uint16, hook PacketHook) Registration

    // BeforeAll registers a pre-handler for ALL packets.
    BeforeAll(hook PacketHook) Registration

    // AfterAll registers a post-handler for ALL packets.
    AfterAll(hook PacketHook) Registration
}

type PacketHook func(ctx PacketContext) bool

type PacketContext struct {
    // SessionID identifies the sending session.
    SessionID string
    // HeaderID is the packet header identifier.
    HeaderID uint16
    // Payload is the raw packet payload bytes (read-only).
    Payload []byte
    // Realm is the target realm name.
    Realm string
}
```

#### Integration with gateway ingress

The packet interception point lives in the gateway's `handleBinary` path (`pkg/http/ws/ingress.go`):

```
WebSocket frame
  → codec.SplitFrames
  → protocol.DecodeC2S
  → interceptor.RunBefore(packetCtx)     ← plugin hook
    → if cancelled: skip
  → bus.Publish(packet.c2s.<realm>.<session>)
  → interceptor.RunAfter(packetCtx)      ← plugin hook
```

For realm-side packet handling (e.g., auth's `Subscriber.handlePacket` in `internal/auth/adapters/transport/subscriber.go`), the realm's handler emits events through the EventBus before/after processing, giving plugins a second interception layer at the domain level.

### RoomService

The room service facade enforces ECS single-writer ownership. All mutations go through the room worker's inbox channel (`RoomWorker.inbox chan Envelope` from 004-ecs-ark.md).

```go
type RoomService interface {
    // Snapshot returns an immutable view of room ECS state.
    // Returns nil if the room is not loaded on this process.
    Snapshot(roomID int64) (*RoomSnapshot, error)

    // SendCommand enqueues a typed command to the room worker inbox.
    SendCommand(roomID int64, command RoomCommand) error

    // BroadcastPacket sends an encoded packet to all sessions in a room.
    BroadcastPacket(roomID int64, headerID uint16, payload []byte) error
}

type RoomSnapshot struct {
    RoomID      int64
    Tick        uint64
    EntityCount int
    Avatars     []AvatarSnapshot
    Items       []ItemSnapshot
    Bots        []BotSnapshot
    Pets        []PetSnapshot
}

type AvatarSnapshot struct {
    UserID   int64
    RoomUnit int32
    X, Y     float32
    Z        float32
    Posture  uint8
}

type RoomCommand struct {
    Type    string
    Payload any
}
```

#### Why snapshots, not live references

- ECS components are value types stored in archetype tables. Returning pointers would allow mutation outside the tick loop, violating single-writer ownership.
- `ark-serde` serialization produces an immutable view safe to read from any goroutine.
- Snapshot cost at 200 entities is ~50-100 µs — acceptable for plugin queries at reasonable rates.

### RouteRegistrar

Plugins can mount HTTP endpoints under a scoped prefix.

```go
type RouteRegistrar interface {
    // Group returns a Fiber route group scoped to /api/v1/plugins/<pluginName>/.
    // The group inherits the server's API key middleware.
    Group() fiber.Router
}
```

- Prevents plugins from shadowing core realm routes (`/api/v1/auth/tickets`).
- Only available if the process runs an HTTP-serving role (`all`, `gateway`, `api`).
- If `HTTP()` returns nil (no HTTP listener on this role), the plugin must handle this gracefully.

### PluginStore

Scoped key-value storage backed by Redis with automatic key prefixing.

```go
type PluginStore interface {
    // Get retrieves a value by key.
    Get(ctx context.Context, key string) ([]byte, error)
    // Set stores a value with optional TTL.
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    // Delete removes a value.
    Delete(ctx context.Context, key string) error
}
```

Keys are automatically prefixed: `plugin:<pluginName>:<key>`. Plugins cannot access other plugins' or core realm keys.

---

## Plugin Loading

### Discovery

```
plugins/
├── custom-commands.so
├── anti-flood.so
└── custom-commands/
    └── config.yml          ← optional per-plugin config
```

The loader scans `plugins/` for `.so` files. Each must export:

```go
// Required symbol in every plugin .so
var NewPlugin func() plugin.Plugin
```

### Dependency resolution

1. Call `NewPlugin()` on each `.so` → collect `Metadata()`.
2. Build a DAG from `DependsOn` fields.
3. Topological sort. Cycle → fatal error at startup.
4. `OnEnable()` in topological order. `OnDisable()` in reverse.

### Auto-deregistration

The `API` implementation wraps every `Registration` returned by `Events().On()`, `Packets().Before()`, etc. When `OnDisable()` completes (or the process shuts down), all tracked registrations are automatically unsubscribed — even if the plugin forgets to clean up.

```go
// Internal tracking wrapper (not exposed to plugins)
type trackedAPI struct {
    inner         API
    registrations []Registration
}

func (t *trackedAPI) Events() EventBus {
    return &trackedEventBus{inner: t.inner.Events(), tracker: t}
}

func (t *trackedAPI) cleanup() {
    for _, reg := range t.registrations {
        reg.Unsubscribe()
    }
}
```

### Error isolation

- If `OnEnable()` returns an error, the plugin is marked disabled and skipped. Other plugins continue loading.
- If a plugin event handler panics, the panic is recovered, logged, and the handler is deregistered. The room tick / packet flow continues.
- A plugin that panics 3 times within 60 seconds is auto-disabled for the rest of the process lifetime.

---

## Realm Extension Points

Each realm fires events at meaningful domain boundaries. This section maps every realm to its plugin hooks.

### Auth realm

| Extension Point | Mechanism | Where in Code |
|---|---|---|
| Ticket validated | EventBus: `auth.ticket.validated` | `internal/auth/adapters/transport/subscriber.go` after `ValidateTicket()` |
| Custom auth providers | PacketInterceptor: `Before` on SSO ticket header | Gateway ingress, before bus publish |
| HTTP endpoints | RouteRegistrar: `/api/v1/plugins/<name>/` | Plugin `OnEnable()` |

### Game realm (room workers)

The game realm is the richest extension surface. Events fire from within the room worker goroutine.

| Extension Point | Mechanism | Where in Code |
|---|---|---|
| Room enter/leave | EventBus: `room.user.join` (cancellable), `room.user.leave` | `RoomWorker.handleCommand()` |
| Chat messages | EventBus: `room.user.chat` (cancellable) | Chat handler before broadcast |
| Movement | EventBus: `room.user.move` (cancellable) | Before pathfinding |
| Furniture interaction | EventBus: `room.item.interact` (cancellable) | ItemInteractionSystem |
| Tick hooks | EventBus: `room.tick.pre`, `room.tick.post` | Inside `RoomWorker.tick()` |
| Custom chat commands | EventBus: `room.command.custom` | Chat handler `:` prefix detection |
| Room state reads | RoomService: `Snapshot()` | Any goroutine, any time |
| Room mutations | RoomService: `SendCommand()` | Enqueued to worker inbox |

#### Tick hook execution model

```go
// Inside RoomWorker.tick()
func (w *RoomWorker) tick() {
    w.eventBus.Emit(Event{Name: "room.tick.pre", RoomID: w.roomID, Tick: w.tickCount})

    MovementSystem(w.rw)
    ArrivalSystem(w.rw)
    RollerSystem(w.rw)
    ItemInteractionSystem(w.rw)
    PetAISystem(w.rw)
    BotAISystem(w.rw)
    ChatCooldownSystem(w.rw, w.tickCount)
    WiredSystem(w.rw)
    BroadcastSystem(w.rw, w.sessions)

    w.eventBus.Emit(Event{Name: "room.tick.post", RoomID: w.roomID, Tick: w.tickCount})
}
```

**Critical constraint:** Plugin tick handlers execute synchronously inside the 50ms tick budget. A handler that blocks >5ms should offload to a goroutine and use `SendCommand()` to enqueue mutations back.

### Gateway realm

| Extension Point | Mechanism | Where in Code |
|---|---|---|
| Session lifecycle | EventBus: `session.connected`, `session.disconnected` | `pkg/http/ws/gateway.go` HandleConnection |
| Packet intercept | PacketInterceptor: Before/After any header | `pkg/http/ws/ingress.go` handleBinary |

### Navigator, Social, Catalog, Moderation

These realms follow the same pattern: fire events at domain boundaries.

| Realm | Key Events | Cancellable |
|---|---|---|
| navigator | `navigator.search`, `navigator.room.created` | search: No, create: Yes |
| social | `social.friend.request`, `social.message.sent` | request: Yes, message: Yes |
| catalog | `catalog.purchase`, `catalog.page.opened` | purchase: Yes, page: No |
| moderation | `moderation.ban.issued`, `moderation.report.filed` | ban: No, report: No |

---

## Plugin ↔ Realm Boundary Rules

### What plugins CAN do

- Subscribe to events from any realm running on the same process.
- Cancel cancellable events (deny entry, mute chat, block purchase).
- Read room ECS state through snapshots.
- Enqueue room commands through `SendCommand()`.
- Broadcast packets to room sessions.
- Register HTTP endpoints under `/api/v1/plugins/<name>/`.
- Store persistent data in Redis through `PluginStore`.
- Log through their scoped logger.

### What plugins CANNOT do

- Access `*ecs.World` directly. No mapper, filter, or entity handle is exposed.
- Call internal realm application services directly.
- Publish to the transport bus (`pkg/core/transport.Bus`).
- Modify core route registrations or middleware.
- Access other plugins' storage keys.
- Block the tick loop for extended periods (>5ms triggers runtime warnings).

### What only internal realm code can do

- Own and mutate ECS worlds.
- Subscribe to transport bus topics.
- Register core HTTP routes under `/api/v1/<realm>/`.
- Access PostgreSQL pools directly.
- Define and evolve protocol packet definitions.
- Start background goroutines with lifecycle management.

---

## How Realm Implementation Should Be Guided by Plugin Extensibility

### Realm design checklist

Every realm implementation must follow this checklist to ensure plugin extensibility:

1. **Identify domain events.** Before writing business logic, list the observable state transitions (user joins, item purchased, ban issued). These become EventBus events.

2. **Fire events at domain boundaries.** Events fire in the application service layer — not in adapters, not in domain aggregates. The app service knows when a complete action has occurred.

3. **Make events cancellable where reversible.** If the action hasn't committed side effects yet (no DB write, no broadcast sent), the event should be cancellable.

4. **Accept `EventBus` in Register().** Each realm's `Register()` function receives the shared `EventBus` and passes it to application services that need to fire events.

5. **Expose ports, not implementations.** Interfaces that realm code exposes to the plugin API must be Go interfaces, not concrete structs. This allows the tracked wrapper to intercept calls.

6. **Don't assume plugin presence.** Zero plugins loaded must be the default. Event emission with no listeners is a no-op.

### Updated realm Register() pattern

```go
// internal/<realm>/register.go

func Register(
    ctx context.Context,
    fiberApp *fiber.App,
    bus coretransport.Bus,
    eventBus plugin.EventBus,     // ← shared event bus for plugin hooks
    logger *zap.Logger,
    apiKey string,
) (*Runtime, error) {
    service := app.NewService(store, eventBus)  // inject event bus into app service
    if fiberApp != nil && apiKey != "" {
        httpadapter.RegisterRoutes(fiberApp, service, apiKey)
    }
    subscriber := transportadapter.NewSubscriber(bus, service, logger)
    if err := subscriber.Start(ctx); err != nil {
        return nil, err
    }
    return &Runtime{Service: service}, nil
}
```

### Application service event integration pattern

```go
// internal/<realm>/app/service.go

func (s *Service) HandleDomainAction(args ...) error {
    // 1. Validate domain rules
    result, err := s.domain.DoThing(args...)
    if err != nil {
        return err
    }

    // 2. Fire cancellable event BEFORE committing
    if s.events != nil {
        evt := plugin.Event{Name: "<realm>.<entity>.<action>", Data: result}
        s.events.Emit(evt)
        if evt.Cancelled() {
            return ErrCancelledByPlugin
        }
    }

    // 3. Commit side effects (DB write, broadcast, etc.)
    return s.store.Save(result)
}
```

---

## Caveats and Trade-offs

### Plugin scope vs distributed mode

In distributed mode, a plugin targeting `game` only runs on `--role=game` processes. If the plugin also needs to intercept packets at the gateway, it must declare `Realms: ["gateway", "game"]` — and the gateway process will load it too. This means the plugin's `OnEnable()` must handle being called with different available services depending on the active role.

The API handles this by returning nil for unavailable services:

```go
api.Rooms()   // → nil on gateway process (no game realm)
api.HTTP()    // → nil on game-only process (no HTTP listener)
```

### EventBus is not the transport Bus

This is the most critical architectural distinction:

| | `transport.Bus` | `plugin.EventBus` |
|---|---|---|
| **Scope** | Cross-process (local channels or NATS) | Same-process only |
| **Consumers** | Realm transport adapters | Plugin event handlers |
| **Message format** | `[]byte` payloads on stable topic strings | Typed `Event` structs |
| **Semantics** | At-least-once delivery | Synchronous observer |
| **Who uses it** | Internal realm code only | Plugins (and realm code for firing) |

A plugin never sees or touches the transport bus. The EventBus exists so that plugins can observe what realm code does without coupling to transport internals.

### Packet interception is gateway-only

`PacketInterceptor` hooks run inside the gateway's WebSocket ingress path. In distributed mode, this means the plugin must be loaded on the gateway process. The interceptor runs BEFORE the packet is published to the transport bus — so a cancelled packet never reaches the target realm.

For realm-side interception (e.g., "I want to modify how auth handles this specific packet"), the realm fires EventBus events that plugins subscribe to. This is a different code path than `PacketInterceptor` and happens after transport delivery.

### Hot reload is not supported (Phase 1)

Go's `plugin` package cannot unload shared objects. To update a plugin, the operator must restart the pixelsv process. Future WASM runtime would enable hot reload, but that is out of scope for initial implementation.

---

## Implementation Phases

### Phase 1: Core plugin infrastructure

- `pkg/plugin/` — API interface, Metadata, Plugin interface, Event types, Registration handle.
- `pkg/plugin/eventbus/` — In-process event bus with listener tracking and panic recovery.
- `pkg/plugin/interceptor/` — Packet interceptor with before/after chains.
- `pkg/plugin/loader/` — `.so` discovery, dependency sort, lifecycle management.

### Phase 2: Realm event integration

- Add `EventBus` parameter to each realm `Register()` function.
- Add event emissions at domain boundaries in auth realm (reference implementation).
- Wire event bus creation in `startup.go` before realm registration.

### Phase 3: Game realm facade

- `RoomService` implementation backed by room worker channels.
- `RoomSnapshot` generation using ark-serde.
- Tick event hooks (pre/post) in room worker loop.
- Cancellable events for join/chat/move/interact.

### Phase 4: HTTP and storage extensions

- `RouteRegistrar` backed by Fiber's `Group()` with `/api/v1/plugins/<name>/` prefix.
- `PluginStore` backed by Redis with `plugin:<name>:` key prefix.
- Plugin config file loading from `plugins/<name>/config.yml`.

### Phase 5: Developer experience

- Example plugin repository with build instructions.
- Plugin development guide in `docs/`.
- `pixelsv plugin validate <path>` CLI command for build compatibility checking.

---

## Operational Constraints

### Go plugin runtime limitations

- **Linux and macOS only.** Windows does not support Go's `plugin` package.
- **Same Go toolchain.** Plugin `.so` and host binary must be built with the same Go version and module dependencies. Version mismatch → load failure.
- **No unloading.** Once loaded, a `.so` stays in memory. `OnDisable()` deregisters hooks but does not free the binary.
- **CGO required.** `CGO_ENABLED=1` must be set for both host and plugin builds.

### Future runtime alternatives

| Runtime | Pros | Cons | Status |
|---|---|---|---|
| Go `.so` plugins | Native speed, full Go API | Platform/version coupling, no unload | Phase 1 target |
| WASM (wazero) | Sandboxed, cross-platform, hot reload | Serialization overhead, limited Go stdlib | Future exploration |
| gRPC sidecar | Language-agnostic, strong isolation | Latency, operational complexity | Not planned |

### Performance budget

- Plugin event handlers in tick loop: **<5ms combined** per tick (50ms budget, 10% reserved).
- Snapshot generation: **<1ms** per room at 200 entities.
- Packet interceptors: **<100µs** per packet per hook.
- Runtime warnings logged if budgets exceeded. Auto-disable after repeated violations.

---

## Quality Gates

Any change to plugin contracts must include:

- Unit tests for event bus, interceptor, and loader behavior.
- Integration test: load a test `.so` plugin, verify event subscription, verify auto-cleanup.
- Benchmark test for event dispatch and snapshot generation.
- Documentation updates in this file.
- Backward compatibility analysis (additions OK, removals require major version bump of plugin API).
