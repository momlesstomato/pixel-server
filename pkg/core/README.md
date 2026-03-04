# pkg/core

`pkg/core` contains cross-domain infrastructure building blocks shared by services and other package modules.

## Sub-packages

- `pkg/core/codec`: packet framing and primitive binary codec.
- `pkg/core/config`: typed runtime config loader (Viper-backed).
- `pkg/core/logging`: Zap logger factory and formatting/level config.
- `pkg/core/bus`: NATS + JetStream thin wrapper and infrastructure subjects.
- `pkg/core/testutil`: shared infra test helpers (testcontainers).

## Module policy

- `pkg/core` is a single Go module (`pkg/core/go.mod`).
- Do not create nested modules under `pkg/core/*`.
- New infrastructure utilities must live in the appropriate `pkg/core/<area>` package, not as standalone modules.

## Usage

From other modules, import by package path:

- `pixel-server/pkg/core/codec`
- `pixel-server/pkg/core/config`
- `pixel-server/pkg/core/logging`
- `pixel-server/pkg/core/bus`
- `pixel-server/pkg/core/testutil`
