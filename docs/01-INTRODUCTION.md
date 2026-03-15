# Introduction

## What is Pixel Server?

Pixel Server (`pixelsv`) is a distributed, high-performance implementation of
the [Pixel Protocol](https://momlesstomato.github.io/pixel-protocol/) — the
binary WebSocket protocol used by Habbo Hotel clients (Nitro). It is designed
from the ground up for multi-instance deployment, horizontal scalability, and
extensibility via a plugin SDK.

## Goals

1. **Faithful protocol implementation** — full compatibility with the
   [pixel-protocol specification](https://momlesstomato.github.io/pixel-protocol/),
   covering all packet realms from handshake through gameplay.

2. **Distribution-first** — every design decision assumes multiple server
   instances sharing Redis and PostgreSQL. Session state, hotel broadcasts, and
   connection lifecycle work identically whether running one instance or twenty.

3. **Performance** — native Go binary with zero-copy packet codec, RC4 stream
   encryption, and per-connection goroutine model. Redis Pub/Sub for
   cross-instance messaging with sub-millisecond latency.

4. **Extensibility** — plugin system via Go shared objects (`.so`) with a
   separate SDK module. Typed event bus with priority ordering and cancellation
   allows plugins to intercept, modify, or block any server behavior.

5. **Operational simplicity** — single binary (`pixelsv`) with CLI subcommands
   for serving, database migrations, and SSO token management.

## Prior Art

Pixel Server is inspired by and builds upon the work of several existing Habbo
server implementations:

| Project | Language | Role in our design |
|---------|----------|-------------------|
| Gladiator | Java | Reference plugin system (143 event types, priority-based dispatch, JAR loading). Our event bus follows the same LOWEST→MONITOR priority model. |
| Sodium | C# | Reference DI-based architecture and packet handler patterns. Influenced our hexagonal adapter design. |
| Galaxy | Java | Session management patterns and heartbeat implementation. |

### What We Do Differently

- **No monolithic god objects** — vendors expose `GameClientManager`,
  `RoomManager`, and similar classes that give unrestricted access to everything.
  We use narrow, versioned API interfaces.

- **Redis-backed session state** — vendors use in-memory maps, making
  multi-instance impossible. Our session registry, hotel status, and SSO tokens
  all live in Redis.

- **Cross-instance coordination** — vendors assume a single process. We use
  Redis Pub/Sub for broadcasts, targeted sends, and connection close signals
  across instances.

- **Go plugin SDK** — vendors use Java classloaders or C# DI. We provide a
  zero-dependency Go module that plugin authors import independently.

- **Session leases** — Redis keys have TTL with periodic refresh. If an
  instance crashes, orphan sessions expire automatically. No vendor handles this.

## Quick Start

```bash
# Build
go build -o pixelsv ./cmd/pixel-server

# Run with required environment
export APP_API_KEY="your-secret-key"
export REDIS_ADDRESS="localhost:6379"
export POSTGRES_DSN="postgres://user:pass@localhost:5432/pixel?sslmode=disable"

./pixelsv serve
```

The server starts on `0.0.0.0:3000` by default. WebSocket connections are
accepted at `/ws`. The REST API requires the `X-API-Key` header.

## Documentation Structure

| Document | Contents |
|----------|----------|
| [02-ARCHITECTURE](02-ARCHITECTURE.md) | System design, hexagonal boundaries, distribution model |
| [03-FOUNDATIONS](03-FOUNDATIONS.md) | Core infrastructure: Fiber, WebSocket, Redis, codec, logging |
| [04-CONFIGURATION](04-CONFIGURATION.md) | Complete environment variable reference |
| [01-session/](01-session/) | Authentication flow, SSO tokens, session lifecycle |
| [02-management/](02-management/) | Hotel status, maintenance, broadcasts, moderation |
| [03-plugins/](03-plugins/) | Plugin system, SDK, event bus, handler lifecycle |
| [04-user/](04-user/) | User identity, settings, social features, name changes |
| [05-permissions/](05-permissions/) | Permission groups, resolution, client perks |
