# Pixel Server Engineering Contract

This document is mandatory for all contributors and agents working in this repository.

## 1) Core Direction

- Build `pixelsv` as a single binary with multiple runtime roles.
- Do not design or implement microservices for core server features.
- Follow Linux/HashiCorp-style operational design: one executable, role-driven startup (`api`, `cli`, workers) through commands/flags.
- Keep domain boundaries strict even inside one repository and one binary.

## 2) Architecture Rules (Non-Negotiable)

- Use DDD for domain modeling and bounded contexts.
- Use Hexagonal Architecture (Ports and Adapters) as the baseline structure.
- Domain logic must not depend on infrastructure/framework implementations.
- ECS is mandatory for runtime simulation concerns (room entities, movement, state updates).
- Packages must be highly decoupled and reusable.
- Prefer composition over direct concrete coupling.
- `pkg/` must stay domain-agnostic and reusable.
- Do not place business-domain entities, repositories, or use-case-specific contracts in `pkg/` (examples: `UserRepository`, `RoomRepository`).
- Domain-specific ports/adapters belong to realm packages under `internal/<realm>/...`.
- Runtime wiring (CLI/bootstrap/composition) belongs to `pkg/core/...`, separated from realm business modules.
- `internal/` contains realm bounded contexts directly (`internal/<realm>/...`), not an extra `internal/realms/` layer.

## 3) Documentation and Commenting Rules

- Every exported and non-trivial unexported function signature, struct, interface, and struct field must have proper GoDoc-style comments.
- Inline comments inside function bodies are not allowed.
- Code must be self-explanatory through naming, small functions, and clear composition.
- Markdown file naming must use `UPPERCASE` file names with `.md` lowercase extension (example: `005-PATHFINDING.md`).
- `architecture/` is planning and intended design only.
- `docs/` describes what is already implemented and validated in code.
- `docs/` must keep detailed usage documentation in:
  - `docs/core/`
  - `docs/realms/<realm>/...`
- Every code modification must review and update corresponding Markdown documentation in `architecture/` and/or `docs/` as applicable.

## 4) File and Code Size Constraints

- Target a hard limit of 150 lines per source file.
- Split files before they exceed the limit.
- Major packages must be split into smaller packages/modules before they grow too wide.
- Maximum allowed core source files per package: `5` (unit test files are excluded from this count).
- Allowed exceptions:
  - generated files
  - migrations
  - test data fixtures

## 5) Testing Policy

- TDD is required for domain logic and reusable packages.
- Maintain high unit test coverage on pure/domain behavior.
- Maintain high end-to-end coverage for protocol, API, and runtime flows.
- New behavior requires tests in the same change set.
- Regressions must be reproduced by tests before fix completion.
- E2E tests in `e2e/` must use concise numeric prefixes (`01_`, `02_`, ...), not verbose `step_*` names.
- Do not split e2e tests into per-step folders unless a step needs multiple files or a file would exceed line limits.

## 6) Runtime/API/CLI Policy

- The binary must serve both API (HTTP/WebSocket) and CLI use cases.
- WebSocket implementation preference is GoFiber-compatible middleware (`fiber/v3` + websocket middleware) unless a documented benchmark justifies deviation.
- Runtime modes must share the same domain core and application services.
- Duplicated code should preferably not exist; shared behavior must be extracted into reusable packages/components.

## 6.1) Performance Policy (Strict)

- All code must target low-latency, predictable runtime performance.
- Every runtime hot path must have explicit performance goals and benchmark coverage.
- 20Hz simulation budget is non-negotiable: one tick must complete within `50ms`; infrastructure overhead should remain below `5ms` per tick under normal load.
- Transport goals:
  - local in-process bus publish-to-handler path must target low microsecond latency and avoid unnecessary allocations.
  - distributed transport (NATS) must be benchmarked and tuned for low single-digit millisecond publish/consume latency in normal network conditions.
- Any change that affects hot paths must include updated benchmarks and performance notes in corresponding docs.

## 7) Configuration Policy

- All new configuration keys must be declared in `.env.example` and updated in the same change set.
- Configuration must be structured and extendable.
- Every configuration struct field must follow this contract:
  - has `default:"..."` tag when optional with fallback value
  - has no `default` tag when required
- Required fields (no `default` tag) must be enforced by validation.
- A package config must contain only package-related settings.
- If settings belong to another package/domain, define a dedicated `config/config.go` in that package.
- Final runtime configuration composition is the responsibility of `cmd` entrypoints, based on command intent.
- Logging configuration must expose:
  - output format: `json` or `console` (pretty console)
  - log level: `debug`, `info`, `warn`, `error`, etc.
- Logging implementation must live in a dedicated package (`pkg/log`), not inside `pkg/config`.
- Administrative HTTP endpoints must be protected by an API key configured via environment (`API_KEY`).

## 8) Vendor and Repository Hygiene

- `vendor/` is read-only reference material.
- Vendor currently contains reference implementations:
  - server references (3): `Arcturus-Community`, `PlusEMU`, `comet-v2`
  - client reference (1): `nitro-renderer`
- Do not couple production code to vendor internals.
- Keep `.gitignore` rules aligned with repository policy, including vendor handling and build artifacts.

## 9) Enforcement Checklist (PR/Change Gate)

- GoDoc coverage complete for required symbols and struct fields.
- No inline function-body comments.
- No domain-to-infrastructure dependency violations.
- `pkg/` contains only reusable implementation-independent primitives, not domain repositories/entities.
- File length policy respected (or exception justified).
- Unit + e2e tests added/updated and passing.
- `architecture/` updates for planned changes.
- `docs/` updates for implemented behavior.
- `.env.example` updated for every new configuration item.
- Package file-count policy validated (`<=5` core files per package unless exception).
