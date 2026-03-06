# STORAGE

## Overview

Storage follows ports-and-adapters:

- Ports: `pkg/storage/interfaces`
- PostgreSQL adapter: `pkg/storage/postgres`
- Redis adapter: `pkg/storage/redis`

## PostgreSQL

- Connection pool service in `postgres.Service`
- Generic query helper in `postgres.FetchOne`
- Domain repositories must be implemented in realm adapter packages (`internal/<realm>/adapters/postgres/`), not in `pkg/storage/postgres`

## Redis

- Connection service in `redis.Service`
- Generic key/value adapter via `redis.KVStore`
- Domain stores must be implemented in realm adapter packages (`internal/<realm>/adapters/redis/`), not in `pkg/storage/redis`

## Role-Aware Access

Each role process only connects to the storage backends it needs:
- `gateway` needs Redis only (sessions, ban cache)
- `game` needs both PostgreSQL and Redis
- `catalog` needs PostgreSQL only

## E2E Steps

- `e2e/01_config_e2e_test.go`
- `e2e/02_storage_e2e_test.go`
