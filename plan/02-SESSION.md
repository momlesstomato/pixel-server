# 02 - Session & Connection Realm

## Overview

The Session & Connection realm manages the post-authentication lifecycle of a
connected client. It owns hotel availability signaling, heartbeat keep-alive,
latency measurement, hotel maintenance broadcasts, generic error/alert delivery,
moderation notifications, and graceful/forced disconnection with reason codes.

This realm activates **after** `authentication.ok` (handshake realm) and persists
until the WebSocket closes. Many packets in this realm are server-initiated
notifications that require no client response.

**Distribution model**: This realm is inherently multi-instance. Hotel status,
broadcasts, moderation actions, and alert delivery must work when connections
are spread across N server instances behind a load balancer. Redis Pub/Sub is
the backbone for cross-instance coordination; the existing `CloseSignalBus`
pattern is generalized into a broader message bus.

---

## Current Status (2026-03-12)

Several packets already live in the handshake realm because they participate in
the authentication flow. These are **shared** with this realm:

| Packet | Location | Notes |
|--------|----------|-------|
| `client.ping` (3928) | `pkg/handshake/packet/session/` | DONE |
| `client.pong` (2596) | `pkg/handshake/packet/session/` | DONE |
| `client.disconnect` (2445) | `pkg/handshake/packet/session/` | DONE |
| `client.latency_test` (295) | `pkg/handshake/packet/telemetry/` | DONE |
| `client.latency_response` (10) | `pkg/handshake/packet/telemetry/` | DONE |
| `disconnect.reason` (4000) | `pkg/handshake/packet/authentication/` | DONE |
| Heartbeat use case | `pkg/handshake/application/sessionflow/` | DONE |
| Disconnect use case | `pkg/handshake/application/sessionflow/` | DONE |

The remaining packets in this plan are **new** and belong to a dedicated session
realm package.

---

## Vendor Cross-Reference

