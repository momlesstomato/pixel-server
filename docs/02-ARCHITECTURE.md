# Architecture

## Design Principles

Pixel Server follows **Domain-Driven Design (DDD)** with **Hexagonal
Architecture** (ports and adapters). Every module is organized around business
domains (realms) rather than technical layers.

### Why Hexagonal?

Traditional Habbo implementations use a layered architecture where protocol
handling, business logic, and storage are interleaved. A single packet handler
might decode a packet, query the database, mutate game state, and send a
response — all in one function.

Hexagonal architecture separates these concerns:

```
              ┌──────────────────────────────┐
              │         Application          │
              │   (use cases / orchestration)│
              │                              │
  Adapters ───┤         Domain               ├─── Adapters
  (inbound)   │   (entities, value objects,  │    (outbound)
              │    port interfaces)          │
              │                              │
              │         Infrastructure       │
              │   (database, Redis, etc.)    │
              └──────────────────────────────┘
```

- **Domain** defines entities and port interfaces (what the system needs)
- **Application** implements use cases (what the system does)
- **Adapters** connect external systems (how the system integrates)
  - Inbound: WebSocket handlers, HTTP routes, CLI commands
  - Outbound: Redis stores, PostgreSQL repositories, broadcast bus

This separation means the handshake flow can be tested without a WebSocket
connection, the SSO store can be swapped from Redis to an in-memory map for
tests, and new transport adapters can be added without touching business logic.

## Module Organization

### Two-Layer Structure

```
core/           Platform infrastructure (transport-agnostic, realm-agnostic)
pkg/            Realm-specific business logic (handshake, session, auth, etc.)
cmd/            Binary entrypoints
sdk/            Plugin SDK module (separate go.mod, zero dependencies)
```

**`core/`** owns reusable infrastructure that any realm may depend on:

| Package | Responsibility |
|---------|---------------|
| `core/app` | Application identity and network config |
| `core/config` | Viper-based configuration loading with validation |
| `core/codec` | Binary protocol frame encoding/decoding |
| `core/connection` | Connection abstraction, session registry |
| `core/broadcast` | Cross-instance Pub/Sub (Redis + local adapters) |
| `core/crypto` | Diffie-Hellman, RSA, RC4 stream encryption |
| `core/http` | Fiber HTTP module, WebSocket, OpenAPI |
| `core/redis` | Redis client factory |
| `core/postgres` | GORM client, migration/seed manager |
| `core/logging` | Zap structured logging |
| `core/cli` | Cobra command tree |
| `core/initializer` | Startup stage orchestration |
| `core/status` | Hotel status configuration |
| `core/plugin` | Plugin manager and event dispatcher |

**`pkg/`** owns realm-specific logic following hexagonal boundaries:

| Package | Realm | Hexagonal Layers |
|---------|-------|-----------------|
| `pkg/handshake` | Handshake & Security | `application/{authflow,cryptoflow,sessionflow}`, `adapter/realtime`, `packet/{bootstrap,security,authentication,crypto,session,telemetry}` |
| `pkg/authentication` | SSO Token Management | `domain`, `application`, `infrastructure/redisstore`, `adapter/{httpapi,command}` |
| `pkg/session` | Session & Connection | `application/{postauth,navigation,notification}`, `packet/{availability,hotel,navigation,notification,error}` |
| `pkg/status` | Hotel Status | `domain`, `application/hotelstatus`, `infrastructure/redisstore` |
| `pkg/user` | User Management | `domain`, `application`, `infrastructure/{model,store}` |

### Realm Boundaries

Each realm is a bounded context in DDD terms. Realms communicate through
well-defined port interfaces, never by importing each other's internal packages.

```
pkg/handshake/                   pkg/authentication/
  application/                     domain/
    authflow/                        ticket.go (SSO model)
      contracts.go ──────────────>   config.go
      usecase.go                   application/
  adapter/                           service.go
    realtime/                      infrastructure/
      handler.go                     redisstore/ (Redis adapter)
                                   adapter/
                                     httpapi/ (REST API)
                                     command/ (CLI)
```

The handshake realm's `authflow` use case depends on a `TicketValidator`
interface. The authentication realm's `application/service.go` implements it.
The wiring happens in `core/cli/serve.go` at startup — not through direct
package imports.

## Distribution Model

### Multi-Instance Architecture

