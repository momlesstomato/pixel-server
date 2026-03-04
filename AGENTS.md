# AGENTS — pixel-server

## Identity

pixel-server is a modern server for Pixels, written in Go 1.23+.
It implements the Pixel Protocol defined in `vendor/pixel-protocol/spec/protocol.yaml`.

---

## Golden rules (never violate)

1. **The spec is the source of truth.** All packet structures come from `vendor/pixel-protocol/spec/protocol.yaml`. Never invent field orders, types, or packet IDs.
2. **Generated code is not hand-edited.** Files under `pkg/protocol/` are produced by `tools/protogen`. Manual changes are overwritten. If logic is needed around a packet, write it in the service handler.
3. **One ECS World per room.** Each room goroutine owns its `*ecs.World` exclusively. No goroutine reads or writes another room's World.
4. **No globals, no singletons.** Every dependency is injected via constructor parameters or interfaces. No `init()` functions that register side effects.
5. **Domain owns its state.** External input to a room arrives via `chan Envelope`; never call room methods from outside the room goroutine.
6. **Fixed-tick simulation at 20 Hz.** The game loop runs on a 50 ms ticker. I/O is never blocking inside the tick.
7. **All cross-service communication goes through NATS.** No direct HTTP/gRPC calls between services in the hot path.
8. **Extensibility is mandatory.** New domain features must define explicit extension points in `pkg/plugin` and corresponding events/interceptors where appropriate.
9. **Top-level module boundaries only.** Under `pkg/`, each top-level domain/infrastructure area is a module root. `pkg/core` is a single shared module; nested modules under `pkg/core/*` are not allowed.
10. **Markdown file naming convention.** All Markdown files use an UPPERCASE stem and a lowercase `.md` extension (e.g. `README.md`, `AGENTS.md`, `OVERVIEW.md`). Docs realm folders under `docs/` are lowercase with a numeric prefix (e.g. `docs/01-handshake/OVERVIEW.md`). Mixed-case stems or uppercase `.MD` extensions are rejected in CI review.
11. **Every server feature wires plugin hooks.** When adding a new domain action (player joins, packet received, room loaded, etc.), emit an `event.Bus` event and support `intercept.Interceptor` hooks. No feature ships without Observable extension points.

---

## Technology stack (locked)

| Component | Choice | Import path / version |
|---|---|---|
| Language | Go 1.23+ | — |
| ECS | Ark | `github.com/mlange-42/ark` v0.7.1 |
| WebSocket | gobwas/ws | `github.com/gobwas/ws` |
| Message bus | NATS JetStream | `github.com/nats-io/nats.go` |
| Database | PostgreSQL 16 | `github.com/jackc/pgx/v5` |
| Cache | Redis 7 | `github.com/redis/go-redis/v9` |
| Configuration | Viper | `github.com/spf13/viper` |
| Migrations | Atlas | `ariga.io/atlas` |
| Logging | Zap | `go.uber.org/zap` |
| Testing | stdlib `testing` + testify | `github.com/stretchr/testify` |
| Integration tests | testcontainers-go | `github.com/testcontainers/testcontainers-go` |

Do **not** introduce new dependencies without justification in a PR description.

---

## Repository layout

