# Handshake — Overview

The **handshake-security** realm is the first thing a connecting client
encounters. Its sole purpose is to prove who the client is before any
game state is touched.

A Nitro client opens a WebSocket to the gateway, exchanges a series of packets
to negotiate encryption parameters (Diffie-Hellman stubs), and finally submits
an SSO ticket. If the ticket is valid the server sends a three-packet post-auth
burst and publishes a NATS event so downstream services can prepare the session.

---

## Documents in this realm

| File | What it covers |
|---|---|
| [PACKETS.MD](PACKETS.MD) | Every C2S and S2C packet: header ID, fields, handler summary |
| [AUTH-FLOW.MD](AUTH-FLOW.MD) | Step-by-step walkthrough of the full authentication sequence |
| [ERROR-HANDLING.MD](ERROR-HANDLING.MD) | Every error path, what is logged, what reaches the client |

---

## End-to-end sequence

```
Client                  Gateway                Auth                 Game
  │                       │                      │                    │
  │──── WS Upgrade ──────►│                      │                    │
  │◄─── WS Accept ────────│                      │                    │
  │                       │                      │                    │
  │ release_version (4000)│                      │                    │
  │──────────────────────►│── handshake.c2s.* ──►│                    │
  │                       │                      │  (log, no reply)   │
  │                       │                      │                    │
  │ client_variables(1053)│                      │                    │
  │──────────────────────►│── handshake.c2s.* ──►│                    │
  │                       │                      │  (debug log only)  │
  │                       │                      │                    │
  │ init_diffie (3110)    │                      │                    │
  │──────────────────────►│── handshake.c2s.* ──►│                    │
  │◄──── init_diffie ─────│◄─ session.output.* ──│                    │
  │      (2512)           │                      │                    │
  │                       │                      │                    │
  │ complete_diffie (773) │                      │                    │
  │──────────────────────►│── handshake.c2s.* ──►│                    │
  │◄── complete_diffie ───│◄─ session.output.* ──│                    │
  │      (1740)           │                      │                    │
  │                       │                      │                    │
  │ sso_ticket (2419)     │                      │                    │
  │──────────────────────►│── handshake.c2s.* ──►│                    │
  │                       │                      │  validate ticket   │
  │◄── auth.ok (2491) ────│◄─ session.output.* ──│                    │
  │◄── avail_status(2033) │◄─ session.output.* ──│                    │
  │◄── client_ping (3928) │◄─ session.output.* ──│                    │
  │                       │                      │                    │
  │                       │◄── session.authenticated ────────────────►│
  │                       │    (string sessionID + int32 userID)       │
  │                       │                      │                    │
  │                       │                      │      login bundle  │
  │◄── user_info ─────────│◄─ session.output.* ◄────────────────────-│
  │◄── permissions ───────│   …(9 packets)        │                    │
```

---

## Services involved

| Service | Role |
|---|---|
| `services/gateway` | Accepts WebSocket connections; classifies and forwards pre-auth packets to `handshake.c2s.<sessionID>` via NATS |
| `services/auth` | Owns the `Handler` that processes handshake packets; validates SSO tickets; sends post-auth burst; publishes `session.authenticated` |
| `services/game` | Subscribes to `session.authenticated`; builds the login bundle (Phase 2) |

---

## Key design decisions

- **Diffie-Hellman is stubbed.** The exchange happens but placeholder strings are
  used. Real cryptography is deferred. `ServerClientEncryption` is always `false`.
- **Tickets are one-time use.** The in-memory `TicketStore` deletes a ticket on
  first successful `Validate()` call.
- **No repeated auth.** Once the gateway marks a session as authenticated the
  pre-auth header IDs are no longer routed to `handshake.c2s.*`, so replaying
  a ticket has no effect.
