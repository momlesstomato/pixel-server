# Event System

## Overview

The event system is the primary communication channel between the server core
and plugins. Events are type-routed, priority-ordered, and optionally
cancellable. The dispatcher uses reflection to match handler function
signatures to concrete event types.

## Event Traits

| Type | Trait | Description |
|------|-------|-------------|
| `Event` | interface marker | Base contract; all events implement `event()` |
| `Cancellable` | interface | Adds `Cancelled() bool` and `Cancel()` |
| `BaseEvent` | embed struct | Non-cancellable base |
| `BaseCancellable` | embed struct | Cancellable base with atomic cancel flag |

## Priority System

Handlers execute in priority order within each event type bucket.

| Constant | Value | Use Case |
|----------|-------|----------|
| `PriorityLowest` | 0 | Logging, metrics |
| `PriorityLow` | 25 | Default background work |
| `PriorityNormal` | 50 | Standard handlers (default) |
| `PriorityHigh` | 75 | Validation, pre-processing |
| `PriorityHighest` | 100 | Security checks, access control |
| `PriorityMonitor` | 127 | Observability; always runs even if cancelled |

## Subscription

```go
server.Events().Subscribe(func(e *sdk.AuthCompleted) {
    server.Logger().Printf("user %d logged in", e.UserID)
})
```

Handlers are functions with exactly one pointer-to-struct parameter and no
return values. The dispatcher resolves the event type via reflection.

### Options

```go
server.Events().Subscribe(handler, sdk.WithPriority(sdk.PriorityHigh))
server.Events().Subscribe(handler, sdk.SkipCancelled())
```

| Option | Description |
|--------|-------------|
| `WithPriority(p)` | Set execution priority |
| `SkipCancelled()` | Skip this handler if the event is already cancelled |

`Subscribe` returns a cancel function that unregisters the handler.

## Cancellation

Cancellable events can be stopped mid-dispatch:

```go
server.Events().Subscribe(func(e *sdk.PacketReceived) {
    if e.PacketID == 9999 {
        e.Cancel()
    }
})
```

When an event is cancelled:
- Handlers with `SkipCancelled()` are skipped.
- `PriorityMonitor` handlers always run regardless of cancellation.
- The event source checks `Cancelled()` after dispatch to decide whether to
  proceed with the operation.

## Built-in Events

### Connection Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `ConnectionOpened` | No | ConnID |
| `ConnectionClosed` | No | ConnID, Reason |

### Authentication Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `AuthValidating` | **Yes** | ConnID, UserID, Ticket |
| `AuthCompleted` | No | ConnID, UserID |
| `DuplicateKick` | **Yes** | OldConnID, NewConnID, UserID |

### Session Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `SessionDisconnecting` | **Yes** | ConnID, UserID, Reason |
| `PongTimeout` | No | ConnID, UserID |

### Protocol Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `PacketReceived` | **Yes** | ConnID, PacketID, Body |
| `PacketSending` | **Yes** | ConnID, PacketID, Body |

### System Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `HotelStatusChanged` | No | OldState, NewState |

## Domain Events

Domain-scoped events live in `sdk/events/` subpackages.

### User Events (`sdk/events/user/`)

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `FigureChanged` | **Yes** | ConnID, UserID, OldFigure, NewFigure, Gender |
| `MottoChanged` | **Yes** | ConnID, UserID, OldMotto, NewMotto |
| `NameChanged` | **Yes** | ConnID, UserID, OldName, NewName |
| `Ignored` | **Yes** | ConnID, UserID, IgnoredUserID |
| `Unignored` | **Yes** | ConnID, UserID, IgnoredUserID |
| `Respected` | **Yes** | ActorConnID, ActorUserID, TargetUserID |

### Permission Events (`sdk/events/permission/`)

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `PermissionChecked` | No | UserID, Permission, Granted |
| `UserGroupChanged` | **Yes** | UserID, OldGroupID, NewGroupID, OldGroupIDs, NewGroupIDs |

## Dispatcher Internals

The `Dispatcher` is shared across all plugins. Thread safety is provided by
`sync.RWMutex` — reads during `Fire`, writes during `Subscribe` and
`RemoveByOwner`.

| Operation | Behavior |
|-----------|----------|
| `Subscribe` | Assigns a monotonic handler ID, sorts handlers by priority within the event type bucket |
| `Fire` | Looks up handlers by `reflect.TypeOf(event)`, iterates in priority order, wraps each call in `recover()` |
| `RemoveByOwner` | Bulk-removes all handlers for a plugin name (used during shutdown) |

Handler resolution accepts `func(*ConcreteEventType)` — the handler must be a
function with exactly one pointer-to-struct parameter.
