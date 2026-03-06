# API

Core HTTP/WebSocket runtime implementation is provided by `pkg/http`.

## Core Endpoints

- `GET /health` returns liveness JSON.
- `GET /ready` returns readiness JSON.
- `GET /openapi.json` returns OpenAPI 3.1 JSON.
- `GET /swagger` returns Swagger UI HTML bound to `/openapi.json`.
- `GET /ws` upgrades to WebSocket for binary protocol sessions.

## WebSocket Runtime Flow

- WebSocket endpoint is implemented through Fiber websocket middleware.
- Incoming binary payloads are parsed as one or many protocol frames (`uint32` length + `uint16` header + payload).
- Each frame is decoded with `pkg/protocol` c2s registry.
- Decoded packet realm is routed through transport topic:
  - `packet.c2s.<realm>.<sessionID>`
- Outbound payloads are consumed from:
  - `session.output.<sessionID>`
- Session lifecycle publishes disconnect events on:
  - `session.disconnected`

## Administrative Endpoints

- `GET /api/v1/admin/ping`
- `POST /api/v1/auth/tickets`
- `DELETE /api/v1/auth/tickets/:ticket`

Administrative endpoints require header:

- `X-API-Key: <API_KEY>`

## OpenAPI and Swagger

- OpenAPI JSON is served at `GET /openapi.json`.
- Swagger UI HTML is served at `GET /swagger` and points to `/openapi.json`.
- Spec includes:
  - `ApiKeyAuth` security scheme (`X-API-Key`)
  - core probes (`/health`, `/ready`)
  - admin endpoints (`/api/v1/*`)
  - realtime endpoint documentation (`/ws`)

## Role-Aware Endpoint Availability

- HTTP listener enabled roles: `all`, `gateway`, `api`, `jobs`.
- HTTP listener disabled roles: `game`, `auth`, `social`, `navigator`, `catalog`, `moderation`.
- Current behavior: any role that enables HTTP serves the same core endpoints listed above.

## CLI

- `pixelsv serve` starts all roles.
- `pixelsv serve --role=api` starts the HTTP runtime for API process deployment.
- `pixelsv serve --role=gateway` starts the HTTP/WebSocket runtime for gateway deployment.
- `pixelsv serve --env-file .env` loads environment from a specific file.
