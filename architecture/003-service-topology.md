# Runtime Topology (Multi-Role Single Binary)

## Overview

`pixelsv` runs as one binary. Logical service names (gateway, auth, game, social, navigator, catalog, moderation) are **realm modules** inside `internal/`. Each can be activated independently via the `--role` flag for failure isolation and horizontal scaling.

## Deployment Modes

### All-in-One (Development / Small Deployment)

```
pixelsv serve
```

One process runs all realm modules. Inter-module communication is in-process (Go channels, direct function calls). No external message broker required.

```
                 pixelsv (single process)
┌──────────────────────────────────────────────────────────┐
│ Fiber HTTP + WebSocket                                   │
│                                                          │
│  gateway ←→ auth ←→ game ←→ social/navigator/catalog/mod│
│         (all in-process through ports and channels)      │
└──────────────────────────────────────────────────────────┘
```

### Distributed (Production / High Load)

```
pixelsv serve --role=gateway
pixelsv serve --role=game
pixelsv serve --role=game          # second game worker instance
pixelsv serve --role=social,navigator,catalog,moderation
pixelsv serve --role=api
pixelsv serve --role=jobs
```

Multiple processes of the same binary, each running specific roles. Inter-module communication uses NATS with the same contract topics. Room distribution uses consistent hashing or registry-based assignment.

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ pixelsv      │  │ pixelsv      │  │ pixelsv      │
│ --role=gw    │  │ --role=game  │  │ --role=game  │
│ WS sessions  │  │ rooms 1-500  │  │ rooms 501+   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                  │
       └─────────┬───────┴──────────────────┘
                 │
         ┌───────┴───────┐
         │     NATS      │
         └───────┬───────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
┌───┴──────┐ ┌──┴─────────┐ ┌┴────────────┐
│ social+  │ │ api        │ │ jobs        │
│ nav+cat+ │ │ REST admin │ │ maintenance │
│ mod      │ │ Swagger    │ │ partitions  │
└──────────┘ └────────────┘ └─────────────┘
```

## Module Catalogue

| Module | Role Flag | Responsibility | Transport In | Transport Out |
|---|---|---|---|---|
| `gateway` | `gateway` | WebSocket session lifecycle, packet ingress/egress | WS frames | packet.c2s.<realm>.<sessionID>, session.disconnected |
| `auth` | `auth` | Handshake, identity, SSO ticket validation | packet.c2s.handshake-security.* | session.authenticated, session.output |
| `game` | `game` | ECS room workers, 20 Hz tick, pathfinding | room.input | session.output (broadcasts) |
| `social` | `social` | Friend graph, messaging, notifications | session.authenticated | session.output |
| `navigator` | `navigator` | Room discovery and search | navigator topics | session.output |
| `catalog` | `catalog` | Offers, purchases, economy | catalog topics | session.output |
| `moderation` | `moderation` | Bans, tickets, safety | moderation topics | moderation.ban.issued |
| `api` | `api` | REST admin, Swagger UI, OpenAPI | HTTP requests | HTTP responses |
| `jobs` | `jobs` | Partition creation, leaderboard refresh | Timer-driven | DB writes |

## Transport Adapter Pattern

Realm modules communicate through **ports** (Go interfaces). The transport implementation behind those ports changes based on deployment mode:

```go
// Defined by the consumer (e.g. game realm)
type SessionWriter interface {
    Send(sessionID string, data []byte) error
}

type EventPublisher interface {
    Publish(topic string, payload any) error
}
```

| Deployment Mode | SessionWriter Implementation | EventPublisher Implementation |
|---|---|---|
| All-in-one | Direct WebSocket write via pointer | In-process channel bus |
| Distributed | NATS publish to session.output.{id} | NATS publish to topic |

The domain code is identical in both modes. Only `cmd/pixelsv` wiring changes.

## Room Worker Distribution (Distributed Mode)

When multiple `--role=game` instances run, rooms must be assigned to specific instances.

### Option A: Consistent Hashing (Default)

Gateway maintains a hash ring of known game instances. Room ID hashes determine the owning game worker. When instances join/leave, affected rooms migrate.

### Option B: Registry-Based

Each game instance registers its rooms in Redis (`HSET room:owner <roomID> <instanceID>`). Gateway looks up the owning instance before routing. More control, slightly more latency.

### Room State Recovery

When a game instance crashes and restarts:
1. Process manager restarts the `pixelsv --role=game` process.
2. On startup, the game worker checks Redis for ark-serde ECS snapshots of its assigned rooms.
3. Rooms with active sessions are restored from snapshot. Rooms with zero sessions are lazily re-created on next entry.

## Concurrency Model

- WebSocket ingress is handled by Fiber middleware and handed to application ports.
- Gateway decodes binary packets before publish to packet realm topics.
- Game simulation uses room-owned goroutines with fixed 20 Hz tick loops.
- ECS world mutation stays single-writer per room worker.
- Cross-module handoff is via typed Go interfaces — in-process when co-located, NATS when distributed.

## Storage and IO

- PostgreSQL and Redis are external adapters.
- In all-in-one mode, realm modules share the same pgx pool and Redis client.
- In distributed mode, each process creates its own connections (only the roles it needs).
- No domain package imports DB/cache/transport libraries.

## Failure Boundaries

- A module panic must be contained and surfaced through supervisor/restart policy.
- In distributed mode, a crashed role process does not affect other roles.
- Backpressure must be explicit on queues/channels.
- Slow IO must fail fast with context deadlines at adapter edges.

## Configuration

```env
# Role selection
PIXELSV_ROLE=all                    # or: gateway, game, social, api, jobs, etc.
PIXELSV_INSTANCE_ID=game-worker-1   # unique per process in distributed mode

# Infrastructure (shared)
POSTGRES_URL=postgres://...
REDIS_URL=redis://...
NATS_URL=nats://nats:4222           # only needed in distributed mode

# HTTP
HTTP_ADDR=:8080
API_KEY=change-me-in-production

# Logging
LOG_FORMAT=json
LOG_LEVEL=info
```

Roles that don't need a dependency skip its initialization:
- `gateway` doesn't connect to PostgreSQL.
- `catalog` doesn't connect to Redis.
- All-in-one mode doesn't connect to NATS.
