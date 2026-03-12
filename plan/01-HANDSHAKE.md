# 01 - Handshake & Security Realm

## Overview

The Handshake & Security realm owns the WebSocket connection lifecycle from TCP
upgrade through authenticated session. It covers cryptographic key exchange
(Diffie-Hellman + RSA + RC4), SSO token authentication via Redis, machine
fingerprinting, keep-alive heartbeat, and graceful disconnection.

This realm is the **gateway boundary** - no game packet may be processed until
`authentication.ok` has been sent to the client.

---

## Vendor Cross-Reference

Analysis of four reference implementations (pixels-emulator Go, PlusEMU C#,
Arcturus-Community Java, comet-v2 Java) and the pixel-protocol YAML spec.

### Handshake Sequence (All Vendors Agree)

```
Client                                    Server
  |                                          |
  +--- release_version (4000) ------------->  |  Phase 1: Pre-Auth
  +--- client_variables (1053) ------------>  |
  +--- security.machine_id (2490) --------->  |
  |<------------ security.machine_id (1488)  |
  |                                          |
  +--- init_diffie (3110) ----------------->  |  Phase 2: Crypto (optional)
  |<------------ init_diffie (1347) ------   |
  +--- complete_diffie (773) -------------->  |
  |<------------ complete_diffie (3885) --   |  RC4 installed both sides
  |                                          |
  +--- security.sso_ticket (2419) --------->  |  Phase 3: Auth
  |<------------ authentication.ok (2491) -  |
  |<------------ identity_accounts (3523) -  |
  |<------------ client.ping (3928) ------   |  Phase 4: Session
  |                                          |
  +--- client.pong (2596) ----------------->  |  Every ~30s ping, 60s timeout
  |                                          |
  +--- client.disconnect (2445) ----------->  |  Graceful close
```

---

## Packet Registry

### Client-to-Server (C2S)

| ID   | Name                            | Phase    | Fields                                                               | Status          |
|------|---------------------------------|----------|----------------------------------------------------------------------|-----------------|
| 4000 | `handshake.release_version`     | pre-auth | `releaseVersion: string`, `clientType: string`, `platform: int32`, `deviceCategory: int32` | DONE            |
| 1053 | `handshake.client_variables`    | pre-auth | `clientId: int32`, `clientUrl: string`, `externalVariablesUrl: string` | DONE            |
| 2490 | `security.machine_id`          | pre-auth | `machineId: string(64 hex)`, `fingerprint: string`, `capabilities: string` | DONE            |
| 3110 | `handshake.init_diffie`        | crypto   | _(empty)_                                                            | DEFERRED        |
| 773  | `handshake.complete_diffie`    | crypto   | `encryptedPublicKey: string`                                         | DEFERRED        |
| 2419 | `security.sso_ticket`          | auth     | `ticket: string`, `timestamp: int32 (optional)`                      | DONE            |
| 2596 | `client.pong`                  | session  | _(empty)_                                                            | DONE            |
| 295  | `client.latency_test`          | session  | `requestId: int32`                                                   | DONE            |
| 2445 | `client.disconnect`            | session  | _(empty)_                                                            | DONE            |

### Server-to-Client (S2C)

| ID   | Name                            | Phase    | Fields                                                               | Status          |
|------|---------------------------------|----------|----------------------------------------------------------------------|-----------------|
| 1347 | `handshake.init_diffie`        | crypto   | `encryptedPrime: string`, `encryptedGenerator: string`               | DEFERRED        |
| 3885 | `handshake.complete_diffie`    | crypto   | `encryptedPublicKey: string`, `serverClientEncryption: bool`         | DEFERRED        |
| 1488 | `security.machine_id`          | auth     | `machineId: string(64 hex)`                                         | DONE            |
| 2491 | `authentication.ok`            | auth     | _(empty)_                                                            | DONE            |
| 3523 | `handshake.identity_accounts`  | auth     | `count: int32`, `accounts: [{id: int32, name: string}]`             | DONE            |
| 3928 | `client.ping`                  | session  | _(empty)_                                                            | DONE            |
| 10   | `client.latency_response`      | session  | `requestId: int32`                                                   | DONE            |

