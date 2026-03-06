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
- `POST /api/v1/tickets` — create SSO ticket
- `DELETE /api/v1/tickets/:ticket` — revoke SSO ticket

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

- `--role=api` or `--role=all`: serves REST admin endpoints, Swagger, OpenAPI.
- `--role=gateway` or `--role=all`: serves `/ws` WebSocket endpoint.
- All roles: serve `/health` and `/ready` probes.

## CLI

- `pixelsv serve` starts all roles (HTTP + WebSocket + all realm modules).
- `pixelsv serve --role=api` starts only the REST admin API.
- `pixelsv serve --role=gateway` starts only the WebSocket gateway.
- `pixelsv serve --env-file .env` loads environment from a specific file.
