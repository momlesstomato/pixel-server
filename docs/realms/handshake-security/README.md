# HANDSHAKE-SECURITY

## Implemented Scope

Phase 3 scaffold for auth realm is implemented in:

- `internal/auth/domain`
- `internal/auth/app`
- `internal/auth/adapters/http`
- `internal/auth/adapters/transport`
- `internal/auth/adapters/memory`
- `internal/auth/register.go`

## Implemented Behavior

- Admin ticket routes:
  - `POST /api/v1/auth/tickets`
  - `DELETE /api/v1/auth/tickets/:ticket`
- Transport subscriber consumes:
  - `packet.c2s.handshake-security.<sessionID>`
- Supported handshake packet action:
  - `security.sso_ticket` validates/consumes one ticket.
- On success:
  - publish `session.authenticated` event (`sessionID + userID` payload)
  - publish `authentication.ok` (`header 2491`) to `session.output.<sessionID>`

## References

- `internal/auth/register.go`
- `internal/auth/adapters/transport/subscriber.go`
- `internal/auth/adapters/http/routes.go`
- tests:
  - `internal/auth/register_test.go`
  - `internal/auth/adapters/transport/subscriber_test.go`
  - `internal/auth/adapters/http/routes_test.go`
  - `e2e/06_auth_phase3_e2e_test.go`
