# CONFIGURATION

## Overview

Configuration is composed from one shared Viper instance.

- `pkg/config` owns shared app-level config.
- `pkg/log` owns logging config.
- `pkg/http` owns HTTP/WebSocket runtime config.
- `pkg/storage/postgres` owns postgres adapter config.
- `pkg/storage/redis` owns redis adapter config.

Configuration defaults are sourced from struct tags using the `default:"..."` convention.
Fields without a `default` tag are treated as required and must pass validation.

## Environment Variables

- `APP_ENV` default: `development`
- `HTTP_ADDR` default: `:8080`
- `HTTP_READ_TIMEOUT_SECONDS` default: `10`
- `OPENAPI_PATH` default: `/openapi.json`
- `SWAGGER_PATH` default: `/swagger`
- `API_KEY` required
- `LOG_FORMAT` default: `console` values: `console`, `json`
- `LOG_LEVEL` default: `info` values: zap-compatible levels
- `POSTGRES_URL` required
- `POSTGRES_MIN_CONNS` default: `1`
- `POSTGRES_MAX_CONNS` default: `10`
- `REDIS_URL` required
- `REDIS_KEY_PREFIX` default: `pixelsv`
- `REDIS_SESSION_TTL_SECONDS` default: `3600`

## Source Order

1. Defaults
2. `.env` file when present
3. Environment variables

## Composition Usage

```go
v, err := config.NewViper(config.DefaultLoadOptions())
if err != nil {
    return err
}
if err := log.BindViper(v); err != nil {
    return err
}
if err := httppkg.BindViper(v); err != nil {
    return err
}
if err := postgres.BindViper(v); err != nil {
    return err
}
if err := redis.BindViper(v); err != nil {
    return err
}
baseCfg, err := config.FromViper(v)
if err != nil {
    return err
}
_ = baseCfg
```

The same Viper instance is reused by package-level `BindViper`/`FromViper` functions so `cmd` composition remains explicit and extendable.

## Ownership Rule

Only package-owned settings are allowed in each package config.

Command entrypoints are responsible for composing final runtime config.
