# Architecture Overview

This document describes pixel-server's system design, service topology,
communication model, and cross-cutting infrastructure. It is the single
reference for understanding how all services and packages fit together.

---

## Design Goals

1. **Protocol fidelity.** Every packet structure comes from
   `vendor/pixel-protocol/spec/protocol.yaml`; no fields are invented.
2. **Extensibility.** Every domain action emits an `event.Bus` event and
   supports `intercept.Interceptor` hook points. No feature ships without
   observable extension points.
3. **Correctness over speed.** Fixed 20 Hz simulation prevents drift.
   I/O never blocks inside a tick.
4. **Testability.** Domain packages carry no external dependencies. All
   repositories are interfaces; in-memory implementations exist for tests.

---

## Service Topology

```
WebSocket Client
      │ ws://
      ▼
┌─────────────┐   NATS    ┌──────────────┐   NATS   ┌──────────────┐
│   gateway   │◄─────────►│     auth     │          │     game     │
│  (WS↔NATS) │           │ (handshake + │◄─────────►│  (ECS rooms) │
└─────────────┘           │    SSO)      │           └──────────────┘
                          └──────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                 ▼
       ┌──────────┐     ┌────────────┐     ┌──────────────┐
       │ social   │     │ navigator  │     │  moderation  │
       │(friends) │     │  (rooms)   │     │ (bans/kicks) │
       └──────────┘     └────────────┘     └──────────────┘
              │
       ┌──────────┐
       │ catalog  │
       │ (store)  │
       └──────────┘
```

All cross-service communication is over NATS JetStream. No service calls
another service over HTTP or gRPC.

---

## Services

| Service | Path | Responsibility |
|---|---|---|
| `gateway` | `services/gateway` | WebSocket upgrade; routes raw packets to NATS; routes NATS responses back to client |
| `auth` | `services/auth` | DH handshake (stub), SSO token validation, login bundle assembly |
| `game` | `services/game` | ECS room simulation; 20 Hz tick; packet dispatch; plugin system |
| `social` | `services/social` | Friends list, messenger |
| `navigator` | `services/navigator` | Room search and discovery |
| `catalog` | `services/catalog` | Store, purchases, economy |
| `moderation` | `services/moderation` | Bans, mutes, tickets |

---

## Domain Packages

| Package | Responsibility | External deps |
|---|---|---|
| `pkg/core/codec` | Binary `Reader`/`Writer` for the Pixel wire protocol | none |
| `pkg/core/config` | Viper-backed typed config loader | `spf13/viper` |
| `pkg/core/logging` | Zap logger factory | `go.uber.org/zap` |
| `pkg/core/bus` | NATS thin wrapper + infrastructure NATS subjects | `nats-io/nats.go` |
| `pkg/core/testutil` | Shared test helpers: `StartPostgres`, `StartRedis`, `StartNATS`, `MockSession` | `testcontainers-go` |
| `pkg/protocol` | **Generated** packet structs (do not edit manually) | `pkg/core/codec` |
| `pkg/pathfinding` | 3D A* router — pure computation | none |
| `pkg/plugin` | Plugin registry, loader | `go.uber.org/zap` |
| `pkg/plugin/event` | `EventBus` interface + implementation | none |
| `pkg/plugin/intercept` | `PacketInterceptor` interface + chain | `pkg/plugin/event` |
| `pkg/plugin/roomsvc` | Room service abstraction for plugin `.so` access | `pkg/plugin/event` |
| `pkg/user` | User domain types + repository interfaces | none |
| `pkg/room` | Room domain types, ECS components, Repository interface | `mlange-42/ark` |
| `pkg/item` | Item/furniture domain types + interfaces | none |
| `pkg/social` | Messenger, friends domain types + interfaces | none |
| `pkg/navigator` | Room discovery domain types + interfaces | none |
| `pkg/catalog` | Store, economy domain types + interfaces | none |
| `pkg/moderation` | Bans, tickets domain types + interfaces | none |

