# Plugin System (Phase 1)

This document describes implemented plugin core infrastructure in code.

## Implemented Packages

- `pkg/plugin`
  - Stable plugin contracts: `Metadata`, `Plugin`, `API`, `Event`, `PacketInterceptor`, `RoomService`, `PluginStore`.
- `pkg/plugin/eventbus`
  - In-process `EventBus` implementation.
  - Supports subscribe/unsubscribe and event emission.
  - Recovers handler panics and auto-deregisters panicking handlers.
- `pkg/plugin/interceptor`
  - Packet before/after hook chain implementation.
  - Supports header-specific and global hooks.
  - Recovers panics and removes panicking hooks.
- `pkg/plugin/loader`
  - `.so` discovery from directory (`Discover`).
  - Shared object loading (`OpenSharedObject`) with `NewPlugin` symbol resolution.
  - Dependency topological sort (`SortByDependencies`).
  - Role/realm-aware lifecycle registry (`Registry` with `EnableAll`, `DisableAll`, `Status`).

## Current Constraints

- Plugin runtime discovery/loading is not yet wired into serve startup orchestration.
- Shared in-process plugin event bus is initialized in startup and injected into auth realm registration.
- Auth realm emits handshake/auth events:
  - `auth.handshake.release_version.received`
  - `auth.handshake.diffie.initialized`
  - `auth.handshake.diffie.completed`
  - `auth.handshake.machine_id.received`
  - `auth.ticket.validated`
- Gateway packet interception integration remains pending.
- Route registrar, storage, and room service concrete plugin adapters are pending integration phases.

## Quality Status

- Unit tests cover:
  - event bus behavior and panic isolation
  - interceptor behavior and panic isolation
  - discovery filtering and error propagation
  - dependency sorting and lifecycle gating
