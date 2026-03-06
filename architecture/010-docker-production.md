# Docker Production Guide (Multi-Role Single Binary)

## Overview

Deploy one `pixelsv` container image with role-specific command overrides, plus infrastructure dependencies.

## Baseline Stack

- `pixelsv` (single binary, multi-role via `--role` flag)
- PostgreSQL 16
- Redis 7 (Valkey)
- Reverse proxy (Nginx/Traefik for TLS termination)
- NATS (only in distributed mode)

## Runtime Commands

```bash
pixelsv serve                        # all roles (dev/small deploy)
pixelsv serve --role=gateway         # WebSocket sessions only
pixelsv serve --role=game            # room workers only
pixelsv serve --role=social,navigator,catalog,moderation  # combined support roles
pixelsv serve --role=api             # REST admin + Swagger only
pixelsv serve --role=jobs            # maintenance scheduler
pixelsv migrate                      # run Atlas migrations
pixelsv admin create-user ...        # CLI administration
```

## Compose: All-in-One (Development)

```yaml
services:
  pixelsv:
    build: .
    command: ["serve"]
    ports:
      - "8080:8080"
    environment:
      PIXELSV_ROLE: all
      POSTGRES_URL: postgres://pixelsv:pixelsv@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
      API_KEY: dev-key
      LOG_FORMAT: console
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: pixelsv
      POSTGRES_USER: pixelsv
      POSTGRES_PASSWORD: pixelsv
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pixelsv"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

## Compose: Distributed (Production)

```yaml
services:
  gateway:
    image: pixelsv:latest
    command: ["serve", "--role=gateway"]
    ports:
      - "443:8080"
    environment:
      PIXELSV_ROLE: gateway
      PIXELSV_INSTANCE_ID: gateway-1
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
      HTTP_ADDR: ":8080"
      API_KEY: ${API_KEY}
    depends_on: [redis, nats]

  game-worker-1:
    image: pixelsv:latest
    command: ["serve", "--role=game"]
    environment:
      PIXELSV_ROLE: game
      PIXELSV_INSTANCE_ID: game-worker-1
      POSTGRES_URL: postgres://pixelsv:${PG_PASS}@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on: [postgres, redis, nats]

  game-worker-2:
    image: pixelsv:latest
    command: ["serve", "--role=game"]
    environment:
      PIXELSV_ROLE: game
      PIXELSV_INSTANCE_ID: game-worker-2
      POSTGRES_URL: postgres://pixelsv:${PG_PASS}@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on: [postgres, redis, nats]

  support:
    image: pixelsv:latest
    command: ["serve", "--role=social,navigator,catalog,moderation"]
    environment:
      PIXELSV_ROLE: social,navigator,catalog,moderation
      PIXELSV_INSTANCE_ID: support-1
      POSTGRES_URL: postgres://pixelsv:${PG_PASS}@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on: [postgres, redis, nats]

  api:
    image: pixelsv:latest
    command: ["serve", "--role=api"]
    ports:
      - "8081:8080"
    environment:
      PIXELSV_ROLE: api
      PIXELSV_INSTANCE_ID: api-1
      POSTGRES_URL: postgres://pixelsv:${PG_PASS}@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
      HTTP_ADDR: ":8080"
      API_KEY: ${API_KEY}
    depends_on: [postgres, redis]

  jobs:
    image: pixelsv:latest
    command: ["serve", "--role=jobs"]
    environment:
      PIXELSV_ROLE: jobs
      PIXELSV_INSTANCE_ID: jobs-1
      POSTGRES_URL: postgres://pixelsv:${PG_PASS}@postgres:5432/pixelsv
      REDIS_URL: redis://redis:6379
    depends_on: [postgres, redis]

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: pixelsv
      POSTGRES_USER: pixelsv
      POSTGRES_PASSWORD: ${PG_PASS}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pixelsv"]

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]

  nats:
    image: nats:2-alpine
    command: ["--jetstream"]
    healthcheck:
      test: ["CMD", "nats-server", "--signal", "ldm"]

volumes:
  pgdata:
```

## Dockerfile

```dockerfile
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum go.work ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pixelsv ./cmd/pixelsv

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /pixelsv /usr/local/bin/pixelsv
ENTRYPOINT ["pixelsv"]
CMD ["serve"]
```

## Health Checks

- `GET /health` — liveness (returns 200 if the process is running)
- `GET /ready` — readiness (returns 200 when all role dependencies are connected)

Roles that don't serve HTTP (`game`, `social`, etc.) expose health on a separate port configured via `HEALTH_ADDR` (defaults to `:9090`).

## Operational Principles

- Keep rollout atomic around one binary version across all role instances.
- Keep bounded contexts decoupled in code, not as separate deployables.
- Scale game workers first (they're the bottleneck); other roles scale rarely.
- In distributed mode, deploy NATS as infrastructure alongside PostgreSQL and Redis.
- Use process managers (systemd, Kubernetes, Nomad) to restart crashed role instances automatically.
- Room state survives crashes via Redis ECS snapshots.
