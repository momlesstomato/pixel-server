# Inter-Service Messaging with NATS JetStream

## Why NATS JetStream?

Alternatives considered:

| Option | Reason rejected |
|---|---|
| Redis Streams | Works, but Redis is already the session store; mixing durable messaging with cache complicates ops |
| Kafka | Operationally heavy for this scale; JVM dependency; per-topic partition pre-allocation overhead |
| RabbitMQ | Weaker replay semantics; AMQP protocol adds latency for game-critical paths |
| gRPC direct | Couples services; requires service discovery; no buffer on burst; harder to scale game workers independently |
| In-process channels | Not viable once services run as separate pods |

NATS JetStream provides:
- **At-least-once delivery** with consumer acknowledgements.
- **Durable subscriptions** that survive pod restarts.
- **Fan-out and work-queue** patterns from the same infrastructure.
- **Sub-millisecond latency** for in-cluster pub/sub.
- **Golang-native client** (`nats.go`) with JetStream API.
- **Single binary deployment** — `nats-server` is 20 MB, no JVM, no ZooKeeper.

---

## Stream topology

### Streams (JetStream persistent subjects)

| Stream name | Subjects | Consumers | Retention |
|---|---|---|---|
| `HANDSHAKE` | `handshake.c2s.>` | auth-svc (work-queue) | Work (deleted on ack) |
| `SESSION_EVENTS` | `session.>` | gateway, game, social | Interest (delivered to all) |
| `ROOM_INPUT` | `room.input.>` | game-svc (work-queue, 1 per room) | Work |
| `SESSION_OUTPUT` | `session.output.>` | gateway (per session) | Work |
| `SOCIAL_EVENTS` | `social.>` | social-svc, gateway | Interest |
| `CATALOG_EVENTS` | `catalog.>` | catalog-svc, inventory, game | Interest |
| `NAVIGATOR_SYNC` | `navigator.>` | navigator-svc | Interest |

### Subjects reference

```
# Pre-auth packets (gateway → auth)
handshake.c2s.<sessionID>

# Auth results (auth → all)
session.authenticated          payload: SessionAuthenticatedEvent{sessionID, userID, ...}
session.disconnected           payload: SessionDisconnectedEvent{sessionID, reason}

# In-room packets (gateway → game)
room.input.<roomID>            payload: RoomInputEnvelope{sessionID, headerID, rawPayload}

# Outbound frames (game → gateway)
session.output.<sessionID>     payload: []byte  (raw framed packet, ready to write to socket)

# Social
social.friend_request.<userID>
social.friend_accepted.<userID>
social.notification.<userID>   payload: NotificationEvent{type, message, …}
social.message.<conversationID>

# Navigator
navigator.room_updated.<roomID>  payload: RoomMetaEvent{name, userCount, score, …}
navigator.room_deleted.<roomID>

# Catalog / economy
catalog.purchase_completed     payload: PurchaseEvent{userID, offerID, items[]}
inventory.updated.<userID>     payload: InventoryUpdateEvent{items[]}

# Moderation
moderation.ban_issued.<userID>   (triggers Redis PUBLISH for immediate disconnect)
moderation.ban_lifted.<userID>
```

---

## Message envelope format

All NATS messages use **protobuf-encoded** envelopes (not JSON) for the hot path. The protobuf schemas live in `pkg/bus/events/`. JSON is available for debugging via NATS CLI but is not used in production payloads.

For `session.output.*` (raw packet frames), the payload is already a binary-framed Habbo packet; no additional envelope wrapping.

---

## Consumer patterns

### Work-queue (one processor per message)

Used for: `HANDSHAKE`, `ROOM_INPUT`, `SESSION_OUTPUT`.

```go
js, _ := nats.Connect(natsURL)
js.Subscribe("room.input.*", func(msg *nats.Msg) {
    envelope := decodeEnvelope(msg.Data)
    supervisor.Route(envelope)
    msg.Ack()
}, nats.Durable("game-room-input"), nats.ManualAck())
```

### Fan-out (all consumers receive each message)

Used for: `SESSION_EVENTS`, `SOCIAL_EVENTS`.

Each service creates a named durable consumer with its own delivery subject:

```
consumer: gateway-session-events   → listens on SESSION_EVENTS
consumer: game-session-events      → listens on SESSION_EVENTS
consumer: social-session-events    → listens on SESSION_EVENTS
```

### Room routing (consistent hash)

`ROOM_INPUT` subjects are scoped by roomID: `room.input.<roomID>`. Within a game-svc pod, the supervisor routes by roomID to the correct room goroutine. Across pods, the gateway resolves which pod owns a given room via Redis key `room:node:<roomID>` → `<podName>`.

If the responsible pod is unavailable, the gateway holds the message in the NATS stream; when a new pod picks up the room it drains the backlog.

---

## Delivery guarantees per use case

| Use case | Required guarantee | Mechanism |
|---|---|---|
| Login / SSO | Exactly-once (idempotent) | auth-svc deduplicates by sessionID in Redis |
| Walk command | At-least-once (idempotent) | path recalculated on re-delivery; no visible effect |
| Chat message | At-least-once | client-side dedup by messageID prevents double display |
| Item purchase | Exactly-once | catalog-svc uses PostgreSQL transaction + NATS ack |
| Room packet broadcast | Best-effort | outbound `session.output` is fire-and-forget once framed |
| Ban enforcement | At-least-once + Redis PUBLISH | Redis PUBLISH ensures socket is closed immediately |

---

## Backpressure and flow control

Each NATS consumer sets `MaxAckPending` to control concurrent in-flight messages:

- `ROOM_INPUT` per room: `MaxAckPending=64` — prevents room inbox overflow.
- `HANDSHAKE`: `MaxAckPending=256` auth-svc capacity.
- `SESSION_OUTPUT`: `MaxAckPending=128` — gateway drains outbound per session.

When a consumer hits `MaxAckPending`, NATS pauses delivery automatically. The producer (gateway) sees that its `session.output` publish blocks and can start dropping non-critical packets (e.g. typing-indicator updates) in favour of critical ones (e.g. disconnect notifications).

---

## Observability

NATS JetStream exposes metrics via its HTTP monitoring port. These are scraped by Prometheus and displayed in Grafana:

- Consumer lag (`num_pending`) per stream — detects build-up in `ROOM_INPUT` stream.
- Message delivery rate per subject prefix.
- Ack latency histogram.

A high `num_pending` on `ROOM_INPUT.<roomID>` means a room goroutine is stalling — alerts trigger a postmortem.
