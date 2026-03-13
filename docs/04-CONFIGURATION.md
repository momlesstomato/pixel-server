# Configuration

All configuration is loaded via Viper from a `.env` file and/or environment
variables. Environment variables take precedence over `.env` values.

Variable names are derived from the config struct field path using
`SECTION_FIELD` naming (e.g., `APP_PORT`, `REDIS_ADDRESS`). If an `--env-prefix`
flag is set, variables are prefixed (e.g., `PIXELSV_APP_PORT`).

Fields without a `default` tag are **required** — the server will not start if
they are missing.

---

## Application

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `APP_BIND_IP` | string | `0.0.0.0` | No | Network interface address to bind |
| `APP_PORT` | int | `3000` | No | HTTP/WebSocket listener port |
| `APP_NAME` | string | `pixel-server` | No | Logical service name (used in logs) |
| `APP_ENVIRONMENT` | string | `development` | No | Runtime environment (`development`, `production`) |
| `APP_API_KEY` | string | — | **Yes** | Shared secret for API route authentication |

**Recommended values:**

| Environment | `APP_BIND_IP` | `APP_PORT` | `APP_ENVIRONMENT` |
|-------------|---------------|------------|-------------------|
| Development | `127.0.0.1` | `3000` | `development` |
| Production | `0.0.0.0` | `3000` | `production` |
| Docker | `0.0.0.0` | `3000` | `production` |

---

## Redis

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `REDIS_ADDRESS` | string | — | **Yes** | Redis endpoint (`host:port`) |
| `REDIS_PASSWORD` | string | `""` | No | Redis authentication password |
| `REDIS_DB` | int | `0` | No | Logical database index (0–15) |
| `REDIS_POOL_SIZE` | int | `20` | No | Maximum pooled connections |

**Recommended values:**

| Environment | `REDIS_ADDRESS` | `REDIS_POOL_SIZE` |
|-------------|-----------------|-------------------|
| Development | `localhost:6379` | `10` |
| Production (single) | `redis:6379` | `20` |
| Production (cluster) | `redis-sentinel:26379` | `50` |

> **Note**: Redis 6.2+ is required for `GETDEL` support used by SSO token
> validation.

---

## PostgreSQL

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `POSTGRES_DSN` | string | — | **Yes** | Connection string (`postgres://user:pass@host:port/db?sslmode=disable`) |
| `POSTGRES_MAX_OPEN_CONNS` | int | `30` | No | Maximum open connections |
| `POSTGRES_MAX_IDLE_CONNS` | int | `10` | No | Maximum idle connections |
| `POSTGRES_CONN_MAX_LIFETIME_SECONDS` | int | `300` | No | Connection max lifetime (seconds) |
| `POSTGRES_MIGRATION_AUTO_UP` | bool | `false` | No | Run migrations on startup |
| `POSTGRES_SEED_AUTO_UP` | bool | `false` | No | Run seeds on startup |
| `POSTGRES_MIGRATION_TABLE` | string | `schema_migrations` | No | Migration tracking table name |
| `POSTGRES_SEED_TABLE` | string | `schema_seeds` | No | Seed tracking table name |

**Recommended values:**

| Environment | `POSTGRES_MAX_OPEN_CONNS` | `MIGRATION_AUTO_UP` | `SEED_AUTO_UP` |
|-------------|--------------------------|---------------------|----------------|
| Development | `10` | `true` | `true` |
| Production | `30` | `false` | `false` |

> **Production note**: Run migrations explicitly via `pixelsv db migrate`
> before deploying. Do not enable auto-migration in production.

---

## Logging

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `LOGGING_FORMAT` | string | `console` | No | Output format: `console` or `json` |
| `LOGGING_LEVEL` | string | `info` | No | Minimum log level: `debug`, `info`, `warn`, `error` |

**Recommended values:**

| Environment | `LOGGING_FORMAT` | `LOGGING_LEVEL` |
|-------------|-----------------|-----------------|
| Development | `console` | `debug` |
| Production | `json` | `info` |
| Debugging | `console` | `debug` |

---

