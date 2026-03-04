# Realm: Handshake & Security

> **Position:** 10 | **Phase:** 1 (Connection) | **Packets:** 13 (8 c2s, 5 s2c)
> **Services:** gateway, auth | **Status:** Partially implemented

---

## Overview

The Handshake & Security realm governs everything that happens between WebSocket connection open and the moment a session becomes authenticated. It covers encryption negotiation (Diffie-Hellman + RC4), client identification (release version, machine ID), and SSO token validation. This is the first realm a connecting client interacts with and the most security-critical.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 1

---

## Packet Inventory

### C2S (Client to Server) -- 8 packets

| ID | Name | Phase | Fields | Summary |
|----|------|-------|--------|---------|
| 4000 | `handshake.release_version` | pre-auth | `releaseVersion:string`, `clientType:string`, `platform:int32`, `deviceCategory:int32` | Advertise Nitro release and client platform metadata |
| 1053 | `handshake.client_variables` | pre-auth | `clientId:int32`, `clientUrl:string`, `externalVariablesUrl:string` | Send client resource metadata (usually discarded) |
| 3110 | `handshake.init_diffie` | pre-auth | _(none)_ | Request server DH parameters for encryption setup |
| 773 | `handshake.complete_diffie` | crypto | `publicKey:string` | Send client DH public key to complete key exchange |
| 2419 | `security.sso_ticket` | auth | `ticket:string`, `elapsedTime:int32` | Authenticate with SSO token |
| 2490 | `security.machine_id` | pre-auth | `machineId:string`, `fingerprint:string`, `flashVersion:string` | Client device fingerprint for session tracking |
| 96 | `handshake.client_latency_measure` | pre-auth | _(none)_ | Initial latency measurement before auth |
| 26979 | `handshake.client_policy` | pre-auth | _(none)_ | Client security policy request |

### S2C (Server to Client) -- 5 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1347 | `handshake.init_diffie` | `signedPrime:string`, `signedGenerator:string` | RSA-signed DH prime and generator |
| 3885 | `handshake.complete_diffie` | `publicKey:string`, `serverEncryption:boolean` | Server DH public key + encryption flag |
| 1488 | `security.machine_id` | `machineId:string` | Echo machine ID back to client |
| 2491 | `authentication.ok` | _(none)_ | Authentication succeeded |
| 3523 | `handshake.identity_accounts` | _(none)_ | Identity/account data (multi-account support stub) |

---

## Architecture Mapping

### Service Ownership

```
Client ──WebSocket──▶ Gateway ──NATS(handshake.c2s.<sid>)──▶ Auth Service
                         ◀──NATS(session.output.<sid>)───────┘
                         ◀──NATS(session.authenticated)──────┘
```

- **Gateway** handles WebSocket accept, frame parsing, and initial routing. Packets with IDs 4000, 1053, 3110, 773, 2419, 2490 are forwarded to `handshake.c2s.<sessionID>` via NATS.
- **Auth service** consumes handshake NATS subjects, performs DH key exchange, validates SSO tokens against Redis, and publishes `session.authenticated` on success.
- **Gateway** listens for `session.authenticated` and promotes the session to full packet routing.

### Database Tables

| Table | Usage |
|-------|-------|
| `users` | SSO ticket resolves to `user_id` via Redis lookup, then user row fetched |
| `bans` | Checked during SSO validation to reject banned accounts |

### Redis Keys

| Key Pattern | Usage |
|-------------|-------|
| `sso:<ticket>` | Maps SSO ticket to user ID (TTL ~30s, single-use) |
| `session:<sessionID>` | Stores session metadata post-auth |
| `ban:<userID>` | Quick ban check during authentication |

### NATS Subjects

| Subject | Direction | Consumer |
|---------|-----------|----------|
| `handshake.c2s.<sessionID>` | gateway -> auth | auth (work-queue) |
| `session.output.<sessionID>` | auth -> gateway | gateway (fan-out) |
| `session.authenticated` | auth -> all | gateway, game, social (fan-out) |

---

## Implementation Analysis

### Connection Flow (Step-by-Step)

1. **WebSocket Open** -- Gateway accepts connection, allocates UUIDv7 session ID, starts read loop.
2. **`handshake.release_version` (4000)** -- Client sends build string (e.g., `NITRO-1-6-6`). Gateway can reject incompatible versions. Most implementations read and discard.
3. **`handshake.client_variables` (1053)** -- Client resource URLs. Safely discarded by server.
4. **`handshake.init_diffie` (3110)** -- Client requests DH parameters. Auth service generates prime/generator, signs with RSA private key, responds with `handshake.init_diffie` (1347).
5. **`handshake.complete_diffie` (773)** -- Client sends DH public key. Auth computes shared secret, derives RC4 key. Responds with `handshake.complete_diffie` (3885). Gateway installs RC4 cipher on the session if `serverEncryption=true`.
6. **`security.machine_id` (2490)** -- Client fingerprint stored on session for tracking. Server echoes back with `security.machine_id` (1488).
7. **`security.sso_ticket` (2419)** -- Client sends SSO token. Auth validates against Redis (single-use, TTL). On success: publishes `session.authenticated`, sends `authentication.ok` (2491) + `handshake.identity_accounts` (3523).
8. **Session promoted** -- Gateway begins routing all subsequent packets to game service via `room.input.<sessionID>`.

