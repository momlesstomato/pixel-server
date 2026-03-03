# pixel-server — Architecture Overview

## Vision

`pixel-server` is a next-generation, cloud-native implementation of a Habbo-style hotel server ("emulator"). It replaces the monolithic, singleton-heavy JVM/CLR paradigm with a modular Go workspace composed of narrow, independently deployable services coordinated through a shared message bus.

The design goals are:  
- **Correctness first** – every packet from the 922-entry `pixel-protocol` spec is derived from verified sources (Nitro renderer + reference servers).  
- **Horizontal scalability** – room simulation, gateway, and social layers scale independently.  
- **Determinism** – game simulation uses a fixed-tick loop decoupled from I/O; pathfinding and ECS state transitions are pure functions testable without a running server.  
- **Observability** – structured logging, distributed tracing (OpenTelemetry), and metrics out of the box.  
- **Modern storage** – PostgreSQL 16 for persistence, Redis 7 for ephemeral session and pub/sub.  

---

## Why the existing emulators fall short

| Caveat | Root cause | Impact |
|---|---|---|
| God-object `Emulator` singleton | All subsystems wired through one static class | Impossible to test in isolation; deployment is always the full monolith |
| Per-room `scheduleAtFixedRate(500 ms)` | Each `Room implements Runnable` is independently scheduled | Thread pool exhaustion at scale; jitter between rooms; no backpressure |
| Synchronous JDBC on cycle threads | DB calls inside `Room#cycle()` | A slow query pauses the entire room tick; GC pressure from unbounded blocking |
| `ConcurrentSet` / `THashMap` everywhere | Optimistic concurrency with no domain ownership | Race conditions on shared entity state; hard to reason about visibility |
| 2-D only pathfinding | `RoomLayout.findPath` operates on `(x, y)` pairs; Z checked only as a hard step cut-off | Cannot represent multi-level rooms, staircase traversal costs, or flying entities |
| Message handling via reflection | `MessageHandlerManager` resolves handlers by header ID with dynamic dispatch | Cannot generate typed router code statically; header collisions silently drop packets |
| Hardcoded worker counts | `HabboExecutorService` wraps a fixed-size pool | Cannot tune per-environment; no work-stealing for asymmetric loads |
| No protocol contract | Packet shapes live only in handler code | Protocol drift between client and server; no shared schema to generate from |

---

## Technology choices

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.23+ | Goroutine-per-connection model fits WebSocket fan-out; excellent concurrency primitives; fast compile times; workspace support for monorepo |
| ECS framework | `mlange-42/ark` (Ark ECS) v0.7.1 | Archetype-based; cache-friendly; entity relationships; event system; zero deps; `ark-serde` for world serialization |
| Pathfinding | Custom 3D A* + HPA* layering | Hierarchical pre-computation for large rooms; JPS for open-floor optimisation |
| WebSocket | `gobwas/ws` (zero-alloc) | Avoids gorilla/websocket allocations on every frame; pairs with `epoll`-based multiplexing |
| Message bus | NATS JetStream | Persistent, exactly-once delivery between services; fan-out to room workers |
| Persistent storage | PostgreSQL 16 + `pgx/v5` | Native Go driver; pipeline mode; prepared statements; row streaming for large result sets |
| Cache / ephemeral | Redis 7 (Valkey) | Session tokens, rate limits, pub/sub for presence and chat replication |
| Code generation | Custom `protogen` (extends `generate-artifacts.mjs` model) | Generates Go structs + encode/decode + handler stubs from `spec/protocol.yaml` |
| Observability | OpenTelemetry → Grafana / Loki / Tempo | Structured logs, metrics, traces in one pipeline |
| Container | Docker + Compose for dev; Kubernetes manifests for prod | Services scale independently |

---

## High-level component map

