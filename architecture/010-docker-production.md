# Docker Production Guide (Single Binary)

## Overview

Deploy one `pixelsv` container plus infrastructure dependencies.

## Baseline Stack

- `pixelsv` (single binary, HTTP/WebSocket + CLI roles)
- PostgreSQL
- Redis
- Reverse proxy (optional)

NATS is optional and only valid when enabled through a messaging adapter requirement.

## Runtime Commands

Examples:

- `pixelsv serve` for API/WebSocket runtime.
- `pixelsv jobs` for background schedulers.
- `pixelsv migrate` for schema migrations.

Use one image and role-specific command overrides.

## Compose Direction

- One application service (`pixelsv`), not multiple per bounded context.
- Environment variables configure enabled roles and adapter implementations.
- Health checks target `/health` and `/ready`.

## Operational Principles

- Keep rollout atomic around one binary version.
- Keep bounded contexts decoupled in code, not as separate deployables.
- Scale vertically first; introduce multi-instance routing only when load requires it.
