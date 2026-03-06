# API

Core HTTP/WebSocket runtime implementation is provided by `pkg/http`.

## Core Endpoints

- `GET /health` returns liveness JSON.
- `GET /ready` returns readiness JSON.
- `GET /openapi.json` returns OpenAPI 3.1 JSON.
- `GET /swagger` returns Swagger UI HTML bound to `/openapi.json`.
- `GET /ws` upgrades to WebSocket and echoes frames.

## Administrative Endpoint

- `GET /api/v1/admin/ping`

Administrative endpoints require header:

- `X-API-Key: <API_KEY>`

## OpenAPI and Swagger

- OpenAPI JSON is served at `GET /openapi.json`.
- Swagger UI HTML is served at `GET /swagger` and points to `/openapi.json`.
- Spec includes:
  - `ApiKeyAuth` security scheme (`X-API-Key`)
  - core probes (`/health`, `/ready`)
  - admin endpoint (`/api/v1/admin/ping`)
  - realtime endpoint documentation (`/ws`)

## Implementation Caveat

Current implementation uses Fiber v2 runtime adapters (`fiberzap` and websocket contrib) for compatibility in this stage, while preserving the same endpoint contract planned in architecture.

## CLI

- `pixelsv serve` starts the core HTTP/WebSocket runtime.
- `pixelsv serve --env-file .env` loads environment from a specific file.
