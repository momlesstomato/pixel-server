# Session — Plugin Hooks

How the plugin system observes and intercepts session lifecycle events.

---

## Plugin events emitted

These events are published to the `event.Bus` injected into the game service's
`Listener`. They run synchronously on the NATS handler goroutine.

### player.joined — `event.PlayerJoined`

Emitted immediately after a new authenticated session is registered in the game
service's session map (before the login bundle is sent).

| `event.Event` field | Value |
|---|---|
| `Name` | `"player.joined"` |
| `EntityID` | `int64(userID)` |
| `Payload` | `string` — the session ID |
| `RoomID` | `0` (not in a room yet) |
| `Tick` | `0` |

---

### player.left — `event.PlayerLeft`

Emitted after the session has been removed from the game service's session map.

| `event.Event` field | Value |
|---|---|
| `Name` | `"player.left"` |
| `EntityID` | `0` (user ID not retained after session is deleted) |
| `Payload` | `string` — the session ID |
| `RoomID` | `0` |
| `Tick` | `0` |

---

### packet.in — `event.PacketIn`

Emitted for every inbound post-auth packet that passes the Before interceptor
chain without being cancelled.

| `event.Event` field | Value |
|---|---|
| `Name` | `"packet.in"` |
| `EntityID` | `0` |
| `Payload` | `*intercept.PacketContext` — see below |

`*intercept.PacketContext` fields:

| Field | Type | Description |
|---|---|---|
| `SessionID` | `string` | Session that sent the packet |
| `HeaderID` | `uint16` | Packet header ID |
| `Payload` | `[]byte` | Raw packet body (after header bytes). Hooks may replace this. |
| `Cancel` | `bool` | Set by a Before hook to drop the packet |
| `Direction` | `string` | Always `"c2s"` for inbound packets |

---

## Packet interceptors

Registered via `interceptor.Before(headerID, fn)` and `interceptor.After(headerID, fn)`.
Both run synchronously on the NATS handler goroutine.

### Before hooks

Called **before** the packet is dispatched to any handler.  
Setting `ctx.Cancel = true` drops the packet entirely — no handler runs, no
`packet.in` event is emitted.

### After hooks

Called **after** the handler returns (for both handled and unhandled packets).  
Useful for logging, metrics, or injecting side effects after dispatch.

---

## Constraints for plugin authors

- **No blocking I/O** inside event handlers or interceptor hooks. These run on
  the NATS handler goroutine. Blocking starves NATS processing.
- **No channel sends** that could block. Use buffered channels or goroutines if
  you need async work.
- **`packet.in` Payload is shared** — if a Before hook modifies `ctx.Payload`,
  the modified slice reaches the handler. Deep-copy if you need the original.

---

## Realm relations

| Realm | How this realm depends on it |
|---|---|
| Handshake | The `session.authenticated` NATS event produced by auth is what triggers `player.joined` in the game service |
| User-Profile | After `player.joined`, `identity.SendLoginBundle` runs; user-profile packets follow immediately |
| Room | Post-auth packets arrive via `room.input.*` subject; this is where room entry requests will be routed in Phase 3 |

---

## Permissions / Guards

- **Authentication gate** (gateway): Post-auth packets from an unauthenticated
  session are dropped before they reach NATS. Plugins never see them.
- **Session isolation**: `session.output.<sessionID>` is subscribed only by the
  gateway instance that owns the connection. Cross-session writes are impossible
  through normal API.
- **Disconnect idempotency**: Session close is protected by a `done` channel
  checked once. Double-close is a no-op.