## Users

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `USERS_JWT_SECRET` | string | — | **Yes** | HMAC signing key for JWT tokens |
| `USERS_PASSWORD_COST` | int | `12` | No | bcrypt cost factor for password hashing |
| `USERS_SESSION_TTL_SECONDS` | int | `86400` | No | User session token lifetime (seconds) |

**Recommended values:**

| Environment | `USERS_PASSWORD_COST` | `USERS_SESSION_TTL_SECONDS` |
|-------------|----------------------|----------------------------|
| Development | `4` | `86400` (24h) |
| Production | `12` | `86400` (24h) |

> **Security note**: `USERS_JWT_SECRET` should be a random string of at least
> 32 characters. Never reuse across environments.

---

## Authentication (SSO)

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `AUTHENTICATION_DEFAULT_TTL_SECONDS` | int | `300` | No | Default SSO ticket lifetime (seconds) |
| `AUTHENTICATION_MAX_TTL_SECONDS` | int | `1800` | No | Maximum allowed SSO ticket lifetime |
| `AUTHENTICATION_KEY_PREFIX` | string | `sso` | No | Redis key prefix for SSO tickets |

**Recommended values:**

| Environment | `DEFAULT_TTL_SECONDS` | `MAX_TTL_SECONDS` |
|-------------|----------------------|-------------------|
| Development | `3600` (1h) | `7200` (2h) |
| Production | `300` (5m) | `1800` (30m) |

> **Note**: Shorter TTLs are more secure. The client must use the SSO ticket
> within the TTL window or it expires.

---

## Hotel Status

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `STATUS_OPEN_HOUR` | int | `0` | No | UTC hour when hotel opens daily |
| `STATUS_OPEN_MINUTE` | int | `0` | No | UTC minute when hotel opens |
| `STATUS_CLOSE_HOUR` | int | `23` | No | UTC hour when hotel closes daily |
| `STATUS_CLOSE_MINUTE` | int | `59` | No | UTC minute when hotel closes |
| `STATUS_REDIS_KEY` | string | `hotel:status` | No | Redis key for hotel state |
| `STATUS_BROADCAST_CHANNEL` | string | `broadcast:all` | No | Pub/Sub channel for hotel broadcasts |
| `STATUS_COUNTDOWN_TICK_SECONDS` | int | `60` | No | Interval for closing countdown ticks |
| `STATUS_DEFAULT_MAINTENANCE_DURATION_MINUTES` | int | `15` | No | Default maintenance window duration |

**Recommended values:**

| Environment | `OPEN_HOUR` | `CLOSE_HOUR` | `COUNTDOWN_TICK_SECONDS` |
|-------------|-------------|--------------|--------------------------|
| Development | `0` | `23` | `10` |
| Production | `6` | `2` | `60` |
| 24/7 Server | `0` | `23` (close_minute=59) | `60` |

---

## Example .env File

```env
# Application
APP_API_KEY=change-me-to-a-secure-random-string
APP_PORT=3000
APP_ENVIRONMENT=development

# Redis
REDIS_ADDRESS=localhost:6379

# PostgreSQL
POSTGRES_DSN=postgres://pixel:pixel@localhost:5432/pixel?sslmode=disable
POSTGRES_MIGRATION_AUTO_UP=true
POSTGRES_SEED_AUTO_UP=true

# Logging
LOGGING_FORMAT=console
LOGGING_LEVEL=debug

# Users
USERS_JWT_SECRET=change-me-to-a-secure-random-string-32chars

# Authentication (SSO)
AUTHENTICATION_DEFAULT_TTL_SECONDS=3600

# Hotel Status
STATUS_OPEN_HOUR=0
STATUS_CLOSE_HOUR=23
STATUS_CLOSE_MINUTE=59
```

---

## CLI Flags

These flags are passed to `pixelsv serve` and override specific behaviors:

| Flag | Default | Description |
|------|---------|-------------|
| `--env-file` | `.env` | Path to environment file |
| `--env-prefix` | _(none)_ | Prefix for all environment variables |
| `--ws-path` | `/ws` | WebSocket endpoint path |
| `--api-key-header` | `X-API-Key` | HTTP header name for API key |
