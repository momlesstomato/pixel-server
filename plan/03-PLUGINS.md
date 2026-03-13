# 03 - Plugin System

## Overview

The plugin system provides an extensible API for third-party developers to
modify server behavior via Go shared object (`.so`) files. Plugins subscribe
to typed events with priority ordering and cancellation, inject packets into
connections, register custom packet handlers, and query session state - all
through a narrow, versioned SDK module.

The SDK lives in a separate Go module within the same workspace (`go.work`),
ensuring plugins depend only on stable interface types with zero transitive
dependencies.

---

## Vendor Cross-Reference

### Gladiator (Java) - Reference Architecture

Gladiator has the most complete plugin system among all vendors:

- **Loading**: Scans `plugins/` for `.jar` files, isolated `URLClassLoader` per JAR
- **Lifecycle**: `onEnable()` / `onDisable()` abstract methods
- **Events**: 143 concrete event classes, `@EventHandler` annotation with 6 priority
  levels (LOWEST..MONITOR), cancellation via `Event.setCancelled()`
- **Dispatch**: Synchronous reflection-based; built-in handlers first, then plugins
- **Registration**: `pluginManager.registerEvents(plugin, listener)` scans methods
- **Isolation**: ClassLoader per plugin (separate classpath, no memory isolation)
- **God objects**: `Emulator.getGameEnvironment().getRoomManager()` chains expose
  everything; no API boundary between core and plugins

### Sodium (C#) - DI-Based

- **Loading**: Compile-time DI registration, no runtime loading
- **Lifecycle**: `IPlugin.Start()` only (no disable)
- **Events**: Standard C# delegates, no priority, no cancellation
- **Isolation**: `AssemblyLoadContext`, same process

### Galaxy (Java) - None

No plugin architecture.

---

## Design Decisions

### Go `.so` Plugins

Go's `plugin` package loads `.so` files compiled with `go build -buildmode=plugin`.

**Constraints:**
- Same Go toolchain version as host binary (exact match required)
- Same dependency versions for shared packages
- Linux and macOS only (no Windows)
- No unloading (stays in memory until process exit)
- No memory isolation (same process space)
- CGO required for plugin loading

These constraints are acceptable because plugin authors target a specific server
version, and the SDK module minimizes shared dependency surface.

### Go Workspace for SDK Separation

The plugin SDK must be a separate Go module so that:
1. Plugin authors import only the SDK, not the full server
2. The SDK has **zero transitive dependencies** (no zap, redis, fiber)
3. Version coupling is minimized to SDK interface compatibility
4. The server implements SDK interfaces internally

**Workspace layout:**

```
pixel-server/
  go.work                 <- workspace: use . and ./sdk
  go.mod                  <- server module
  sdk/
    go.mod                <- SDK module (github.com/momlesstomato/pixel-sdk)
    plugin.go             <- Plugin, Manifest, Server interfaces
    event.go              <- Event, Cancellable, base types, concrete events
    bus.go                <- Priority, HandlerOption, EventBus interface
    api.go                <- SessionAPI, PacketAPI, Logger, SessionInfo
    codec.go              <- Reader/Writer for packet body encoding
```

The SDK module exports **only interfaces and value types**. The server's
`core/plugin/` package implements these interfaces and wires them to real
infrastructure.

### Go-Idiomatic API Design

The API follows Go conventions rather than mimicking Java/Minecraft patterns:

1. **Functional options** over annotation-driven registration
2. **Interface composition** over inheritance hierarchies
3. **Concrete event types** via struct embedding, not string-keyed maps
4. **`context.Context`** propagation for cancellation and deadlines
5. **Error returns** over panic/exception patterns
6. **Reflect for type dispatch** — `reflect.TypeOf` for map key lookup only;
   handler invocation uses direct function calls

---

## Architecture

### Package Layout

