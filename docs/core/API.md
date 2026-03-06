# API

Core HTTP/WebSocket runtime implementation is provided by `pkg/http`.

## Core Endpoints

- `GET /health` returns liveness JSON.
- `GET /ready` returns readiness JSON.
- `GET /openapi.json` returns OpenAPI 3.1 JSON.
- `GET /swagger` returns Swagger UI HTML bound to `/openapi.json`.
- `GET /ws` upgrades to WebSocket for binary protocol sessions.

## Administrative Endpoints

- `GET /api/v1/admin/ping`

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
- Phase 0 behavior: any role that enables HTTP serves the same core endpoints listed above.

## CLI

- `pixelsv serve` starts all roles.
- `pixelsv serve --role=api` starts the HTTP runtime for API process deployment.
- `pixelsv serve --role=gateway` starts the HTTP/WebSocket runtime for gateway deployment.
- `pixelsv serve --env-file .env` loads environment from a specific file.
