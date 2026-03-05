# LOGGING

## Overview

`pkg/log` provides logging configuration validation and zap logger construction.

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
