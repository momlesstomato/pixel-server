# Messaging Contracts (Internal-First)

## Goal

Keep module boundaries decoupled while preserving single-binary deployment.

## Default Transport

- In-process event bus/queues are the default transport.
- Contracts are defined as typed envelopes and topic names.
- External brokers are optional adapters and must not leak into domain logic.

## Contract Topics

```
handshake.c2s.<sessionID>
session.authenticated
session.disconnected
room.input.<roomID>
session.output.<sessionID>
social.notification.<userID>
navigator.room_updated.<roomID>
catalog.purchase.completed
moderation.ban.issued.<userID>
```

Topic names remain stable even if transport changes.

## Envelope Rules

- Use typed Go structs for application/domain envelopes.
- Avoid raw `map[string]any` payloads.
- Include correlation metadata for tracing (`sessionID`, `userID`, `roomID`, `requestID`).

## Delivery and Idempotency

- Domain handlers must be idempotent for retry-safe operations.
- Critical side effects (purchases, balance changes) require transactional guarantees at adapter layer.
- Non-critical fan-out events are best-effort.

## Backpressure

- Each queue/channel must define capacity explicitly.
- Overflow policies must be explicit (`drop`, `retry`, `reject`) per topic.
- Gateway must protect game loop integrity by rejecting abusive ingress rates.

## External Broker Policy

If an external broker is introduced later:

- Keep it behind adapter ports.
- Preserve topic contracts.
- Do not move domain logic into transport handlers.
