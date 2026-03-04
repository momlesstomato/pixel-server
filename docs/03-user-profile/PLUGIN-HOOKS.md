# User Profile — Plugin Hooks

Extension points available to plugins within the user-profile realm.

---

## Overview

The user-profile realm does not currently emit standalone event bus events.
Extension is achieved through the **packet interceptor** chain, which runs
before and after every C2S packet handler in the game service.

All interceptors are registered via `intercept.Interceptor.Before` and
`intercept.Interceptor.After`. They execute synchronously on the room/game
goroutine. **Never block or perform I/O inside an interceptor.**

---

## Packet Interceptor Hooks

### `intercept.Interceptor.Before(headerID uint16, fn)`

Called before the dispatch handler runs for any matching packet.

```go
i.Before(4000, func(ctx *intercept.PacketContext) {
    // Inspect or mutate ctx.Payload before the handler reads it.
    // Call ctx.Cancel() to suppress the handler entirely.
})
```

| Field | Type | Description |
|---|---|---|
| `ctx.SessionID` | `string` | Session identifier of the sending client |
| `ctx.HeaderID` | `uint16` | Incoming packet header ID |
| `ctx.Payload` | `[]byte` | Raw packet body (may be mutated) |
| `ctx.Direction` | `string` | Always `"C2S"` for incoming packets |
| `ctx.Cancel()` | `func()` | Prevents the handler from executing |

### `intercept.Interceptor.After(headerID uint16, fn)`

Called after the dispatch handler runs for any matching packet.

```go
i.After(4000, func(ctx *intercept.PacketContext) {
    // Inspect ctx.Payload after handler has run.
    // ctx.Cancel() has no effect here.
})
```

---

## Interceptable Profile Packets

Any of the following C2S header IDs can be intercepted. Refer to the
[PACKETS.MD](PACKETS.MD) page for field layouts.

| Packet name | Header ID |
|---|---|
| `get_info` | 2755 |
| `change_figure` | 2508 |
| `change_motto` | 2521 |
| `update_settings` | 2609 |
| `get_wardrobe` | 2690 |
| `save_wardrobe_outfit` | 2834 |
| `get_badges` | 2753 |
| `change_username` | 2977 |
| `change_email` | 2622 |
| `get_relationship_status_info` | 2717 |
| `load_ignored_users` | 2512 |
| `ignore_user` | 2517 |
| `unignore_user` | 2518 |
| `get_identity_settings` | 2840 |
| `get_profile` | 2751 |
| `get_stats` | 2754 |
| `get_home_room` | 2641 |

---

## Realm Relations

| Realm | How it relates |
|---|---|
| **SESSION** | Supplies `SessionID` and the authenticated `UserID`; profile handlers fail fast if the session is not authenticated |
| **HANDSHAKE** | Completes before any profile packet is valid; auth gate is enforced by the session layer |
| **ROOM** | `home_room` field on the User model is consumed by the room join flow |

---

## Permissions and Guards

| Guard | Enforced by |
|---|---|
| Session must be authenticated | `identitySvc.GetSession` returns `ErrNotFound` for unauthenticated sessions; handler returns early |
| Username change requires `AllowNameChange == true` | `change_username` (2977) handler checks the User field before proceeding |
| Users cannot ignore themselves | `ignore_user` (2517) compares `targetID == session.UserID` and rejects the request |
| Settings update is own-user only | Handler derives `userID` from session — client cannot supply a foreign ID |

---

## Configuration

The following env vars from the game service affect profile handling:

| Variable | Config field | Default | Effect |
|---|---|---|---|
| `GAME_HOST` | `game.host` | `"0.0.0.0"` | WebSocket listen address |
| `GAME_PORT` | `game.port` | `"3002"` | WebSocket listen port |
| `NATS_URL` | `nats_url` | (required) | Bus used to publish profile events |

Profile handlers use the NATS bus to publish the authentication envelope and
listen for login bundle responses. No profile-specific knobs presently exist.