### Status Legend

- **PLANNED** - Will be implemented in this realm
- **DEFERRED** - Not in initial milestone (see reason below)
- **DONE** - Implemented and tested

### Deferred Packets - Rationale

| Packet                     | Reason                                                                                                |
|----------------------------|-------------------------------------------------------------------------------------------------------|
| `handshake.init_diffie`    | RC4+RSA+DH encryption is optional; Nitro client works without it. Adds significant crypto complexity (RSA key management, RC4 stream cipher). Implement after core auth flow is stable. |
| `handshake.complete_diffie`| Same as above - part of the crypto handshake pair.                                                    |
| `handshake.client_policy`  | Not present in any vendor implementation analyzed. Likely deprecated or client-only.                  |

---

## Architecture

### Package Layout

```
pkg/handshake/packet/
  bootstrap/            <- Release negotiation packets (4000, 1053)
  security/             <- Machine and SSO packets (2490, 1488, 2419)
  authentication/       <- Auth success/account list packets (2491, 3523)
  session/              <- Lifecycle heartbeat/disconnect packets (3928, 2596, 2445)
  telemetry/            <- Latency test/response packets (295, 10)

core/connection/
  conn.go               <- Connection abstraction (read, write, close)
  session_registry.go   <- Session registry interface + Redis-backed impl

core/codec/
  frame.go              <- Frame encode/decode (wire header + body)
  primitives.go         <- Typed payload reader/writer primitives

core/redis/
  client.go             <- Redis client factory
  stage.go              <- Redis initializer stage
```

### Domain Model

```go
// Session represents an authenticated WebSocket connection.
type Session struct {
    ConnID    string
    UserID    int       // zero until authenticated
    MachineID string    // 64-char hex from client
    State     SessionState
    CreatedAt time.Time
}

type SessionState int
const (
    StateConnected SessionState = iota  // TCP up, no auth
    StateAuthenticated                   // SSO validated
    StateDisconnecting                   // graceful close in progress
)
```

### Port Interfaces

```go
// SSOStore manages single-use SSO tokens with expiration.
type SSOStore interface {
    Store(ctx context.Context, token string, userID int, ttl time.Duration) error
    Validate(ctx context.Context, token string) (userID int, err error) // GET + DEL atomic
}

// SessionRegistry tracks active sessions for duplicate detection.
type SessionRegistry interface {
    Register(session *Session) error
    FindByUserID(userID int) (*Session, bool)
    Remove(connID string)
}

// Transport sends packets to a specific connection.
type Transport interface {
    Send(connID string, packet Packet) error
    Close(connID string) error
}
```

---

## SSO Token Design

### Generation

SSO tokens are generated **outside the WebSocket flow** via two paths:

1. **REST API** - `POST /api/v1/sso` with body `{"user_id": 123, "ttl_seconds": 300}`
   - Protected by `X-API-Key` header (already implemented)
   - Returns `{"ticket": "<uuid-v4>", "expires_at": "2026-03-11T12:05:00Z"}`

2. **CLI Command** - `pixelsv sso --user-id 123 --ttl 5m`
   - Prints ticket to stdout for development/debugging
   - Uses same Redis store as API path

### Redis Storage

```
Key:     sso:<ticket>
Value:   <user_id>  (string representation of int)
TTL:     configurable (default 300s / 5 minutes)
Command: SET sso:<ticket> <user_id> EX <ttl>
```

### Validation (Atomic)

```
GETDEL sso:<ticket>
```

Single atomic operation: reads the user ID and deletes the key in one step.
This guarantees single-use tokens - a race between two connections with the
same ticket will result in exactly one winner.

**Why GETDEL over GET+DEL pipeline:** `GETDEL` is atomic at the Redis command
level (Redis 6.2+). A `GET` then `DEL` pipeline has a race window where two
clients could both `GET` successfully before either `DEL` executes.

### Configuration