```
pixel-server/
├── go.work
├── Makefile
├── AGENTS.md                   ← this file
├── architecture/               ← design docs (read-only reference)
├── docker/
│   └── compose.yml
├── pkg/
│   ├── core/                   ← feature-independent infrastructure
│   │   ├── codec/              ← binary Reader/Writer
│   │   ├── config/             ← Viper-backed runtime config loader
│   │   ├── logging/            ← Zap logger factory
│   │   ├── bus/                ← NATS thin wrapper + infrastructure subjects
│   │   └── testutil/           ← shared test helpers (testcontainers, mocks)
│   ├── protocol/               ← GENERATED packet structs
│   ├── pathfinding/            ← 3D A* + JPS + HPA*
│   ├── plugin/                 ← plugin framework (registry, loader)
│   │   ├── event/              ← EventBus interface + implementation
│   │   ├── intercept/          ← PacketInterceptor interface + chain
│   │   └── roomsvc/            ← Room service abstraction for plugins
│   ├── user/                   ← user domain: models, repository interfaces
│   │   └── memory/             ← in-memory user repository (tests/dev)
│   ├── room/                   ← room domain: models, ECS components, repository
│   │   └── memory/             ← in-memory room repository (tests/dev)
│   ├── item/                   ← item/furniture domain
│   ├── social/                 ← messenger, friends domain
│   ├── navigator/              ← room discovery domain
│   ├── catalog/                ← store, economy domain
│   └── moderation/             ← bans, tickets domain
├── examples/
│   └── plugins/
│       └── hello-world/        ← reference plugin (buildmode=plugin)
├── plugins/                    ← runtime .so drop directory (gitignored)
├── services/
│   ├── gateway/                ← WebSocket → NATS bridge
│   ├── auth/                   ← handshake + SSO
│   ├── game/                   ← room simulation, ECS
│   ├── social/                 ← messenger, friends
│   ├── navigator/              ← room discovery
│   ├── catalog/                ← store, economy
│   └── moderation/             ← bans, tickets
├── docs/                       ← wiki-like project documentation (what IS built)
│   ├── ARCHITECTURE.md         ← cross-cutting system overview
│   ├── 01-handshake/           ← handshake-security realm
│   │   ├── OVERVIEW.md
│   │   ├── PACKETS.md
│   │   ├── AUTH-FLOW.md
│   │   └── ERROR-HANDLING.md
│   ├── 02-session/             ← session-connection realm
│   │   ├── OVERVIEW.md
│   │   ├── LIFECYCLE.md
│   │   ├── PACKETS.md
│   │   └── PLUGIN-HOOKS.md
│   ├── 03-user-profile/        ← user-profile realm
│   │   ├── OVERVIEW.md
│   │   ├── LOGIN-BUNDLE.md
│   │   ├── PACKETS.md
│   │   ├── DATA-MODELS.md
│   │   └── PLUGIN-HOOKS.md
│   ├── 04-room/                ← room domain, ECS components, movement
│   │   ├── OVERVIEW.md
│   │   ├── DATA-MODELS.md
│   │   ├── ECS-COMPONENTS.md
│   │   ├── SYSTEMS.md
│   │   └── PLUGIN-HOOKS.md
│   └── 05-pathfinding/         ← 3D A*, heightmap layout, integration
│       ├── OVERVIEW.md
│       ├── LAYOUT.md
│       ├── ALGORITHM.md
│       └── INTEGRATION.md
├── tools/
│   ├── protogen/               ← YAML → Go code generator
│   └── packageguard/           ← CI: enforces max file count per package
└── vendor/                     ← upstream references (read-only)
```

---

## Documentation (`docs/`)

`docs/` contains wiki-like documentation describing **what is currently implemented** — not plans or aspirations. This is the single source of truth for understanding the running system.

### Rules

1. **Only document what exists.** Every statement in `docs/` must correspond to working, tested code in the repository. Planned features belong in `architecture/`, never in `docs/`.
2. **Update on every change.** When a feature is added, modified, or removed, the corresponding `docs/` page must be updated in the same PR. Stale documentation is treated as a bug.
3. **One folder per realm, numbered by implementation order.** Each implemented protocol realm gets its own lowercase subdirectory with a two-digit numeric prefix (e.g. `docs/04-room/`). Inside, create focused files named with an UPPERCASE stem and lowercase `.md` extension: `OVERVIEW.md`, `PACKETS.md`, `DATA-MODELS.md`, `PLUGIN-HOOKS.md`, etc. New realms receive the next available number. Cross-cutting concerns get a single standalone `docs/ARCHITECTURE.md` file.
4. **Detail level: exhaustive wiki-quality.** Each page MUST include all of the following sections:
   - **Purpose** — what the realm is responsible for.
   - **Packet table** — every C2S and S2C packet: header ID, struct name, field list, when it is sent.
   - **Handler walkthrough** — per-packet handler: what it reads, what it validates, what it publishes/returns.
   - **NATS subjects** — every subject published or subscribed, with message layout (field-by-field).
   - **Data models** — every domain struct and ECS component involved, with all fields and types.
   - **Plugin hooks** — every `event.Bus` event emitted and every `intercept.Interceptor` hook point, with Payload field descriptions.
   - **Realm relations** — which other realms this realm depends on and how (NATS, shared domain types).
   - **Permissions / guards** — any auth or role checks enforced before processing.
   - **Error handling** — every error path, what is logged, and what is returned to the client.
   - **Configuration knobs** — env vars and config struct fields that control this realm's behavior.
   A new contributor must be able to understand and extend the feature from the doc alone.
