# Go Module and Workspace Baseline

## Objective

Bootstrap `pixelsv` with Go module/workspace support while keeping a single-binary architecture.

## Baseline Files

### `go.mod` (root)

```go
module pixelsv

go 1.25.5
```

### `go.work` (root)

```go
go 1.25.5

use .
```

## Why Keep `go.work` with One Module?

- Future-proofs local multi-module tooling without changing developer workflow.
- Keeps room for tools or generated modules if they appear later.
- Maintains an explicit workspace root for IDE/task consistency.

## Layout Direction

```
cmd/pixelsv/                      <- binary entrypoint
internal/runtime/cli/             <- command graph and runtime composition
internal/runtime/transport/       <- local + NATS transport adapters
internal/runtime/supervisor/      <- goroutine lifecycle, panic recovery
internal/realms/<realm>/          <- domain-focused bounded contexts
internal/realms/<realm>/domain/   <- entities, aggregates, domain services
internal/realms/<realm>/app/      <- use cases and orchestration ports
internal/realms/<realm>/adapters/ <- transport/storage adapters per realm
pkg/                              <- reusable packages shared across realms
e2e/                              <- end-to-end tests
```

Current shared package baseline:

- `pkg/config` for structured runtime configuration loading and validation.
- `pkg/log` for zap logging configuration and logger construction.
- `pkg/codec` for binary protocol Reader/Writer.
- `pkg/protocol` for generated packet structs.
- `pkg/pathfinding` for 3D A* + JPS + HPA*.
- `pkg/storage/interfaces` for generic persistence ports.
- `pkg/storage/postgres` and `pkg/storage/redis` for persistence adapters.
- `pkg/http` for core Fiber runtime, Swagger routes, API-key middleware, and WebSocket endpoint.
- `pkg/plugin` for plugin framework (EventBus, PacketInterceptor, Registry).
- `cmd/pixelsv` + Cobra command graph for single-binary runtime entry.
- `e2e/` for step-based end-to-end coverage.

## Dependency Direction

- `internal/realms/<realm>/domain` depends on nothing infra-specific.
- `internal/realms/<realm>/app` depends on realm domain + ports.
- `internal/realms/<realm>/adapters` depends on realm app/domain ports and concrete libraries.
- `internal/runtime/cli` composes reusable packages and realm modules.
- `internal/runtime/transport` implements transport ports (local channels or NATS).
- `cmd/pixelsv` wires dependencies based on `--role` flag.

## Multi-Role Wiring

The `cmd/pixelsv` entrypoint reads the `--role` flag and initializes only the requested realm modules:

```go
// Pseudo-code
roles := parseRoles(cfg.Role) // "all", "gateway", "game", "api", ...

if roles.Has("gateway") || roles.Has("all") {
    gatewayRealm := gateway.New(deps)
    gatewayRealm.MountWS(app)
}
if roles.Has("game") || roles.Has("all") {
    gameRealm := game.New(deps)
    gameRealm.Start(ctx)
}
// ... etc
```

When `--role=all` (default), transport adapters use in-process channels. When roles are split across processes, transport adapters use NATS with the same contract topics.

## Testing Direction

- Unit tests: domain + reusable packages first.
- Integration tests: adapters.
- End-to-end tests: CLI + HTTP/WebSocket flows through the same binary.
