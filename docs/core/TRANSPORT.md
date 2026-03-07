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

Implemented topic constants are split by realm packages:

- generic packet ingress:
  - `packet.c2s.<realm>.<sessionID>` in `pkg/core/transport`
- auth realm:
  - `internal/auth/messaging`
- session-connection realm:
  - `internal/sessionconnection/messaging`
  - `session.connected`
  - `session.disconnected`
  - `session.output.<sessionID>`
  - `session.disconnect.<sessionID>`

## Performance Direction

- Local bus targets minimal publish overhead for 20Hz runtime loops.
- Local wildcard matching supports NATS-like `*` and `>` semantics.
- NATS adapter preserves the same topic contract for distributed mode.
- Local bus benchmark exists in `pkg/core/transport/local/bus_benchmark_test.go`.
- WebSocket ingress decode/publish benchmark exists in `pkg/http/ws/benchmark_test.go`.

Current baseline (`go test -run '^$' -bench BenchmarkHandleBinary -benchmem ./pkg/http/ws`):

- `~232 ns/op`
- `352 B/op`
- `9 allocs/op`

## Usage

```go
roles, _ := newRoleSet("all")
bus, _ := factory.New(factory.Config{NATSURL: "", ForceLocal: roles.forceLocalTransport()})
defer bus.Close()
```

Runtime startup wiring is in `pkg/core/cli/startup.go`.

Topic parsing helpers:

- `transport.ParsePacketC2STopic(topic string) (realm string, sessionID string, ok bool)`
- `sessionconnection/messaging.ParseOutputTopic(topic string) (sessionID string, ok bool)`
- `sessionconnection/messaging.ParseDisconnectTopic(topic string) (sessionID string, ok bool)`