```
sdk/                          <- separate Go module (zero dependencies)
  go.mod                      <- github.com/momlesstomato/pixel-sdk, go 1.25.5
  plugin.go                   <- Plugin interface + Manifest + Server interface
  event.go                    <- Event, Cancellable, BaseEvent, 11 concrete types
  bus.go                      <- Priority constants + HandlerOption + EventBus
  api.go                      <- SessionAPI + PacketAPI + Logger + SessionInfo
  codec.go                    <- Reader/Writer for Habbo protocol encoding

core/plugin/                  <- server-side implementation
  dispatcher.go               <- Typed event dispatch (priority + cancellation)
  manager.go                  <- Load, Enable, Disable, Shutdown lifecycle
  loader.go                   <- .so file discovery + symbol lookup
  api_impl.go                 <- Server/Session/Packet/Logger/EventBus wiring
  stage.go                    <- Initializer stage for serve.go integration

core/plugin/tests/
  dispatcher_test.go          <- Priority ordering, cancellation, recovery, unsubscribe
```

### SDK Interfaces

```go
package sdk

// Plugin defines the contract for a loadable server extension.
type Plugin interface {
    // Manifest returns plugin metadata.
    Manifest() Manifest
    // Enable is called when the plugin is activated.
    // The Server value provides access to event registration and server APIs.
    Enable(srv Server) error
    // Disable is called when the server is shutting down.
    Disable() error
}

// Manifest describes plugin identity.
type Manifest struct {
    Name    string
    Author  string
    Version string
}

// Server is the entry point for plugin interaction with the server.
// Each method returns a focused API surface; no god object.
type Server interface {
    // Logger returns a logger scoped to the calling plugin.
    Logger() Logger
    // Events returns the event subscription API.
    Events() EventBus
    // Sessions returns the session query and control API.
    Sessions() SessionAPI
    // Packets returns the packet send and handler registration API.
    Packets() PacketAPI
}
```

### Event System

```go
package sdk

// Event is the base contract for all dispatchable events.
type Event interface {
    event()  // unexported marker method; only SDK-defined types satisfy this
}

// Cancellable extends Event with cancellation support.
type Cancellable interface {
    Event
    Cancelled() bool
    Cancel()
}

// BaseEvent is embedded by all concrete event types.
type BaseEvent struct{}
func (BaseEvent) event() {}

// BaseCancellable is embedded by cancellable event types.
type BaseCancellable struct {
    cancelled bool
}
func (BaseCancellable) event() {}
func (e *BaseCancellable) Cancelled() bool { return e.cancelled }
func (e *BaseCancellable) Cancel()         { e.cancelled = true }
```

**Concrete event types** (11 defined in `sdk/event.go`):

| Event | Embedding | Cancellable | Fields |
|---|---|---|---|
| `ConnectionOpened` | `BaseEvent` | No | ConnID |
| `ConnectionClosed` | `BaseEvent` | No | ConnID, Reason |
| `AuthValidating` | `BaseCancellable` | Yes | ConnID, UserID, Ticket |
| `AuthCompleted` | `BaseEvent` | No | ConnID, UserID |
| `DuplicateKick` | `BaseCancellable` | Yes | OldConnID, NewConnID, UserID |
| `SessionDisconnecting` | `BaseCancellable` | Yes | ConnID, UserID, Reason |
| `PongTimeout` | `BaseEvent` | No | ConnID, UserID |
| `DesktopView` | `BaseCancellable` | Yes | ConnID, UserID |
| `HotelStatusChanged` | `BaseEvent` | No | OldState, NewState |
| `PacketReceived` | `BaseCancellable` | Yes | ConnID, PacketID, Body |
| `PacketSending` | `BaseCancellable` | Yes | ConnID, PacketID, Body |

### Event Bus