Analysis of pixels-emulator (Go), PlusEMU (C#), Arcturus-Community (Java),
comet-v2 (Java), and the pixel-protocol YAML spec.

### Post-Authentication Burst (All Vendors Agree)

After `authentication.ok`, the server sends a burst of session packets:

```
Server                                    Client
  |                                          |
  +--- authentication.ok (2491) ---------->  |  (from handshake realm)
  +--- availability.status (2033) -------->  |  Hotel open/closed state
  +--- client.ping (3928) --------------->  |  Start keepalive
  +--- first_login_of_day (793) --------->  |  Daily reward trigger (optional)
  +--- motd_messages (2035) ------------->  |  Message of the day (optional)
  +--- info_feed_enable (3284) ---------->  |  Social feed toggle (optional)
  |                                          |
  [session active - game packets flow]       |
  |                                          |
  +--- hotel.will_close (1050) ---------->  |  Maintenance warning (if scheduled)
  +--- hotel.maintenance (1350) --------->  |  Maintenance details
  +--- disconnect.reason (4000) --------->  |  Before forced close
```

---

## Packet Registry

### Client-to-Server (C2S)

| ID   | Name                                   | Fields                   | Status   |
|------|----------------------------------------|--------------------------|----------|
| 105  | `session.desktop_view`                 | _(empty)_                | PLANNED  |
| 295  | `client.latency_test`                  | `requestId: int32`       | DONE (handshake) |
| 1160 | `session.peer_users_classification`    | _(empty)_                | DEFERRED |
| 2313 | `session.client_toolbar_toggle`        | _(empty)_                | DEFERRED |
| 2445 | `client.disconnect`                    | _(empty)_                | DONE (handshake) |
| 2596 | `client.pong`                          | _(empty)_                | DONE (handshake) |
| 3226 | `session.render_room`                  | _(empty)_                | DEFERRED |
| 3230 | `session.tracking_performance_log`     | _(empty)_                | DEFERRED |
| 3457 | `session.event_tracker`                | _(empty)_                | DEFERRED |
| 3847 | `session.tracking_lag_warning_report`  | _(empty)_                | DEFERRED |

### Server-to-Client (S2C)

| ID   | Name                                   | Fields                                                                     | Status   |
|------|----------------------------------------|----------------------------------------------------------------------------|----------|
| 10   | `client.latency_response`              | `requestId: int32`                                                         | DONE (handshake) |
| 426  | `session.restore_client`               | _(empty)_                                                                  | DEFERRED |
| 600  | `availability.time`                    | `isOpen: int32`, `minutesUntilChange: int32`                               | DEFERRED |
| 793  | `session.first_login_of_day`           | `isFirstLogin: bool`                                                       | PLANNED  |
| 1004 | `connection.error`                     | `messageId: int32`, `errorCode: int32`, `timestamp: string`                | PLANNED  |
| 1050 | `hotel.will_close`                     | `minutes: int32`                                                           | PLANNED  |
| 1350 | `hotel.maintenance`                    | `isInMaintenance: bool`, `minutesUntilChange: int32`, `duration: int32`    | PLANNED  |
| 1600 | `session.generic_error`                | `errorCode: int32`                                                         | PLANNED  |
| 1663 | `session.hotel_merge_name_change`      | _(empty)_                                                                  | DEFERRED |
| 1890 | `session.moderation_caution`           | `message: string`, `detail: string`                                        | PLANNED  |
| 2033 | `availability.status`                  | `isOpen: bool`, `onShutdown: bool`, `isAuthentic: bool`                    | PLANNED  |
| 2035 | `session.motd_messages`                | `count: int32`, `messages: [string]`                                       | DEFERRED |
| 2771 | `hotel.closes_and_opens_at`            | `openHour: int32`, `openMinute: int32`, `userThrownOutAtClose: bool`       | PLANNED  |
| 3284 | `session.info_feed_enable`             | `enabled: bool`                                                            | DEFERRED |
| 3523 | `session.desktop_view`                 | _(empty)_                                                                  | PLANNED  |
| 3728 | `hotel.closed_and_opens`               | `openHour: int32`, `openMinute: int32`                                     | PLANNED  |
| 3801 | `session.generic_alert`                | `message: string`                                                          | PLANNED  |
| 3928 | `client.ping`                          | _(empty)_                                                                  | DONE (handshake) |
| 3945 | `session.epic_popup`                   | `assetUri: string`                                                         | DEFERRED |
| 4000 | `disconnect.reason`                    | `reason: int32`                                                            | DONE (handshake) |

### Status Legend

- **DONE (handshake)** - Already implemented in the handshake realm
- **PLANNED** - Will be implemented in this realm
- **DEFERRED** - Deferred intentionally (see reason below)

### Deferred Packets - Rationale

| Packet                               | Reason                                                                            |
|---------------------------------------|-----------------------------------------------------------------------------------|
| `session.peer_users_classification`   | Social classification system; requires friends/ignore list (social realm)          |
| `session.client_toolbar_toggle`       | UI state preference only; no server-side behavior in any vendor                   |
| `session.render_room`                 | PlusEMU explicitly `throw new NotImplementedException()`; not needed until rooms   |
| `session.tracking_performance_log`    | Client telemetry; no vendor implements server-side handling                        |
| `session.event_tracker`               | Client telemetry; PlusEMU returns `Task.CompletedTask` (no-op)                    |
| `session.tracking_lag_warning_report` | Client telemetry; no vendor implements server-side handling                        |
| `session.restore_client`              | Reconnection/resume system; requires session persistence beyond disconnect         |
| `availability.time`                   | Scheduled open/close times; requires hotel scheduling system                       |
| `session.hotel_merge_name_change`     | Business event notification; not applicable to private servers                     |
| `session.motd_messages`               | Requires admin MOTD management system                                             |
| `session.info_feed_enable`            | Social feed toggle; requires social realm                                         |
| `session.epic_popup`                  | Marketing/promotional popup; requires admin content management                    |

---

## Architecture

### Package Layout

```
pkg/session/
  packet/
    availability/         <- availability.status, availability.time
    hotel/                <- hotel.will_close, hotel.maintenance, hotel.closes_*, hotel.closed_*
    notification/         <- generic_error, generic_alert, moderation_caution
    navigation/           <- desktop_view (c2s + s2c), first_login_of_day
    error/                <- connection.error
  application/
    postauth/             <- Post-authentication burst orchestration
    hotelstatus/          <- Hotel open/close/maintenance lifecycle
    notification/         <- Generic error/alert dispatch service
  adapter/
    realtime/             <- Fiber WebSocket runtime integration

core/broadcast/
  bus.go                  <- Broadcast port interface
  redis_bus.go            <- Redis Pub/Sub broadcast adapter
  local_bus.go            <- In-process broadcast (single-instance / tests)
```

### Integration with Handshake Realm

The session realm **does not duplicate** heartbeat, disconnect, or latency logic.
It imports these from the handshake realm via their port interfaces:

```
pkg/handshake/application/sessionflow/  <- owns heartbeat, disconnect, latency
pkg/session/application/postauth/       <- owns post-auth burst (calls sessionflow)
pkg/session/application/hotelstatus/    <- owns hotel lifecycle (independent)
```

The post-auth burst is triggered by the handshake realm's `AuthenticateUseCase`
result. When authentication succeeds, the handshake adapter calls into the
session realm's post-auth orchestrator.

---

## Distribution Model

### Problem Statement

When multiple server instances share the same Redis and serve connections behind
a load balancer, several session operations must cross instance boundaries:

1. **Hotel broadcasts** (maintenance, close) must reach all connections on all
   instances
2. **Targeted sends** (moderation caution, generic alert) must reach a user
   whose connection may live on any instance
3. **Hotel status** must be consistent across instances (one source of truth)
4. **Audit records** (moderation actions, alerts) must persist regardless of
   which instance handled them

### Cross-Instance Message Bus

Generalize the existing `CloseSignalBus` (Redis Pub/Sub on
`handshake:close:{connID}`) into a broader broadcast bus:

```go
// Broadcaster defines cross-instance packet delivery.
type Broadcaster interface {
    // Publish sends a message to a named channel.
    Publish(ctx context.Context, channel string, payload []byte) error
    // Subscribe returns a channel that receives messages for a topic.
    Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error)
}
```

**Channel topology:**

| Channel Pattern              | Purpose                              | Publisher         | Subscribers        |
|------------------------------|--------------------------------------|-------------------|--------------------|
| `broadcast:all`              | Hotel-wide packets (maintenance, close) | Any instance   | All instances      |
| `broadcast:conn:{connID}`    | Targeted packet to specific connection | Any instance    | Owning instance    |
| `broadcast:user:{userID}`    | Targeted packet to specific user     | Any instance      | Owning instance    |
| `hotel:status`               | Hotel state change notification      | Status owner      | All instances      |

Each server instance subscribes to `broadcast:all` and `broadcast:conn:{connID}`
for every local connection at upgrade time. When a connection closes, its
per-connection subscription is disposed.

The existing `CloseSignalBus` becomes a specialized case of the broadcast bus
(`broadcast:conn:{connID}` with close-specific payload).

### Hotel Status Consistency

Hotel status (OPEN/CLOSING/CLOSED) is stored as a Redis key:

```
Key:   hotel:status
Value: JSON { "state": "open", "closeAt": null, "reopenAt": null }
TTL:   none (persistent until changed)
```

Any instance can read current status. Only the instance that initiates a state
transition publishes to `hotel:status` channel. Other instances update their
local cache from the notification.

### Session Lease / TTL

**Current gap**: Redis session keys have no TTL. If an instance crashes, its
sessions become orphans in Redis forever.

**Solution**: Add a TTL to session keys and require the owning instance to
refresh them periodically:

```
session:conn:{connID}  TTL = 120s (refreshed every 60s by heartbeat goroutine)
session:user:{userID}  TTL = 120s (refreshed alongside conn key)
```

If an instance crashes, orphan sessions expire after 120s. New login attempts
for the same user will find no existing session (or an expired one) and proceed
normally.

This also handles the edge case where a connection's server process is killed
without graceful shutdown - the session naturally expires.

### Instance-Aware Connection Tracking

Add an `InstanceID` field to the session record:

```go
type Session struct {
    ConnID     string
    UserID     int
    MachineID  string
    State      SessionState
    InstanceID string      // identifies owning server instance
    CreatedAt  time.Time
}
```

`InstanceID` is generated at startup (random UUID or hostname:port). This
allows any instance to determine whether a session belongs to it (local close)
or another instance (publish to bus).

---

## Disconnect Reason Codes (Complete Registry)

The `disconnect.reason` packet (4000) is already implemented in the handshake
realm. This section documents the **full reason code registry** for reference:

| Code | Name                    | Trigger                                      | Vendor Source      |
|------|-------------------------|----------------------------------------------|--------------------|
| 0    | LOGOUT                  | Normal user-initiated logout                 | All vendors        |
| 1    | JUST_BANNED             | User was just banned during session          | Arcturus, PlusEMU  |
| 2    | CONCURRENT_LOGIN        | Duplicate login detected, old session kicked | All vendors        |
| 3    | CONNECTION_LOST_TO_PEER | Connection lost to peer server               | Protocol spec      |
| 10   | STILL_BANNED            | User attempted login while banned            | Arcturus, PlusEMU  |
| 12   | HOTEL_CLOSED            | Hotel closed while user was connected        | Arcturus           |
| 13   | DUAL_LOGIN_BY_IP        | Multiple connections from same IP            | Protocol spec      |
| 17   | NO_LOGIN_PERMISSION     | User lacks login permission                  | Protocol spec      |
| 18   | DUPLICATE_CONNECTION    | Duplicate connection detected                | Protocol spec      |
| 22   | INVALID_LOGIN_TICKET    | SSO ticket invalid, expired, or missing      | All vendors        |
| 112  | IDLE_TIMEOUT            | No activity timeout                          | Protocol spec      |
| 113  | PONG_TIMEOUT            | Heartbeat pong not received in time          | Comet-v2           |
| 114  | IDLE_NOT_AUTHENTICATED  | Auth timeout (custom, our addition)          | pixelsv            |

Already defined constants in `pkg/handshake/packet/authentication/disconnect_reason.go`:
- `DisconnectReasonConcurrentLogin = 2`
- `DisconnectReasonInvalidLoginTicket = 22`
- `DisconnectReasonPongTimeout = 113`
- `DisconnectReasonIdleNotAuthenticated = 114`

New constants needed for session realm:
- `DisconnectReasonLogout = 0`
- `DisconnectReasonJustBanned = 1`
- `DisconnectReasonStillBanned = 10`
- `DisconnectReasonHotelClosed = 12`

---

## Availability & Hotel Status System

### availability.status (2033) - Post-Auth Burst

Sent immediately after `authentication.ok`. All vendors hardcode this:

```go
type AvailabilityStatusPacket struct {
    IsOpen      bool  // Hotel is open for play
    OnShutdown  bool  // Shutdown in progress
    IsAuthentic bool  // User is authenticated (not guest)
}
```

**Vendor consensus**: Always `{true, false, true}` in normal operation.

**Our approach**: Read from `hotel:status` Redis key. `OnShutdown` becomes `true`
when a maintenance window is scheduled. Every instance reads from the same key,
ensuring consistency across the cluster.

### Hotel Lifecycle State Machine

```
         schedule_close()           countdown_done()
OPEN ─────────────────────> CLOSING ──────────────────> CLOSED
  ^                           |                           |
  |                           | cancel_close()            |
  |                           v                           |
  +─────── reopen() ──────── OPEN <──── reopen() ────────+
```

**State transitions are performed by any instance** via Redis CAS (compare-and-swap
using `WATCH`/`MULTI`/`EXEC`). The transitioning instance publishes the new state
to `hotel:status` channel. All instances update their local cache.

**OPEN state:**
- `availability.status` sends `{isOpen: true, onShutdown: false, ...}`
- Normal gameplay

**CLOSING state (maintenance scheduled):**
- `availability.status` sends `{isOpen: true, onShutdown: true, ...}`
- `hotel.will_close` (1050) broadcast periodically with decreasing minutes
- `hotel.closes_and_opens_at` (2771) broadcast with scheduled reopen time
- Countdown is driven by a single instance (the one that initiated the close);
  if it crashes, any instance can take over by reading Redis state and resuming

**CLOSED state:**
- `availability.status` sends `{isOpen: false, onShutdown: false, ...}`
- `hotel.closed_and_opens` (3728) sent with reopen time
- New connections on all instances receive `disconnect.reason(12)` and are closed
- Existing connections either stay (read-only) or get kicked depending on
  `userThrownOutAtClose` flag

### hotel.maintenance (1350)

```go
type HotelMaintenancePacket struct {
    IsInMaintenance     bool   // Currently in maintenance
    MinutesUntilChange  int32  // Minutes until maintenance starts or ends
    Duration            int32  // Expected duration (default 15 if absent)
}
```

Sent when admin triggers maintenance mode. Client displays a maintenance banner.
Broadcast via `broadcast:all` to reach every connected client on every instance.

---

## Persistent Action Records

Certain session actions must persist for auditing and deferred delivery
regardless of which instance processed them. These records live in PostgreSQL
(not Redis) because they are permanent.

### Audit Log Schema (Future, Not Part of This Milestone)

| Action              | Stored Fields                                          | Delivery                    |
|---------------------|--------------------------------------------------------|-----------------------------|
| Moderation caution  | `userID`, `moderatorID`, `message`, `detail`, `time`   | Immediate if online, else on next login |
| Generic alert       | `userID`, `message`, `source`, `time`                  | Immediate if online only    |
| Hotel maintenance   | `initiatorID`, `scheduledAt`, `duration`, `reopenAt`   | Broadcast to all on trigger |
| Ban                 | `userID`, `moderatorID`, `reason`, `expiresAt`         | Immediate kick + persist    |

**Ownership**: Each record has a `createdBy` field (user ID or "system") for
future admin panel attribution. The `source` field distinguishes API-triggered
actions (`api`), CLI-triggered (`cli`), and plugin-triggered (`plugin:<name>`).

**Deferred delivery**: When a moderation caution targets an offline user, the
record is stored in PostgreSQL. On next login, the post-auth burst queries for
pending cautions and delivers them. The record is marked as `delivered_at` to
prevent re-delivery.

---

## Desktop View (Room Exit)

### Flow

```
Client                                Server
  |                                      |
  +--- session.desktop_view (105) --->   |  User clicks "exit room"
  |                                      |  1. Remove user from room
  |                                      |  2. Notify room occupants
  |                                      |  3. Persist room-exit state
  |<--- session.desktop_view (3523) ---  |  Confirm: show hotel view
```

**Important**: This is NOT a disconnect. The user remains authenticated and
connected. They are simply moved from a room to the hotel lobby view.

**Vendor behavior**: Arcturus removes the Habbo from the room's `RoomUnit`
list, broadcasts `UserRemove` to other occupants, and sends the client
back to navigator/desktop state.

**Dependency**: Requires room system to be meaningful. The packet definition
and wiring can be implemented now, but the room-exit logic is deferred until
the room realm exists.

---

## Connection Error System

### connection.error (1004)

```go
type ConnectionErrorPacket struct {
    MessageID  int32   // Header ID of the offending message
    ErrorCode  int32   // Numeric error code
    Timestamp  string  // Server-side ISO timestamp
}
```

Sent when the server encounters a protocol-level error (unknown packet ID,
malformed body, unexpected packet for current state). This is **informational**
and does **not** cause disconnection by itself.

**Use cases**:
- Unknown packet ID received -> `{messageID: <id>, errorCode: 1, timestamp: ...}`
- Packet decode failure -> `{messageID: <id>, errorCode: 2, timestamp: ...}`
- Packet received in wrong state -> `{messageID: <id>, errorCode: 3, timestamp: ...}`

---

## Generic Error & Alert System

### session.generic_error (1600)

```go
type GenericErrorPacket struct {
    ErrorCode int32
}
```

**Known error codes** (from Arcturus):

| Code    | Meaning                          |
|---------|----------------------------------|
| -3      | Authentication failed            |
| -400    | Server connection failed         |
| 4008    | Kicked out of room               |
| 4009    | Need VIP subscription            |
| 4010    | Room name unacceptable           |
| 4011    | Cannot ban group member          |
| -100002 | Wrong password used              |

### session.generic_alert (3801)

Simple text alert displayed as a modal on the client:

```go
type GenericAlertPacket struct {
    Message string
}
```

**Distribution**: Alerts targeting a specific user are published to
`broadcast:user:{userID}`. The instance owning that connection delivers the
packet. If the user is offline, the alert is discarded (alerts are transient).

### session.moderation_caution (1890)

```go
type ModerationCautionPacket struct {
    Message string
    Detail  string
}
```

Sent to a user when a moderator issues a caution. Increments the user's
caution counter (requires user persistence).

**Distribution**: Published to `broadcast:user:{userID}`. If offline, stored
in PostgreSQL for deferred delivery on next login.

---

## Edge Cases & Security

### 1. Post-Auth Burst Ordering

The post-authentication packet burst must follow a specific order:

1. `authentication.ok` (2491) - **must be first**
2. `availability.status` (2033) - hotel state
3. `first_login_of_day` (793) - daily check (if applicable)
4. `client.ping` (3928) - start heartbeat

If packets arrive out of order, some clients may not properly initialize their
UI state. The `authentication.ok` packet specifically triggers the client's
post-login initialization routine.

### 2. Desktop View Without Room

If `session.desktop_view` (105) is received when the user is not in a room:
- Silently ignore (no error, no response)
- Do NOT send `session.desktop_view` (3523) back

### 3. Hotel Close During Active Sessions

When hotel enters CLOSED state with `userThrownOutAtClose = true`:
1. Broadcast `hotel.will_close` with countdown (5, 3, 1 minutes) via
   `broadcast:all` - all instances receive and forward to local connections
2. At close time, publish `disconnect.reason(12)` + close to `broadcast:all`
3. Each instance closes its local WebSocket connections
4. All instances reject new WebSocket upgrades with HTTP 503

When `userThrownOutAtClose = false`:
1. Broadcast hotel close notification via `broadcast:all`
2. Existing sessions continue but cannot enter rooms
3. New connections are still rejected on all instances

### 4. Generic Error Flooding

Rate-limit `connection.error` sends per connection to prevent error storms
from malformed client implementations. Maximum 10 errors per minute per
connection; after that, close the connection.

### 5. Moderation Caution to Offline User

If a moderator cautions a user who is offline:
- Store the caution in PostgreSQL with `delivered_at = NULL`
- On next login (any instance), post-auth burst queries pending cautions
- Deliver `session.moderation_caution` and set `delivered_at`

### 6. First Login of Day Across Timezones

Server uses UTC for "day" boundary. The `first_login_of_day` check compares
the user's last login date (UTC) against the current UTC date.

### 7. Instance Crash During Hotel Close Countdown

If the instance driving the close countdown crashes:
- Redis `hotel:status` key retains `{state: "closing", closeAt: <timestamp>}`
- Any other instance can resume the countdown by reading `closeAt` and
  computing remaining minutes
- No coordinator election needed; any instance that reads CLOSING state and
  finds no active countdown ticker claims it

### 8. Broadcast Storm After Mass Disconnect

If the hotel closes and kicks 1000+ users simultaneously:
- Each instance processes its local connections only
- Session registry removals are batched in Redis pipelines (10 per pipeline)
- Close bus publishes are local (no cross-instance close needed when
  broadcasting via `broadcast:all`)

---

## Implementation Roadmap

### Milestone 1: Core Broadcast Bus

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 1 | Define `Broadcaster` port interface                 | -          | PENDING |
| 2 | Implement `RedisBroadcaster` (Pub/Sub adapter)      | 1          | PENDING |
| 3 | Implement `LocalBroadcaster` (in-process, for tests)| 1          | PENDING |
| 4 | Add `InstanceID` to `Session` struct                | -          | PENDING |
| 5 | Add TTL to session registry keys (120s, refresh 60s)| -          | PENDING |
| 6 | Refactor `CloseSignalBus` to use `Broadcaster`      | 2          | PENDING |
| 7 | Unit test: pub/sub round-trip                       | 2          | PENDING |
| 8 | Unit test: session TTL expiry                       | 5          | PENDING |

### Milestone 2: Post-Authentication Burst

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 9 | Compose `availability.status` (2033) S2C            | -          | PENDING |
| 10| Compose `session.first_login_of_day` (793) S2C      | -          | PENDING |
| 11| Hotel status Redis key read on auth                 | -          | PENDING |
| 12| Post-auth burst orchestrator (sends 2033 + 3928)    | 9, 11      | PENDING |
| 13| Integration test: burst after auth.ok               | 12         | PENDING |

### Milestone 3: Hotel Status Lifecycle

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 14| Compose `hotel.will_close` (1050) S2C               | -          | PENDING |
| 15| Compose `hotel.maintenance` (1350) S2C              | -          | PENDING |
| 16| Compose `hotel.closes_and_opens_at` (2771) S2C      | -          | PENDING |
| 17| Compose `hotel.closed_and_opens` (3728) S2C         | -          | PENDING |
| 18| Hotel status state machine with Redis CAS            | 2          | PENDING |
| 19| Close countdown ticker with crash-recovery           | 18         | PENDING |
| 20| Broadcast hotel packets via `broadcast:all`          | 2, 14-17   | PENDING |
| 21| Integration test: scheduled close across instances   | 18, 20     | PENDING |

### Milestone 4: Error & Notification System

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 22| Compose `connection.error` (1004) S2C               | -          | PENDING |
| 23| Compose `session.generic_error` (1600) S2C          | -          | PENDING |
| 24| Compose `session.generic_alert` (3801) S2C          | -          | PENDING |
| 25| Compose `session.moderation_caution` (1890) S2C     | -          | PENDING |
| 26| Targeted send via `broadcast:user:{userID}`         | 2          | PENDING |
| 27| Protocol error handler (unknown/malformed packets)  | 22         | PENDING |
| 28| Error rate limiter per connection                   | 27         | PENDING |

### Milestone 5: Navigation & Disconnect Reasons

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 29| Parse `session.desktop_view` (105) C2S              | -          | PENDING |
| 30| Compose `session.desktop_view` (3523) S2C           | -          | PENDING |
| 31| Desktop view use case (stub until room realm)       | 29, 30     | PENDING |
| 32| Add remaining disconnect reason constants           | -          | PENDING |
| 33| Hotel-closed disconnect flow (reason 12)            | 18, 32     | PENDING |
| 34| Ban disconnect flow (reasons 1, 10)                 | 26, 32     | PENDING |

### Milestone 6: E2E & Integration

| # | Task                                                | Depends On | Status  |
|---|-----------------------------------------------------|------------|---------|
| 35| E2E test: full post-auth burst sequence             | 12         | PENDING |
| 36| E2E test: hotel close broadcast + disconnect        | 20, 33     | PENDING |
| 37| E2E test: connection error on unknown packet        | 27         | PENDING |
| 38| E2E test: targeted alert via broadcast bus          | 26         | PENDING |

---

## Vendor Implementation Comparison

| Aspect                     | pixels-emulator (Go) | PlusEMU (C#)        | Arcturus (Java)      | comet-v2 (Java)      |
|----------------------------|----------------------|---------------------|----------------------|----------------------|
| Post-auth burst            | AuthOk only          | ~10 packets         | ~15 packets          | ~10 packets          |
| availability.status values | N/A                  | Hardcoded (T,F,T)   | Configurable 3 bools | Configurable 3 bools |
| Hotel close system         | Not implemented      | Not shown           | HotelWillCloseComposer + timer | HotelMaintenanceComposer |
| connection.error           | Not implemented      | Not shown           | ConnectionErrorComposer | Not shown          |
| Generic error              | Not implemented      | GenericErrorComposer | GenericErrorMessagesComposer | Not shown       |
| Moderation caution         | Not implemented      | ModerationCautionEvent (incoming) | Via moderation system | Not shown    |
| Desktop view               | Not implemented      | Basic handler       | Room exit + broadcast | Basic handler       |
| Disconnect reasons         | Not shown            | Multiple codes      | Multiple codes       | Multiple codes       |
| Multi-instance             | None                 | None                | None                 | None                 |

### Our Design Choices vs Vendors

| Decision                      | Our Choice                    | Rationale                                                       |
|-------------------------------|-------------------------------|-----------------------------------------------------------------|
| Post-auth burst               | Minimal (status + ping)       | Add packets as features are implemented; avoid sending stubs    |
| Hotel status                  | Redis-backed state machine    | Consistent across instances; crash-recoverable countdown        |
| Broadcast                     | Redis Pub/Sub bus             | No vendor supports multi-instance; we design for it from day one|
| Error rate limiting           | 10/min per connection         | No vendor implements this; prevents error storms                |
| Session TTL                   | 120s with 60s refresh         | Handles instance crashes; no vendor addresses orphan sessions   |
| Audit persistence             | PostgreSQL + deferred delivery| Moderation actions survive restarts; offline delivery on login  |
| Desktop view timing           | Packet now, logic deferred    | Packet codec is trivial; room-exit logic needs room realm       |
| Disconnect reason registry    | Centralized constants file    | All reason codes in one place for cross-realm usage             |

---

## Caveats & Technical Notes

### Broadcast Bus vs Close Bus Unification

The existing `CloseSignalBus` in `pkg/handshake/adapter/realtime/close_bus.go`
is a specialized Redis Pub/Sub wrapper. The new `Broadcaster` in `core/broadcast/`
is a generalized version. Once the broadcast bus is implemented, `CloseSignalBus`
should be refactored to delegate to it, using `broadcast:conn:{connID}` as the
channel. This avoids maintaining two parallel Pub/Sub systems.

### Packet ID Collision: desktop_view

`session.desktop_view` exists as both C2S (105) and S2C (3523). These are
**different packet IDs** despite sharing a name in the protocol spec. The C2S
packet is the client requesting to exit a room; the S2C packet is the server
confirming the client should show the hotel view.

### Packet ID Collision: disconnect.reason vs release_version

Both `disconnect.reason` (S2C) and `handshake.release_version` (C2S) use
packet ID **4000**. This is valid because packet direction disambiguates:
the server never receives a `disconnect.reason` and the client never receives
a `release_version`.

### Post-Auth Burst vs Lazy Loading

Vendors send 10-15 packets in the post-auth burst (user profile, permissions,
inventory, navigator settings, etc.). We adopt a **lazy loading** approach:
only send packets for implemented features. As new realms are added (navigator,
inventory, etc.), their initialization packets are added to the burst via
the event system (see 03-PLUGINS.md).

### Redis Pub/Sub vs Streams

Redis Pub/Sub is fire-and-forget with no persistence. This is correct for
real-time session packets (if a message is missed because no subscriber exists,
it is irrelevant). Redis Streams would add delivery guarantees at the cost of
complexity and cleanup burden. Pub/Sub is the right choice for this use case.

### Hotel Status Key Contention

Multiple instances may attempt state transitions simultaneously (e.g., two
admins scheduling close at the same time). Redis `WATCH`/`MULTI`/`EXEC`
provides optimistic locking: the first transaction wins, the second retries
or fails. This is acceptable for admin-initiated operations which are
infrequent.
