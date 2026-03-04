# Realm: Session & Connection

> **Position:** 20 | **Phase:** 1 (Connection) | **Packets:** 30 (10 c2s, 20 s2c)
> **Services:** gateway (primary), game (secondary) | **Status:** Partially implemented

---

## Overview

The Session & Connection realm manages the lifecycle of an authenticated session: keep-alive (ping/pong), latency measurement, availability status, hotel maintenance windows, error reporting, and session restoration. Most packets in this realm are handled entirely within the gateway service with no NATS round-trip, making it the lowest-latency realm.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 1

---

## Packet Inventory

### C2S (Client to Server) -- 10 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 295 | `client.latency_test` | `field1:int32` | Client responds to latency probe |
| 2596 | `client.pong` | `field1:int32` | Client responds to ping (keep-alive) |
| 2445 | `client.disconnect` | _(none)_ | Client requests graceful disconnect |
| 105 | `session.desktop_view` | _(none)_ | Client navigated to hotel view (left room) |
| 1160 | `session.peer_users_classification` | _(none)_ | Peer classification request |
| 2313 | `session.client_toolbar_toggle` | `field1:string` | Client toggled a toolbar element |
| 3226 | `session.render_room` | _(none)_ | Client finished rendering current room |
| 3230 | `session.tracking_performance_log` | _(none)_ | Client performance telemetry |
| 3457 | `session.event_tracker` | _(none)_ | Client-side analytics event |
| 3847 | `session.tracking_lag_warning_report` | _(none)_ | Client lag report |

### S2C (Server to Client) -- 20 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 3928 | `client.ping` | _(none)_ | Server keep-alive ping |
| 10 | `client.latency_response` | `field1:int32` | Server latency measurement response |
| 2033 | `availability.status` | `field1:boolean`, `field2:boolean`, `field3:boolean` | Hotel availability flags (open, shutting down, closed) |
| 600 | `availability.time` | `field1:int32` | Seconds until maintenance begins |
| 1004 | `connection.error` | `field1:int32` | Connection error with reason code |
| 4000 | `disconnect.reason` | `field1:int32` | Disconnect with reason code |
| 1050 | `hotel.will_close_minutes` | `field1:int32` | Minutes until hotel closes |
| 1350 | `hotel.maintenance` | `field1:boolean`, `field2:int32` | Maintenance mode toggle + duration |
| 3728 | `hotel.closed_and_opens` | `field1:int32` | Hotel closed, reopens at timestamp |
| 2771 | `hotel.closes_and_opens_at` | `field1:int32`, `field2:int32` | Close/reopen timestamps |
| 122 | `session.desktop_view` | _(none)_ | Acknowledge hotel view navigation |
| 426 | `session.restore_client` | _(none)_ | Restore client after reconnection |
| 793 | `session.first_login_of_day` | _(none)_ | First daily login flag |
| 1600 | `session.generic_error` | `field1:int32` | Generic error code |
| 1663 | `session.hotel_merge_name_change` | _(none)_ | Hotel merge name change notification |
| 1890 | `session.moderation_caution` | `field1:string`, `field2:string` | Moderation warning message |
| 2035 | `session.motd_messages` | `field1:int32` | Message of the day |
| 3284 | `session.info_feed_enable` | `field1:boolean` | Enable/disable info feed |
| 3801 | `session.generic_alert` | `field1:string` | Generic alert popup |
| 3945 | `session.epic_popup` | `field1:string`, `field2:string` | Full-screen promotional popup |

---

## Architecture Mapping

### Service Ownership

```
Client ◄──ping/pong────▶ Gateway (inline, no NATS)
Client ──desktop_view──▶ Gateway ──NATS──▶ Game Service (room leave)
Client ◄──availability──  Gateway (computed locally or from config)
```

**Key design principle:** The vast majority of session-connection packets are handled inline in the gateway with zero NATS overhead. Only `session.desktop_view` (which triggers room leave) and `session.render_room` (which informs the game service) require NATS communication.

### Gateway State Per Session

```go
type Session struct {
    ID            string
    conn          net.Conn
    userID        int32
    authenticated atomic.Bool
    lastPong      atomic.Int64  // unix timestamp
    latencyMs     atomic.Int32  // last measured RTT
    createdAt     time.Time
    machineID     string
}
```

### Telemetry Packets (Read-Only)

Packets 3230, 3457, and 3847 are telemetry from the client. In production Habbo, these feed analytics pipelines. For pixel-server:
- **Phase 1:** Read and discard (log at debug level).
- **Future:** Optionally forward to an external analytics sink via NATS.

---

## Implementation Analysis

### Ping/Pong Keep-Alive

The server sends `client.ping` (3928) at a configurable interval (default: 30 seconds). The client must respond with `client.pong` (2596) within a timeout window.

```
Gateway:
  ticker := time.NewTicker(30 * time.Second)
  for range ticker.C:
      session.SendPacket(ClientPingOutPacket{})
      session.lastPingSent = time.Now()

  // Separate timeout checker:
  if time.Since(session.lastPong) > 90*time.Second:
      session.Disconnect(ReasonPingTimeout)
```

