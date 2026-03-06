# TRANSPORT

## Overview

Runtime modules communicate through `pkg/core/transport` ports.

- `transport.EventPublisher`
- `transport.EventSubscriber`
- `transport.Bus`

Topic contracts are stable constants and helpers in:

- `pkg/core/transport/topics.go`

## Implemented Adapters

- Local bus: `pkg/core/transport/local`
- NATS bus: `pkg/core/transport/nats`
- Selector factory: `pkg/core/transport/factory`

Factory behavior:

- `NATS_URL` empty: local bus
- `NATS_URL` set and role split: NATS bus
- `--role=all`: force local bus

## Topic Contract

Implemented topic constants include:

- `handshake.c2s.<sessionID>`
- `session.authenticated`
- `session.disconnected`
- `room.input.<roomID>`
- `session.output.<sessionID>`
- `social.notification.<userID>`
- `navigator.room_updated.<roomID>`
- `catalog.purchase.completed`
- `moderation.ban.issued.<userID>`

## Performance Direction

- Local bus targets minimal publish overhead for 20Hz runtime loops.
- Local wildcard matching supports NATS-like `*` and `>` semantics.
- NATS adapter preserves the same topic contract for distributed mode.
- Local bus benchmark exists in `pkg/core/transport/local/bus_benchmark_test.go`.

## Usage

```go
roles, _ := newRoleSet("all")
bus, _ := factory.New(factory.Config{NATSURL: "", ForceLocal: roles.forceLocalTransport()})
defer bus.Close()
```

Runtime startup wiring is in `pkg/core/cli/startup.go`.