```go
type AuthenticationConfig struct {
    DefaultTTLSeconds int    `mapstructure:"default_ttl_seconds" default:"300"`
    MaxTTLSeconds     int    `mapstructure:"max_ttl_seconds" default:"1800"`
    KeyPrefix         string `mapstructure:"key_prefix" default:"sso"`
}
```

Added as a nested field in the existing `Config` struct under `authentication`.

### Redis Dependency

Redis client dependency is already present and the adapter is implemented in
`pkg/authentication` (`RedisStore`) with `SET` and atomic `GETDEL`.

---

## Edge Cases & Security

### 1. Duplicate Login (User Already Connected)

When `security.sso_ticket` arrives and the user ID is already in the session
registry:

1. Look up existing session by user ID via `SessionRegistry.FindByUserID()`
2. Send a close frame to the **old** connection via `Transport.Close(oldConnID)`
3. Remove old session from registry
4. Proceed with new authentication normally

**Vendor consensus:** PlusEMU explicitly kicks via `DisconnectCurrentOnlineHabboTask`.
Arcturus implicitly disposes old `GameClient`. We follow PlusEMU's explicit approach.

**No dedicated "kick" packet exists.** All vendors simply close the WebSocket
connection. The Nitro client handles `onclose` gracefully.

### 2. Invalid / Expired / Non-Existent SSO Token

When `GETDEL sso:<ticket>` returns `redis.Nil`:

1. Log warning with connection ID and (sanitized) ticket prefix
2. Close WebSocket connection immediately (close frame 4001 "Unauthorized")
3. No `authentication.ok` is sent

**No dedicated "auth failed" packet exists across any vendor.** All vendors
simply close/dispose the connection on auth failure.

### 3. Empty or Malformed Ticket

- Whitespace-only ticket: strip and treat as empty -> close connection
- Ticket exceeding 128 chars: reject immediately (comet-v2 validates 8-128)
- Ticket with non-printable characters: reject

### 4. Machine ID Validation

Following Arcturus pattern:
- Must be exactly 64 hex characters
- If starts with `~` or wrong length: server generates a new random 64-char hex
  and returns it via `security.machine_id` (S2C 1488)
- Machine ID stored on session for later ban checks

### 5. Connection Without Authentication

If a connection sends game packets before completing SSO auth:
- Packets are silently dropped
- Session state machine enforces ordering: `Connected -> Authenticated`
- After 30 seconds with no `sso_ticket`, server closes connection

### 6. Ping/Pong Timeout

- Server sends `client.ping` (3928) every 30 seconds
- Client must reply with `client.pong` (2596) within 60 seconds
- On timeout: close connection, remove session, clean up

### 7. Graceful Disconnect

When `client.disconnect` (2445) received:
1. Set session state to `Disconnecting`
2. Remove from session registry
3. Close WebSocket connection
4. _(Future: persist user state, notify rooms, etc.)_

### 8. Abrupt Disconnect (Network Drop)

WebSocket `onclose`/`onerror` handler:
1. Remove from session registry
2. Clean up resources
3. _(Future: same as graceful but with potential reconnect window)_

---

## Encryption Decision: DEFERRED

RC4 + RSA + Diffie-Hellman encryption is **deferred** to a later milestone.

### Rationale

1. **Nitro client works without encryption** - the `encryption.forced` config
   in Arcturus/comet defaults to `false`
2. **WebSocket already provides TLS** - wss:// gives transport encryption that
   the original Flash TCP socket lacked
3. **Complexity** - RSA key management, DH parameter generation (128-bit
   primes), RC4 stream cipher installation on every packet is significant
   crypto code
4. **No security benefit over TLS** - RC4 is deprecated (RFC 7465) and the
   original protocol used it only because Flash TCP had no TLS

### When to Implement

If/when supporting non-TLS WebSocket connections or legacy Flash clients. The
packet IDs (3110, 773, 1347, 3885) are reserved and the handshake flow is
designed to slot encryption in between pre-auth and auth phases without
breaking changes.

---

## Implementation Roadmap

