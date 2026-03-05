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
cmd/pixelsv/             <- binary entrypoint and command graph
internal/domain/         <- entities, aggregates, domain services
internal/app/            <- use cases and orchestration
internal/adapters/       <- DB, cache, transport adapters
pkg/                     <- reusable packages shared across modules
```

Current shared package baseline:

- `pkg/config` for structured runtime configuration loading and validation.
- `pkg/log` for zap logging configuration and logger construction.
- `pkg/storage/interfaces` for persistence ports.
- `pkg/storage/postgres` and `pkg/storage/redis` for persistence adapters.
- `e2e/01_config_e2e_test.go` and `e2e/02_storage_e2e_test.go` for step-based e2e coverage growth.

## Dependency Direction

- `internal/domain` depends on nothing infra-specific.
- `internal/app` depends on domain + ports.
- `internal/adapters` depends on app/domain ports and concrete libraries.
- `cmd/pixelsv` wires dependencies.

## Testing Direction

- Unit tests: domain + reusable packages first.
- Integration tests: adapters.
- End-to-end tests: CLI + HTTP/WebSocket flows through the same binary.
