# Hotel Status Management

## Overview

The hotel status system controls whether the server is open for play, in a
maintenance window, or closed. Status changes are broadcast to all connected
clients across all server instances via Redis Pub/Sub.

## State Machine

```
              schedule_close()           countdown_done()
    OPEN ─────────────────────> CLOSING ──────────────────> CLOSED
      ^                           │                           │
      │                           │ cancel_close()            │
      │                           v                           │
      └─────── reopen() ──────── OPEN <──── reopen() ────────┘
```

| State | Description | Client Experience |
|-------|-------------|-------------------|
| **OPEN** | Normal operation | Full gameplay access |
| **CLOSING** | Shutdown scheduled, countdown active | Warning banners, countdown notifications |
| **CLOSED** | Hotel is shut down | New connections rejected; existing may be kicked |

## Redis Persistence

Hotel status is stored as a persistent Redis key:

```json
// hotel:status
{
    "state": "open",
    "closeAt": null,
    "reopenAt": null,
    "thrownOutAtClose": true
}
```

All instances read from this key to determine the current hotel state. When an
instance transitions the state, it updates the key using Redis
`WATCH`/`MULTI`/`EXEC` (optimistic locking) and publishes to the broadcast
channel.

## Packets

### availability.status (S2C 2033)

Sent in the post-authentication burst to every new connection:

| Field | Type | Description |
|-------|------|-------------|
| `isOpen` | bool | Hotel is currently open |
| `onShutdown` | bool | Shutdown is scheduled (CLOSING state) |
| `isAuthentic` | bool | User is authenticated (always true) |

**Values by state:**

| State | isOpen | onShutdown | isAuthentic |
|-------|--------|------------|-------------|
| OPEN | `true` | `false` | `true` |
| CLOSING | `true` | `true` | `true` |
| CLOSED | `false` | `false` | `true` |

### hotel.will_close (S2C 1050)

Broadcast periodically during CLOSING state with a countdown:

| Field | Type | Description |
|-------|------|-------------|
| `minutes` | int32 | Minutes remaining until close |

Sent at intervals defined by `STATUS_COUNTDOWN_TICK_SECONDS` (default 60s).
The client displays a countdown warning banner.

### hotel.closes_and_opens_at (S2C 2771)

Sent once when CLOSING state is entered to announce the schedule:

| Field | Type | Description |
|-------|------|-------------|
| `openHour` | int32 | UTC hour when hotel will reopen (0-23) |
| `openMinute` | int32 | UTC minute when hotel will reopen (0-59) |
| `userThrownOutAtClose` | bool | Whether users are forcefully disconnected at close |

### hotel.closed_and_opens (S2C 3728)

Sent when hotel is in CLOSED state to inform about reopening:

| Field | Type | Description |
|-------|------|-------------|
| `openHour` | int32 | UTC hour when hotel will reopen |
| `openMinute` | int32 | UTC minute when hotel will reopen |

### hotel.maintenance (S2C 1350)

Sent when maintenance mode is triggered:

| Field | Type | Description |
|-------|------|-------------|
| `isInMaintenance` | bool | Currently in maintenance |
| `minutesUntilChange` | int32 | Minutes until maintenance starts or ends |
| `duration` | int32 | Expected maintenance duration in minutes |

## Closing Sequence

When an administrator schedules a hotel close:

```
Time    Action
─────   ──────
T-5m    Broadcast hotel.will_close(5) + hotel.closes_and_opens_at
T-4m    Broadcast hotel.will_close(4)
T-3m    Broadcast hotel.will_close(3)
T-2m    Broadcast hotel.will_close(2)
T-1m    Broadcast hotel.will_close(1)
T-0     State → CLOSED
        If userThrownOutAtClose:
          Broadcast disconnect.reason(12) to all sessions
          Close all WebSocket connections
        All instances reject new connections (HTTP 503)
```

The countdown is driven by a ticker goroutine on the instance that initiated
the close. If that instance crashes, any other instance can resume by reading
the `closeAt` timestamp from Redis and computing remaining time.

## Broadcast Mechanism

All hotel packets are published to `broadcast:all` via Redis Pub/Sub. Every
server instance subscribes to this channel and forwards received packets to
all local WebSocket connections.

```
Admin triggers close on Instance A
  → Instance A updates Redis hotel:status key
  → Instance A publishes hotel.will_close to broadcast:all
  → Instance B receives via Pub/Sub subscription
  → Instance B sends to all its local WebSocket connections
  → Instance C receives via Pub/Sub subscription
  → Instance C sends to all its local WebSocket connections
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `STATUS_OPEN_HOUR` | `0` | Daily open hour (UTC) |
| `STATUS_OPEN_MINUTE` | `0` | Daily open minute |
| `STATUS_CLOSE_HOUR` | `23` | Daily close hour (UTC) |
| `STATUS_CLOSE_MINUTE` | `59` | Daily close minute |
| `STATUS_REDIS_KEY` | `hotel:status` | Redis key for state |
| `STATUS_BROADCAST_CHANNEL` | `broadcast:all` | Pub/Sub channel |
| `STATUS_COUNTDOWN_TICK_SECONDS` | `60` | Countdown interval |
| `STATUS_DEFAULT_MAINTENANCE_DURATION_MINUTES` | `15` | Default maintenance duration |

## Edge Cases

### Instance Crash During Countdown

Redis retains the `closeAt` timestamp. Any instance can resume the countdown
by reading the timestamp and calculating remaining minutes. No coordinator
election needed.

### Concurrent State Transitions

Two administrators scheduling close simultaneously: Redis `WATCH`/`MULTI`/`EXEC`
ensures only one transaction succeeds. The other retries or fails gracefully.

### Close with `userThrownOutAtClose = false`

Users remain connected but cannot enter rooms. The client displays a "hotel
closed" overlay. New connections are still rejected.

### Reopen While Users Are Connected

If the hotel reopens while users are still connected (from a
`userThrownOutAtClose = false` close), they receive a new `availability.status`
with `{isOpen: true}` and can resume normal gameplay.
