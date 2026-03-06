# SESSIONS

## Overview

WebSocket session management is implemented in:

- `pkg/core/session`
- `pkg/http/ws`

`pkg/core/session` is a reusable runtime port that tracks active session ids and outbound binary writers.

## Session Manager

Core types:

- `session.Connection` (`WriteBinary`, `Close`)
- `session.Writer` (`Send`)
- `session.Manager` (`Register`, `Send`, `Remove`, `IDs`, `Count`)

Behavior:

- session ids are required and unique.
- writes are serialized per connection.
- remove closes and unregisters the session.

## HTTP Gateway Integration

`pkg/http/ws.Gateway` provides:

- websocket upgrade middleware for `/ws`
- connection lifecycle registration/removal
- frame decode and packet-topic publish
- transport subscription for `session.output.<sessionID>`

Session ids are generated as monotonic base36 values (`1`, `2`, `3`, ...).
