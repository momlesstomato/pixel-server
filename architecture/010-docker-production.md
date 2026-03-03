# Docker Production Guide

## Overview

pixel-server is designed to run on a **single affordable VPS** for small hotel operators while remaining horizontally scalable when traffic grows. This document covers Docker Compose deployment — no Kubernetes required.

A typical hotel running 50–500 concurrent users fits comfortably on a **2 vCPU / 4 GB RAM** machine with under €20/month.

---

## Target stack

| Service | Image | Minimum RAM |
|---|---|---|
| PostgreSQL 16 | `postgres:16-alpine` | 256 MB |
| Redis 7 | `redis:7-alpine` | 64 MB |
| NATS JetStream | `nats:2.10-alpine` | 64 MB |
| `gateway` | pixel-server build | 128 MB |
| `auth-svc` | pixel-server build | 64 MB |
| `game-svc` | pixel-server build | 256 MB (per instance) |
| `social-svc` | pixel-server build | 64 MB |
| `navigator-svc` | pixel-server build | 64 MB |
| `catalog-svc` | pixel-server build | 64 MB |
| `moderation-svc` | pixel-server build | 64 MB |
| Nginx | `nginx:alpine` | 32 MB |
| **Total** | | **~1.1 GB** |

---

## Repository layout

```
docker/
  compose.yml
  compose.override.yml        ← local dev overrides (not committed in prod)
  nginx/
    habbo.conf
  postgres/
    init/
      00-schema.sql            ← applied once on first boot (Atlas exported)
  nats/
    server.conf
  .env.example
.env                           ← gitignored; actual secrets
```

---

## `docker/compose.yml`

```yaml
name: pixel-server

services:

  # ─── Infrastructure ─────────────────────────────────────────────────────────

  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_DB:       ${POSTGRES_DB}
      POSTGRES_USER:     ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./postgres/init:/docker-entrypoint-initdb.d:ro
    deploy:
      resources:
        limits:
          memory: 512m
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --save 60 1 --loglevel warning --maxmemory 128mb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    deploy:
      resources:
        limits:
          memory: 192m
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  nats:
    image: nats:2.10-alpine
    restart: always
    command: ["-c", "/etc/nats/server.conf"]
    volumes:
      - ./nats/server.conf:/etc/nats/server.conf:ro
      - nats-data:/data/jetstream
    deploy:
      resources:
        limits:
          memory: 128m
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://127.0.0.1:8222/healthz"]
      interval: 10s
      timeout: 3s
      retries: 5

  # ─── pixel-server services ───────────────────────────────────────────────────

  gateway:
    build:
      context: ..
      target: gateway
    restart: always
    ports:
      - "2096:2096"    # WebSocket game port (exposed to clients)
    environment:
      NATS_URL:         nats://nats:4222
      REDIS_URL:        redis://redis:6379
      LISTEN_ADDR:      ":2096"
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      nats:    { condition: service_healthy }
      redis:   { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 192m
    healthcheck:
      test: ["CMD", "/app/gateway", "--healthcheck"]
      interval: 15s
      timeout: 5s
      retries: 3

  auth-svc:
    build:
      context: ..
      target: auth-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_URL:        redis://redis:6379
      JWT_SECRET:       ${JWT_SECRET}
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      nats:     { condition: service_healthy }
      redis:    { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 96m

  game-svc:
    build:
      context: ..
      target: game-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_URL:        redis://redis:6379
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      nats:     { condition: service_healthy }
      redis:    { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 384m
    # Scale game-svc replicas when traffic grows:
    # docker compose up --scale game-svc=3 -d
    # game-svc instances claim rooms via NATS work-queue; no coordination needed.

  social-svc:
    build:
      context: ..
      target: social-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_URL:        redis://redis:6379
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      nats:     { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 96m

  navigator-svc:
    build:
      context: ..
      target: navigator-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_URL:        redis://redis:6379
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      redis:    { condition: service_healthy }
      nats:     { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 96m

  catalog-svc:
    build:
      context: ..
      target: catalog-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      nats:     { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 96m

  moderation-svc:
    build:
      context: ..
      target: moderation-svc
    restart: always
    environment:
      NATS_URL:         nats://nats:4222
      DATABASE_URL:     postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_URL:        redis://redis:6379
      LOG_LEVEL:        ${LOG_LEVEL:-info}
    depends_on:
      postgres: { condition: service_healthy }
      nats:     { condition: service_healthy }
      redis:    { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 96m

  # ─── Reverse proxy ───────────────────────────────────────────────────────────

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/habbo.conf:/etc/nginx/conf.d/default.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro   # certbot certificates
      - nginx-logs:/var/log/nginx
    depends_on:
      - gateway
    deploy:
      resources:
        limits:
          memory: 48m

volumes:
  postgres-data:
  redis-data:
  nats-data:
  nginx-logs:
```

