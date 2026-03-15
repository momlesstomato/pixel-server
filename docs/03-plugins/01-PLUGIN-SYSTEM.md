# Plugin System

## Overview

Pixel Server supports Go native plugins (`.so` shared objects) that extend
server behavior at runtime. Plugins receive a sandboxed SDK surface and
interact with the server exclusively through events, sessions, packets, and
permissions. Direct access to the database, filesystem, or HTTP layer is not
exposed.

## SDK Module

Plugin authors import the SDK as an independent Go module:

```
github.com/momlesstomato/pixel-sdk
```

The SDK has zero external dependencies. It defines the `Plugin` interface,
event types, bus options, and the `Server` facade.

## Plugin Contract

Every plugin must implement `sdk.Plugin`:

```go
type Plugin interface {
    Manifest() Manifest
    Enable(Server) error
    Disable() error
}
```

| Method | Purpose |
|--------|---------|
| `Manifest()` | Returns plugin identity (name, author, version) |
| `Enable(server)` | Called at load time; register event handlers here |
| `Disable()` | Called at shutdown; clean up resources here |

The shared object must export a factory variable:

```go
var NewPlugin = func() sdk.Plugin { return &myPlugin{} }
func main() {} // required for .so build
```

## Lifecycle

```
discover (.so scan, alphabetical)
    │
    ▼
load (plugin.Open → lookup NewPlugin symbol)
    │
    ▼
enable (factory() → validate manifest → call Enable(server))
    │
    ▼
[runtime — events flowing, handlers active]
    │
    ▼
disable (reverse order, panic-safe)
    │
    ▼
cleanup (RemoveByOwner removes all handlers)
```

| Phase | Behavior |
|-------|----------|
| **Discover** | Scans the plugin directory for `.so` files, sorted alphabetically |
| **Load** | Opens each `.so` and looks up the `NewPlugin` symbol |
| **Enable** | Instantiates the plugin, validates the manifest, creates a scoped `Server`, calls `Enable` |
| **Disable** | Iterates plugins in reverse order (LIFO); each `Disable` call is wrapped in `recover()` |
| **Cleanup** | `RemoveByOwner()` bulk-removes all event handlers registered by the plugin |

Duplicate plugin names are rejected at load time. Empty manifest names are
rejected. If the plugin directory does not exist, zero plugins are loaded
without error.

## Server API Surface

When `Enable(server)` is called, the plugin receives five sub-APIs:

```
sdk.Server
├── Logger()      → sdk.Logger        (Printf, Errorf)
├── Events()      → sdk.EventBus      (Subscribe)
├── Sessions()    → sdk.SessionAPI    (FindByUserID, FindByConnID, Kick, Count)
├── Packets()     → sdk.PacketAPI     (Send, Broadcast, Handle)
└── Permissions() → sdk.PermissionAPI (HasPermission, GetGroup)
```

### Logger

Scoped to `plugin.<name>` via Zap. Provides `Printf` and `Errorf`.

### Sessions

| Method | Description |
|--------|-------------|
| `FindByUserID(id)` | Returns `SessionInfo` for a connected user |
| `FindByConnID(id)` | Returns `SessionInfo` by connection ID |
| `Kick(connID, reason)` | Disconnects a session with a reason string |
| `Count()` | Returns total connected session count |

`SessionInfo` is a read-only snapshot: `ConnID`, `UserID`, `MachineID`,
`InstanceID`. Plugins cannot mutate internal session state.

### Packets

| Method | Description |
|--------|-------------|
| `Send(connID, packet)` | Send one packet to a specific connection |
| `Broadcast(packet)` | Send one packet to all connections |
| `Handle(packetID, handler)` | Register a handler for an incoming packet ID |

### Permissions

| Method | Description |
|--------|-------------|
| `HasPermission(userID, permission)` | Check if user holds a permission |
| `GetGroup(userID)` | Returns `GroupInfo` for the user's primary group |

When `EmitPermissionChecked` is enabled in server config, every
`HasPermission` call fires a `PermissionChecked` event visible to all plugins.

## Example: Login Logger

```go
type loginLogger struct{ server sdk.Server }

func (p *loginLogger) Manifest() sdk.Manifest {
    return sdk.Manifest{Name: "login-logger", Author: "pixel", Version: "1.0.0"}
}

func (p *loginLogger) Enable(server sdk.Server) error {
    p.server = server
    server.Events().Subscribe(func(e *sdk.AuthCompleted) {
        server.Logger().Printf("user %d authenticated on %s", e.UserID, e.ConnID)
    })
    return nil
}

func (p *loginLogger) Disable() error { return nil }
```

## Example: Packet Filter

```go
func (p *packetFilter) Enable(server sdk.Server) error {
    server.Events().Subscribe(func(e *sdk.PacketReceived) {
        if e.PacketID == 9999 {
            e.Cancel()
            server.Logger().Printf("blocked packet 9999 from %s", e.ConnID)
        }
    })
    return nil
}
```

## Security Model

| Concern | Mechanism |
|---------|-----------|
| Isolation | Each plugin gets its own scoped logger and owner-tagged handlers |
| Panic containment | Handler execution and `Disable` are wrapped in `recover()` |
| Ownership tracking | All subscriptions tagged with plugin name; bulk cleanup on shutdown |
| Duplicate prevention | Duplicate plugin names and duplicate packet handler IDs are rejected |
| Read-only sessions | Plugins receive `SessionInfo` snapshots, not mutable internals |
| No escape hatches | SDK exposes only Logger, Events, Sessions, Packets, Permissions |
| Broadcast scoping | Packet sending goes through the `Broadcaster` abstraction |
