# OpenAPI And Swagger

## Purpose

This document defines the API documentation surface for runtime endpoints.

## Routes

- OpenAPI JSON: `/openapi.json`
- Swagger UI: `/swagger`
Both routes are public and do not require API key.

## Current Endpoint Coverage

- `POST /api/v1/sso`
- `GET /ws` (or configured websocket route)
- `GET /openapi.json`
- `GET /swagger`

## Ownership

- Core OpenAPI document assembly and docs routes: `core/http/openapi`
- Authentication HTTP route specs: `pkg/authentication/adapter/httpapi`

## Update Rule

When an HTTP endpoint is added, changed, or removed:

1. Update endpoint implementation.
2. Update its OpenAPI path item in the owning module.
3. Keep Swagger UI and OpenAPI routes valid.