---

## `docker/nats/server.conf`

```
port: 4222
http_port: 8222

jetstream {
  store_dir: /data/jetstream
  max_mem_store: 64mb
  max_file_store: 2gb
}

# Disable account auth for internal cluster (not exposed to internet)
no_auth_user: pixel
accounts: {
  pixel: { users: [{ user: pixel, password: "" }] }
}
```

---

## `docker/nginx/habbo.conf`

```nginx
# Redirect HTTP → HTTPS
server {
    listen 80;
    server_name hotel.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name hotel.example.com;

    ssl_certificate     /etc/letsencrypt/live/hotel.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/hotel.example.com/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    # WebSocket game connection
    location /ws {
        proxy_pass          http://gateway:2096;
        proxy_http_version  1.1;
        proxy_set_header    Upgrade    $http_upgrade;
        proxy_set_header    Connection "upgrade";
        proxy_set_header    X-Real-IP  $remote_addr;
        proxy_read_timeout  3600s;
        proxy_send_timeout  3600s;
    }

    # Static hotel client assets (optional — can serve from CDN instead)
    location / {
        root  /var/www/client;
        try_files $uri $uri/ =404;
    }
}
```

---

## `.env.example`

```bash
# Database
POSTGRES_DB=pixel
POSTGRES_USER=pixel
POSTGRES_PASSWORD=changeme_strong_password

# Auth
JWT_SECRET=changeme_64_char_random_secret

# Logging
LOG_LEVEL=info
```

Copy to `.env` and fill in real values. Never commit `.env`.

---

## Multi-stage Dockerfile

Each service produces a minimal image. Root `Dockerfile` at the repo root:

```dockerfile
# ── Builder ──────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.work go.work.sum ./
COPY pkg/        pkg/
COPY services/   services/
COPY tools/      tools/
# Build all service binaries in one layer
RUN go build -o /out/gateway      ./services/gateway && \
    go build -o /out/auth-svc     ./services/auth    && \
    go build -o /out/game-svc     ./services/game    && \
    go build -o /out/social-svc   ./services/social  && \
    go build -o /out/navigator-svc ./services/navigator && \
    go build -o /out/catalog-svc  ./services/catalog && \
    go build -o /out/moderation-svc ./services/moderation

# ── Final target images ───────────────────────────────────────────────────────
FROM alpine:3.20 AS base-runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

FROM base-runtime AS gateway
COPY --from=builder /out/gateway .
EXPOSE 2096
ENTRYPOINT ["/app/gateway"]

FROM base-runtime AS auth-svc
COPY --from=builder /out/auth-svc .
ENTRYPOINT ["/app/auth-svc"]

FROM base-runtime AS game-svc
COPY --from=builder /out/game-svc .
ENTRYPOINT ["/app/game-svc"]

FROM base-runtime AS social-svc
COPY --from=builder /out/social-svc .
ENTRYPOINT ["/app/social-svc"]

FROM base-runtime AS navigator-svc
COPY --from=builder /out/navigator-svc .
ENTRYPOINT ["/app/navigator-svc"]

FROM base-runtime AS catalog-svc
COPY --from=builder /out/catalog-svc .
ENTRYPOINT ["/app/catalog-svc"]

FROM base-runtime AS moderation-svc
COPY --from=builder /out/moderation-svc .
ENTRYPOINT ["/app/moderation-svc"]
```

Each service calls `docker compose build --target <service>` independently so CI only rebuilds changed services.

---

## Initial setup on a fresh VPS

```bash
# 1. Install Docker (Debian/Ubuntu)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER && newgrp docker

# 2. Install certbot for TLS
sudo apt-get install -y certbot
sudo certbot certonly --standalone -d hotel.example.com

# 3. Clone repo and configure
git clone https://github.com/your-org/pixel-server.git
cd pixel-server
cp docker/.env.example docker/.env
# Edit docker/.env with real passwords

# 4. Build images
docker compose -f docker/compose.yml build

# 5. First boot: apply schema
docker compose -f docker/compose.yml up -d postgres
sleep 5
docker compose -f docker/compose.yml exec postgres \
    psql -U pixel -d pixel -f /docker-entrypoint-initdb.d/00-schema.sql

# 6. Start everything
docker compose -f docker/compose.yml up -d

# 7. Follow logs
docker compose -f docker/compose.yml logs -f --tail=100
```

