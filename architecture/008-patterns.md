# Architectural Patterns

## What to use and what to avoid

This document examines common patterns (DDD, TDD, Hexagonal, CQRS, Event Sourcing, Clean Architecture) and gives a concrete verdict on each for pixel-server. The rule is: **adopt a pattern only when it solves a real problem present in this codebase** — do not adopt patterns as prestige signals.

---

## Hexagonal Architecture (Ports & Adapters) — ADOPT

### Why it fits

The central problem in legacy emulators is that business logic (room simulation, pathfinding, item interaction) is directly coupled to:
- JDBC/MySQL calls inside cycle() methods.
- Static `Emulator.*` accessor calls everywhere.
- Direct Netty/Mina channel writes inside message handlers.

Hexagonal architecture severs these couplings by wrapping infrastructure behind interfaces (ports) that the domain core depends on:

```
            ┌──────────────────────────┐
            │        Domain            │
            │  (room simulation, ECS,  │
            │   pathfinding, WIRED)    │
            │                          │
            │  depends on →            │
            │    RoomRepository (port) │
            │    EventPublisher (port) │
            │    Clock (port)          │
            └──────────┬───────────────┘
                       │ implemented by
          ┌────────────┼────────────────┐
          ▼            ▼                ▼
   pgx adapter    NATS adapter    time.Now() adapter
```

### Implementation in pixel-server

In practice this means every service's `internal/` package contains:

```
internal/
  domain/           ← pure business logic; no framework imports
    room/
      room.go       ← Room struct, tick(), handleInput()
      repository.go ← interface RoomRepository
    user/
      user.go
      repository.go
  application/      ← use-case functions; orchestrates domain + ports
    room_entry.go
    purchase.go
  adapters/
    postgres/       ← implements domain repository interfaces
    nats/           ← implements EventPublisher
    redis/          ← implements SessionStore
  cmd/
    main.go         ← wires adapters into application
```

The `domain` package has **no** imports from `adapters`, `pgx`, `nats`, or any framework. It can be tested with in-memory fakes.

### What this is NOT

Hexagonal architecture does NOT require:
- A mandatory `usecase`, `presenter`, `interactor` layer with separate structs for each operation.
- Mapping DTOs at every boundary (only at genuine seams: HTTP/NATS boundary, DB boundary).
- Strict "one interface per repository method" (a single `RoomRepository` interface with 5 methods is fine).

---

## Domain-Driven Design (DDD) — ADOPT SELECTIVELY

### Aggregate boundaries

DDD's most useful concept for this project is the **aggregate** — a cluster of entities that is always read/written as a whole, with one root enforcing invariants.

Correct aggregate roots for pixel-server:

| Aggregate root | Members | Invariant enforced |
|---|---|---|
| `Room` | RoomUnit list, item grid, layout | Max occupancy; item stacking rules |
| `User` | Profile, stats, currency | Non-negative credits |
| `Friendship` | User A, User B, status | No duplicate friendships |
| `Trade` | Two users, two item sets, accept flags | Items owned by participants |

Incorrect use of DDD (bloat to avoid):
- A `RoomEntry` value object with a `RoomEntryFactory` — this is just a function call.
- `DomainEvent` base classes with generic `DomainEventPublisher<T extends DomainEvent>` — use typed NATS events directly.
- Separate `RoomDTO` / `RoomCommand` / `RoomQuery` structs when the Go struct itself is already sufficient.
- Repository methods per aggregate field accessor (`FindRoomByOwnerID`, `FindRoomByName`, `FindRoomByCategory` — fine to have on the interface, but do not create separate repositories).

### Bounded Contexts

The service boundaries in [003-service-topology.md](003-service-topology.md) are bounded contexts. They do not share database schemas. Cross-context references use IDs only (e.g., `game-svc` knows the `userID` of an avatar but does not embed a full `User` struct from auth-svc's domain).

---

## Test-Driven Development (TDD) — ADOPT FOR PURE LOGIC

TDD is valuable where the function is determinate and testable in isolation. It becomes friction when the test setup requires a running database, network, or 500-line fake.

### Apply TDD to

- `pkg/pathfinding` — path correctness is a pure function of heightmap + start + goal.
- `pkg/codec` — encode/decode round-trip is deterministic.
- `pkg/protocol` (generated code) — generator output is tested with golden files.
- ECS systems (MovementSystem, ChatCooldownSystem, etc.) — tick logic is a pure transformation of world state.
- WIRED logic engine — condition/effect evaluation is a pure function.
- Rate limiting (Redis Lua scripts can be tested against a mock pipeline).

### Do not apply TDD to

- Database adapter code — test with integration tests against a real PostgreSQL container (testcontainers-go).
- NATS consumer/producer code — test with a real embedded NATS server.
- The gateway's epoll loop — test with load/integration tests, not unit tests.
- Service wiring (main.go) — no business logic lives here.

---

## CQRS — SKIP

Command-Query Responsibility Segregation (separate read/write models) is not warranted:

- The read load is handled by PostgreSQL index-optimised queries + Redis caches.
- There is no separate reporting or analytics database.
- The complexity cost (dual model sync, eventual consistency for reads) outweighs the benefit at this scale.

**Exception:** The navigator service is read-heavy and caches room listings in Redis. This is a lightweight variant of CQRS without a fully separate model — appropriate here.

---

## Event Sourcing — SKIP

Event sourcing (storing events as the source of truth instead of current state) would complicate item placement, currency changes, and user state. The operational overhead (event store, projection rebuilds, snapshotting) is not justified. PostgreSQL tables with `created_at` columns and `chat_log` partitioned tables provide sufficient audit trail.

---

## Clean Architecture — SKIP (use Hexagonal instead)

"Clean Architecture" as described by Uncle Bob adds concentric layers (`Entities`, `Use Cases`, `Interface Adapters`, `Frameworks & Drivers`) that map almost identically to Hexagonal for this use case, but with more mandatory ceremony. Hexagonal achieves the same isolation with fewer required files.

---

## Patterns summary

| Pattern | Verdict | Scope |
|---|---|---|
| Hexagonal (Ports & Adapters) | **ADOPT** | All services |
| DDD Aggregates | **ADOPT** for Room, User, Trade | game, social, catalog |
| DDD Bounded Contexts | **ADOPT** | service boundaries |
| TDD | **ADOPT** | pkg/pathfinding, pkg/codec, ECS systems, WIRED |
| Integration tests | **ADOPT** | DB adapters, NATS adapters |
| CQRS | **SKIP** | — |
| Event Sourcing | **SKIP** | — |
| Clean Architecture | **SKIP** (Hexagonal is enough) | — |

---

## Go-specific conventions

- **Interfaces belong to the consumer, not the implementer.** Define `RoomRepository` in `domain/room/`, not in `adapters/postgres/`.
- **Constructors are functions, not factories.** `room.New(id, layout, opts)` is enough; no `RoomFactory` struct.
- **Errors are values.** Use `fmt.Errorf("room %d: %w", id, err)` consistently. Define domain error types (`ErrRoomFull`, `ErrNotAuthenticated`) as exported `errors.New()` vars.
- **No init().** All initialisation happens at explicit call sites in `main.go`.
- **Context propagation.** Every function that does I/O (DB, NATS, Redis) receives a `context.Context` as the first argument. Deadlines are set at the service entry point, not deep inside domain code.
- **Table-driven tests.** All unit tests use `testing.T` and `[]struct{name, input, want}` tables. No test framework beyond the standard library and `github.com/stretchr/testify/assert`.
