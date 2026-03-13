# SSO Authentication

## Overview

SSO (Single Sign-On) tickets are the authentication mechanism for WebSocket
connections. A ticket is a single-use, time-limited token that maps to a user
ID. Tickets are generated via the REST API or CLI, stored in Redis, and
consumed during the WebSocket handshake.

## Ticket Lifecycle

```
1. Generate   API/CLI creates ticket → stored in Redis with TTL
2. Deliver    External system passes ticket to Nitro client
3. Consume    Client sends ticket via security.sso_ticket packet
4. Validate   Server atomically reads and deletes from Redis (GETDEL)
5. Expire     If not consumed within TTL, Redis auto-deletes
```

## Generating Tickets

### REST API

```
POST /api/v1/sso
Header: X-API-Key: <your-api-key>
Content-Type: application/json

{
    "user_id": 42,
    "ttl_seconds": 300
}
```

**Response:**

```json
{
    "ticket": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "user_id": 42,
    "expires_in_seconds": 300
}
```

**Validation rules:**
- `user_id` must be > 0
- `ttl_seconds` must be between 1 and `AUTHENTICATION_MAX_TTL_SECONDS`
- If `ttl_seconds` is omitted, `AUTHENTICATION_DEFAULT_TTL_SECONDS` is used

### CLI

```bash
pixelsv sso issue --user-id 42 --ttl 5m
```

Prints the ticket to stdout. Useful for development and debugging.

## Redis Storage

```
Key:    sso:<ticket-uuid>
Value:  <user-id>  (string representation of integer)
TTL:    <configured seconds>
```

Example:

```
SET sso:a1b2c3d4-e5f6-7890-abcd-ef1234567890 42 EX 300
```

## Validation (Atomic)

When the client sends `security.sso_ticket` (2419), the server validates:

```
GETDEL sso:<ticket>
```

`GETDEL` is a single atomic Redis command (requires Redis 6.2+) that reads the
value and deletes the key in one operation. This guarantees:

- **Single-use**: If two connections race with the same ticket, exactly one
  gets the user ID and the other gets `nil`
- **No race window**: Unlike `GET` then `DEL`, there is no time gap where a
  second client could also `GET` successfully

### Validation Outcomes

| Outcome | Server Behavior |
|---------|----------------|
| Valid ticket (user ID returned) | Proceed with authentication |
| Missing/expired ticket (`nil`) | Send `disconnect.reason(22)`, close connection |
| Empty or whitespace ticket | Close connection immediately |
| Ticket > 128 characters | Close connection immediately |

## Authentication Flow

After successful ticket validation:

```go
// 1. Check for duplicate login
existing, found := sessions.FindByUserID(userID)
if found && existing.ConnID != request.ConnID {
    // Kick the old session
    transport.Send(existing.ConnID, disconnectReasonPacket(2))  // concurrent login
    transport.Close(existing.ConnID, 1008, "duplicate login")
    sessions.Remove(existing.ConnID)
}

// 2. Register new session
sessions.Register(Session{
    ConnID:    request.ConnID,
    UserID:    userID,
    MachineID: request.MachineID,
    State:     StateAuthenticated,
})

// 3. Send success packets
transport.Send(connID, authenticationOK)
transport.Send(connID, identityAccounts(userID))
```

### Duplicate Login Handling

When a user authenticates and already has an active session (on any instance):

1. **Find existing session** via `sessions.FindByUserID()` — this searches Redis,
   so it works across instances
2. **Send disconnect reason** (code 2 = concurrent login) to the OLD connection
3. **Close the old connection** — if the old connection is on another instance,
   the close signal is published via Redis Pub/Sub and the owning instance
   closes it
4. **Remove old session** from the registry
5. **Register new session** and proceed with authentication

The new connection always wins. The old connection is always kicked.

### identity_accounts (S2C 3523)

After `authentication.ok`, the server sends the account list:

| Field | Type | Description |
|-------|------|-------------|
| `count` | int32 | Number of accounts (always 1 currently) |
| `accounts[].id` | int32 | User ID |
| `accounts[].name` | string | Display name |

Until the user realm is fully implemented, the name is stubbed as
`Player#<userID>`.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTHENTICATION_DEFAULT_TTL_SECONDS` | `300` | Default ticket lifetime |
| `AUTHENTICATION_MAX_TTL_SECONDS` | `1800` | Maximum allowed ticket lifetime |
| `AUTHENTICATION_KEY_PREFIX` | `sso` | Redis key prefix |

## Security Considerations

- **Tickets are single-use** — once consumed, they cannot be reused
- **Tickets expire** — unclaimed tickets are automatically deleted by Redis TTL
- **API key required** — ticket generation requires the `X-API-Key` header
- **No ticket content in logs** — only a sanitized prefix is logged on failure
- **Atomic validation** — `GETDEL` prevents race conditions
- **Cross-instance duplicate detection** — Redis-backed registry catches
  duplicates across all server instances