---

## Backup strategy

### Automated daily PostgreSQL backup

```bash
#!/usr/bin/env bash
# /etc/cron.daily/pixel-backup
set -euo pipefail
BACKUP_DIR="/var/backups/pixel"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"

docker compose -f /opt/pixel-server/docker/compose.yml exec -T postgres \
    pg_dump -U pixel pixel | gzip > "$BACKUP_DIR/postgres_$DATE.sql.gz"

# Keep 14 days
find "$BACKUP_DIR" -name "postgres_*.sql.gz" -mtime +14 -delete
```

### Redis backup

Redis is configured with `save 60 1` (snapshot every 60 s if ≥ 1 key changed). The `redis-data` Docker volume is snapshotted by the host backup cron. Redis loss means session eviction — all players must re-login, which is acceptable. **No game state lives only in Redis.**

---

## Horizontal scaling for game-svc

When concurrent users exceed ~1 000, add more `game-svc` replicas:

```bash
docker compose -f docker/compose.yml up --scale game-svc=3 -d
```

Each replica subscribes to the `rooms.assign` NATS work queue. When a room is first loaded, one replica claims it and all subsequent traffic for that room routes to that instance via `room.worker.<roomID>` subject. No sticky sessions at the Nginx level are needed — the gateway routes per-packet over NATS, not per-connection.

**When to scale:**

| Concurrent users | Recommended game-svc instances |
|---|---|
| < 500 | 1 |
| 500–1 500 | 2 |
| 1 500–3 000 | 3–4 |
| > 3 000 | Add more; also consider separating PostgreSQL to a managed instance |

All other services (gateway, auth, social, navigator, catalog, moderation) are effectively stateless and can be scaled similarly.

---

## Resource tuning guide

### PostgreSQL on a small VPS

Add to `docker/compose.yml` under `postgres.command`:

```yaml
command: >
  postgres
  -c shared_buffers=128MB
  -c work_mem=4MB
  -c maintenance_work_mem=64MB
  -c max_connections=50
  -c wal_level=minimal
  -c checkpoint_completion_target=0.9
```

### Redis memory policy

Already set in compose: `--maxmemory 128mb --maxmemory-policy allkeys-lru`. Redis is a cache; eviction of least-recently-used keys is acceptable.

### NATS JetStream

`max_mem_store: 64mb` is sufficient. Message retention is `WorkQueuePolicy` (delete on ack) so messages do not accumulate.

---

## Monitoring (lightweight)

For a minimal self-hosted setup use **Prometheus + Grafana via Docker**:

```yaml
# Add to compose.yml
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    ports: ["9090:9090"]

  grafana:
    image: grafana/grafana:latest
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-admin}
    ports: ["3000:3000"]
    volumes:
      - grafana-data:/var/lib/grafana
```

Each pixel-server binary exposes `/metrics` on a dedicated internal port. Metrics to alert on:

| Metric | Alert threshold |
|---|---|
| `game_rooms_active` | > 80 % of configured max |
| `gateway_connections` | > 90 % of expected max |
| `nats_publish_errors_total` | Any increase |
| `postgres_pool_wait_seconds_p99` | > 100 ms |
| `log_entries_dropped_total` | Any increase |

---

## Security hardening

- PostgreSQL and Redis are **not** bound to `0.0.0.0` — they are only accessible within the compose network.
- Only ports **80**, **443**, and **2096** are exposed to the host.
- The WebSocket game port (2096) should be rate-limited at the firewall level to ~10 new connections/second per IP:

```bash
# UFW example
sudo ufw limit 2096/tcp comment 'pixel-server game ws'
```

- Gateway inspects the `X-Real-IP` header (set by Nginx) for IP bans. Nginx is the TLS terminator; plain WebSocket on 2096 can optionally also be exposed directly for clients that do not use HTTPS.
- Rotate `JWT_SECRET` and `POSTGRES_PASSWORD` periodically. After rotating `JWT_SECRET` all sessions are invalidated — players must re-login.