5. **No duplication with `architecture/`.** `architecture/` is the *design intent*; `docs/` is the *as-built reality*. When the implementation diverges from the architecture doc, `docs/` is authoritative for the current state.
6. **File naming convention enforced.** Markdown files anywhere in the repository must have an UPPERCASE stem and a lowercase `.md` extension. Realm subdirectories under `docs/` must be lowercase with a numeric prefix. Any file with an uppercase `.MD` extension or a mixed-case folder name is treated as a bug.

---

## Architecture documents

All design decisions are recorded in `architecture/`. These are the authoritative references:

| Document | Governs |
|---|---|
| `000-overview.md` | Master index, technology table, guiding principles |
| `001-go-workspace.md` | Module boundaries, `go.work`, Makefile targets |
| `002-protocol-codegen.md` | How `tools/protogen` reads the spec and emits Go code |
| `003-service-topology.md` | Service decomposition, NATS subjects, connection flow |
| `004-ecs-ark.md` | Ark v0.7.1 API, component definitions, system tick order |
| `005-pathfinding-3d.md` | 3D A* with Z-cost, JPS, HPA*, flying entities |
| `006-storage.md` | PostgreSQL schema (all tables), Redis patterns, async log writer |
| `007-messaging.md` | NATS JetStream subjects, stream topology, backpressure |
| `008-patterns.md` | Hexagonal architecture, DDD aggregates, TDD scope |
| `009-packet-roadmap.md` | 13-phase implementation order for 922 packets |
| `010-docker-production.md` | Docker Compose production deployment guide |
| `011-plugin-system.md` | ECS-aligned plugin API: EventBus, PacketInterceptor, Registry |

When implementing a feature, **read the relevant architecture doc first**.
When a decision contradicts an architecture doc, update the doc before changing code.

---

## Coding conventions

### Naming
- Package names: lowercase, single word (`codec`, `bus`, `pathfinding`).
- Exported types: PascalCase.
- File names: snake_case (`walk_path.go`, `room_worker.go`).
- Test files: `<file>_test.go` in the same package for unit tests, `<file>_integration_test.go` with build tag for integration.
- Packet handler files: match the packet name (`handshake_release_version.go`).

### Documentation
- Every exported function, method, interface, struct, and exported struct field must have GoDoc-style comments.
- Inside function bodies, avoid obvious comments; add comments only for non-trivial reasoning, constraints, or invariants.
- Every major module (`pkg/*`, `services/*`, `tools/*`) must have a `README.md`.
- Module/package `README.md` files must be accurate, actionable, and kept in sync with code changes in the same PR.
- Placeholder READMEs with only vague one-line descriptions are not acceptable.
- At minimum, each module/package `README.md` must document: purpose and scope, key entry points/APIs, invariants/constraints, and operational commands (build/test/generate where relevant).
- The repository root must include a public-oriented `README.md` describing architecture, design goals, and operational commands.
- Do not place production `.go` files directly under `pkg/`; source files must live inside a specific module package (for example `pkg/core/config`, `pkg/room`, `pkg/user`).
- **Markdown file naming: UPPERCASE stem, lowercase `.md` extension.** Examples: `README.md`, `AGENTS.md`, `OVERVIEW.md`. Docs realm folders are lowercase with a numeric prefix: `docs/01-handshake/PACKETS.md`.

### Configuration
- Service/runtime configuration must load through `pkg/core/config` (Viper-backed) with struct schemas.
- Config fields must use `mapstructure` and `env` tags.
- If a field has `default:"..."`, that default is applied when env is absent.
- If a field has no `default` tag, it is required and startup must fail when missing.
- All services/tools use one root `.env.example` as the canonical variable source.

### Errors
- Wrap with `fmt.Errorf("context: %w", err)`.
- Never swallow errors silently. Log at the boundary, return inside the domain.
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for domain-level conditions.

### Context
- Every public function that does I/O takes `context.Context` as first parameter.
- Room tick systems do **not** take context — they are pure computation.

### Logging
- Use `pkg/core/logging` (Zap-backed) for all executables.
- Log format must be configurable between `json` and `pretty` through env (`LOG_FORMAT`).
- Log levels: `debug`, `info`, `warn`, `error`; default `info`.
- Logs must be essential and structured: lifecycle/info state, debug traces (including packet debug when enabled), recoverable warnings, and actionable errors.

### Concurrency
- Each room goroutine is single-threaded for its ECS world.
- Use channels for communication between goroutines, not mutexes on shared state.
- `sync.Pool` for reusable scratch buffers (pathfinding, codec).

