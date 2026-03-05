# REST API Layer

## Overview

Every pixel-server service exposes a REST API alongside its NATS-based packet processing. The REST layer serves administrative operations — creating SSO tickets, querying session counts, managing domain entities — that do not belong in the binary game protocol.

The REST layer uses **Go Fiber v3** with a shared setup in `pkg/core/httpserver`. All services use the same authentication mechanism (bearer token), logging middleware (Zap), and health-check endpoint. The binary WebSocket protocol for game clients remains unchanged.

---

## Design Decisions

### Why REST alongside NATS?

NATS subjects are ideal for high-throughput, low-latency inter-service communication in the game loop. REST is added for:

1. **External integration** — CMS panels, web dashboards, mobile apps, and third-party tools interact via standard HTTP/JSON.
2. **Administrative operations** — Creating users, generating SSO tickets, banning players, and querying system state are request/response operations that map naturally to REST.
3. **Observability** — Health checks (`GET /health`) enable container orchestrators (Docker, Kubernetes) to probe service liveness.

### Why Fiber v3?

| Requirement | Fiber v3 |
|---|---|
| Performance | Built on fasthttp — low allocation, high throughput |
| Middleware stack | Express-style chaining, easy to compose |
| JSON handling | Native `c.JSON()` / `c.Bind().JSON()` |
| WebSocket upgrade | Fiber can coexist with raw TCP listeners |
| Go-native | Pure Go, no CGO, small dependency tree |

### Why static bearer token?

For the initial implementation, a single `API_TOKEN` environment variable provides:
- Zero infrastructure (no OAuth server, no JWT key management).
- Sufficient security for private/internal APIs behind a firewall.
- Easy rotation by restarting services with a new token.

Future iterations can replace the `AuthMiddleware` with JWT or OAuth2 without changing the Fiber route handlers.

---

## Shared Infrastructure (`pkg/core/httpserver`)

### Server Factory

`httpserver.New(cfg, logger)` returns a `*Server` with:

1. **Fiber app** configured with JSON error handler.
2. **Zap request logging** middleware — logs method, path, status, latency, IP.
3. **Bearer-token auth** middleware — validates `Authorization: Bearer <token>`, skips `/health`.
4. **Health endpoint** — `GET /health` returns `{"status":"ok"}`.

### Configuration

```go
type Config struct {
    HTTPAddr string `mapstructure:"http_addr" env:"HTTP_ADDR" default:":8080"`
    APIToken string `mapstructure:"api_token" env:"API_TOKEN"`
}
```

- `HTTP_ADDR` — listen address (default `:8080`).
- `API_TOKEN` — required bearer token. Empty token rejects all requests (fail-closed).

### Lifecycle

```go
httpSrv := httpserver.New(cfg.HTTP, logger)
// Register service-specific routes on httpSrv.App
go httpSrv.ListenAndServe(ctx)
```

The HTTP server runs in a background goroutine. On context cancellation, `Fiber.Shutdown()` drains connections gracefully.

### Fiber Startup Banner

The Fiber startup banner is **always disabled** (`DisableStartupMessage: true` in `ListenConfig`). All startup and runtime logging goes through Zap.

---

## Per-Service Endpoints

### Gateway

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |
| `GET` | `/api/v1/sessions` | Active WebSocket session count |

The gateway keeps its raw TCP WebSocket listener on `:2096` for the binary game protocol. The Fiber HTTP server runs on `HTTP_ADDR` (default `:8080`).

### Auth

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |
| `POST` | `/api/v1/tickets` | Create a one-time SSO ticket for a user |
| `DELETE` | `/api/v1/tickets/:ticket` | Revoke an existing SSO ticket |

The ticket creation endpoint is the primary REST integration point: external systems (CMS, web client) call it to generate a ticket, then pass the ticket to the game client for WebSocket authentication.

### Game

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |

Future endpoints: room listing, player kick, room config.

### Catalog / Navigator / Social / Moderation

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |

Domain-specific REST endpoints will be added as these services mature.

---

## OpenAPI Specification

All REST endpoints are documented in OpenAPI 3.1 format under `api/`:

| File | Contents |
|---|---|
| `api/openapi.yaml` | Aggregated spec — all services, all endpoints |
| `api/gateway.yaml` | Gateway-specific spec |
| `api/auth.yaml` | Auth-specific spec |

### Workflow

1. Implement the endpoint in Go (Fiber handler).
2. Add it to the service-specific `api/<service>.yaml`.
3. Add it to the aggregated `api/openapi.yaml`.
4. Update `docs/06-rest-api/ENDPOINTS.md`.

### Validation

```bash
# Using spectral (npm i -g @stoplight/spectral-cli)
spectral lint api/openapi.yaml
```

---

## Authentication Flow

```
Client                   Service
  │                        │
  ├── GET /api/v1/...     ─┤
  │   Authorization:       │
  │   Bearer <API_TOKEN>   │
  │                        ├── AuthMiddleware checks token
  │                        │   ├── Missing → 401 {"error":"missing authorization header"}
  │                        │   ├── Bad scheme → 401 {"error":"invalid authorization scheme"}
  │                        │   ├── Wrong token → 403 {"error":"invalid token"}
  │                        │   └── Valid → next handler
  │◄── 200 JSON response ─┤
```

Token comparison uses `crypto/subtle.ConstantTimeCompare` to prevent timing attacks.

---

## Error Handling

All errors are returned as JSON:

```json
{"error": "description of what went wrong"}
```

HTTP status codes follow standard semantics:
- `400` — Bad request (invalid body, missing fields).
- `401` — Missing or malformed Authorization header.
- `403` — Valid header format but wrong token.
- `404` — Resource not found.
- `500` — Internal server error.

Fiber's custom `ErrorHandler` ensures even framework-level errors (404 for unmatched routes) return JSON instead of plain text.

---

## Deployment Topology

### Development

All services share `HTTP_ADDR=:8080`. Since services run as separate binaries, each must use a **different port** when running locally:

```bash
HTTP_ADDR=:8080  ./gateway    # REST on 8080, WS on 2096
HTTP_ADDR=:8081  ./auth       # REST on 8081
HTTP_ADDR=:8082  ./game       # REST on 8082
# etc.
```

### Production (Docker / Kubernetes)

Each service container exposes port 8080 internally. The orchestrator assigns unique external ports or uses service discovery. A reverse proxy (Nginx, Traefik) can aggregate all service APIs under a single domain using path-based routing.

### Health Checks

Every service responds to `GET /health` without authentication. Docker and Kubernetes health probes use this endpoint:

```yaml
# Docker Compose
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://127.0.0.1:8080/health"]
  interval: 10s
  timeout: 3s
  retries: 5
```

```yaml
# Kubernetes
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  periodSeconds: 10
```