```go
package sdk

// Priority controls handler execution order.
// Lower values execute first. Built-in server handlers run at PriorityNormal.
type Priority int

const (
    PriorityLowest  Priority = 0
    PriorityLow     Priority = 25
    PriorityNormal  Priority = 50
    PriorityHigh    Priority = 75
    PriorityHighest Priority = 100
    PriorityMonitor Priority = 127 // always executes, even if cancelled
)

// HandlerConfig holds resolved handler options. Exported fields.
type HandlerConfig struct {
    Priority      Priority
    SkipCancelled bool
}

// HandlerOption configures event handler behavior.
type HandlerOption func(*HandlerConfig)

// WithPriority sets the handler execution priority.
func WithPriority(p Priority) HandlerOption {
    return func(c *HandlerConfig) { c.Priority = p }
}

// SkipCancelled causes the handler to be skipped if the event is already
// cancelled by a higher-priority handler.
func SkipCancelled() HandlerOption {
    return func(c *HandlerConfig) { c.SkipCancelled = true }
}

// EventBus allows subscribing to typed events.
type EventBus interface {
    // Subscribe registers a handler for events of type T.
    // Returns an unsubscribe function.
    Subscribe(handler any, opts ...HandlerOption) (unsubscribe func())
}
```

**Handler type**: Handlers are `func(T)` where `T` is a concrete event type.
The dispatcher uses `reflect.TypeOf` for map key lookup and calls the handler
function via `reflect.Value.Call`. The `any` parameter is validated at
registration time and panics with a clear message if the signature is wrong.

**Usage from a plugin:**

```go
func (p *MyPlugin) Enable(srv sdk.Server) error {
    srv.Events().Subscribe(func(e *sdk.AuthCompleted) {
        srv.Logger().Printf("user %d authenticated on %s", e.UserID, e.ConnID)
    })

    srv.Events().Subscribe(func(e *sdk.PacketReceived) {
        if e.PacketID == 105 {
            e.Cancel() // block desktop_view for this user
        }
    }, sdk.WithPriority(sdk.PriorityHigh), sdk.SkipCancelled())

    return nil
}
```

### Session API

```go
package sdk

// SessionInfo provides read-only session data.
type SessionInfo struct {
    ConnID     string
    UserID     int
    MachineID  string
    InstanceID string
}

// SessionAPI provides session query and control.
type SessionAPI interface {
    // FindByUserID returns session info for an online user.
    FindByUserID(userID int) (SessionInfo, bool)
    // FindByConnID returns session info for a connection.
    FindByConnID(connID string) (SessionInfo, bool)
    // Kick disconnects a session with a reason code.
    Kick(connID string, reason int32) error
    // Count returns the number of authenticated sessions.
    Count() int
}
```

### Packet API

```go
package sdk

// PacketAPI provides packet injection and custom handler registration.
type PacketAPI interface {
    // Send writes an encoded packet to a specific connection.
    Send(connID string, packetID uint16, body []byte) error
    // Broadcast sends a packet to all authenticated sessions.
    Broadcast(packetID uint16, body []byte) error
    // Handle registers a handler for a custom inbound packet ID.
    // If a handler already exists for the ID, it returns an error.
    Handle(packetID uint16, handler PacketHandler) error
}

// PacketHandler processes an inbound packet from a connection.
type PacketHandler func(connID string, body []byte) error
```

### Logger Interface

```go
package sdk

// Logger provides structured logging for plugins.
// Intentionally simple to avoid coupling to zap or any specific library.
type Logger interface {
    Printf(format string, args ...any)
    Errorf(format string, args ...any)
}
```

The server implementation wraps `*zap.SugaredLogger` behind this interface.

### Codec Utilities

The SDK includes minimal codec helpers so plugins can encode/decode packet
bodies without importing the server's `core/codec` package:

```go
package sdk

// Reader reads Habbo protocol primitive types from a byte slice.
type Reader struct { /* ... */ }
func NewReader(data []byte) *Reader
func (r *Reader) ReadInt32() (int32, error)
func (r *Reader) ReadString() (string, error)
func (r *Reader) ReadBool() (bool, error)
func (r *Reader) Remaining() int

// Writer builds Habbo protocol primitive types into a byte slice.
type Writer struct { /* ... */ }
func NewWriter() *Writer
func (w *Writer) WriteInt32(v int32)
func (w *Writer) WriteString(v string)
func (w *Writer) WriteBool(v bool)
func (w *Writer) Bytes() []byte
```

---

## Event Dispatch Engine

### Dispatch Flow

