# Go Workspace Layout

## Why a Go workspace?

A Go workspace (`go.work`) lets several modules coexist in one repository while resolving local replace directives automatically. This gives pixel-server:

- **Independent versioning** per module (protocol, game, gateway…) without a single giant `go.mod`.
- **Shared internal packages** (`pkg/*`) referenced from any service without a published registry.
- **Isolated dependency trees** per service – a socket-heavy gateway does not pull in ORM libraries used only in auth.
- **Parallel `go build ./...`** across modules in CI.

---

## `go.work` (root)

```go
go 1.23

use (
    ./pkg/protocol
    ./pkg/codec
    ./pkg/ecs
    ./pkg/pathfinding
    ./pkg/storage
    ./pkg/bus
    ./services/gateway
    ./services/auth
    ./services/game
    ./services/social
    ./services/navigator
    ./services/catalog
    ./services/moderation
    ./tools/protogen
)
```

---

## Module responsibilities

### `pkg/protocol`  `pixel-server/protocol`

**Auto-generated. Do not edit manually.**

Contains one Go file per realm (e.g. `handshake.go`, `room.go`, `navigator.go`) produced by `tools/protogen` from `spec/protocol.yaml`.

Exports:
- Typed packet structs (`ReleaseVersionPacket`, `MoveAvatarPacket`, …)
- `Encode(w io.Writer, pkt Packet) error`
- `Decode(r io.Reader, id uint16) (Packet, error)`
- `HeaderID(pkt Packet) uint16` – returns compile-time ID constant
- A `RouterTable` mapping header IDs → handler interfaces

No business logic. No imports other from `pkg/codec`.

### `pkg/codec`  `pixel-server/codec`

Binary encoding primitives matching the spec wire format:

```
uint32 big-endian length prefix
uint16 big-endian message ID
payload bytes
```

Provides:
- `Reader` – wraps `io.Reader`, exposes `ReadBool`, `ReadInt32`, `ReadUint16`, `ReadUint32`, `ReadString`, `ReadBytes`
- `Writer` – wraps `bytes.Buffer`, exposes symmetric `Write*` methods plus `Frame(id uint16)` that prepends length + ID
- RC4 stream cipher wrapper for post-handshake encryption (`RC4Reader`, `RC4Writer`)

No external dependencies.

### `pkg/ecs`  `pixel-server/ecs`

Thin opinionated wrapper around `github.com/mlange-de/arche`.

Defines the canonical components used across all rooms:
- `Position{X, Y, Z float32}`
- `TileRef{X, Y int16}` – grid-snapped tile address
- `Velocity{DX, DY, DZ float32}`
- `WalkPath{Steps []PathStep, Index int}`
- `EntityKind{Kind uint8}` – Avatar/Bot/Pet/Item
- `Status{Posture uint8; Effects uint32}` – bit-packed posture + active effects
- `ChatCooldown{Counter int32}`

Does **not** import any service-specific code.

### `pkg/pathfinding`  `pixel-server/pathfinding`

Self-contained 3D A* + HPA* (Hierarchical Path-finding A*) implementation.  
See [005-pathfinding-3d.md](005-pathfinding-3d.md) for full design.

Zero external dependencies. Fully unit-testable with table-driven tests.

### `pkg/storage`  `pixel-server/storage`

Repository interfaces + implementations:

```
storage/
  interfaces/         ← Go interfaces only (no driver imports)
    user.go
    room.go
    item.go
    …
  postgres/           ← pgx/v5 implementations
  redis/              ← go-redis/v9 implementations
  migrations/         ← SQL migration files (Atlas or goose)
```

Services depend on `storage/interfaces`; only the DI wiring layer imports `storage/postgres`.

### `pkg/bus`  `pixel-server/bus`

Thin wrapper around NATS JetStream:
- `Publisher` – publishes typed events to known subjects
- `Consumer` – subscribes with at-least-once delivery guarantee and back-pressure
- Subject constants (see [007-messaging.md](007-messaging.md))

### `services/gateway`  `pixel-server/gateway`

Responsibilities:
1. Accepts raw WebSocket connections (`gobwas/ws` + `epoll`).
2. Assigns a `sessionID` (UUIDv7).
3. Frames inbound bytes into packets using `pkg/codec`.
4. Dispatches pre-auth packets (`handshake.*`, `security.sso_ticket`) to `auth-svc` via NATS.
5. After auth: attaches the session to a game room by forwarding all packets to `game-svc` via NATS subject `room.<roomID>.input`.
6. Subscribes to `session.<sessionID>.output` to fan outbound packets back to the socket.

No game logic. No database writes. Purely I/O.

### `services/auth`  `pixel-server/auth`

Responsibilities:
1. Implements the Diffie-Hellman + RC4 handshake protocol.
2. Validates SSO tickets against a signed token store (Redis).
3. Resolves user identity from PostgreSQL.
4. Publishes `session.authenticated` events.

Stateless beyond Redis session store; horizontally scalable.

### `services/game`  `pixel-server/game`

The core game simulation. The heaviest service.

Internal structure:

```
game/
  supervisor/       ← manages room worker pool, creates/destroys room goroutines
  room/
    world.go        ← Arche ECS world per room
    tick.go         ← 20 Hz tick loop
    systems/        ← one Go file per ECS system
      movement.go
      interaction.go
      chat.go
      roller.go
      wired.go
      pet_ai.go
      …
  pathfinding/      ← delegates to pkg/pathfinding with room heightmap
  handlers/         ← generated handler stubs, filled with business logic
```

Each room runs in its own goroutine. All input arrives via a `chan Envelope`. All output is published to NATS `session.<id>.output`.

### `services/social`  `pixel-server/social`

Friends, messenger, room invitations, user search, console social notifications.  
Runs independently of game simulation.

### `services/navigator`  `pixel-server/navigator`

Room listing, search, favourites, home room, promoted rooms. Read-heavy; PostgreSQL + Redis cache.

### `services/catalog`  `pixel-server/catalog`

Store page tree, offers, purchases, gift wrapping, voucher redemption, marketplace listing. Transactional; PostgreSQL only.

### `services/moderation`  `pixel-server/moderation`

Ban management, mute, mod-tool interactions, ticket queue. Stateless; PostgreSQL + Redis.

### `tools/protogen`  `pixel-server/tools/protogen`

A Go program (invoked by `go generate` or `make generate`) that:
1. Reads `vendor/pixel-protocol/spec/protocol.yaml`.
2. Parses all packet definitions.
3. Emits `pkg/protocol/*.go` files.

See [002-protocol-codegen.md](002-protocol-codegen.md).

---

## Makefile targets

```makefile
generate:          ## Regenerate protocol package from spec
    go run ./tools/protogen -spec vendor/pixel-protocol/spec/protocol.yaml -out pkg/protocol

build:             ## Build all modules
    go build $(go list -f '{{.Dir}}' ./...)

test:              ## Run all tests
    go test ./...

lint:              ## golangci-lint
    golangci-lint run ./...
```

---

## CI checklist per PR

1. `go generate ./...` produces no git diff (generated files are committed).
2. `go vet ./...` clean.
3. `golangci-lint run ./...` clean.
4. `go test ./...` passes with `-race`.
5. `go build ./...` produces no error.
