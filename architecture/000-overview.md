# pixelsv Architecture Overview

## Vision

`pixelsv` is a modern Habbo-style server implemented as a **single binary** with clear domain boundaries and reusable packages.

The runtime is modular, but not microservice-based. Bounded contexts exist as internal modules that can be started as roles through CLI commands.

## Non-Negotiable Principles

- DDD for domain boundaries and invariants.
- Hexagonal architecture for dependency direction.
- TDD for domain/core logic.
- ECS as mandatory runtime model for real-time simulation.
- Domain code must never depend on infrastructure adapters.
- API and CLI must share the same application/core layer.

## Runtime Model

- One executable: `pixelsv`.
- One process can run API, game loop, schedulers, and admin commands.
- Optional role flags/subcommands decide which modules start.
- Internal contracts are message-driven, but transport remains in-process by default.

## Web Layer Direction

- HTTP and WebSocket are served by GoFiber v3.
- WebSocket transport should use the GoFiber websocket middleware path unless benchmark evidence requires an alternative.
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
│   │   └── cli/
│   └── realms/
│       ├── user/
│       ├── room/
│       └── chat/
├── pkg/
│   └── ... reusable libraries
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
| [003-service-topology.md](003-service-topology.md) | Single-binary runtime topology |
| [007-messaging.md](007-messaging.md) | Internal messaging contracts |
| [008-patterns.md](008-patterns.md) | DDD/Hex/TDD conventions |
| [012-rest-api.md](012-rest-api.md) | Fiber HTTP + WebSocket API plan |