```
Core code calls: dispatcher.Fire(&sdk.AuthValidating{...})

Dispatcher:
  1. Look up handlers registered for *sdk.AuthValidating
  2. Sort by priority (stable sort at registration, not per-fire)
  3. For each handler:
     a. If event implements Cancellable AND handler has SkipCancelled AND event.Cancelled():
        skip (unless PriorityMonitor)
     b. If priority == PriorityMonitor:
        always execute (regardless of cancellation)
     c. Execute handler inside recover() wrapper
  4. Return to caller (checks Cancelled() if applicable)
```

### Registration Internals

When `Subscribe(func(e *sdk.AuthValidating) { ... })` is called:

1. Reflect on the function signature to extract the event type
2. Validate it is `func(T)` with exactly one parameter and zero returns
3. Assign a unique monotonic `uint64` ID via `sync/atomic`
4. Store as `handlerEntry{id, eventType, priority, skipCancelled, fn, owner}`
   in a sorted slice (stable sort by priority)
5. Return an unsubscribe closure that removes the entry by ID

**Type dispatch** at fire time uses a `reflect.TypeOf(event)` lookup into a
`map[reflect.Type][]handlerEntry`. This is O(1) lookup + O(n) iteration over
handlers for that type. Handler invocation uses `reflect.Value.Call`.

### Panic Recovery

Every handler invocation is wrapped:

```go
func (d *Dispatcher) invoke(e *handlerEntry, event sdk.Event) {
    defer func() {
        if r := recover(); r != nil {
            d.logger.Error("plugin handler panicked",
                zap.Any("panic", r),
                zap.String("event", e.eventType.String()),
                zap.String("owner", e.owner))
        }
    }()
    e.fn(event)
}
```

A panicking handler does not crash the server, does not affect other handlers
in the chain, and does not cancel the event.

---

## Plugin Loading

### Directory Structure

```
plugins/
  my-plugin.so
  another-plugin.so
```

### Plugin Author Template

```go
package main

import "github.com/momlesstomato/pixel-sdk"

type MyPlugin struct {
    logger sdk.Logger
}

func (p *MyPlugin) Manifest() sdk.Manifest {
    return sdk.Manifest{Name: "my-plugin", Author: "dev", Version: "1.0.0"}
}

func (p *MyPlugin) Enable(srv sdk.Server) error {
    p.logger = srv.Logger()
    srv.Events().Subscribe(func(e *sdk.AuthCompleted) {
        p.logger.Printf("user %d connected", e.UserID)
    })
    return nil
}

func (p *MyPlugin) Disable() error {
    p.logger.Printf("shutting down")
    return nil
}

// NewPlugin is the symbol looked up by the plugin loader.
var NewPlugin = func() sdk.Plugin { return &MyPlugin{} }
```

**Build:**

```bash
go build -buildmode=plugin -o plugins/my-plugin.so ./my-plugin
```

### Loading Process

1. Scan `plugins/` directory for `*.so` files (alphabetical order)
2. `plugin.Open(path)` loads the shared object
3. `plug.Lookup("NewPlugin")` finds the factory symbol
4. Type-assert to `*PluginFactory` or `*func() sdk.Plugin`
5. Call factory: `p := factory()`
6. Validate manifest: name non-empty, unique across loaded plugins
7. Call `p.Enable(serverAPI)` with the wired server implementation
8. Append to manager's plugin list
9. On any error: log and skip, do not crash

### Shutdown

1. Iterate plugins in **reverse** load order
2. Call `p.Disable()` inside `recover()` wrapper
3. Remove all event handlers registered by the plugin via `RemoveByOwner`
4. Log errors but continue to next plugin

---

## Serve Integration

Plugin loading is wired into `core/cli/serve.go` via `core/plugin.Stage`:

```go
pluginStage := coreplugin.Stage{
    Dir:    "plugins",
    Logger: runtime.Logger,
    Deps: coreplugin.ServerDependencies{
        Registry:         svc.registry,
        Broadcaster:      svc.broadcaster,
        BroadcastChannel: runtime.Config.Status.BroadcastChannel,
    },
}
pluginManager, err := pluginStage.Initialize()
if err != nil { return err }
defer pluginManager.Shutdown()
```

