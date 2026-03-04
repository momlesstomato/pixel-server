# Handshake — Error Handling

Every error path that can occur during the handshake sequence, what is logged,
and what (if anything) the client receives.

---

## Gateway errors

### WebSocket upgrade failure

**Trigger**: Malformed HTTP request, client does not support WebSocket.  
**Log**: `warn` — `"WebSocket upgrade failed"` with error.  
**Client receives**: HTTP error response from the OS TCP stack. No WS frame.

### NATS publish failure (forwarding C2S packet)

**Trigger**: NATS connection is down or subject unreachable.  
**Log**: `error` — `"publish handshake packet"` with sessionID and error.  
**Client receives**: Nothing. The session remains open; the auth service never
sees the packet. The client may time out waiting for a response.

### NATS subscribe failure for session output

**Trigger**: NATS connection error when setting up `session.output.<sessionID>`.  
**Log**: `error` — recorded at session creation time.  
**Effect**: The session is still opened but the client will never receive any
server-initiated packets. Auth responses will be silently lost.

---

## Auth errors

### Envelope decode failure (sessionID / headerID / payload)

**Trigger**: NATS message body is malformed or truncated.  
**Log**: `error` — `"decode envelope sessionID"` / `"decode envelope headerID"` /
`"decode envelope payload"` with error.  
**Client receives**: Nothing. Handler returns immediately without responding.

### release_version decode failure

**Trigger**: Packet payload cannot be decoded into `HandshakeReleaseVersionInPacket`.  
**Log**: `error` — `"decode release_version"`.  
**Client receives**: Nothing.

### complete_diffie decode failure

**Trigger**: Packet payload cannot be decoded.  
**Log**: `error` — `"decode complete_diffie"`.  
**Client receives**: Nothing. The DH exchange stalls; the client will typically
disconnect or time out.

### sso_ticket decode failure

**Trigger**: Packet payload cannot be decoded.  
**Log**: `error` — `"decode sso_ticket"`.  
**Client receives**: Nothing. No auth.ok, no login bundle.

### Invalid SSO ticket

**Trigger**: `TicketStore.Validate` returns `false` (ticket not found or already
consumed).  
**Log**: `warn` — `"invalid SSO ticket"` with sessionID.  
**Client receives**: Nothing. No `authentication.ok`. The client UI will stall
or show a connection error.

### Post-auth burst send failure

**Trigger**: NATS `Publish` fails for any of the three burst packets.  
**Log**: `error` — `"send auth burst"` with packet name and error.  
**Effect**: Depending on which packet fails, the client may be partially
initialised. The `session.authenticated` event is still published if the failure
is transient — auth does not gate the event on burst delivery.

### Unknown handshake header

**Trigger**: Footer from the client does not match any known C2S header.  
**Log**: `warn` — `"unknown handshake header"` with header ID and sessionID.  
**Client receives**: Nothing.

---

## Summary table

| Scenario | Log level | Client impact |
|---|---|---|
| WS upgrade failure | warn | Connection refused |
| NATS forward failure | error | Packet lost, session stalls |
| Malformed envelope | error | Packet ignored |
| Decode failure (any packet) | error | Packet ignored |
| Invalid SSO ticket | warn | No authentication, client stalls |
| Post-auth burst send failure | error | Partial initialisation possible |
| Unknown handshake header | warn | Packet ignored |