### Milestone 1: Foundation (Core Infrastructure)

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 1 | Add `go-redis/v9` dependency                 | -          | DONE    |
| 2 | Implement Redis client initializer in `core/` | 1          | DONE    |
| 3 | Define packet codec (header + body encoding) | -          | DONE    |
| 4 | Define connection abstraction in `core/connection/` | -     | DONE    |
| 5 | Implement Redis-backed session registry      | 4          | DONE    |

### Milestone 2: SSO Token Management

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 6 | Implement `SSOStore` Redis adapter (SET+GETDEL) | 2       | DONE    |
| 7 | Add `POST /api/v1/sso` endpoint              | 6          | DONE    |
| 8 | Add `pixelsv sso` CLI command                | 6          | DONE    |
| 9 | SSO validation unit tests                    | 6          | DONE    |
| 10| SSO integration tests with Redis             | 6          | DONE    |

### Milestone 3: Handshake Packets (Pre-Auth)

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 11| Parse `release_version` (4000) C2S           | 3          | DONE    |
| 12| Parse `client_variables` (1053) C2S          | 3          | DONE    |
| 13| Parse `security.machine_id` (2490) C2S       | 3          | DONE    |
| 14| Compose `security.machine_id` (1488) S2C     | 3          | DONE    |
| 15| Machine ID validation logic (64 hex chars)   | 13, 14     | DONE    |

### Milestone 4: Authentication Flow

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 16| Parse `security.sso_ticket` (2419) C2S       | 3          | DONE    |
| 17| `AuthenticateUseCase` (validate + register)  | 5, 6, 16   | DONE    |
| 18| Duplicate login detection + kick             | 5, 17      | DONE    |
| 19| Compose `authentication.ok` (2491) S2C       | 3          | DONE    |
| 20| Compose `identity_accounts` (3523) S2C       | 3          | DONE    |
| 21| Auth timeout (30s no-ticket -> close)        | 17         | DONE    |

### Milestone 5: Session Lifecycle

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 22| Compose `client.ping` (3928) S2C             | 3          | DONE    |
| 23| Parse `client.pong` (2596) C2S               | 3          | DONE    |
| 24| Heartbeat goroutine (30s ping, 60s timeout)  | 22, 23     | PENDING |
| 25| Parse `client.disconnect` (2445) C2S         | 3          | DONE    |
| 26| `DisconnectUseCase` (cleanup + close)        | 5, 25      | PENDING |
| 27| Abrupt disconnect handler (onclose/onerror)  | 5          | PENDING |

### Milestone 6: Latency & Polish

| # | Task                                         | Depends On | Status  |
|---|----------------------------------------------|------------|---------|
| 28| Parse `client.latency_test` (295) C2S        | 3          | DONE    |
| 29| Compose `client.latency_response` (10) S2C   | 3          | DONE    |
| 30| Latency measurement handler                  | 28, 29     | PENDING |
| 31| E2E test: full handshake flow                | 17, 24     | PENDING |
| 32| E2E test: duplicate login kick               | 18         | PENDING |
| 33| E2E test: expired SSO rejection              | 17         | PENDING |

### Future Milestone: Encryption (Deferred)

| # | Task                                         | Status   |
|---|----------------------------------------------|----------|
| - | RSA key pair generation/management           | DEFERRED |
| - | DH parameter generation (128-bit primes)     | DEFERRED |
| - | RC4 stream cipher (per-connection)            | DEFERRED |
| - | Parse `init_diffie` (3110) C2S               | DEFERRED |
| - | Compose `init_diffie` (1347) S2C             | DEFERRED |
| - | Parse `complete_diffie` (773) C2S            | DEFERRED |
| - | Compose `complete_diffie` (3885) S2C         | DEFERRED |

### User System Dependency Note

Tasks 17-21 require a user ID from the SSO token but **do not require** a full
user entity or user repository. The SSO store maps `ticket -> user_id (int)`.
The `identity_accounts` packet (3523) needs `id` and `name` - this will be
stubbed with `{id: userID, name: "Player#<userID>"}` until the user realm is
implemented. The session stores the user ID as an integer, no user model needed.