The stage creates a `Dispatcher`, a `Manager`, calls `LoadAll`, and returns the
manager. Each plugin receives a `serverImpl` that wraps:

- `pluginLogger` → `zap.SugaredLogger`
- `pluginEventBus` → `Dispatcher` with owner context
- `pluginSessionAPI` → `SessionRegistry` (FindByUserID, FindByConnID, Kick, Count via ListAll)
- `pluginPacketAPI` → `Broadcaster` (Send, Broadcast, Handle via sync.Map)

---

## Distribution Considerations

### Events Are Local

Event dispatch happens in-process only. When `dispatcher.Fire()` is called on
instance A, only handlers registered on instance A execute. This is correct
because:

1. Events relate to local connection state (a packet arrived on THIS connection)
2. Plugin handlers need low-latency access (synchronous dispatch)
3. Cross-instance coordination uses the broadcast bus, not events

### Cross-Instance Operations via API

When a plugin calls `srv.Packets().Send(connID, id, body)`:
- Publishes packet to `broadcast:conn:{connID}` via broadcaster

When a plugin calls `srv.Packets().Broadcast(id, body)`:
- Publishes to configured broadcast channel (default `broadcast:all`)
- All instances receive and forward to their connections

When a plugin calls `srv.Sessions().Kick(connID, reason)`:
- Removes the session from the registry directly

The plugin author does not need to know whether a connection is local or remote.
The `SessionAPI` and `PacketAPI` implementations handle routing transparently.

### Plugin Loading on Multiple Instances

Each instance loads plugins independently from its own `plugins/` directory.
Plugins should be deployed identically across instances. If instance A has
plugin X but instance B does not, events on instance B will not trigger
plugin X handlers - this is expected and consistent.

---

## Preventing God Objects

### Problem

Gladiator's plugins access everything via `Emulator.getGameEnvironment()` chains.
This creates tight coupling between plugins and server internals, making both
harder to evolve.

### Our Approach

1. **`Server` is the single entry point** with focused sub-APIs. Each method
   returns a narrow interface (`EventBus`, `SessionAPI`, `PacketAPI`, `Logger`).
   No sub-API exposes another sub-API.

2. **Value types in events** - event fields are primitive types and copies.
   `PacketReceived.Body` is a byte slice copy, not a reference to the
   connection's buffer.

3. **Packet injection over state mutation** - plugins send packets via
   `PacketAPI.Send()`. They do not mutate session state, room state, or user
   state directly.

4. **API grows per realm** - as new realms are implemented, the `Server`
   interface gains new methods. Each realm decides what to expose. Example:

   ```go
   // v1 (handshake + session)
   type Server interface {
       Logger() Logger
       Events() EventBus
       Sessions() SessionAPI
       Packets() PacketAPI
   }

   // v2 (+ rooms) - extends via interface embedding
   type ServerV2 interface {
       Server
       Rooms() RoomAPI
   }
   ```

   Plugins compiled against v1 SDK continue to work with v2 server because
   `ServerV2` embeds `Server`.

5. **SDK has zero dependencies** - the SDK module imports nothing. No `zap`,
   no `redis`, no `fiber`. This means plugins only couple to interface
   definitions, not implementation details.

---

## Go Workspace Setup

### Workspace File

```
// go.work
go 1.25.5

use (
    .
    ./sdk
)
```

### SDK Module

```
// sdk/go.mod
module github.com/momlesstomato/pixel-sdk

go 1.25.5
```

No `require` directives. Zero dependencies.

### Server Module Reference

The server's `go.mod` does not add an explicit `require` for the SDK. With
`go.work`, the local `./sdk` directory satisfies import resolution during
development. For releases, the SDK is tagged and published independently, and
the server's `go.mod` would then add a versioned require.

### Plugin Author's Module

```
// my-plugin/go.mod
module example.com/my-plugin

go 1.25.5

require github.com/momlesstomato/pixel-sdk v0.1.0
```

Plugin authors depend only on the SDK. They never import the server module.

---

## Edge Cases & Security

