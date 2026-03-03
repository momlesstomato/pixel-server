# Service Topology

## Overview

pixel-server is decomposed into focused services. Each service has one primary responsibility, one database schema prefix, and one NATS subject namespace. Services do **not** call each other synchronously (no gRPC in the hot path); all cross-service communication is through NATS JetStream events.

---

## Service catalogue

| Service | Role | Scales with |
|---|---|---|
| `gateway` | WebSocket accept + packet framing | Connected clients |
| `auth` | Handshake, SSO, session lifecycle | Login rate |
| `game` | Room simulation, ECS, pathfinding | Active rooms |
| `social` | Friends, messenger, notifications | User count |
| `navigator` | Room discovery, search | Read QPS |
| `catalog` | Store, purchases, economy | Purchase QPS |
| `moderation` | Bans, tickets, mod-tool | Moderation load |

---

## Runtime topology

### Development (single host)

```
docker-compose up
  ├── nats          (nats:2 JetStream)
  ├── postgres      (postgres:16-alpine)
  ├── redis         (redis:7-alpine)
  ├── gateway       (port 2096 / 443 TLS)
  ├── auth
  ├── game
  ├── social
  ├── navigator
  ├── catalog
  └── moderation
```

All services share one NATS cluster but are fully isolated as Go binaries.

### Production (Kubernetes)

Each service is a separate `Deployment`. Horizontal pod autoscalers scale on:
- `gateway` – active WebSocket connection count
- `game` – active room count per pod
- `auth`, `navigator` – RPS

`game` uses consistent hashing (by `roomID`) to ensure all sessions in the same room route to the same pod. Redis is the source of truth for which `game` pod currently owns a room.

---

## Connection flow (new client)

```
1. Client opens WebSocket to gateway.
2. Gateway assigns sessionID (UUIDv7), stores in Redis with TTL=60s.
3. Gateway reads first packets; all go to NATS subject:
        handshake.c2s.<sessionID>
4. auth-svc consumes, drives Diffie-Hellman + SSO handshake.
5. auth-svc publishes:
        session.authenticated { sessionID, userID, username, figures, … }
6. gateway subscribes to session.authenticated, updates Redis session,
   begins forwarding all subsequent packets to:
        room.input.<roomID>   (after NAVIGATE_TO_HOTEL_VIEW or room entry)
7. game-svc consumes room.input.<roomID>.
8. game-svc publishes outbound per client to:
        session.output.<sessionID>
9. gateway subscribes to session.output.<sessionID>, writes frames to socket.
```

---

## NATS subject hierarchy

```
handshake.c2s.<sessionID>           ← gateway → auth  (pre-auth packets)
session.authenticated               ← auth → all      (broadcast on login)
session.disconnected                ← gateway → all   (cleanup)

room.input.<roomID>                 ← gateway → game  (after auth)
session.output.<sessionID>          ← game → gateway  (outbound frames)

social.friend_request.<userID>      ← game,catalog → social
social.notification.<userID>        ← social → gateway
social.message.<conversationID>     ← client → social → delivery

navigator.room_updated.<roomID>     ← game → navigator  (metadata sync)

catalog.purchase_completed          ← catalog → inventory,game
inventory.updated.<userID>          ← inventory → game (item grant)
```

Full subject reference in [007-messaging.md](007-messaging.md).

---

## Gateway design

The gateway is the only service directly exposed to the internet.

### Connection handling

Uses `gobwas/ws` (zero-allocation WebSocket) with a `netpoll`/`epoll`-based event loop (via `mailru/easygo` or a direct epoll wrapper):

```
for {
    events := epoll.Wait()
    for _, event := range events {
        if event.readable {
            // Read one packet frame; dispatch to NATS
        }
        if event.writable {
            // Flush pending outbound queue for this connection
        }
    }
}
```

This avoids one goroutine per connection. At 10 k concurrent connections a goroutine-per-conn model wastes ~80 MB just on stack space; epoll multiplexing stays flat.

### Session lifecycle in Redis

```
HSET session:<id>  userID <id>  roomID <id>  gameNode <podName>  encKey <hex>
EXPIRE session:<id> 3600
```

On disconnect, gateway publishes `session.disconnected` and deletes the Redis key.

---

## Game service — room worker pool

The `game` service maintains a pool of room goroutines:

```
supervisor
  ├── roomWorker{id: 1, world: arche.World}
  ├── roomWorker{id: 2, world: arche.World}
  └── …
```

Each `roomWorker` runs in its own goroutine with a `chan Envelope` for inbound messages. The room is loaded from PostgreSQL on first entry; disposed 30 seconds after the last player leaves. This replaces the `scheduleAtFixedRate(500 ms)` per-room scheduler with a **single select loop**:

```go
func (w *roomWorker) run() {
    ticker := time.NewTicker(50 * time.Millisecond) // 20 Hz
    defer ticker.Stop()
    for {
        select {
        case msg := <-w.inbox:
            w.handleMessage(msg)      // synchronous, no concurrency issues
        case <-ticker.C:
            w.tick()                  // ECS systems run here
        case <-w.shutdown:
            w.dispose()
            return
        }
    }
}
```

No `ConcurrentMap`, no `synchronized`, no lock contention inside the simulation. All writes to ECS world happen on one goroutine.

### Backpressure

`w.inbox` is a buffered channel with configurable capacity (default 256). If the buffer fills (i.e. the simulation tick is slower than inbound message rate), **the gateway drops packets with a rate-limit error** rather than blocking the WebSocket handler. Legitimate clients never send faster than the simulation can process under normal conditions.

---

## Auth service — Diffie-Hellman + SSO

Replicates the handshake sequence confirmed in the spec:

1. Client sends `handshake.init_diffie` (c2s id 3110).
2. Auth generates RSA-signed DH parameters, responds with `handshake.init_diffie` (s2c id 1347).
3. Client sends encrypted public key via `handshake.complete_diffie` (c2s id 773).
4. Auth computes shared secret, installs RC4 on the session's gateway-side ciphers, responds with `handshake.complete_diffie` (s2c id 3885).
5. Client sends `security.sso_ticket` (c2s id 2419); auth validates, resolves user.
6. Auth publishes `session.authenticated`.

All session keys are stored in Redis (not in-process state), so any auth pod can handle the subsequent steps.

---

## Inter-service failures

| Scenario | Mitigation |
|---|---|
| `game` pod dies mid-room | Redis key expires, clients reconnect, new pod loads room from DB |
| NATS temporarily unavailable | Gateway buffers up to N frames per session; drops if buffer full |
| PostgreSQL slow | game/auth use connection pools (pgx) with context deadlines; slow reads surface as room load failures, not stalls |
| auth pod overloaded | NATS consumer group distributes handshake work across multiple auth pods |
