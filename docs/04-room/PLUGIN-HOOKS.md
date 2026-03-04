# Room — Plugin Hooks

Extension points planned and available in the room realm.

---

## Current Status

The room realm exposes **packet interceptors** for any room-related C2S packets
that pass through the game service. Dedicated room lifecycle events on the
`event.Bus` are planned but not yet emitted — the placeholders below describe
the intended API shape.

---

## Packet Interceptor Hooks (Available)

Any room-related C2S packet can be intercepted via
`intercept.Interceptor.Before` / `intercept.Interceptor.After`.

```go
interceptor.Before(headerID, func(ctx *intercept.PacketContext) {
    // Inspect or mutate ctx.Payload.
    // ctx.Cancel() suppresses the handler.
})
```

| `ctx` Field | Type | Description |
|---|---|---|
| `SessionID` | `string` | Session of the sending player |
| `HeaderID` | `uint16` | Incoming packet header ID |
| `Payload` | `[]byte` | Raw packet body (may be mutated) |
| `Direction` | `string` | Always `"C2S"` |
| `Cancel()` | `func()` | Prevents handler execution |

---

## Planned Event Bus Events

The following events will be emitted on `event.Bus` once room goroutines are
fully wired. Plugins subscribe with:

```go
bus.Subscribe(event.RoomLoaded, func(payload any) {
    data := payload.(event.RoomLoadedPayload)
    _ = data.RoomID
})
```

### `event.RoomLoaded`

Fired when a room goroutine boots up and is ready to accept players.

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | ID of the room that was loaded |

### `event.RoomUnloaded`

Fired when the last player leaves and the room goroutine is shut down.

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | ID of the room that was unloaded |

### `event.RoomTick`

Fired once per tick (20 Hz) after all systems have run.

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | Room ID |
| `Tick` | `uint64` | Monotonic tick counter |

### `event.EntitySpawned`

Fired after `SpawnAvatar`, `SpawnBot`, `SpawnPet`, or `SpawnItem` completes.

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | Room the entity was spawned in |
| `Kind` | `uint8` | Kind constant (`KindAvatar`, `KindBot`, etc.) |
| `EntityID` | `int64` | `AvatarID.UserID`, bot/pet DB ID, or furni ID depending on kind |

### `event.EntityRemoved`

Fired when `RemoveEntity` is called for any entity.

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | Room the entity was removed from |
| `Kind` | `uint8` | Kind constant |
| `EntityID` | `int64` | Same semantics as `EntitySpawned.EntityID` |

### `event.PlayerWalk`

Fired when a player's `WalkPath` is assigned (i.e. walk request accepted).

| Payload field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | Room ID |
| `UserID` | `int64` | Walking player's user ID |
| `Steps` | `int` | Number of path steps in the assigned walk |

---

## Plugin Constraints

- **No blocking in event handlers.** Handlers run synchronously on the room
  goroutine during the tick. Blocking even briefly delays the entire simulation.
- **No I/O in event handlers.** No NATS publishes, no database reads inside an
  event callback. Schedule work on a separate goroutine if needed.
- **Do not mutate `RoomWorld` from a plugin.** Plugins do not receive a
  `*RoomWorld` reference. Mutation must be requested via the room's input
  channel.

---

## Realm Relations

| Realm | Dependency |
|---|---|
| **SESSION** | Room enter/leave maps to player joined/left session events |
| **PATHFINDING** | `FindPath` is called before assigning `WalkPath` |
| **USER-PROFILE** | `AvatarID.UserID` references the user domain |

---

## Permissions and Guards

| Guard | Enforced by |
|---|---|
| Player must be authenticated | Room enter handler rejects unauthenticated sessions |
| Occupancy limit | Enter handler checks `Room.UsersNow < Room.UsersMax` |
| Password check | Handler compares provided password when `Room.State == "password"` |
| Club-only model | Enter handler checks `Model.ClubOnly` against user's club level |
