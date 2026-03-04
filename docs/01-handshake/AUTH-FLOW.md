# Handshake — Authentication Flow

Detailed walkthrough of every step in the connection and authentication sequence,
from TCP connect through to the login bundle.

---

## Step 1 — WebSocket Upgrade (Gateway)

The gateway's `net.Listener` accepts a raw TCP connection.
`gobwas/ws` performs the HTTP→WebSocket upgrade.

- If upgrade fails: connection is closed, warning logged. Client sees HTTP 400.
- On success: a `Session` is created with a 32-char hex ID (`crypto/rand`).
- Session is added to `SessionStore`.
- A NATS subscription is opened on `session.output.<sessionID>` so the gateway
  can forward server-initiated response packets to the client.
- The read pump goroutine starts.

---

## Step 2 — release_version + client_variables (Auth)

The client sends `release_version` (4000) then `client_variables` (1053).

Both packets are forwarded by the gateway to `handshake.c2s.<sessionID>`.
Auth logs the version at `info` level and the variables at `debug`. No response.

These packets do not change any state — they are purely informational.

---

## Step 3 — Diffie-Hellman Exchange (Auth)

### init_diffie (3110 → 2512)

Client requests DH parameters. Auth responds with:

```
HandshakeInitDiffieOutPacket{
    EncryptedPrime:     "placeholder_prime",
    EncryptedGenerator: "placeholder_generator",
}
```

Packet is sent to `session.output.<sessionID>`. Gateway forwards it to the
WebSocket client.

### complete_diffie (773 → 1740)

Client sends its public key (ignored as a stub). Auth responds with:

```
HandshakeCompleteDiffieOutPacket{
    EncryptedPublicKey:       "placeholder_public_key",
    ServerClientEncryption:   false,
}
```

`ServerClientEncryption: false` tells the Nitro client not to encrypt traffic.
No actual cryptography is performed on either side.

---

## Step 4 — SSO Ticket Validation (Auth)

Client sends `sso_ticket` (2419).

Auth calls `TicketStore.Validate(ticket)`:

```
tickets map[string]int32
```

The `TicketStore` is a mutex-protected in-memory map. `Validate` removes the
ticket on first call — it cannot be replayed.

| Outcome | What happens |
|---|---|
| Ticket not found | Warning logged: `"invalid SSO ticket"`. No response to client. |
| Ticket valid | Proceed to post-auth burst (Step 5) |

---

## Step 5 — Post-Auth Burst (Auth → Client)

Three packets are published in order to `session.output.<sessionID>`:

1. `authentication.ok` (2491) — empty acknowledgement
2. `availability_status` (2033) — `{IsOpen:true, OnShutdown:false, IsAuthentic:true}`
3. `client_ping` (3928) — empty, starts pong cycle

---

## Step 6 — session.authenticated (Auth → NATS)

After the burst, Auth publishes to `session.authenticated`:

```
payload: WriteString(sessionID) + WriteInt32(userID)
```

Encoded with `pkg/core/codec.Writer`.

**Gateway** receives this event and:
1. Sets `session.authenticated = true`  
2. Stores `userID` on the session struct

Future post-auth packets from this client will now be routed to
`room.input.<sessionID>` instead of being dropped.

---

## Step 7 — Login Bundle (Game)

The game service's `Listener.handleAuthenticated` receives the same
`session.authenticated` NATS message.

1. Creates a `natsSession{sessionID, userID}` and adds it to the sessions map
2. Emits `event.PlayerJoined` on the plugin event bus
3. Calls `identity.SendLoginBundle` which sends 9 packets to the client

The 9 login bundle packets complete the connection setup. See the
[User-Profile](../USER-PROFILE/LOGIN-BUNDLE.MD) documentation for detail.

---

## NATS Subjects Summary

| Subject | Publisher | Subscriber | When |
|---|---|---|---|
| `handshake.c2s.<sessionID>` | Gateway | Auth | Every pre-auth C2S packet |
| `session.output.<sessionID>` | Auth | Gateway | Every response packet to client |
| `session.authenticated` | Auth | Gateway + Game | SSO ticket validated successfully |

---

## TicketStore — In-Memory SSO Store

`services/auth/ticket.go`:

```go
type TicketStore struct {
    mu      sync.Mutex
    tickets map[string]int32   // ticket → userID
}

func (ts *TicketStore) Register(ticket string, userID int32)
func (ts *TicketStore) Validate(ticket string) (int32, bool) // deletes on hit
```

Tickets must be pre-registered before the client connects (e.g. by a login API
that is out of scope for this service). In production this becomes a Redis store
with TTL expiry. The interface is identical.
