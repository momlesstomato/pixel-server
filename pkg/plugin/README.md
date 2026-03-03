# pkg/plugin

`pkg/plugin` is the host/plugin contract used to extend services without violating ECS room ownership rules.

## Package structure

- `pkg/plugin`: `Plugin` interface, metadata, loader, registry, host API.
- `pkg/plugin/event`: synchronous in-process event bus with cancellable events.
- `pkg/plugin/intercept`: packet before/after interception pipeline.
- `pkg/plugin/roomsvc`: room facade that avoids exposing mutable ECS internals.

## Lifecycle

1. `Registry.LoadAll()` discovers `*.so` and `*.dylib` in the configured plugins directory.
2. Registry sorts plugins by declared `Meta.Depends`.
3. `OnEnable(api)` runs in dependency order.
4. `OnDisable()` runs in reverse order.
5. Event and packet hook subscriptions are auto-cancelled on disable.

## Plugin interface

Plugins must export:

```go
func NewPlugin() plugin.Plugin
```

And implement:

- `Meta() Meta`
- `OnEnable(api API) error`
- `OnDisable() error`

## Host API available to plugins

- `Scope()`: service name, node ID, version.
- `Events()`: in-process synchronous event bus.
- `Packets()`: packet interceptors (`Before`/`After`).
- `Rooms()`: room snapshot + broadcast facade.
- `Logger()`: plugin-scoped Zap logger.
- `Config()`: raw bytes from `<plugin-name>.yml` if present.

## Operational constraints

- Event handlers and packet hooks run synchronously on the caller goroutine.
- In `game`, handlers may execute on the room tick goroutine; avoid blocking and I/O.
- Runtime unload is not supported by Go's `plugin` package.
- Loading shared-object plugins is supported on Linux and macOS only.

## Build a plugin

```bash
go build -buildmode=plugin -o plugins/my_plugin.so ./path/to/plugin
```