### 1. Plugin Compilation Mismatch

`plugin.Open()` fails with a descriptive error if Go version or dependency
versions differ. The manager logs the error and skips the plugin.

### 2. Handler Panic

All handler invocations are `recover()`-protected. A panicking handler:
- Does not crash the server
- Does not affect subsequent handlers in the chain
- Does not cancel the event
- Is logged with event type and owner name

### 3. Blocking Handlers

Synchronous dispatch means a blocking handler blocks the event chain. This is
documented as a plugin author responsibility. Guidelines:
- Handlers should complete in < 1ms
- For slow work, spawn a goroutine and return immediately
- Future: optional per-handler timeout via `HandlerOption`

### 4. Packet Handler Conflicts

`PacketAPI.Handle()` returns an error if a handler is already registered for
the given packet ID. This prevents plugins from silently overriding core
handlers. To intercept core packets, use `PacketReceived` events instead.

### 5. Event Loop Prevention

`PacketSending` events are NOT fired for packets injected via `PacketAPI.Send()`
or `PacketAPI.Broadcast()`. This prevents infinite loops where a send-handler
injects another packet that triggers another send-handler.

### 6. Hot Reload

Go `.so` plugins cannot be unloaded. Hot reload requires server restart. The
manifest version field helps operators verify loaded plugin versions.

### 7. Plugin Load Order

Plugins load in alphabetical filesystem order. If ordering matters, plugins
should use event priorities. There is no inter-plugin dependency system.

### 8. Subscription Leak

If a plugin subscribes to events in `Enable()` but does not unsubscribe in
`Disable()`, the handlers remain active. The manager clears all handlers
associated with a plugin via `RemoveByOwner` during shutdown, but plugins that
subscribe dynamically (outside `Enable()`) should store and call their
unsubscribe functions.

---

## Implementation Roadmap

### Milestone 1: SDK Module

| # | Task                                                | Depends On | Status |
|---|-----------------------------------------------------|------------|--------|
| 1 | Create `sdk/` directory with `go.mod`               | -          | DONE   |
| 2 | Create `go.work` workspace file                     | 1          | DONE   |
| 3 | Define `Plugin`, `Manifest`, `Server` interfaces    | 1          | DONE   |
| 4 | Define `Event`, `Cancellable`, `BaseEvent` types    | 1          | DONE   |
| 5 | Define `Priority` type and constants                | 1          | DONE   |
| 6 | Define `HandlerOption` functional options            | 1          | DONE   |
| 7 | Define `EventBus`, `SessionAPI`, `PacketAPI`, `Logger` | 1       | DONE   |
| 8 | Define all concrete event types (11 initial)        | 4          | DONE   |
| 9 | Implement codec `Reader`/`Writer` in SDK            | 1          | DONE   |
| 10| Verify SDK has zero `require` directives            | all        | DONE   |

### Milestone 2: Event Dispatcher

| # | Task                                                | Depends On | Status |
|---|-----------------------------------------------------|------------|--------|
| 11| Implement `Dispatcher` with typed handler registry  | 4, 5       | DONE   |
| 12| Priority-sorted insertion at registration time      | 11         | DONE   |
| 13| Cancellation propagation + SkipCancelled logic      | 11         | DONE   |
| 14| PriorityMonitor always-execute behavior             | 11         | DONE   |
| 15| Panic recovery per handler                          | 11         | DONE   |
| 16| Unit test: priority ordering                        | 12         | DONE   |
| 17| Unit test: cancellation chain                       | 13         | DONE   |
| 18| Unit test: panic does not affect chain               | 15         | DONE   |
| 19| Unit test: unsubscribe removes handler              | 11         | DONE   |
| 20| Unit test: RemoveByOwner clears plugin handlers     | 11         | DONE   |

### Milestone 3: Plugin Loader + Manager

