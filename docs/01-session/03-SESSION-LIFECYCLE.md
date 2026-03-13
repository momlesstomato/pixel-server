# Session Lifecycle

## Overview

A session tracks a WebSocket connection from upgrade through disconnection.
Sessions transition through three states and are persisted in Redis for
cross-instance visibility.

## Session States

```
                    auth success
  CONNECTED ─────────────────────> AUTHENTICATED
      │                                   │
      │ auth timeout (30s)                │ disconnect packet
      │ or invalid SSO                    │ or pong timeout
      v                                   v
  [connection closed]              DISCONNECTING
                                          │
                                          v
                                   [connection closed]
```

| State | Meaning | Duration |
|-------|---------|----------|
| `Connected` | TCP/WebSocket established, no auth yet | Max 30 seconds |
| `Authenticated` | SSO validated, game packets accepted | Until disconnect |
| `Disconnecting` | Graceful close in progress | Brief (< 1 second) |

## Session Record

Each session is stored in Redis as JSON with two index keys:

```json
// session:conn:abc123def456...
{
    "ConnID": "abc123def456...",
    "UserID": 42,
    "MachineID": "e4f5a6b7...",
    "State": 1,
    "InstanceID": "server-01:a1b2c3",
    "CreatedAt": "2026-03-12T14:30:00Z"
}
```

```
// session:user:42
abc123def456...
```

Both keys have a TTL of 120 seconds, refreshed every 60 seconds by the
heartbeat goroutine. This provides automatic cleanup if the owning instance
crashes.

## Heartbeat (Keep-Alive)

The heartbeat runs as a per-connection goroutine after authentication:

```
Server                    Client
  │                          │
  │── client.ping (3928) ──>│  Every 30 seconds
  │                          │
  │<── client.pong (2596) ──│  Must arrive within 60 seconds
  │                          │
  │   [refresh session TTL]  │
  │                          │
```

### Timeout Handling

If no pong arrives within 60 seconds:

1. Send `disconnect.reason` (4000) with reason 113 (pong timeout)
2. Close WebSocket with code 1006 (abnormal closure)
3. Remove session from registry

The heartbeat also calls `Touch(connID)` to refresh the session TTL in Redis.

## Latency Measurement

```
Client                    Server
  │                          │
  │── latency_test (295) ──>│  requestId: <int32>
  │                          │
  │<── latency_response (10)│  requestId: <same int32>
  │                          │
```

The server echoes the request ID immediately. The client measures round-trip
time by comparing send and receive timestamps locally.

## Post-Authentication Burst

Immediately after `authentication.ok`, the server sends a burst of packets to
initialize the client state:

| Order | Packet | Purpose |
|-------|--------|---------|
| 1 | `authentication.ok` (2491) | Triggers client post-login initialization |
| 2 | `identity_accounts` (3523) | Account list |
| 3 | `availability.status` (2033) | Hotel open/close/shutdown state |
| 4 | `first_login_of_day` (793) | Daily login bonus trigger |
| 5 | `client.ping` (3928) | Starts heartbeat cycle |

**Order matters.** The client expects `authentication.ok` first because it
triggers the UI initialization routine. `availability.status` must come before
game packets because it determines whether the hotel is open.

## Graceful Disconnect

When the client sends `client.disconnect` (2445):

1. Update session state to `Disconnecting`
2. Remove session from registry (Redis keys deleted)
3. Close WebSocket with code 1000 (normal closure)

The server sends `disconnect.reason` (4000) with reason 0 (logout) before
closing.

## Abrupt Disconnect

When the WebSocket closes unexpectedly (network drop, browser close):

1. The `onclose` handler fires
2. Session is removed from registry
3. Resources are cleaned up (heartbeat goroutine stopped, close bus
   subscription disposed)

No disconnect reason packet is sent because the connection is already gone.

## Disconnect Reason Codes

The `disconnect.reason` packet (4000) is sent before closing a connection to
explain why:

| Code | Name | Cause |
|------|------|-------|
| 0 | Logout | Normal user-initiated disconnect |
| 1 | Just Banned | User was banned during active session |
| 2 | Concurrent Login | New login on another connection kicked this one |
| 10 | Still Banned | Login rejected because account is banned |
| 12 | Hotel Closed | Hotel closed while user was connected |
| 22 | Invalid Login Ticket | SSO ticket was invalid, expired, or missing |
| 113 | Pong Timeout | Heartbeat pong not received within 60 seconds |
| 114 | Idle Not Authenticated | No SSO ticket received within 30 seconds |

## Cross-Instance Disconnect

When instance A needs to close a connection on instance B:

1. Instance A publishes a close signal to `broadcast:conn:{connID}` via
   Redis Pub/Sub
2. Instance B's close bus subscriber receives the signal
3. Instance B closes the WebSocket locally

This is used by duplicate login detection (the old session may be on a
different instance) and moderation actions (ban from any instance).

## Session Registry Operations

| Operation | Use Case |
|-----------|----------|
| `Register(session)` | On auth success; updates both conn and user indexes |
| `FindByUserID(id)` | Duplicate login detection (works cross-instance) |
| `FindByConnID(id)` | Session lookup for packet handling |
| `Touch(connID)` | Heartbeat TTL refresh (every 60s) |
| `Remove(connID)` | On disconnect; cleans both conn and user indexes |
