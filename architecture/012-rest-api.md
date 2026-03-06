# HTTP and WebSocket API (Fiber-First)

## Scope

`pixelsv` exposes API and realtime endpoints from the same binary. In distributed mode, the `api` role serves REST and the `gateway` role serves WebSocket; in all-in-one mode, one Fiber app serves everything.

## Framework Direction

- Use GoFiber v3 for HTTP server concerns.
- Prefer Fiber-compatible websocket middleware for realtime endpoints.
- Keep API/WebSocket handlers thin; call application ports and use cases.

## Current Caveat

The current core implementation may temporarily run on Fiber v2-compatible middleware packages while keeping the same route and contract surface. The upgrade target remains Fiber v3 once middleware compatibility is fully aligned.

## Why Fiber for WebSocket Path

- Unified middleware and lifecycle in one server runtime.
- Consistent auth/logging/error handling between HTTP and WebSocket.
- Simpler operations for a single-binary architecture.
- Shared TLS termination, proxy header handling, and rate limiting.

## API Surface

### All Roles

- `GET /health` (no auth) — liveness probe
- `GET /ready` (no auth) — readiness probe

### API Role (`--role=api` or `--role=all`)

- `GET /openapi.json` — OpenAPI 3.1 spec
- `GET /swagger/*` — Swagger UI
- `/api/v1/*` — authenticated administrative endpoints

### Gateway Role (`--role=gateway` or `--role=all`)

- `GET /ws` — WebSocket upgrade for binary protocol sessions

## Administrative REST Endpoints

| Method | Path | Auth | Realm | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/tickets` | API key | auth | Create SSO ticket |
| `DELETE` | `/api/v1/tickets/:ticket` | API key | auth | Revoke SSO ticket |
| `GET` | `/api/v1/sessions` | API key | gateway | List active sessions |
| `POST` | `/api/v1/users` | API key | auth | Create user |
| `GET` | `/api/v1/users/:id` | API key | auth | Get user |
| `PUT` | `/api/v1/users/:id` | API key | auth | Update user |
| `GET` | `/api/v1/rooms` | API key | navigator | List rooms |
| `POST` | `/api/v1/bans` | API key | moderation | Issue ban |
| `DELETE` | `/api/v1/bans/:id` | API key | moderation | Revoke ban |
| `GET` | `/api/v1/admin/ping` | API key | core | Admin ping |

Each realm module registers its routes via `MountRoutes(app *fiber.App)` during startup.

## Auth Direction

- Administrative HTTP routes use API key authentication (`X-API-Key` header).
- Session auth for `/ws` follows handshake/auth realm contracts (SSO ticket via binary protocol).
- Replaceable auth adapters (JWT/OAuth) are allowed without changing domain logic.

## Error Contract

- JSON errors for HTTP routes: `{"error": "message"}`.
- Structured close/error events for WebSocket routes.
- Domain errors are mapped in transport adapters, not created there.

## Operational Requirements

- Disable framework banner output in production.
- All logs flow through the project logger (`pkg/log`).
- Graceful shutdown must stop HTTP and WebSocket listeners with context deadlines.
- Proxy header support (`X-Forwarded-For`, `CF-Connecting-IP`) via Fiber's built-in `ProxyHeader` and `TrustedProxies` config.

## OpenAPI Spec

One aggregated OpenAPI 3.1 spec served at `/openapi.json`. Includes:
- `ApiKeyAuth` security scheme (`X-API-Key` header)
- All administrative endpoints across all realm modules
- Core probes (`/health`, `/ready`)
- Realtime endpoint documentation (`/ws`)

The spec is maintained in `api/openapi.yaml` and served as JSON at runtime.

## References

- Fiber websocket contrib docs: <https://docs.gofiber.io/contrib/next/websocket/>
- Fiber websocket middleware source: <https://github.com/gofiber/fiber/tree/main/middleware/websocket>
