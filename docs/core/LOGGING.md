# LOGGING

## Overview

`pkg/log` provides logging configuration validation and zap logger construction.

Runtime logging policy:

- lifecycle events are emitted at `info` level (startup, service activation, listening address, shutdown)
- Fiber per-request access logs are enabled only when `LOG_LEVEL=debug`
- error logs remain enabled for Fiber/app failures

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
