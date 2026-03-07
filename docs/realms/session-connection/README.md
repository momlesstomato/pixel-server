# session-connection

The session-connection realm is implemented in `internal/sessionconnection` and owns authenticated session lifecycle behavior after handshake.

## Implemented Behavior

- Subscribed lifecycle topics:
  - `session.connected`
  - `session.disconnected`
  - `session.authenticated`
- Subscribed packet ingress topic:
  - `packet.c2s.session-connection.<sessionID>`
- Auth bootstrap output:
  - `2033` `availability.status` sent immediately after `session.authenticated`
- Concurrent login protection:
  - previous active session for the same user is disconnected with reason `2`
- Keepalive:
  - periodic `3928` `client.ping`
  - stale sessions disconnected with reason `4`

## Implemented C2S Handlers

- `295` `client.latency_test` -> responds with `10` `client.latency_response`
- `2596` `client.pong` -> updates pong timestamp
- `2445` `client.disconnect` -> emits `4000` `disconnect.reason` and disconnect control
- `105` `session.desktop_view` -> responds with `122` `session.desktop_view`
- `1160`, `2313`, `3226` -> accepted no-op in phase-1 core
- `3230`, `3457`, `3847` -> telemetry accepted with per-header throttle

## Realm-Owned Contracts

- Topics and payload codecs:
  - `internal/sessionconnection/messaging/topics.go`
  - `internal/sessionconnection/messaging/packet_topics.go`
  - `internal/sessionconnection/messaging/authenticated.go`
- Plugin events:
  - `sessionconnection.session.connected`
  - `sessionconnection.session.disconnected`
  - `sessionconnection.session.authenticated`
  - `sessionconnection.packet.received`
  - `sessionconnection.client.pong.received`
  - `sessionconnection.client.latency_test.received`
  - `sessionconnection.session.desktop_view.received`

## Disconnect Contract

- On server-side protocol/runtime packet failures, gateway attempts to send `disconnect.reason` (`4000`) before socket teardown.
- On process shutdown (`Ctrl+C` / context cancellation), gateway disconnects all active websocket sessions with maintenance reason (`5`) and publishes `session.disconnected` for cleanup subscribers.
- `session.disconnected` is the canonical cleanup trigger for session-bound state; cleanup handlers must be idempotent to tolerate repeated close paths.
- Runtime observability: session-connection transport logs an info line (`session-connection cleanup handled`) when disconnect cleanup executes.

## Tests

- Unit:
  - `internal/sessionconnection/app/service_test.go`
  - `internal/sessionconnection/messaging/topics_test.go`
  - `internal/sessionconnection/messaging/authenticated_test.go`
- Transport integration:
  - `internal/sessionconnection/adapters/transport/subscriber_test.go`
- E2E:
  - `e2e/08_session_connection_e2e_test.go`