**Configuration points:**
- `PING_INTERVAL` -- seconds between pings (default: 30)
- `PONG_TIMEOUT` -- seconds before declaring session dead (default: 90)

### Latency Measurement

`client.latency_test` (295) is a round-trip measurement:
1. Server sends `client.latency_response` (10) with a timestamp token.
2. Client echoes back `client.latency_test` (295) with the same token.
3. Server computes `RTT = now - sentTime`.

Store the rolling average RTT on the session for adaptive timeout tuning.

### Availability Status

`availability.status` (2033) has three boolean flags:
- `field1` -- hotel is open for login
- `field2` -- hotel is shutting down (read-only mode)
- `field3` -- hotel is fully closed

Sent immediately after `authentication.ok`. Gateway should read these from a global config (Redis key or environment variable) that can be toggled by operations without restarting.

### Hotel Maintenance Windows

The suite of hotel maintenance packets (`hotel.will_close_minutes`, `hotel.maintenance`, `hotel.closed_and_opens`, `hotel.closes_and_opens_at`) supports scheduled downtime announcements:

1. Ops sets maintenance window in admin panel (or Redis key).
2. Gateway broadcasts `hotel.will_close_minutes` (1050) to all sessions at T-30, T-15, T-5.
3. At maintenance time, broadcast `hotel.maintenance` (1350) with `enabled=true`.
4. Reject new connections with `availability.status` `open=false`.
5. After maintenance, broadcast `hotel.closed_and_opens` (3728) with reopen timestamp.

### Desktop View (Hotel View)

`session.desktop_view` (105 c2s) signals the client navigated to the hotel lobby view. This must trigger:
1. If user is in a room: send room leave command to game service via NATS.
2. Update user's state to "in lobby" in Redis.
3. Respond with `session.desktop_view` (122 s2c) acknowledgement.

### Session Restoration

`session.restore_client` (426) is sent when a user's connection drops and they reconnect within a grace period. The gateway should:
1. Check Redis for existing session with same `userID`.
2. If found and within grace period: restore state (current room, etc.).
3. If not found: treat as new login.

---

## Caveats & Edge Cases

### 1. Ping/Pong and Mobile Clients
Mobile browsers may suspend WebSocket connections when the app is backgrounded. A strict 90-second pong timeout will disconnect mobile users. Consider a longer timeout for mobile `deviceCategory` (from handshake).

### 2. Disconnect Reason Codes
`disconnect.reason` (4000) and `connection.error` (1004) use integer reason codes. These must be documented and consistent. Known codes from reference implementations:
- `0` -- generic
- `1` -- banned
- `2` -- concurrent login (kicked by newer session)
- `3` -- hotel closed
- `4` -- idle timeout
- `5` -- maintenance

### 3. Concurrent Login Handling
When a user logs in from a second device, the first session must be terminated. The auth service publishes `session.authenticated` with the user ID; the gateway must check if another session exists for the same user and disconnect it with reason code 2 before promoting the new session.

### 4. Telemetry Packet Flooding
Malicious clients could flood telemetry packets (3230, 3457, 3847) to waste gateway CPU. Rate-limit these packets to at most 1 per second per session, discarding excess silently.

### 5. Generic Alert Injection
`session.generic_alert` (3801) and `session.epic_popup` (3945) are server-to-client only, but the content must be sanitized if it originates from admin input. Never render raw HTML in these packets.

### 6. First Login of Day Tracking
`session.first_login_of_day` (793) requires tracking the last login date per user. Use the `users.last_login` column and compare against the current date (server timezone). This flag drives daily reward systems in later phases.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Keep-alive** | Fixed 30s interval, no adaptation | Configurable interval; RTT-adaptive timeout possible |
| **Concurrent login** | In-memory session check (single-node only) | Redis-backed session store; works across gateway replicas |
| **Maintenance windows** | Hard restart required | Graceful degradation with progressive warnings |
| **Telemetry** | Discarded entirely | Optional NATS forwarding to analytics sink |
| **Session restore** | Not supported | Redis grace-period session with state recovery |
| **Rate limiting** | None | Per-packet-type rate limiting on gateway |

---

## Dependencies

- **Phase 1 handshake** -- session must be authenticated before most session packets are meaningful
- **pkg/core/bus** -- NATS for desktop_view -> game service room leave
- **Redis** -- session store, availability config, concurrent login detection
- **pkg/core/config** -- ping interval, pong timeout, maintenance window settings

---

## Testing Strategy

### Unit Tests
- Ping/pong timer logic (mock clock)
- Latency measurement RTT computation
- Concurrent login detection (two sessions, same user)
- Availability status flag combinations

### Integration Tests
- Full ping/pong cycle with real WebSocket (testcontainers)
- Session timeout after missed pongs
- Desktop view triggers room leave in game service
- Concurrent login disconnects first session

### E2E Tests
- Connected client survives idle for 60+ seconds with ping/pong
- Hotel maintenance announcement reaches all connected clients
- Second login from same user kicks first session
