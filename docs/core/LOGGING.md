# LOGGING

## Overview

`pkg/log` provides logging configuration validation and zap logger construction.

Runtime logging policy:

- lifecycle events are emitted at `info` level (startup, service activation, listening address, shutdown)
- Fiber per-request access logs are enabled only when `LOG_LEVEL=debug`
- error logs remain enabled for Fiber/app failures
- relevant runtime internals emit `debug` diagnostics in console mode, including packet ingress/egress and storage query/command traces

## Settings

- `logging.format` values: `console`, `json`
- `logging.level` zap-compatible levels

Both settings provide defaults via config struct tags.

## Usage

```go
logCfg, err := log.FromViper(v)
if err != nil {
    return err
}
logger, err := log.New(logCfg)
if err != nil {
    return err
}
defer logger.Sync()
```

## HTTP Logging Behavior

- `debug`: request/response access logs are emitted by Fiber middleware.
- `info` and above: request/response access logs are suppressed to keep logs clean.
- Fiber and application errors are still logged.
- 4xx client errors (for example `404 Cannot GET /`) are treated as client diagnostics and are logged only at `debug`, not as `error`.
- 5xx and unexpected server failures remain `error` logs.

## Storage and Packet Debug Behavior

- PostgreSQL:
  - `pgx` query tracing is enabled when `LOG_LEVEL=debug`.
  - helper query wrappers emit debug lines with query text and args.
- Redis:
  - client command hooks emit debug lines for commands and pipelines.
  - key-value store operations (`set/get/del`) emit debug lines with namespaced keys.
- WebSocket and auth transport:
  - inbound packet decode and topic routing emit debug lines.
  - outbound packet writes and session disconnect control publishes emit debug lines.
- Session-connection runtime:
  - keepalive output (`client.ping`) and timeout disconnect publishes emit debug lines.
  - accepted telemetry packets emit throttled debug diagnostics per session/header.
