# Session — Packets

The session layer owns a small set of packets that are handled entirely within
the gateway, never forwarded to a backend service.

---

## Client → Server (C2S)

### client_pong — header 2596

*(No fields — empty body.)*

Keep-alive response. Sent periodically by the client after receiving a
`client_ping` (3928) from the server. The gateway acknowledges it silently;
nothing is published to NATS.

---

### latency_test — header 295

*(No fields — empty body.)*

The client sends this to measure round-trip latency. The gateway immediately
echoes `latency_response` (header 10 + `int32(0)`) back on the same WebSocket
connection without touching NATS.

---

### client_disconnect — header 2445

*(No fields — empty body.)*

The client signals it is intentionally disconnecting. The gateway initiates the
disconnect sequence. See [LIFECYCLE.MD](LIFECYCLE.MD) for the full sequence.

---

## Server → Client (S2C)

### client_ping — header 3928

*(No fields — empty body.)*

Published by `services/auth` to `session.output.<sessionID>` as the third
packet in the post-auth burst. The gateway forwards it to the WebSocket client.

This starts the keep-alive cycle: client is expected to respond with `client_pong`
(2596) periodically.

---

### availability_status — header 2033

| Field | Type | Value |
|---|---|---|
| `IsOpen` | `bool` | `true` — hotel is open |
| `OnShutdown` | `bool` | `false` — not shutting down |
| `IsAuthentic` | `bool` | `true` — this is an authentic server |

Published by `services/auth` as the second packet in the post-auth burst.

---

### latency_response — header 10

| Field | Type | Value |
|---|---|---|
| *(unnamed)* | `int32` | Always `0` |

Sent by the gateway inline (no NATS) in response to `latency_test` (295).

---

## NATS subjects (session layer)

| Subject | Publisher | Subscriber | Payload encoding |
|---|---|---|---|
| `session.output.<sessionID>` | Auth, Game | Gateway | `uint16 headerID` + encoded packet body |
| `session.authenticated` | Auth | Gateway, Game | `WriteString(sessionID)` + `WriteInt32(userID)` |
| `session.disconnected` | Gateway | Game | Raw `[]byte(sessionID)` |
| `handshake.c2s.<sessionID>` | Gateway | Auth | `WriteString(sessionID)` + `WriteUint16(headerID)` + `WriteBytes(payload)` |
| `room.input.<sessionID>` | Gateway | Game | `WriteString(sessionID)` + `WriteUint16(headerID)` + `WriteBytes(payload)` |
