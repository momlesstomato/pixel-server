# pixelsv Architecture Overview

## Vision

`pixelsv` is a modern Habbo-style server implemented as a **single binary** with clear domain boundaries and reusable packages.

The runtime is modular, but not microservice-based. Bounded contexts exist as internal realm modules that can be activated as roles through CLI flags.

## Non-Negotiable Principles

- DDD for domain boundaries and invariants.
- Hexagonal architecture for dependency direction.
- TDD for domain/core logic.
- ECS as mandatory runtime model for real-time room simulation.
- Domain code must never depend on infrastructure adapters.
- API and CLI must share the same application/core layer.

## Runtime Model

- One executable: `pixelsv`.
- One process can run all roles (API, game loop, schedulers, admin commands) or a subset.
- Role flags (`--role=gateway,game,social,...`) decide which realm modules start.
- When all roles run in one process, inter-module transport is in-process (Go channels and function calls).
- When roles run as separate processes, inter-module transport uses NATS (same contract topics, different adapter).

## Multi-Role Architecture

`pixelsv` follows the HashiCorp pattern (Consul, Nomad, Vault): one binary, multiple deployment modes.

| Role | Responsibility | Needs DB | Needs Redis | Listens HTTP | Listens WS |
|---|---|---|---|---|---|
| `gateway` | WebSocket sessions, packet routing | No | Yes | Health only | Yes |
| `game` | ECS room workers, simulation | Yes | Yes | No | No |
| `auth` | Handshake, SSO, identity | Yes | Yes | No | No |
| `social` | Friends, messenger, notifications | Yes | Yes | No | No |
| `navigator` | Room discovery, search | Yes | Yes | No | No |
| `catalog` | Store, economy, purchases | Yes | No | No | No |
| `moderation` | Bans, tickets, safety | Yes | Yes | No | No |
| `api` | REST admin endpoints, Swagger | Yes | Yes | Yes | No |
| `jobs` | Maintenance scheduler | Yes | Yes | Health only | No |
| `all` | All of the above (dev/small deploy) | Yes | Yes | Yes | Yes |

### Scaling Model

- **Dev / small deployment**: `pixelsv serve` (all roles, one process, local transport).
- **Medium deployment**: Separate gateway + game + api processes with NATS transport.
- **Large deployment**: Multiple game worker instances with consistent hashing or registry-based room distribution.

### Failure Isolation

- Each role runs as a separate OS process when deployed distributed.
- A crashed game worker only affects rooms on that instance.
- Gateway, social, API, and other roles remain available.
- Process managers (systemd, Kubernetes, Nomad) restart crashed role instances.
- Room state survives crashes via Redis ECS snapshots (ark-serde).

## Web Layer Direction

- HTTP and WebSocket are served by GoFiber v3.
- Prefer Fiber-compatible websocket middleware for realtime endpoints.
- REST and WebSocket both map into application ports, never directly into domain state mutation.

## Repository Intent

```
pixel-server/
├── AGENTS.md               <- engineering contract
├── go.mod                  <- module: pixelsv
├── go.work                 <- workspace root
├── architecture/           <- planning and target design
├── docs/                   <- implemented behavior and usage
│   ├── core/
│   └── realms/<realm>/
├── cmd/
│   └── pixelsv/
├── internal/
│   ├── runtime/
│   │   ├── cli/            <- Cobra command graph
│   │   ├── transport/      <- local + NATS adapters
│   │   └── supervisor/     <- goroutine lifecycle, panic recovery
│   └── realms/
│       ├── auth/
│       │   ├── domain/
│       │   ├── app/
│       │   └── adapters/
│       ├── game/
│       │   ├── domain/
│       │   ├── app/
│       │   └── adapters/
│       ├── social/
│       ├── navigator/
│       ├── catalog/
│       └── moderation/
├── pkg/
│   ├── config/             <- Viper config loading
│   ├── log/                <- Zap logger factory
│   ├── codec/              <- binary Reader/Writer
│   ├── protocol/           <- GENERATED packet structs
│   ├── pathfinding/        <- 3D A* + JPS + HPA*
│   ├── http/               <- Fiber setup, middleware, Swagger
│   ├── storage/
│   │   ├── interfaces/     <- generic persistence ports
│   │   ├── postgres/       <- pgx adapter
│   │   └── redis/          <- go-redis adapter
│   └── plugin/             <- plugin framework
├── e2e/                    <- end-to-end tests
├── tools/
│   └── protogen/           <- YAML -> Go code generator
└── vendor/                 <- read-only references
```

## Documentation Contract

- `architecture/` documents intended design and future decisions.
- `docs/` documents real implemented behavior only.
- If code changes behavior, update `docs/`.
- If direction changes before implementation, update `architecture/`.

## Vendor Contract

`vendor/` is reference-only and not part of the core architecture. Current canonical references:

- Server references: `Arcturus-Community`, `PlusEMU`, `comet-v2`
- Client reference: `nitro-renderer`

## Architecture Index

| File | Scope |
|---|---|
| [001-go-workspace.md](001-go-workspace.md) | Go module/workspace baseline |
| [002-protocol-codegen.md](002-protocol-codegen.md) | Protocol YAML to Go code generation |
| [003-service-topology.md](003-service-topology.md) | Multi-role runtime topology |
| [004-ecs-ark.md](004-ecs-ark.md) | ECS component model and tick systems |
| [005-pathfinding-3d.md](005-pathfinding-3d.md) | 3D A*, JPS, HPA* algorithms |
| [006-storage.md](006-storage.md) | PostgreSQL schema, Redis patterns |
| [007-messaging.md](007-messaging.md) | Transport adapter contracts |
| [008-patterns.md](008-patterns.md) | DDD/Hex/TDD conventions |
| [009-packet-roadmap.md](009-packet-roadmap.md) | 13-phase packet implementation order |
| [010-docker-production.md](010-docker-production.md) | Docker deployment guide |
| [011-plugin-system.md](011-plugin-system.md) | Plugin API and ECS safety |
| [012-rest-api.md](012-rest-api.md) | Fiber HTTP, WebSocket, and REST API |
