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
internal/realms/<realm>/          <- domain-focused bounded contexts (user, room, chat, ...)
internal/realms/<realm>/domain/   <- entities, aggregates, domain services
internal/realms/<realm>/app/      <- use cases and orchestration ports
internal/realms/<realm>/adapters/ <- transport/storage adapters per realm
pkg/                              <- reusable packages shared across realms
```

Current shared package baseline:

- `pkg/config` for structured runtime configuration loading and validation.
- `pkg/log` for zap logging configuration and logger construction.
- `pkg/storage/interfaces` for persistence ports.
- `pkg/storage/postgres` and `pkg/storage/redis` for persistence adapters.
- `pkg/http` for core Fiber runtime, Swagger routes, API-key middleware, and WebSocket endpoint.
- `cmd/pixelsv` + Cobra command graph for single-binary runtime entry.
- `e2e/01_config_e2e_test.go`, `e2e/02_storage_e2e_test.go`, and `e2e/03_api_e2e_test.go` for step-based e2e coverage growth.

## Dependency Direction

- `internal/realms/<realm>/domain` depends on nothing infra-specific.
- `internal/realms/<realm>/app` depends on realm domain + ports.
- `internal/realms/<realm>/adapters` depends on realm app/domain ports and concrete libraries.
- `internal/runtime/cli` composes reusable packages and realm modules.
- `cmd/pixelsv` wires dependencies.

## Testing Direction

- Unit tests: domain + reusable packages first.
- Integration tests: adapters.
- End-to-end tests: CLI + HTTP/WebSocket flows through the same binary.
