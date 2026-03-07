# E2E STEPS

- `01_config_e2e_test.go`: validates end-to-end configuration composition flow.
- `02_storage_e2e_test.go`: validates storage adapter round-trips and optional live PostgreSQL ping.
- `03_api_e2e_test.go`: validates HTTP runtime composition with API-key protected admin endpoint.
- `04_transport_e2e_test.go`: validates transport factory selection and local bus round-trip flow.
- `05_protocol_e2e_test.go`: validates websocket binary protocol ingress decode/publish and session output fan-out.
- `06_auth_phase3_e2e_test.go`: validates phase 3 auth realm scaffold (ticket create/validate and authentication.ok frame).
- `07_handshake_security_e2e_test.go`: validates full handshake-security packet flow (diffie, machine id, release/client metadata, ticket auth).

Run all e2e tests:

```bash
go test ./e2e/...
```
