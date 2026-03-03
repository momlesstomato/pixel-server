# pixel-server

`pixel-server` is a modern, modular server for Pixels/Habbo-like clients built in Go. It replaces legacy monolith patterns with a service topology, deterministic ECS room simulation, protocol-driven packet generation, and explicit extension points for plugins.

## What this repository contains

- **Protocol codegen pipeline** from `vendor/pixel-protocol/spec/protocol.yaml` into typed Go packets.
- **Realtime gateway/auth/game foundations** for connection and handshake flow.
- **ECS-first simulation model** (one world per room goroutine at fixed 20 Hz tick).
- **Cross-service eventing via NATS JetStream** (no direct service-to-service RPC in hot paths).
- **Plugin framework** for packet interception and in-process room events.

## Architecture at a glance

- **gateway**: WebSocket ingress/egress and packet framing
- **auth**: handshake + ticket validation + session auth events
- **game**: room ownership, ECS tick systems, plugin execution context
- **social/navigator/catalog/moderation**: domain services (scaffolded)

Core decisions are documented in `architecture/` (start with `000-overview.md`).

## Configuration model

All executables must load config through `pkg/config` (Viper-backed):

- Schema defined as Go structs.
- `mapstructure` + `env` tags on fields.
- `default:"..."` tag means optional with fallback.
- No `default` tag means required (startup fails if missing).

Use a single root `.env.example` as canonical env source.

## Logging model

All executables use `pkg/logging` (Zap-backed):

- `LOG_FORMAT=json|pretty`
- `LOG_LEVEL=debug|info|warn|error`

Logs should remain essential and structured: lifecycle state, packet debug (when enabled), warnings, and actionable errors.

## Quick start

1. Copy env template:
   - `cp .env.example .env`
2. Start infra:
   - `make docker-up`
3. Generate protocol (if needed):
   - `make generate`
4. Build all modules:
   - `make build`

## Testing commands

- Unit: `make test`
- Integration: `make test-integration`
- E2E: `make test-e2e`
- Package split enforcement: `make check-package-split`

CI expectations: all three layers (unit, integration, e2e) are mandatory.

## Repository structure

- `architecture/` design and constraints
- `pkg/` reusable libraries (`codec`, `config`, `logging`, `plugin`, etc.)
- `services/` executable services
- `tools/` executable tooling (`protogen`)
- `vendor/` read-only upstream references
