# Session — Overview

The **session-connection** realm manages the transport layer that keeps a client
connected throughout its time on the server. Every Nitro client gets a unique
session the moment its WebSocket connection is accepted.

The gateway service owns all sessions. It is the only process that ever touches
WebSocket frames directly — every other service communicates with the client
exclusively through NATS.

---

## Documents in this realm

| File | What it covers |
|---|---|
| [LIFECYCLE.MD](LIFECYCLE.MD) | Session creation, keep-alive, disconnect, authentication gate |
| [PACKETS.MD](PACKETS.MD) | C2S and S2C packets managed by the session layer |
| [PLUGIN-HOOKS.MD](PLUGIN-HOOKS.MD) | Plugin events, packet interceptors, realm relations, permissions |

---

## Responsibilities

- Accept WebSocket connections and generate unique session IDs
- Forward pre-auth packets to `services/auth` via NATS
- Forward post-auth packets to `services/game` via NATS
- Deliver server-initiated packets from any backend service to the correct client
- Track authentication state and enforce the pre-auth / post-auth gate
- Handle keep-alive (ping/pong) and latency measurement without forwarding
- Publish `session.disconnected` when a client leaves

---

## Services involved

| Service | Role |
|---|---|
| `services/gateway` | Owns the `SessionStore`, `Session` struct, WebSocket read pump, `routePacket` |
| `services/auth` | Subscribes to `session.authenticated` subject it publishes itself |
| `services/game` | Subscribes to `session.disconnected` to remove sessions from its map |

---

## Session identity

Each `Session` has a **32-character lowercase hex string** ID generated from
`crypto/rand` (16 bytes → hex). This ID is stable for the lifetime of the
connection and is used as the routing key in all NATS subjects.

```
session.output.6f2a8b3c1d4e5f6a7b8c9d0e1f2a3b4c   ← outbound to client
room.input.6f2a8b3c1d4e5f6a7b8c9d0e1f2a3b4c       ← inbound post-auth packets
handshake.c2s.6f2a8b3c1d4e5f6a7b8c9d0e1f2a3b4c   ← inbound pre-auth packets
```
