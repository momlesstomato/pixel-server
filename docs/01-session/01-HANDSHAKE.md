# Handshake Flow

## Overview

The handshake is the first phase of every WebSocket connection. It establishes
the client's identity, optionally negotiates encryption, and authenticates via
an SSO ticket. No game packets are processed until the handshake completes
successfully.

## Connection Sequence

```
Client                                    Server
  |                                          |
  |─── WebSocket Upgrade GET /ws ──────────>│ HTTP 101 Switching Protocols
  |                                          │ Generate connID (32 hex chars)
  |                                          │ Start auth timeout (30s)
  │                                          │
  │─── release_version (4000) ─────────────>│ Read client build info
  │─── client_variables (1053) ────────────>│ Read client URLs (informational)
  │─── security.machine_id (2490) ─────────>│ Validate machine fingerprint
  │<──────── security.machine_id (1488) ────│ Return validated/regenerated ID
  │                                          │
  │    [Optional: Encryption]                │
  │─── init_diffie (3110) ─────────────────>│ Request DH parameters
  │<──────── init_diffie (1347) ────────────│ RSA-signed prime & generator
  │─── complete_diffie (773) ──────────────>│ RSA-encrypted client DH public key
  │<──────── complete_diffie (3885) ────────│ Server DH public key; RC4 installed
  │                                          │
  │─── security.sso_ticket (2419) ─────────>│ Authenticate with SSO token
  │                                          │ Validate ticket (Redis GETDEL)
  │                                          │ Check duplicate login
  │                                          │ Register session
  │<──────── authentication.ok (2491) ──────│ Auth success
  │<──────── identity_accounts (3523) ──────│ Account list
  │<──────── availability.status (2033) ────│ Hotel state (post-auth burst)
  │<──────── client.ping (3928) ────────────│ Start heartbeat cycle
  │                                          │
  │    [Session active]                      │
```

## Pre-Authentication Packets

### release_version (C2S 4000)

The first packet the client sends. Contains the Nitro build version, client
type, platform identifier, and device category.

| Field | Type | Description |
|-------|------|-------------|
| `releaseVersion` | string | Nitro build string (e.g., `"PRODUCTION-202401"`) |
| `clientType` | string | Client type identifier |
| `platform` | int32 | Platform code |
| `deviceCategory` | int32 | Device category code |

The server reads and logs this information. No response is sent.

### client_variables (C2S 1053)

Client metadata including resource URLs.

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | int32 | Client instance ID |
| `clientUrl` | string | Client application URL |
| `externalVariablesUrl` | string | External variables URL |

Informational only. The server reads and discards.

### security.machine_id (C2S 2490 / S2C 1488)

The client sends its machine fingerprint. The server validates it:

**Inbound fields (C2S):**

| Field | Type | Description |
|-------|------|-------------|
| `machineId` | string | 64-character hex string |
| `fingerprint` | string | Browser/device fingerprint |
| `capabilities` | string | Client capability flags |

**Validation rules:**
- Must be exactly 64 hexadecimal characters
- If it starts with `~` or has wrong length, the server generates a new one
- The validated ID is stored on the session for future ban checks

**Response (S2C):**

| Field | Type | Description |
|-------|------|-------------|
| `machineId` | string | Validated or regenerated 64-char hex ID |

## Encryption (Optional)

Encryption uses Diffie-Hellman key exchange with RSA-signed parameters and
RC4 stream ciphers. It is optional — the Nitro client works without it when
connecting over `wss://` (TLS).

### Phase 1: init_diffie (C2S 3110 / S2C 1347)

Client sends an empty packet requesting DH parameters. Server responds with
RSA-signed prime and generator:

| Field (S2C) | Type | Description |
|-------------|------|-------------|
| `encryptedPrime` | string | RSA-signed DH prime (128-bit) |
| `encryptedGenerator` | string | RSA-signed DH generator |

### Phase 2: complete_diffie (C2S 773 / S2C 3885)

Client sends its DH public key encrypted with RSA. Server decrypts it,
computes the shared key, and activates RC4:

| Field (C2S) | Type | Description |
|-------------|------|-------------|
| `encryptedPublicKey` | string | RSA-encrypted client DH public key |

| Field (S2C) | Type | Description |
|-------------|------|-------------|
| `encryptedPublicKey` | string | RSA-encrypted server DH public key |
| `serverClientEncryption` | bool | Whether server→client is encrypted |

After this exchange, all subsequent packets are RC4-encrypted in both
directions. The cipher uses SHA1 of the shared DH key as the RC4 key material.

## Authentication Timeout

If the client does not send `security.sso_ticket` within 30 seconds of
connecting, the server:

1. Sends `disconnect.reason` (4000) with reason code 114 (idle, not authenticated)
2. Closes the WebSocket connection

This prevents connections from sitting idle indefinitely without authenticating.
