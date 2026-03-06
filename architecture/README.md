# Architecture Directory Contract

This directory is planning-only.

- Use this directory for intended design, decisions, trade-offs, and phased plans.
- Do not describe implementation status here unless explicitly marked as planned or target state.

Global terminology for all files in `architecture/`:

- "module" or "realm" means internal bounded context inside the single `pixelsv` binary (`internal/<realm>/`).
- "contract topic" means internal messaging topic. In all-in-one mode these are in-process channels; in distributed mode they are NATS subjects.
- "role" means a runtime activation flag (`--role=gateway`, `--role=game`, etc.) that decides which realm modules start.

Implemented behavior must be documented under `docs/`.

## Document Index

| Document | Scope |
|---|---|
| [000-overview.md](000-overview.md) | Vision, principles, repository layout |
| [001-go-workspace.md](001-go-workspace.md) | Go module/workspace baseline |
| [002-protocol-codegen.md](002-protocol-codegen.md) | Protocol YAML to Go code generation |
| [003-service-topology.md](003-service-topology.md) | Multi-role runtime topology |
| [004-ecs-ark.md](004-ecs-ark.md) | ECS component model and tick systems |
| [005-pathfinding-3d.md](005-pathfinding-3d.md) | 3D A*, JPS, HPA* algorithms |
| [006-storage.md](006-storage.md) | PostgreSQL schema, Redis patterns, role-aware access |
| [007-messaging.md](007-messaging.md) | Transport adapter contracts (local + NATS) |
| [008-patterns.md](008-patterns.md) | DDD/Hex/TDD/ECS conventions |
| [009-packet-roadmap.md](009-packet-roadmap.md) | 13-phase packet implementation order |
| [010-docker-production.md](010-docker-production.md) | Docker deployment (all-in-one + distributed) |
| [011-plugin-system.md](011-plugin-system.md) | Plugin API and ECS safety |
| [012-rest-api.md](012-rest-api.md) | Fiber HTTP, WebSocket, and REST API |
