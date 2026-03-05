# Runtime Topology (Single Binary)

## Overview

`pixelsv` runs as one process and one binary. Logical service names (gateway, auth, game, social, navigator, catalog, moderation) are treated as **runtime modules/bounded contexts**, not independently deployable services.

## Module Catalogue

| Module | Responsibility |
|---|---|
| `gateway` | Client session lifecycle, packet ingress/egress |
| `auth` | Handshake, identity, session authentication |
| `game` | ECS room simulation, pathfinding, interactions |
| `social` | Friend graph, messaging, notifications |
| `navigator` | Room discovery and search models |
| `catalog` | Offers, purchases, economy orchestration |
| `moderation` | Safety actions and enforcement workflows |

## Process Topology

```
                 pixelsv (single binary)
┌──────────────────────────────────────────────────────────────┐
│ Fiber HTTP + WebSocket                                      │
│ CLI command runtime                                         │
│                                                              │
│  gateway -> auth -> game -> social/navigator/catalog/moderation
│             (all in-process through ports and contracts)     │
└──────────────────────────────────────────────────────────────┘
```

## Concurrency Model

- WebSocket ingress is handled by Fiber middleware and handed to application ports.
- Game simulation uses room-owned goroutines with fixed tick loops.
- ECS world mutation stays single-writer per room worker.
- Cross-module handoff is asynchronous but in-process by default.

## Storage and IO

- PostgreSQL and Redis are external adapters.
- Messaging is internal first; optional external broker is an adapter decision, not a domain rule.
- No domain package imports DB/cache/transport libraries.

## Failure Boundaries

- A module panic must be contained and surfaced through supervisor/restart policy.
- Backpressure must be explicit on queues/channels.
- Slow IO must fail fast with context deadlines at adapter edges.

## Notes on Legacy Wording

Older planning docs may still use "service" or "NATS subject" terminology. In this architecture phase, those references map to internal modules and in-process contract topics unless explicitly marked as external integration.