```
                    ┌─────────────┐
                    │ Load Balancer│
                    └──────┬──────┘
              ┌────────────┼────────────┐
              │            │            │
        ┌─────┴─────┐ ┌───┴─────┐ ┌───┴─────┐
        │ Instance A │ │Instance B│ │Instance C│
        │  pixelsv   │ │ pixelsv  │ │ pixelsv  │
        └─────┬─────┘ └────┬────┘ └────┬────┘
              │            │            │
              └────────────┼────────────┘
                    ┌──────┴──────┐
              ┌─────┴─────┐ ┌────┴─────┐
              │   Redis    │ │PostgreSQL │
              └───────────┘ └──────────┘
```

All instances are stateless processes. Shared state lives in Redis (ephemeral)
and PostgreSQL (persistent):

| State | Storage | Rationale |
|-------|---------|-----------|
| Session registry | Redis (TTL 120s) | Fast lookup, auto-expiry on crash |
| SSO tokens | Redis (TTL configurable) | Single-use, ephemeral |
| Hotel status | Redis (persistent key) | Consistent across instances |
| Broadcast channels | Redis Pub/Sub | Fire-and-forget real-time delivery |
| User accounts | PostgreSQL | Permanent, queryable |
| Migration state | PostgreSQL | Schema versioning |

### Cross-Instance Communication

The `Broadcaster` interface (`core/broadcast/bus.go`) provides Pub/Sub:

```go
type Broadcaster interface {
    Publish(ctx context.Context, channel string, payload []byte) error
    Subscribe(ctx context.Context, channel string) (<-chan []byte, Disposable, error)
}
```

Two implementations exist:
- `RedisBroadcaster` — Redis Pub/Sub for production multi-instance
- `LocalBroadcaster` — in-process channels for tests and single-instance

**Channel topology:**

| Channel | Purpose | Publisher | Subscribers |
|---------|---------|-----------|-------------|
| `broadcast:all` | Hotel-wide packets | Any instance | All instances |
| `broadcast:conn:{id}` | Targeted connection | Any instance | Owning instance |
| `broadcast:user:{id}` | Targeted user | Any instance | Owning instance |
| `hotel:status` | State change | Transitioning instance | All instances |

### Session Leases

Every session key in Redis has a 120-second TTL, refreshed every 60 seconds by
the heartbeat goroutine. If an instance crashes:

1. Heartbeat stops → TTL is not refreshed
2. After 120 seconds, Redis expires the session keys
3. The user can log in again on any instance without hitting a duplicate session

This eliminates the orphan session problem that plagues single-instance servers.

## Startup Sequence

The initializer runner (`core/initializer/runner.go`) executes stages in order:

```
1. Config          Load .env + environment variables, validate
2. Redis           Connect to Redis, verify connectivity
3. Logger          Create Zap logger from config
4. PostgreSQL      Connect to DB, run migrations/seeds if enabled
5. HTTP            Create Fiber app, apply API key middleware
6. WebSocket       Register /ws handler with full handshake pipeline
```

Each stage receives the outputs of previous stages. If any stage fails, the
server exits with a descriptive error.

## Comparison with Other Implementations

| Aspect | Gladiator (Java) | Sodium (C#) | Galaxy (Java) | Pixel Server |
|--------|----------------|-------------|---------------------|-------------|
| Architecture | Monolithic classes | DI + layers | Package-based | DDD + Hexagonal |
| Multi-instance | Not supported | Not supported | Not supported | Redis-backed |
| Session state | `ConcurrentHashMap` | In-memory dict | In-memory map | Redis with TTL |
| Broadcasting | Direct iteration | Direct iteration | N/A | Redis Pub/Sub |
| Plugin isolation | URLClassLoader | AssemblyLoadContext | None | Same-process `.so` |
| Config | Properties file | JSON config | Env vars | Viper (.env + env) |
| DB migrations | Manual SQL | EF migrations | GORM AutoMigrate | gormigrate stages |
| Connection close | Direct socket close | Direct close | Direct close | Broadcast bus (cross-instance) |
| Packet codec | Custom per-vendor | Custom per-vendor | Custom | `core/codec` (shared) |
| Encryption | Optional RC4+RSA+DH | Optional | N/A | Optional RC4+RSA+DH |

The primary architectural advantage is that Pixel Server's hexagonal boundaries
make it straightforward to:
- Add a new realm without touching existing code
- Swap infrastructure (e.g., Redis → another KV store) by implementing a port
- Test business logic in isolation from transport and storage
- Deploy as single instance or distributed cluster with identical code
