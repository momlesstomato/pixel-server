# Realm: HANDSHAKE-SECURITY

## Position and Scope

- Position: `10`
- Phase: `1` (connection/authentication gate)
- Primary owner: `internal/auth`
- Gateway ingress topic: `packet.c2s.handshake-security.<sessionID>`

This realm is now implemented end-to-end for the phase-1 handshake packet set in the current single-binary architecture.

## Implemented Packet Coverage

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

- `1347` `handshake.init_diffie` (signed prime/generator)
- `3885` `handshake.complete_diffie` (server public key + encryption flag)
- `1488` `security.machine_id` (normalized machine id when client value is invalid)
- `2491` `authentication.ok`
- `3523` `handshake.identity_accounts` (currently empty list)
- `1004` `connection.error` on handshake rejection/timeout

## Runtime Flow

1. Gateway decodes WS frames and publishes handshake packets to auth realm topic.
2. Auth transport adapter decodes and dispatches to app service use cases.
3. App service maintains per-session handshake state:
   - release metadata
   - client variables
   - diffie state
   - machine id
   - authentication status
4. Auth publishes output frames to `session.output.<sessionID>`.
5. On invalid handshake or timeout, auth publishes:
   - `connection.error` output frame
   - `session.disconnect.<sessionID>` control topic
6. Gateway closes the target session on disconnect control topic.

## Topology and Contracts

Realm-owned contracts:

- transport topics and parsers: `internal/auth/messaging/topics.go`
- realm events: `internal/auth/messaging/events.go`

Session-connection realm contracts used by auth:

- `session.authenticated`
- `session.connected`
- `session.disconnected`
- `session.output.<sessionID>`
- `session.disconnect.<sessionID>`

## Plugin Extension Surface

Implemented auth events for plugin hooks:

- `auth.handshake.release_version.received`
- `auth.handshake.diffie.initialized`
- `auth.handshake.diffie.completed`
- `auth.handshake.machine_id.received`
- `auth.ticket.validated`

Events are emitted from app-layer boundaries, not transport adapters.

## Vendor Alignment Notes

Implementation behavior was aligned to legacy emulator handshake sequencing patterns from vendor references:

- `Arcturus-Community`
- `PlusEMU`
- `comet-v2`

Adaptations to our architecture:

- realm logic remains in `internal/auth`
- transport is `pkg/core/transport` (local/NATS) instead of direct socket-service coupling
- session closure is done via realm-scoped control topic, not direct socket access from auth adapter

## Security and Performance Behavior

- Diffie handshake state is tracked per session.
- Unauthenticated handshake sessions expire and are disconnected by timeout monitor.
- Machine id normalization enforces 64-char non-prefixed values.
- Reject paths are fail-fast with explicit connection error frame plus disconnect control.

## Remaining Non-Phase-1 Work

- persistent auth backends (Redis/PostgreSQL) replacing in-memory ticket adapter
- enforced diffie-before-auth mode via configuration toggles
- richer identity account payload beyond empty list
- ban and sanction validation in auth path