```
┌─────────────────────────────────────────────────────────────────┐
│                          Clients (Nitro)                         │
└───────────────────────────────┬─────────────────────────────────┘
                                │ WebSocket  (uint32 len + uint16 id + payload)
                                ▼
                    ┌──────────────────────┐
                    │     gateway svc      │   TLS termination
                    │  (gobwas/ws, epoll)  │   packet framing
                    └──────────┬───────────┘   session attach
                               │ NATS JetStream
          ┌────────────────────┼───────────────────────────────┐
          ▼                    ▼                               ▼
  ┌───────────────┐  ┌─────────────────────┐        ┌─────────────────┐
  │   auth svc    │  │    game-core svc     │        │   social svc    │
  │ SSO / Diffie  │  │  room worker pool   │        │ friends/chat    │
  │ session mgmt  │  │  ECS (Ark) world    │        │ invitations     │
  └───────┬───────┘  │  3D A* pathfinder   │        └────────┬────────┘
          │          │  item interactions  │                 │
          │          └──────────┬──────────┘                 │
          │                     │                             │
          └──────────┬──────────┘─────────────────────────────┘
                     │
           ┌─────────┴─────────┐
           │    Redis  7        │   sessions, rate-limits,
           │  (session + PS)    │   room presence, chat fan-out
           └─────────┬─────────┘
                     │
           ┌─────────┴─────────┐
           │  PostgreSQL 16     │   all durable state
           └───────────────────┘
```

Additional services (thinner, can be rolled out in later phases):

- `navigator-svc` – room search, favourites, home room  
- `catalog-svc` – store pages, purchases, subscriptions  
- `inventory-svc` – item ownership, badge grants  
- `moderation-svc` – ban management, ticket queue  

---

## Repository layout (Go workspace)

```
pixel-server/
├── go.work                         ← Go workspace root
├── architecture/                   ← this folder
├── vendor/                         ← upstream references (read-only)
├── services/
│   ├── gateway/                    ← go.mod: pixel-server/gateway
│   ├── auth/                       ← go.mod: pixel-server/auth
│   ├── game/                       ← go.mod: pixel-server/game
│   ├── social/                     ← go.mod: pixel-server/social
│   ├── navigator/                  ← go.mod: pixel-server/navigator
│   ├── catalog/                    ← go.mod: pixel-server/catalog
│   └── moderation/                 ← go.mod: pixel-server/moderation
├── pkg/
│   ├── protocol/                   ← go.mod: pixel-server/protocol  (generated)
│   ├── codec/                      ← go.mod: pixel-server/codec
│   ├── ecs/                        ← go.mod: pixel-server/ecs
│   ├── pathfinding/                ← go.mod: pixel-server/pathfinding
│   ├── storage/                    ← go.mod: pixel-server/storage
│   └── bus/                        ← go.mod: pixel-server/bus
└── tools/
    └── protogen/                   ← go.mod: pixel-server/tools/protogen
        └── main.go
```

See [001-go-workspace.md](001-go-workspace.md) for `go.work` and per-module details.

---

## Guiding principles

1. **No globals, no singletons.** Every subsystem is passed by interface at construction time.  
2. **Domain owns its state.** A room's ECS world is owned exclusively by its game-worker goroutine. No external goroutine writes to it directly; all external input arrives via a channel.  
3. **Generated beats hand-written.** All packet encode/decode and handler stub code is generated from `spec/protocol.yaml`. Manual edits in generated files are CI-blocking errors.  
4. **Fixed-tick simulation, async I/O.** The game loop runs at 20 Hz (50 ms tick). I/O is non-blocking; goroutines for network reads/writes are separated from the simulation goroutine.  
5. **Test the pure core.** Pathfinding, ECS queries, codec round-trips, and WIRED logic are unit-tested without a database or network.  

---

## Document index

| File | Topic |
|---|---|
| [001-go-workspace.md](001-go-workspace.md) | Go workspace layout, module boundaries, `go.work` |
| [002-protocol-codegen.md](002-protocol-codegen.md) | YAML spec → Go code generation |
| [003-service-topology.md](003-service-topology.md) | Service decomposition, NATS subjects, scaling |
| [004-ecs-ark.md](004-ecs-ark.md) | ECS evaluation, Ark v0.7.1 integration in game-core |
| [005-pathfinding-3d.md](005-pathfinding-3d.md) | 3D A* design, HPA* layering, performance |
| [006-storage.md](006-storage.md) | PostgreSQL schema strategy, Redis patterns |
| [007-messaging.md](007-messaging.md) | NATS JetStream subjects, event contracts |
| [008-patterns.md](008-patterns.md) | Hexagonal architecture, DDD boundaries, TDD scope |
| [009-packet-roadmap.md](009-packet-roadmap.md) | Phase-by-phase 922-packet implementation order |
| [011-plugin-system.md](011-plugin-system.md) | ECS-aligned plugin lifecycle, events, interceptors |
