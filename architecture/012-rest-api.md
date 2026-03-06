# HTTP and WebSocket API (Fiber-First)

## Scope

`pixelsv` exposes API and realtime endpoints from the same binary.

## Framework Direction

- Use GoFiber v3 for HTTP server concerns.
- Prefer Fiber-compatible websocket middleware for realtime endpoints.
- Keep API/WebSocket handlers thin; call application ports and use cases.

## Current Caveat

The current core implementation may temporarily run on Fiber v2-compatible middleware packages while keeping the same route and contract surface (`/health`, `/ready`, `/openapi.json`, `/swagger`, `/ws`, `/api/v1/*`). The upgrade target remains Fiber v3 once middleware compatibility is fully aligned.

## Why Fiber for WebSocket Path

- Unified middleware and lifecycle in one server runtime.
- Consistent auth/logging/error handling between HTTP and WebSocket.
- Simpler operations for a single-binary architecture.

## API Surface

- `GET /health` (no auth)
- `GET /ready` (no auth)
- `/api/v1/*` for authenticated administrative endpoints
- `/ws` for binary protocol sessions

## Auth Direction

- Start with token-based auth for administrative HTTP routes.
- Session auth for `/ws` follows handshake/auth module contracts.
- Replaceable auth adapters (JWT/OAuth) are allowed without changing domain logic.

## Error Contract

- JSON errors for HTTP routes.
- Structured close/error events for WebSocket routes.
- Domain errors are mapped in transport adapters, not created there.

## Operational Requirements

- Disable framework banner output in production.
- All logs flow through the project logger.
- Graceful shutdown must stop HTTP and WebSocket listeners with context deadlines.

## References

- Fiber websocket contrib docs: <https://docs.gofiber.io/contrib/next/websocket/>
- Fiber websocket middleware source: <https://github.com/gofiber/fiber/tree/main/middleware/websocket>
