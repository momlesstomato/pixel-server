# Handshake — Packets

Complete reference for every packet that belongs to the handshake-security realm.

---

## Client → Server (C2S)

All C2S packets are forwarded by the gateway to `handshake.c2s.<sessionID>` and
processed by `services/auth/handler.go`.

### release_version — header 4000

Sent immediately after the WebSocket connection is established. Identifies the
client build.

| Field | Type | Description |
|---|---|---|
| `ReleaseVersion` | `string` | Client version string (e.g. `"PRODUCTION-201611291003-338474136"`) |
| `ClientType` | `string` | Client type string (e.g. `"UNITY"`, `"FLASH"`) |
| `Platform` | `int32` | Platform ID |
| `DeviceCategory` | `int32` | Device category ID |

**Handler** (`handleReleaseVersion`): Decodes the packet and logs at `info` level
(`client connected`, version, clientType, platform). No response is sent.

---

### client_variables — header 1053

Sent early in the handshake with client configuration URLs.

| Field | Type | Description |
|---|---|---|
| `ClientId` | `int32` | Internal client identifier |
| `ClientUrl` | `string` | URL where the client was loaded from |
| `ExternalVariablesUrl` | `string` | URL of the external variables config file |

**Handler** (`handleClientVariables`): Debug log only. No response sent.

---

### init_diffie — header 3110

Client requests Diffie-Hellman parameters to begin key negotiation.

*(No fields — empty body.)*

**Handler** (`handleInitDiffie`): Responds immediately with `init_diffie` S2C
(header 2512), carrying placeholder prime and generator strings.

---

### complete_diffie — header 773

Client sends its public key to complete the DH exchange.

| Field | Type | Description |
|---|---|---|
| `EncryptedPublicKey` | `string` | Client's DH public key (currently a stub) |

**Handler** (`handleCompleteDiffie`): Responds with `complete_diffie` S2C
(header 1740), carrying a placeholder server public key and
`ServerClientEncryption = false`.

---

### sso_ticket — header 2419

The pivotal authentication packet. The client presents its SSO ticket.

| Field | Type | Description |
|---|---|---|
| `Ticket` | `string` | Single-use SSO token |

**Handler** (`handleSSOTicket`):
1. Calls `TicketStore.Validate(ticket)` — returns `(userID, ok)`
2. On failure: logs warning `"invalid SSO ticket"`, no response
3. On success: sends post-auth burst (3 packets), publishes `session.authenticated`

---

### machine_id — header 2490

Client sends device identifiers after authentication.

| Field | Type | Description |
|---|---|---|
| `MachineId` | `string` | Hardware machine identifier |
| `Fingerprint` | `string` | Browser/client fingerprint |

**Handler** (`handleMachineID`): Stores `MachineId` on the session (gateway side
via `session.machineID`). Logs at debug level.

---

### unique_id — header 1735

Routed to auth but not currently processed.

**Handler**: Dropped silently. Logged as unknown header at warn level if it falls
through.

---

## Server → Client (S2C)

These packets are sent by the auth service via `session.output.<sessionID>`.

### init_diffie — header 2512

| Field | Type | Description |
|---|---|---|
| `EncryptedPrime` | `string` | DH prime parameter (placeholder string) |
| `EncryptedGenerator` | `string` | DH generator parameter (placeholder string) |

Sent in response to client's `init_diffie` (3110).

---

### complete_diffie — header 1740

| Field | Type | Description |
|---|---|---|
| `EncryptedPublicKey` | `string` | Server's DH public key (placeholder string) |
| `ServerClientEncryption` | `bool` | Whether traffic is encrypted (`false` — stub) |

Sent in response to client's `complete_diffie` (773).

---

### authentication.ok — header 2491

*(Empty body.)*

First packet of the post-auth burst. Confirms SSO validation succeeded.

---

### availability_status — header 2033

| Field | Type | Description |
|---|---|---|
| `IsOpen` | `bool` | Whether the hotel is open (`true`) |
| `OnShutdown` | `bool` | Whether the hotel is shutting down (`false`) |
| `IsAuthentic` | `bool` | Whether this is an authentic hotels server (`true`) |

Second packet of the post-auth burst.

---

### client_ping — header 3928

*(Empty body.)*

Third packet of the post-auth burst. Starts the client's keep-alive pong cycle.

---

## Gateway Routing — Pre-auth Header Classification

The gateway's `routePacket` function classifies by header ID:

| Header IDs | Gateway action |
|---|---|
| 4000, 1053, 3110, 773, 2419, 2490, 1735 | Forward to `handshake.c2s.<sessionID>` |
| 2596 | Pong — acknowledged silently, not forwarded |
| 295 | Latency test — echo `header 10 + int32(0)` back to client |
| 2445 | Client disconnect — close the session |
| All other IDs, session authenticated | Forward to `room.input.<sessionID>` |
| All other IDs, session not authenticated | Drop with warn log |
