# Messaging Contracts (Transport Adapter Pattern)

## Goal

Keep realm module boundaries decoupled while supporting both single-process and distributed deployment.

## Transport Adapter Pattern

Inter-module communication flows through **port interfaces**. The concrete transport is injected at startup based on the deployment mode:

| Mode | Transport | Latency | When |
|---|---|---|---|
| All-in-one (`--role=all`) | In-process Go channels + direct calls | ~10-50 us | Dev, small deploy |
| Distributed (`--role=X`) | NATS JetStream | ~2-5 ms | Production at scale |

Domain code never imports transport libraries. It depends only on port interfaces:

```go
// Defined by consumer realm
type EventPublisher interface {
    Publish(ctx context.Context, topic string, payload []byte) error
}

type EventSubscriber interface {
    Subscribe(ctx context.Context, topic string, handler Handler) (Subscription, error)
}

type Handler func(ctx context.Context, message Message) error

type Message struct {
    Topic string
    Payload []byte
}

type SessionWriter interface {
    Send(sessionID string, data []byte) error
}
```

## Local Transport (In-Process)

```go
// pkg/core/transport/local/bus.go

type Bus struct {
    handlers map[string][]func(any)
    mu       sync.RWMutex
}

func (b *Bus) Publish(topic string, payload any) error {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, h := range b.handlers[topic] {
        h(payload)
    }
    return nil
}

func (b *Bus) Subscribe(topic string, handler func(any)) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers[topic] = append(b.handlers[topic], handler)
    return nil
}
```

## NATS Transport (Distributed)

```go
// pkg/core/transport/nats/bus.go

type Bus struct {
    conn *nats.Conn
}

func (b *Bus) Publish(topic string, payload any) error {
    data, err := encode(payload)
    if err != nil { return err }
    return b.conn.Publish(topic, data)
}

func (b *Bus) Subscribe(topic string, handler func(any)) error {
    _, err := b.conn.Subscribe(topic, func(msg *nats.Msg) {
        payload, _ := decode(msg.Data)
        handler(payload)
    })
    return err
}
```

## Contract Topics

Topic names are stable regardless of transport. They are defined as constants:

```
handshake.c2s.<sessionID>         <- gateway -> auth
session.authenticated             <- auth -> game, social, navigator
session.disconnected              <- gateway -> game, social
room.input.<roomID>               <- gateway -> game worker
session.output.<sessionID>        <- game worker -> gateway
social.notification.<userID>      <- social -> gateway
navigator.room_updated.<roomID>   <- game -> navigator
catalog.purchase.completed        <- catalog -> game
moderation.ban.issued.<userID>    <- moderation -> gateway
```

## Envelope Rules

- Use typed Go structs for application/domain envelopes.
- Avoid raw `map[string]any` payloads.
- Include correlation metadata for tracing (`sessionID`, `userID`, `roomID`, `requestID`).
- In local transport, envelopes are passed by reference (zero-copy).
- In NATS transport, envelopes are serialized to binary (codec-encoded or JSON).

## Delivery and Idempotency

- Domain handlers must be idempotent for retry-safe operations.
- Critical side effects (purchases, balance changes) require transactional guarantees at adapter layer.
- Non-critical fan-out events are best-effort.

## Backpressure

- Each channel must define capacity explicitly.
- Overflow policies must be explicit (`drop`, `retry`, `reject`) per topic.
- Gateway must protect game loop integrity by rejecting abusive ingress rates.
- Room worker inbox channels have bounded capacity; overflow triggers client-visible backpressure (packet drop or disconnect).

## Wiring Example

```go
// cmd/pixelsv (simplified)

func buildTransport(cfg Config) (EventPublisher, EventSubscriber, SessionWriter) {
    if cfg.Role == "all" {
        bus := local.NewBus()
        sw := local.NewSessionWriter(sessionStore)
        return bus, bus, sw
    }
    nc, _ := nats.Connect(cfg.NATS.URL)
    bus := natst.NewBus(nc)
    sw := natst.NewSessionWriter(nc)
    return bus, bus, sw
}
```

## External Broker Policy

- NATS is the only supported external broker.
- It activates automatically when roles run in separate processes.
- Keep it behind adapter ports — no domain code imports `nats.go`.
- Preserve topic contracts exactly.
- Do not move domain logic into transport handlers.
