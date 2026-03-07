# HANDSHAKE-SECURITY

## Implemented Scope

The handshake-security realm is implemented in `internal/auth` with full phase-1 packet handling.

### C2S handled

- `4000` `handshake.release_version`
- `1053` `handshake.client_variables`
- `3110` `handshake.init_diffie`
- `773` `handshake.complete_diffie`
- `2490` `security.machine_id`
- `96` `handshake.client_latency_measure`
- `26979` `handshake.client_policy`
- `2419` `security.sso_ticket`

### S2C produced

- `1347` `handshake.init_diffie`
- `3885` `handshake.complete_diffie`
- `1488` `security.machine_id` (on normalization)
- `2491` `authentication.ok`
- `3523` `handshake.identity_accounts` (empty list)
- `1004` `connection.error` (reject/timeout)

## Package Layout

- `internal/auth/domain`
- `internal/auth/app`
- `internal/auth/adapters/http`
- `internal/auth/adapters/transport`
- `internal/auth/adapters/memory`
- `internal/auth/messaging`
- `internal/auth/register.go`

## Transport Contracts

Auth consumes:

- `packet.c2s.handshake-security.<sessionID>`

Auth publishes:

- `session.output.<sessionID>`
- `session.authenticated`
- `session.disconnect.<sessionID>`

Rejected handshake sessions are removed from auth realm state immediately before disconnect publish to prevent duplicate timeout-disconnect paths.

## Plugin Events

- `auth.handshake.release_version.received`
- `auth.handshake.diffie.initialized`
- `auth.handshake.diffie.completed`
- `auth.handshake.machine_id.received`
- `auth.ticket.validated`

## HTTP Admin Endpoints

- `POST /api/v1/auth/tickets`
- `DELETE /api/v1/auth/tickets/:ticket`

## Tests

- `internal/auth/app/service_test.go`
- `internal/auth/app/handshake_test.go`
- `internal/auth/adapters/transport/subscriber_test.go`
- `internal/auth/adapters/transport/subscriber_handshake_test.go`
- `internal/auth/register_test.go`
- `e2e/06_auth_phase3_e2e_test.go`
- `e2e/07_handshake_security_e2e_test.go`