### Reusability and file structure
- Code must prioritize reusability, performance, and readability.
- Files over ~150 lines should be split when practical into cohesive units (types, adapters, lifecycle, tests) rather than accumulating mixed concerns.

### Package splitting
- Packages must be split by concern once they become broad or mixed-responsibility.
- As a hard threshold, if a package exceeds ~12 non-test files or begins mixing unrelated domains, create sub-packages (`internal/...` or sibling packages) with clear ownership.
- Avoid catch-all files like `utils.go` and `helpers.go` for unrelated concerns.

---

## Testing requirements (mandatory)

Every package and service **must** have tests. PRs without tests for new logic are rejected.
Unit, integration, and e2e test layers are all mandatory in CI.

### Enforcement rules

1. **No untested handler.** Every packet handler must have at least one unit test covering the happy path and one covering the primary error path.
2. **No untested service method.** Every exported method on a domain service (`*Service`) must have table-driven unit tests.
3. **No untested NATS wiring.** Every NATS subscription callback must have a unit test using a mock or in-memory bus proving message dispatch.
4. **Coverage gates.** Core packages (`pkg/core/*`, `pkg/user`, `pkg/room`) target ≥90% line coverage. Service packages (`services/*/internal/*`) target ≥80%.
5. **Integration tests per repository.** Every repository interface method backed by PostgreSQL/Redis must have an integration test using `testcontainers`.
6. **E2E per phase exit.** Each completed phase (per `009-packet-roadmap.md`) must have at least one scenario-based e2e test proving the phase exit criteria.

### Three test levels

#### 1. Unit tests (`go test ./...`)
- **Build tag:** none (always run).
- **Scope:** Pure functions, codec round-trips, ECS queries, pathfinding, domain logic.
- **Dependencies:** None. No database, no network, no file system.
- **Pattern:** Table-driven tests with `t.Run` sub-tests.
- **Assertions:** `github.com/stretchr/testify/assert` and `require`.
- **Coverage target:** Aim for the highest coverage possible, preferably 100% for core packages and all new logic.
- **Coverage bar for changes:** New or modified logic should target 100% statement coverage where practical; any gap must be intentional and explained in the PR notes.

```go
func TestReaderWriterRoundTrip(t *testing.T) {
    tests := []struct {
        name string
        write func(w *codec.Writer)
        read  func(r *codec.Reader) (any, error)
        want  any
    }{
        {"int32", func(w *codec.Writer) { w.WriteInt32(42) },
                  func(r *codec.Reader) (any, error) { return r.ReadInt32() }, int32(42)},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

#### 2. Integration tests (`go test -tags=integration ./...`)
- **Build tag:** `//go:build integration`
- **Scope:** Repository implementations against real PostgreSQL/Redis/NATS.
- **Dependencies:** `testcontainers-go` spins up containers automatically.
- **Pattern:** Each test suite gets a fresh database via `testcontainers`.
- **When to write:** Every repository method, every NATS publish/subscribe pattern.

```go
//go:build integration

func TestUserRepository_Create(t *testing.T) {
    ctx := context.Background()
    pg := testutil.StartPostgres(t, ctx)       // testcontainers helper
    repo := postgres.NewUserRepository(pg.Pool)
    u, err := repo.Create(ctx, &user.CreateParams{...})
    require.NoError(t, err)
    assert.Equal(t, "testuser", u.Username)
}
```

#### 3. End-to-end tests (`go test -tags=e2e ./...`)
- **Build tag:** `//go:build e2e`
- **Scope:** Full client→gateway→auth→game flow. A mock WebSocket client connects, completes handshake, enters a room.
- **Dependencies:** Full Docker Compose stack via testcontainers-go compose module, or pre-started via `make e2e-up`.
- **Pattern:** Scenario-based: `TestClientCanConnectAndEnterRoom`, `TestTwoClientsChat`.
- **When to write:** One per phase exit criteria (see `009-packet-roadmap.md`).

### Test utilities (`pkg/core/testutil/`)
Shared helpers:
- `StartPostgres(t, ctx)` — spins up PostgreSQL via testcontainers, runs migrations, returns pool.
- `StartRedis(t, ctx)` — spins up Redis via testcontainers, returns client.
- `StartNATS(t, ctx)` — spins up NATS via testcontainers, returns connection.
- `MockSession(userID int64)` — creates an in-memory session for unit tests.
- `MustDecode[T](t, data []byte)` — decodes a packet or fails the test.

