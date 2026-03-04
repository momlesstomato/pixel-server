# Session — Lifecycle

How a session is born, lives, and dies.

---

## Creation

```
TCP connection accepted by net.Listener
        │
        ▼
WebSocket upgrade (gobwas/ws)
        │  failure → warn log, connection closed
        ▼
Session{
    ID:             <32-char hex>   // crypto/rand, 16 bytes
    conn:           net.Conn        // underlying WebSocket connection
    authenticated:  atomic.Bool     // false
    createdAt:      time.Now()
    writeMu:        sync.Mutex
    done:           make(chan struct{})
}
        │
        ▼
SessionStore.Add(sess)              // concurrent-safe RWMutex map
        │
        ▼
NATS subscription on session.output.<sessionID>
        │  failure → error log, but session continues (no outbound packets)
        ▼
goroutine: read pump started
```

---

## Session struct

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | 32-character hex session identifier |
| `conn` | `net.Conn` | Underlying WebSocket connection |
| `userID` | `int32` | Set upon `session.authenticated` event; 0 until then |
| `authenticated` | `atomic.Bool` | True once auth service confirms the ticket |
| `createdAt` | `time.Time` | Session creation timestamp |
| `machineID` | `string` | Client hardware ID (set by `machine_id` packet) |
| `writeMu` | `sync.Mutex` | Serialises concurrent WebSocket frame writes |
| `done` | `chan struct{}` | Closed once to signal goroutines to exit |

---

## Read pump

The read pump is a tight loop:
```
for {
    read WebSocket frame
    if error (EOF or read error) → close session
    else → routePacket(frame)
}
```

`routePacket` classifies by header ID:

| Condition | Action |
|---|---|
| Header in pre-auth set (4000, 1053, 3110, 773, 2419, 2490, 1735) | Publish to `handshake.c2s.<sessionID>` |
| Header 2596 (pong) | Silently acknowledge; nothing published |
| Header 295 (latency test) | Write echo directly to WebSocket: `header=10 + int32(0)` |
| Header 2445 (client disconnect) | Close session |
| Any other header, authenticated | Publish to `room.input.<sessionID>` |
| Any other header, not authenticated | Drop; warn log |

---

## Authentication gate

When `session.authenticated` arrives from the auth service:

1. `session.authenticated.Store(true)` — atomic flag
2. `session.userID = userID` — from the NATS payload

After this point all non-pre-auth packets reach the game service. Before this
point they are silently dropped.

---

## Keep-alive

| Direction | Packet | Header | Handling |
|---|---|---|---|
| S→C | `client_ping` | 3928 | Auth sends this as part of the post-auth burst |
| C→S | `client_pong` | 2596 | Gateway acknowledges silently; nothing forwarded |

There is no server-side pong timeout. Dead connections are detected by TCP-level
errors on the next read attempt.

---

## Latency measurement

When the client sends `latency_test` (header 295):

1. Gateway writes a WebSocket frame with `uint16(10) + int32(0)` directly back
2. No NATS publish — the echo is done inline in the read pump

---

## Disconnect

Sessions can close through any of these paths:

| Trigger | Notes |
|---|---|
| Client sends `client_disconnect` (2445) | Intentional; handled in `routePacket` |
| WebSocket read error | EOF (normal close) is logged at debug; other errors at debug |
| Server `context.Cancel` during shutdown | Listener closes, all sessions receive EOF |

**Disconnect sequence** (always the same, regardless of trigger):

```
1. select on done channel → already closed? → skip (idempotent)
2. close(done)
3. SessionStore.Remove(sessionID)
4. Unsubscribe NATS session.output.<sessionID>
5. Publish session.disconnected  (raw sessionID bytes)
6. Close WebSocket connection
```

`session.disconnected` payload is the raw session ID bytes (`[]byte(sessionID)`).

---

## SessionStore

In-memory concurrent map:

```go
type SessionStore struct {
    mu   sync.RWMutex
    sessions map[string]*Session
}

func (s *SessionStore) Add(sess *Session)
func (s *SessionStore) Get(id string) (*Session, bool)
func (s *SessionStore) Remove(id string)
func (s *SessionStore) Count() int
```

No external storage. A gateway restart loses all in-flight sessions.

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:2096` | WebSocket bind address |
| `NATS_URL` | `nats://127.0.0.1:4222` | NATS server URL |
| `LOG_FORMAT` | `pretty` | `"json"` or `"pretty"` |
| `LOG_LEVEL` | `info` | `"debug"` / `"info"` / `"warn"` / `"error"` |