| # | Task                                                | Depends On | Status |
|---|-----------------------------------------------------|------------|--------|
| 21| Implement `.so` file scanner and `plugin.Open`      | 3          | DONE   |
| 22| Implement `NewPlugin` symbol lookup + validation    | 21         | DONE   |
| 23| Implement `Enable` / `Disable` lifecycle            | 22         | DONE   |
| 24| Implement reverse-order shutdown with panic recovery | 23        | DONE   |
| 25| Initializer stage for plugin loading (`stage.go`)   | 23         | DONE   |

### Milestone 4: API Implementation + Serve Integration

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 26| Implement `SessionAPI` (delegates to session registry) | 7       | DONE    |
| 27| Implement `PacketAPI` (broadcaster send + handle)   | 7          | DONE    |
| 28| Implement `Logger` wrapper over zap.SugaredLogger   | 7          | DONE    |
| 29| Implement `EventBus` wrapper with owner tracking    | 7          | DONE    |
| 30| Wire plugin stage into `core/cli/serve.go`          | 25         | DONE    |
| 31| Pass `ServerDependencies` (Registry, Broadcaster)   | 30         | DONE    |

### Milestone 5: Core Event Integration

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 32| Fire `ConnectionOpened` in WebSocket handler         | 11         | DONE    |
| 33| Fire `ConnectionClosed` in dispose handler           | 11         | DONE    |
| 34| Fire `AuthValidating` in auth use case              | 11         | DONE    |
| 35| Fire `AuthCompleted` in auth use case               | 11         | DONE    |
| 36| Fire `DuplicateKick` in auth use case               | 11         | DONE    |
| 37| Fire `SessionDisconnecting` in disconnect use case  | 11         | DONE    |
| 38| Fire `PacketReceived` in packet read loop           | 11         | DONE    |
| 39| Fire `PacketSending` in transport send path         | 11         | DONE    |
| 40| Integration test: cancel auth via plugin            | 34         | DONE    |
| 41| Integration test: cancel disconnect via plugin      | 37         | DONE    |

### Milestone 6: E2E & Example

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 42| Example plugin: login logger                        | 35         | DONE    |
| 43| Example plugin: packet filter                       | 38         | DONE    |
| 44| E2E test: load .so, fire event, verify behavior     | 25, 32     | DONE    |

---

## Caveats & Technical Notes

### Go Plugin Binary Compatibility

The `.so` must be compiled with:
- Exact same Go toolchain version
- Exact same `GOOS`/`GOARCH`
- Exact same versions of any shared module (only the SDK in our case)

To help plugin authors, the server binary should expose its Go version and SDK
version via `pixelsv version --json` for build-time verification.

### SDK Versioning Strategy

The SDK uses semantic versioning. Breaking changes to event types or interfaces
require a major version bump. Adding new event types or new methods to `Server`
is backwards-compatible (minor bump) because:
- New event types do not affect existing subscriptions
- New `Server` methods use interface embedding (`ServerV2 embeds Server`)

### Reflect Usage Justification

`reflect.TypeOf()` is used once per `Fire()` call to look up handlers by event
type. This is O(1) map lookup and adds ~50ns per dispatch. Handler invocation
uses `reflect.Value.Call` which wraps the typed function. The `resolveHandler`
function in `dispatcher.go` extracts the event type from the function parameter
at registration time and stores a `func(sdk.Event)` closure.

### `Subscribe` Signature: `any`

The `Subscribe(handler any, ...)` signature accepts `any` because Go generics
cannot express "function whose single parameter implements Event" without
making `EventBus` generic (which would leak the type parameter to `Server`).
The implementation validates the function signature at registration time and
panics with a clear message if the signature is wrong. This is a one-time
development-time check, not a runtime risk.

### No Inter-Plugin Communication

Plugins cannot directly call each other. If plugin A wants to trigger behavior
in plugin B, it should fire a custom event type (defined in a shared module
between the two plugins). The server's event bus dispatches it like any other
event. This prevents tight coupling between plugins.

### Codec Duplication

The SDK's `Reader`/`Writer` duplicates logic from `core/codec/`. This is
intentional: the SDK must have zero dependencies on the server module. The
implementations are small (~50 lines each) and the wire format is stable
(Habbo protocol hasn't changed). If the codec evolves, both copies are
updated in the same workspace commit.