---

## NATS Subject Topology

| Subject prefix | Owner | Purpose |
|---|---|---|
| `pixel.gateway.*` | `pkg/core/bus` | Raw WebSocket packet forwarding (C2S and S2C) |
| `pixel.auth.*` | `services/auth` | Handshake and login events |
| `pixel.game.*` | `services/game` | Room goroutine input/output |
| `pixel.social.*` | `services/social` | Messenger and friends |
| `pixel.navigator.*` | `services/navigator` | Room search responses |
| `pixel.catalog.*` | `services/catalog` | Store and purchase results |
| `pixel.moderation.*` | `services/moderation` | Ban/mute/ticket results |

Each domain package owns its own subjects in a `subjects.go` file.
Infrastructure subjects live in `pkg/core/bus/subjects.go`.

---

## Plugin System

The plugin system (`pkg/plugin`) allows `.so` shared libraries to extend the
game service at runtime without modifying server internals.

### EventBus

```go
bus.Subscribe(eventName string, fn func(payload any))
bus.Publish(eventName string, payload any)
```

### PacketInterceptor

```go
interceptor.Before(headerID uint16, fn func(*intercept.PacketContext))
interceptor.After(headerID uint16, fn func(*intercept.PacketContext))
interceptor.RunBefore(ctx *PacketContext) bool  // returns false if cancelled
interceptor.RunAfter(ctx *PacketContext)
```

### Emitted Events (currently wired)

| Event name | Payload type | Emitted when |
|---|---|---|
| `event.PlayerJoined` | `event.PlayerJoinedPayload{SessionID, UserID}` | Player completes auth and enters the room |
| `event.PlayerLeft` | `event.PlayerLeftPayload{SessionID}` | Player disconnects or leaves |
| `event.PacketIn` | `event.PacketInPayload{SessionID, HeaderID, Payload}` | C2S packet is dispatched |

### Plugin Constraints

- Plugins may only import `pkg/plugin` and the Go standard library.
- Do not block or perform I/O inside an event handler or interceptor.
- Go's `plugin` package does not support runtime unloading.
- Plugins must be compiled with the same Go toolchain version as the host binary.

---

## Storage

| Store | Use |
|---|---|
| PostgreSQL 16 | All persistent domain data (users, rooms, items, social graph) |
| Redis 7 | Session cache, pub/sub fan-out, throttle counters |
| NATS JetStream | Service-to-service messaging; durable event log |

---

## Configuration

All services load typed config structs through `pkg/core/config` (Viper-backed).
The root `.env.example` is the canonical variable source. Config fields use
`mapstructure` and `env` tags:

```go
type Config struct {
    Host    string `mapstructure:"host" env:"GAME_HOST" default:"0.0.0.0"`
    Port    string `mapstructure:"port" env:"GAME_PORT" default:"3002"`
    NATSUrl string `mapstructure:"nats_url" env:"NATS_URL"` // required
}
```

Fields without `default` tags are required; startup fails with an error when
they are absent.

---

## Testing Strategy

| Level | Tag | Scope |
|---|---|---|
| Unit | (none) | Pure functions, codec round-trips, ECS logic, domain rules |
| Integration | `integration` | Repository methods against real PostgreSQL/Redis/NATS via testcontainers |
| End-to-End | `e2e` | Full client→gateway→auth→game flow; one scenario per phase exit |

Coverage targets: `pkg/core/*` and `pkg/user`/`pkg/room` ≥ 90 %; service
internal packages ≥ 80 %.

---

## Logging

All services use `pkg/core/logging` (Zap). Log format is configured via:

```
LOG_FORMAT=json   # structured JSON (production default)
LOG_FORMAT=pretty # human-readable (development default)
```

Log levels: `debug`, `info`, `warn`, `error` (default `info`). Packet-level
tracing is emitted at `debug` level only.
