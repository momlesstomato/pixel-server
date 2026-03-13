# Management API

The management API provides REST endpoints and CLI commands for administering
sessions, connections, and hotel status across all server instances.

All REST endpoints require the `X-API-Key` header (same key as SSO endpoints).
CLI commands interact with Redis directly and do not require a running server.

---

## Session Management

### REST API

#### List Sessions

```
GET /api/v1/sessions
GET /api/v1/sessions?instance=myserver
```

Returns all active sessions in the cluster. Optionally filter by instance ID.

**Response:**

```json
{
  "sessions": [
    {
      "conn_id": "abc123",
      "user_id": 42,
      "machine_id": "f8a3...",
      "state": "authenticated",
      "instance_id": "myserver:0a1b2c",
      "created_at": "2026-03-12T15:04:05Z"
    }
  ],
  "count": 1
}
```

Session states: `connected` (pre-auth), `authenticated`, `disconnecting`.

#### Get Session

```
GET /api/v1/sessions/:connID
```

Returns one session by connection identifier.

#### Disconnect Session

```
DELETE /api/v1/sessions/:connID
```

Publishes a close signal via the broadcast bus (works cross-instance), then
removes the session from Redis. The owning instance receives the close signal
and terminates the WebSocket connection.

**Response:**

```json
{ "disconnected": "abc123" }
```

### CLI Commands

```bash
# List all sessions
pixelsv session list

# Filter by instance
pixelsv session list --instance myserver

# Disconnect one session
pixelsv session kick abc123
```

---

## Hotel Status Management

### REST API

#### Get Hotel Status

```
GET /api/v1/hotel/status
```

Returns the current hotel state machine snapshot.

**Response:**

```json
{
  "state": "open",
  "close_at": null,
  "reopen_at": null,
  "throw_users": false
}
```

States: `open`, `closing`, `closed`.

#### Schedule Hotel Close

```
POST /api/v1/hotel/close
```

Transitions the hotel to `closing` state. Publishes countdown packets to all
connected clients via the broadcast bus.

**Request:**

```json
{
  "minutes_until_close": 5,
  "duration_minutes": 15,
  "throw_users": true
}
```

- `minutes_until_close` — countdown before closing (0 = immediate)
- `duration_minutes` — maintenance window duration (defaults to `STATUS_DEFAULT_MAINTENANCE_DURATION_MINUTES`)
- `throw_users` — disconnect all users when close time is reached

**Response:** Updated hotel status object.

Returns `409 Conflict` if the state transition is invalid (e.g., already closed).

#### Reopen Hotel

```
POST /api/v1/hotel/reopen
```

Transitions the hotel to `open` state immediately. No request body required.

Returns `409 Conflict` if the state transition is invalid.

### CLI Commands

```bash
# Show current status
pixelsv hotel status

# Schedule close in 5 minutes, 15-minute maintenance, disconnect users
pixelsv hotel close --minutes 5 --duration 15 --throw-users

# Reopen immediately
pixelsv hotel reopen
```

---

## Cross-Instance Behavior

All management operations work across multiple server instances:

- **Session disconnect** uses Redis Pub/Sub to send a close signal to the
  instance that owns the connection. The owning instance terminates the
  WebSocket and removes the session.

- **Hotel status** is stored in Redis with compare-and-swap semantics. State
  transitions are atomic. The countdown ticker runs independently on each
  instance and advances transitions when the scheduled time arrives.

- **Session listing** uses Redis `SCAN` to enumerate all `session:conn:*` keys.
  Each session record includes an `instance_id` field for filtering.

---

## OpenAPI

All management endpoints are documented in the OpenAPI specification at
`/openapi.json` and browsable via the Swagger UI at `/swagger`.