### Diffie-Hellman Implementation

The DH key exchange uses RSA-signed parameters to prevent MITM attacks:

```
Server: Generate DH prime (p), generator (g)
Server: Sign p and g with RSA private key
Server → Client: signed_p, signed_g
Client: Verify signatures with RSA public key
Client: Generate DH keypair (a, g^a mod p)
Client → Server: g^a mod p
Server: Generate DH keypair (b, g^b mod p)
Server → Client: g^b mod p
Both: Compute shared secret = (g^ab mod p)
Both: Derive RC4 key from shared secret
```

**Implementation note:** The DH parameters and RSA keys should be loaded from configuration, not hardcoded. The RSA key pair is typically 1024-bit for compatibility with Nitro client expectations.

### SSO Ticket Validation

```
1. Auth receives security.sso_ticket
2. Redis GET sso:<ticket>
3. If not found → send connection.error, disconnect
4. Redis DEL sso:<ticket> (single-use enforcement)
5. PostgreSQL: SELECT * FROM users WHERE id = <resolvedUserID>
6. PostgreSQL: SELECT * FROM bans WHERE user_id = <resolvedUserID> AND expires_at > NOW()
7. If banned → send disconnect_reason, close socket
8. Update users SET last_login = NOW(), ip_current = <clientIP>
9. Publish session.authenticated with user data
10. Send authentication.ok (2491)
```

---

## Caveats & Edge Cases

### 1. Race Condition: Double SSO Ticket Use
If a user opens two browser tabs simultaneously, both may attempt SSO validation with the same ticket. The Redis `DEL` must be atomic -- use `GETDEL` (Redis 6.2+) instead of separate GET + DEL to prevent race conditions.

### 2. DH Parameter Caching
Generating DH primes is computationally expensive. Pre-generate a pool of DH parameter sets at startup and rotate through them. Reference emulators (Comet v2, Arcturus) hardcode a single prime/generator pair, which is less secure but simpler.

### 3. RC4 Encryption Optional
The `serverEncryption` boolean in `handshake.complete_diffie` (3885) controls whether RC4 is activated. Many deployments run without encryption (Nitro over WSS/TLS makes RC4 redundant). pixel-server should support both modes via configuration.

### 4. Client Version Gating
`handshake.release_version` provides the client build string. While most emulators discard it, pixel-server should:
- Log the version for debugging.
- Optionally reject unknown versions via a configurable allowlist.
- Never reject silently -- send `connection.error` with a reason code.

### 5. Machine ID Abuse Prevention
The `security.machine_id` fingerprint can be spoofed by clients. It should be used as a supplementary tracking signal (e.g., for linking alt accounts) but never as the sole basis for security decisions like IP bans.

### 6. Timeout Between Handshake Steps
If a client sends `handshake.init_diffie` but never sends `handshake.complete_diffie`, the session leaks resources. Implement a handshake timeout (e.g., 15 seconds from connection open to `authentication.ok`) that forcefully disconnects incomplete sessions.

### 7. Gateway-Auth NATS Latency
The handshake round-trips (gateway -> NATS -> auth -> NATS -> gateway) add latency. For the SSO validation path, this is acceptable (one-time cost). However, the DH exchange involves two round-trips. If NATS latency is high, the total handshake time may feel slow to users. Monitor p99 handshake duration.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **SSO validation** | Synchronous DB lookup in connection thread | Async via NATS; gateway never blocks on DB |
| **DH parameters** | Single hardcoded prime | Configurable, rotatable parameter pool |
| **Encryption** | Always RC4 or always disabled | Per-deployment toggle; RC4 redundant with WSS |
| **Session tracking** | In-memory map with no expiry | Redis-backed with TTL; survives gateway restart |
| **Ban checking** | DB query per connection | Redis cache with PUB/SUB invalidation |
| **Multi-tab handling** | Undefined (often crashes) | Atomic `GETDEL` prevents double SSO use |
| **Handshake timeout** | None (leaked sessions) | Configurable timeout with forced disconnect |

---

## Dependencies

- **pkg/core/codec** -- Reader/Writer for packet serialization
- **pkg/core/bus** -- NATS publish/subscribe for gateway <-> auth communication
- **pkg/protocol** -- Generated packet structs (handshake_security.go)
- **Redis** -- SSO ticket store, session metadata, ban cache
- **PostgreSQL** -- User lookup, ban table
- **RSA key pair** -- For signing DH parameters (loaded from config/secret)

---

## Testing Strategy

### Unit Tests
- Codec round-trip for all 13 packet types
- SSO ticket validation logic (mock Redis)
- DH key exchange computation (deterministic test vectors)
- Handshake state machine transitions

### Integration Tests
- Full handshake flow against real Redis + PostgreSQL (testcontainers)
- SSO ticket expiry and single-use enforcement
- Ban rejection during authentication
- Concurrent SSO validation with same ticket

### E2E Tests
- Mock WebSocket client connects, completes full handshake, receives `authentication.ok`
- Client with invalid SSO ticket is disconnected with correct error code
- Banned user cannot authenticate