---

## Caveats & Technical Notes

### Redis

- **GETDEL requires Redis 6.2+** - document as minimum version
- **Connection pooling** - `PoolSize: 20` already in config; sufficient for
  SSO validation which is once-per-connection
- **Key namespace** - all SSO keys prefixed with `sso:` to avoid collisions
- **No persistence needed** - SSO tokens are ephemeral; Redis `RDB` or `AOF`
  persistence is not required for this use case
- **Failure mode** - if Redis is down, no SSO validation is possible and
  session registry reads/writes fail; existing sockets may stay open but
  lifecycle operations that require session state cannot progress

### WebSocket

- **Fiber WebSocket** is already integrated (`core/http/module.go`) with
  `RegisterWebSocket()` and upgrade middleware
- **Binary frames** - Habbo protocol uses binary frames (not text); ensure
  `websocket.BinaryMessage` is used for reads/writes
- **Message size limit** - configure max message size on WebSocket upgrade to
  prevent memory abuse (4KB reasonable for handshake packets)
- **Close codes** - use custom close codes in 4000-4999 range:
  - `4001` - Unauthorized (invalid SSO)
  - `4002` - Duplicate login (kicked)
  - `4003` - Auth timeout
  - `4004` - Pong timeout

### Packet Codec

The Habbo protocol uses a binary format:
- **Header**: 4 bytes big-endian length + 2 bytes big-endian packet ID
- **Body**: variable-length, type-specific encoding
- **Types**: int32 (4 bytes BE), string (2 bytes BE length + UTF-8), bool (1 byte)

This codec must be implemented in `core/codec/` as a reusable library before
any packet can be parsed or composed. All vendors agree on this wire format.

### Concurrency

- **Session registry** must be thread-safe (concurrent WebSocket handlers)
- **Heartbeat** runs as a per-connection goroutine; must be cleaned up on
  disconnect to prevent leaks
- **SSO validation** is inherently safe via Redis atomicity (GETDEL)

---

## Vendor Implementation Comparison

| Aspect                    | pixels-emulator (Go) | PlusEMU (C#)      | Arcturus (Java)    | comet-v2 (Java)    |
|---------------------------|----------------------|--------------------|--------------------|--------------------|
| SSO storage               | PostgreSQL           | Database           | Database           | Database + Cache   |
| Duplicate login           | Reject (>1 result)   | Explicit kick      | Implicit dispose   | Overwrite mapping  |
| Machine ID validation     | Not implemented      | Not shown          | 64 hex + regen     | Not shown          |
| Encryption default        | N/A                  | Config option      | Off by default     | Config option      |
| DH bit size               | N/A                  | 32 bits            | 128 bits           | 128 bits           |
| Ping interval             | Not shown            | ~30s               | ~30s               | ~30s               |
| Pong timeout              | Not enforced         | Likely enforced    | Likely enforced    | 60s                |
| Disconnect packet         | conn.Dispose()       | session.Disconnect | disposeClient()    | session.disconnect |
| Ban checks at auth        | None                 | Not shown          | MAC + IP ban       | Machine ban        |
| Post-auth packet burst    | AuthOk only          | ~10 packets        | ~15 packets        | ~10 packets        |

### Our Design Choices vs Vendors

| Decision                            | Our Choice                | Rationale                                                    |
|-------------------------------------|---------------------------|--------------------------------------------------------------|
| SSO storage                         | **Redis** (not DB)        | Ephemeral tokens don't need persistence; Redis TTL is native |
| SSO validation                      | **GETDEL** (atomic)       | Stronger single-use guarantee than DB query + delete          |
| Duplicate login                     | **Explicit kick** (PlusEMU) | Clear intent, auditable, predictable behavior              |
| Encryption                          | **Deferred**              | TLS via wss:// already provides transport security           |
| Machine ID                          | **Validate + regenerate** (Arcturus) | Best security practice from vendors              |
| Auth failed response                | **Close connection**      | All vendors agree - no error packet exists                   |