### CI pipeline
```
make vet                              # workspace-wide vet across all modules
make lint                             # workspace-wide golangci-lint + package split guard
make check-package-split              # explicit package split guard
make test                             # unit tests
make test-integration                 # integration (needs Docker)
make test-e2e                         # e2e (needs Docker)
```

All three levels run in CI on every PR. Unit tests gate merge; integration and e2e must be kept green for release readiness.

### Executables and tooling quality
- Every executable binary (`services/*`, `tools/*`) must have structured startup/shutdown flow, dependency wiring, configuration validation, and clear logging.
- Auxiliary tooling in this repository must be implemented in Go. Do not introduce Python/Node/Ruby helper scripts for build-critical workflows.

---

## Package dependency rules

### Module topology
- Under `pkg/`, modules must be top-level only (`pkg/core`, `pkg/room`, `pkg/user`, etc.).
- `pkg/core` is the only module root for core infrastructure; do not add `go.mod` files in `pkg/core/*`.
- Sub-packages live under their owning module and must not become standalone modules unless architecture docs are updated first.

```
pkg/core/codec       → (nothing)
pkg/core/config      → github.com/spf13/viper
pkg/core/logging     → go.uber.org/zap
pkg/core/bus         → github.com/nats-io/nats.go
pkg/core/testutil    → testcontainers-go, testify, domain packages
pkg/protocol         → pkg/core/codec
pkg/pathfinding      → (nothing)
pkg/plugin           → go.uber.org/zap, pkg/plugin/event, pkg/plugin/intercept, pkg/plugin/roomsvc
pkg/plugin/event     → (nothing, part of pkg/plugin module)
pkg/plugin/intercept → pkg/plugin/event (part of pkg/plugin module)
pkg/plugin/roomsvc   → pkg/plugin/event (part of pkg/plugin module)
pkg/user             → (nothing — pure domain types + interfaces)
pkg/room             → github.com/mlange-42/ark (ECS components + World)
pkg/item             → (nothing — pure domain types + interfaces)
pkg/social           → (nothing — pure domain types + interfaces)
pkg/navigator        → (nothing — pure domain types + interfaces)
pkg/moderation       → (nothing — pure domain types + interfaces)
pkg/catalog          → (nothing — pure domain types + interfaces)

services/*           → pkg/protocol, pkg/core/codec, pkg/core/bus, pkg/core/config, pkg/core/logging, domain packages
services/game        → pkg/room, pkg/pathfinding, pkg/plugin (additionally)
tools/protogen       → gopkg.in/yaml.v3

examples/plugins/*   → pkg/plugin only (no server internals)
```

Each domain package (`pkg/user`, `pkg/room`, `pkg/item`, etc.) **owns its NATS subjects** in a `subjects.go` file.
Infrastructure-level subjects (handshake, session lifecycle) live in `pkg/core/bus/subjects.go`.

**Never** import a service package from another service. Cross-service data flows through NATS events only.

**Never** import database/cache implementation packages from domain code. Domain packages define repository interfaces; implementations are injected at startup (`main.go`).

**Never** import server-internal packages from a plugin `.so`. Plugin code may only depend on `pkg/plugin` and the Go standard library.

**Never** read environment variables ad-hoc via `os.Getenv` in service startup code. Use typed config structs with `pkg/core/config`.

---

## Commit conventions

- Prefix: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`, `gen:` (for generated code updates).
- One logical change per commit.
- Generated code changes get their own `gen: regenerate protocol` commit.

---

## What NOT to do

- Do **not** use `interface{}` or `any` in packet field types. Use typed structs.
- Do **not** use an ORM. All SQL is explicit `pgx/v5` queries.
- Do **not** call `os.Exit` anywhere except `main()`.
- Do **not** import `vendor/` code — it is read-only reference material.
- Do **not** put business logic in packet decode/encode functions.
- Do **not** use `sync.Mutex` to protect a room's ECS World — use the channel-based goroutine model.
- Do **not** block inside an ECS system tick (no I/O, no channel sends that could block).
- Do **not** skip writing tests. If it has logic, it has tests.
- Do **not** block or perform I/O inside a plugin event handler — it runs synchronously on the tick goroutine.
- Do **not** try to unload a plugin at runtime — Go's `plugin` package does not support it.
- Do **not** compile plugins with a different Go toolchain version than the host service binary.
- Do **not** use an uppercase `.MD` extension — all Markdown files must use a lowercase `.md` extension with an UPPERCASE stem (e.g. `README.md`, not `README.MD`).
- Do **not** add a server feature without wiring `event.Bus` emission and `intercept.Interceptor` hook points.
