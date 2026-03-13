# Foundations

This document covers the core infrastructure components that every realm builds
upon. Understanding these foundations is essential before working with any
specific realm.

## Go & Module Structure

The project uses **Go 1.25.5** as a single module (`pixelsv`) with a Go
workspace (`go.work`) that includes the plugin SDK:

```
pixel-server/
  go.work          workspace: use . and ./sdk
  go.mod           server module
  sdk/go.mod       plugin SDK module (zero dependencies)
```

The binary is built from `cmd/pixel-server/main.go` and produces a single
executable called `pixelsv`.

## Fiber HTTP Framework

The HTTP layer uses [Fiber v3](https://gofiber.io/) — a high-performance web
framework built on `fasthttp`.

### Module

The `core/http` package wraps Fiber into a `Module` struct that manages:

- Fiber app lifecycle (create, listen, shutdown)
- API key authentication middleware
- WebSocket endpoint registration
- Active connection tracking for graceful shutdown

```go
module := http.New(http.Options{
    Logger:    logger,
    APIKey:    "secret",
    APIKeyHeader: "X-API-Key",
})

module.RegisterGET("/health", healthHandler)
module.RegisterWebSocket("/ws", wsHandler)
```

### API Key Protection

All HTTP routes are protected by API key middleware except:

- WebSocket upgrade paths (authenticated via SSO after upgrade)
- `/openapi.json` and `/swagger` (documentation endpoints)

The middleware validates the `X-API-Key` header (or `api_key` query parameter)
using constant-time comparison to prevent timing attacks.

### OpenAPI & Swagger

The server auto-generates an OpenAPI 3.1.0 specification and serves Swagger UI:

- `GET /openapi.json` — machine-readable spec
- `GET /swagger` — interactive Swagger UI

Endpoints are registered dynamically as routes are added. The WebSocket
endpoint is documented as `GET /ws → 101 Switching Protocols`.

## WebSocket Transport

WebSocket connections use `gofiber/contrib/websocket` built on `gorilla/websocket`.

### Connection Lifecycle

```
HTTP GET /ws
  → Fiber upgrade middleware (checks WebSocket headers)
  → WebSocket handler receives *websocket.Conn
  → Handler generates connID (32 random hex chars)
  → Creates Transport wrapping the connection
  → Subscribes to Redis close bus for remote close signals
  → Starts auth timeout (30s)
  → Enters read loop (processes binary frames)
  → On close: cleanup session, dispose resources
```

### Binary Protocol

All WebSocket communication uses **binary frames** (not text). The Habbo
protocol wire format is:

```
┌────────────────┬──────────────┬────────────────────┐
│ Length (4B BE)  │ Packet ID    │ Body (variable)    │
│ uint32         │ (2B BE)      │                    │
│                │ uint16       │                    │
└────────────────┴──────────────┴────────────────────┘
```

The length field covers packet ID + body (does not include itself).

### Graceful Shutdown

When the server shuts down (SIGTERM/SIGINT), for each active WebSocket:

1. Send `disconnect.reason` packet (ID 4000, reason 19 = shutdown)
2. Send WebSocket close frame (code 1001 = Going Away)
3. Close the underlying connection
4. Wait up to 2 seconds for completion

## Protocol Codec

The `core/codec` package provides frame encoding/decoding and typed primitives.

### Frame Operations

```go
// Encode a packet
frame := codec.EncodeFrame(packetID, body)

// Decode a single frame
f, consumed, err := codec.DecodeFrame(data)
// f.PacketID, f.Body

// Decode concatenated frames (buffered reads)
frames, err := codec.DecodeFrames(data)
```

### Primitive Types

The `Reader` and `Writer` handle Habbo protocol types:

```go
// Reading packet body
reader := codec.NewReader(body)
version, _ := reader.ReadString()   // uint16 length-prefixed UTF-8
platform, _ := reader.ReadInt32()   // 4 bytes big-endian
enabled, _ := reader.ReadBool()     // 1 byte (0 or 1)
remaining := reader.Remaining()     // bytes left

// Writing packet body
writer := codec.NewWriter()
writer.WriteString("PRODUCTION-202401")
writer.WriteInt32(1)
writer.WriteBool(true)
body := writer.Bytes()
```

| Type | Wire Format | Size |
|------|-------------|------|
| `int32` | 4 bytes, big-endian, signed | 4B |
| `uint16` | 2 bytes, big-endian, unsigned | 2B |
| `string` | 2B length prefix + UTF-8 payload | 2B + len |
| `bool` | 1 byte (0 = false, 1 = true) | 1B |

## Redis

Redis serves three roles in the server:

1. **Session registry** — tracks active connections with TTL-based leases
2. **SSO token store** — single-use authentication tickets
3. **Pub/Sub broadcast** — cross-instance message delivery

### Client Setup

```go
client, err := redis.NewClient(redis.Config{
    Address:  "localhost:6379",
    Password: "",
    DB:       0,
    PoolSize: 20,
})
```

Timeouts are fixed at 3 seconds for read, write, and dial operations.

### Key Namespaces

| Prefix | Purpose | TTL |
|--------|---------|-----|
| `session:conn:{connID}` | Session JSON record | 120s (refreshed every 60s) |
| `session:user:{userID}` | User → connID index | 120s (refreshed every 60s) |
| `sso:{ticket}` | SSO ticket → userID | Configurable (default 300s) |
| `hotel:status` | Hotel state JSON | Persistent |
| `handshake:close:{connID}` | Close signal channel | Pub/Sub (no persistence) |
| `broadcast:all` | Hotel-wide broadcast | Pub/Sub (no persistence) |

### Pub/Sub Broadcast

The `core/broadcast` package abstracts publish/subscribe:

```go
// Publish to all instances
broadcaster.Publish(ctx, "broadcast:all", packetBytes)

// Subscribe to targeted messages
ch, disposable, err := broadcaster.Subscribe(ctx, "broadcast:conn:abc123")
defer disposable.Dispose()
for payload := range ch {
    // handle incoming packet
}
```

Two implementations:
- **`RedisBroadcaster`** — uses Redis Pub/Sub, adds optional namespace prefix
- **`LocalBroadcaster`** — in-process channels, buffered (8 messages),
  non-blocking publish (drops if buffer full)

## PostgreSQL & Migrations

PostgreSQL stores permanent data (users, audit logs). GORM is the ORM.

### Connection

```go
db, err := postgres.NewClient(postgres.Config{
    DSN:                    "postgres://...",
    MaxOpenConns:           30,
    MaxIdleConns:           10,
    ConnMaxLifetimeSeconds: 300,
})
```

### Migration System

Migrations use `gormigrate/v2` with explicit versioning:

```go
// core/postgres/migrations/01_users.go
var CreateUsers = &gormigrate.Migration{
    ID: "01_create_users",
    Migrate: func(tx *gorm.DB) error {
        return tx.AutoMigrate(&model.User{})
    },
    Rollback: func(tx *gorm.DB) error {
        return tx.Migrator().DropTable("users")
    },
}
```

Migrations and seeds have separate registries and state tables:

| Table | Purpose |
|-------|---------|
| `schema_migrations` | Tracks applied schema migrations |
| `schema_seeds` | Tracks applied seed data |

Auto-run is controlled by config (`migration_auto_up`, `seed_auto_up`).

## Structured Logging

Logging uses [Zap](https://github.com/uber-go/zap) with two format modes:

### Console Format (Development)

```
2026-03-12T14:30:00.000Z  INFO  handshake/handler.go:45  connection authenticated  {"connID": "abc123", "userID": 42}
```

### JSON Format (Production)

```json
{"level":"info","ts":"2026-03-12T14:30:00.000Z","caller":"handshake/handler.go:45","msg":"connection authenticated","connID":"abc123","userID":42}
```

### HTTP Request Logging

Fiber requests are logged via `fiberzap` middleware. At `debug` level, every
request/response is logged. At `info` and above, only errors are logged.

## Connections & Disposables

### Disposable Pattern

The `Disposable` interface is used throughout the codebase for resource cleanup:

```go
type Disposable interface {
    Dispose() error
}

// Wrap a function as Disposable
cleanup := connection.DisposeFunc(func() error {
    return pubSub.Close()
})
```

Used by: WebSocket connections, Redis subscriptions, Pub/Sub listeners,
broadcast subscribers, and the HTTP module itself.

### Connection Abstraction

The `Connection` interface decouples packet I/O from WebSocket specifics:

```go
type Connection interface {
    Disposable
    ID() string
    Read(ctx context.Context) ([]byte, error)
    Write(ctx context.Context, payload []byte) error
}
```

`MemoryConnection` provides an in-process implementation for tests:

```go
conn := connection.NewMemoryConnection("test-conn", 10)

// Simulate inbound packet
conn.PushInbound(packetBytes)

// Read what the server sent
response, err := conn.ReadOutbound(ctx)
```

### Session Registry

The session registry tracks active connections with Redis-backed persistence:

```go
type Session struct {
    ConnID     string
    UserID     int
    MachineID  string
    State      SessionState  // Connected | Authenticated | Disconnecting
    InstanceID string        // identifies owning server instance
    CreatedAt  time.Time
}
```

Key operations:
- `Register(session)` — upsert with automatic user index management
- `FindByUserID(id)` — cross-instance lookup (returns session on any instance)
- `FindByConnID(id)` — direct lookup
- `Touch(connID)` — refresh TTL (called by heartbeat)
- `Remove(connID)` — delete session and user index

## Encryption (Optional)

The server supports the Habbo RC4 + RSA + Diffie-Hellman encryption handshake.
Encryption is optional — the Nitro client works without it over `wss://`.

### Diffie-Hellman Key Exchange

```go
dh, err := crypto.NewDiffieHellman(crypto.WithBitSize(128))
// dh.Prime(), dh.Generator(), dh.PublicKey()

sharedKey := dh.DeriveSharedKey(clientPublicKey)
```

### RSA Decryption

The server generates an RSA key pair and uses it to decrypt the client's
DH public key during the handshake:

```go
privateKey, err := crypto.GeneratePrivateKey(1024, rand.Reader)
clientPubKey, err := crypto.DecodeClientPublicKey(privateKey, encryptedBytes)
```

### RC4 Stream Cipher

After DH exchange, both sides derive an RC4 cipher from the shared key:

```go
cipher, err := crypto.NewStreamCipher(sharedKey)
encrypted := cipher.Encrypt(plaintext)  // outbound
decrypted := cipher.Decrypt(ciphertext) // inbound
```

Each direction has an independent stream state. The cipher is thread-safe
(mutex-protected).

## Initializer System

The startup orchestrator executes typed stages in dependency order:

```go
runner := initializer.NewRunner(
    configStage,       // 1. Load config
    redisStage,        // 2. Connect Redis
    loggerStage,       // 3. Create logger
    postgresStage,     // 4. Connect DB + migrations
    httpStage,         // 5. Create HTTP server
    wsStage,           // 6. Register WebSocket handler
)

runtime, err := runner.Run()
// runtime.Config, runtime.Redis, runtime.Logger, runtime.DB, runtime.HTTP
```

Each stage implements a typed interface and receives outputs from prior stages.
The runner fails fast with the stage name on error.

## CLI

The command tree uses Cobra:

```
pixelsv
  ├── serve              Start the server
  │   ├── --env-file     Path to .env file (default: .env)
  │   ├── --env-prefix   Environment variable prefix
  │   ├── --ws-path      WebSocket endpoint path (default: /ws)
  │   └── --api-key-header  API key header name (default: X-API-Key)
  ├── db
  │   ├── migrate        Run database migrations
  │   └── seed           Run database seeds
  └── sso
      └── issue          Generate an SSO ticket for a user ID
```

Dependencies are injected via a `Dependencies` struct for testability:

```go
type Dependencies struct {
    Output  io.Writer
    ErrOut  io.Writer
    Args    []string
}
```
