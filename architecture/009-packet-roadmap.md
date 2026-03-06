# Packet Implementation Roadmap

## Overview

The `pixel-protocol` spec defines **922 packets** (461 c2s + 461 s2c as of spec v1.0.0) across 21 realms. This document defines the implementation order, grouped into phases that each produce a system that is usable and testable end-to-end even before later phases are complete.

Terminology for this roadmap:

- "Module" means internal realm/bounded context inside the single `pixelsv` binary (`internal/<realm>/`).
- "Contract topic" means internal messaging topic. In all-in-one mode these are in-process channels; in distributed mode they are NATS subjects. The topic names are identical either way.

The phases are ordered by **dependency depth**, not by packet count. A phase is not started until all packets it depends on (for a connected session to function) are implemented.

---

## How to read this document

Each phase lists:
- Target realms and packet counts
- The minimal capability unlocked at phase completion
- Implementation notes specific to pixelsv's architecture
- Entry and exit criteria (what needs to pass in CI before the phase is "done")

**"Implemented"** means: packet is decoded in `pkg/protocol`, handler is registered in the appropriate realm adapter (`internal/<realm>/adapters/ws/`), handler contains correct business logic (not a TODO stub), at least one happy-path integration test passes.

---

## Packet count by realm

| Realm | Total | c2s | s2c |
|---|---|---|---|
| `furniture-items` | 99 | ~50 | ~49 |
| `room` | 90 | ~45 | ~45 |
| `moderation-safety` | 83 | ~42 | ~41 |
| `groups-forums` | 64 | ~32 | ~32 |
| `user-profile` | 56 | ~28 | ~28 |
| `navigator` | 55 | ~28 | ~27 |
| `economy-trading` | 54 | ~27 | ~27 |
| `subscription-offers` | 50 | ~25 | ~25 |
| `games-entertainment` | 49 | ~25 | ~24 |
| `pets` | 41 | ~21 | ~20 |
| `room-entities` | 34 | ~17 | ~17 |
| `quests-campaigns` | 33 | ~17 | ~16 |
| `inventory` | 33 | ~17 | ~16 |
| `session-connection` | 30 | ~15 | ~15 |
| `messenger-social` | 30 | ~15 | ~15 |
| `achievements-talents` | 24 | ~12 | ~12 |
| `notifications-landing` | 22 | ~11 | ~11 |
| `catalog-store` | 21 | ~11 | ~10 |
| `crafting-recycling` | 16 | ~8 | ~8 |
| `handshake-security` | 13 | ~7 | ~6 |
| `other` | 10 | ~5 | ~5 |
| **Total** | **922** | **461** | **461** |

---

## Phase 0 — Foundation (no packets)

**Goal:** Every structural decision in place before a single packet handler is written.

Deliverables:
- Root `go.mod` (`module pixelsv`) and root `go.work` (`use .`).
- `cmd/pixelsv` with Cobra command graph (`serve`, `migrate`, `jobs`).
- `tools/protogen` reads `spec/protocol.yaml` and emits stubs for all 922 packets in `pkg/protocol`. The stubs panic ("not implemented"); they compile.
- `pkg/codec` with `Reader`/`Writer` and round-trip tests for all primitive types.
- `pkg/http` with Fiber setup, Swagger UI, API-key middleware, WebSocket endpoint, health/ready probes.
- `pkg/config` with Viper-backed structured configuration.
- `pkg/log` with Zap logger factory.
- `pkg/storage` generic interfaces with in-memory fakes.
- `pkg/pathfinding` 3D A* with full unit test suite (see [005-pathfinding-3d.md](005-pathfinding-3d.md)).
- `internal/game/domain/` component registration; empty `RoomWorld`; one system skeleton.
- `pkg/core/transport/` local bus implementation.
- Docker Compose: `pixelsv`, `postgres`, `redis`.
- Atlas migration: `users`, `rooms`, `items`, `bans` tables created.
- CI: `go build ./...`, `go test ./...`, `go vet ./...`, lint all pass.

Exit: `go build ./...` green; all tables created; `pkg/pathfinding` 100% test coverage.

---

## Phase 1 — Connection (43 packets)

**Realms:** `handshake-security` (13), `session-connection` (30)
**Modules:** gateway, auth

**Capability unlocked:** A Nitro client can connect, complete Diffie-Hellman, authenticate via SSO ticket, and maintain a session (keep-alive, ping/pong, disconnect cleanly).

### handshake-security (13 packets)

| ID | Direction | Name | Notes |
|---|---|---|---|
| 4000 | c2s | `handshake.release_version` | Read, validate client version string; reject if unknown |
| 1053 | c2s | `handshake.client_variables` | Read and discard (no validation required by spec) |
| 3110 | c2s | `handshake.init_diffie` | Verify encryption module; send s2c response |
| 773  | c2s | `handshake.complete_diffie` | Compute shared secret; install RC4 on session |
| 2419 | c2s | `security.sso_ticket` | Validate token in Redis; resolve userID |
| 1735 | c2s | `handshake.machine_id` | Store client fingerprint on session |
| 1347 | s2c | `handshake.init_diffie` | RSA-signed DH prime + generator |
| 3885 | s2c | `handshake.complete_diffie` | Server DH public key |
| _remaining_ | s2c | Auth error packets | Disconnect with reason codes |

### session-connection (30 packets)

Key packets:
- `session.ping` / `session.pong` — keep-alive (handled inline by gateway, no cross-module round-trip)
- `session.latency_measure` (c2s + s2c) — round-trip time tracking
- `session.disconnect` (s2c) — graceful close with reason code
- `availability.status` (s2c) — hotel open/closed flag on login
- `connection.error` (s2c) — general error envelope

Implementation note: Most session-connection packets are handled entirely in the gateway module with no cross-module communication. The gateway handles ping/pong inline; only authenticated session state changes are published to contract topics.

Exit: A Nitro client connects to `pixelsv` WebSocket endpoint, completes handshake, receives the `availability.status` packet, and maintains an idle connection without disconnecting.

---

## Phase 2 — Identity (56 packets)

**Realms:** `user-profile` (56)
**Modules:** auth, game (thin profile load)

**Capability unlocked:** A logged-in user has a name, figure, motto, credits, activity points, and subscription status. The client receives the `user.authenticated` data composite and loads the hotel view.

Key packets:
- `user.authenticated` (s2c) — sends user data after SSO
- `user.figure_update` (c2s + s2c) — appearance change
- `user.motto_update` (c2s + s2c) — motto change
- `user.credits` (s2c) — credit balance
- `user.activity_points` (s2c) — activity / pixel points balance
- `user.subscription` (s2c) — HC/VIP status
- `user.settings` (c2s + s2c) — client preferences
- `user.ignore_list` (s2c) — ignored user IDs on login
- `user.wardrobe` (s2c) — saved outfits
- `user.badges` (s2c) — badge collection subset

Implementation note: Profile data is loaded through domain-owned repositories (`internal/auth/domain/`). After `session.authenticated` is published, the game module prepares a `user.authenticated` composite and writes it to the session.

### Permissions parity plan (Phase 2 entry requirement)

Before Phase 2 is marked complete, permission behavior must be aligned to vendor baselines (Comet-v2, Arcturus, PlusEMU).

Exit: After login, the client renders the hotel view with its username, figure, and credits displayed.

---

## Phase 3 — Room Entry & Movement (124 packets)

**Realms:** `room` (90), `room-entities` (34)
**Modules:** game (room worker, ECS, pathfinding), navigator (room metadata)

**Capability unlocked:** Players can enter rooms, see each other, walk, chat, and express basic postures. The ECS world and 20 Hz tick loop are live.

Implementation notes:
- On `room.enter`, the room worker is created (or woken from idle) and the ECS entity for the entering player is spawned.
- `room_entities.move` triggers 3D A* path computation; `WalkPath` component is updated.
- Each tick, `MovementSystem` advances the path; `BroadcastSystem` writes state updates to all sessions in the room.
- Chat goes through `ChatCooldownSystem` (rate limiting) before being broadcast.
- In distributed mode, gateway routes `room.input.<roomID>` to the correct game worker instance.

Exit: Two logged-in clients can enter the same room, see each other's avatars, walk to arbitrary tiles, and chat.

---

## Phase 4 — Navigator (55 packets)

**Realms:** `navigator` (55)
**Modules:** navigator

Exit: Navigator window works end-to-end; rooms appear in search results; room creation redirects client to the new room.

---

## Phase 5 — Social & Messenger (30 packets)

**Realms:** `messenger-social` (30)
**Modules:** social

Exit: Friend list loads; messages sent from one client arrive in real-time on a second client's messenger window.

---

## Phase 6 — Furniture & Items (99 packets)

**Realms:** `furniture-items` (99)
**Modules:** game (item interaction, WIRED)

Exit: A room with a full furniture layout loads correctly; players can interact with items.

---

## Phase 7 — Economy, Catalog & Inventory (108 packets)

**Realms:** `economy-trading` (54), `catalog-store` (21), `inventory` (33)
**Modules:** catalog, game (trading session)

Exit: Player can buy a furniture item, see it in inventory, place it in a room, and trade it.

---

## Phase 8 — Subscriptions & Offers (50 packets)

**Realms:** `subscription-offers` (50)
**Modules:** catalog, auth

Exit: HC subscription purchase flow completes; HC badge appears on user profile.

---

## Phase 9 — Pets (41 packets)

**Realms:** `pets` (41)
**Modules:** game (PetAI ECS system)

Exit: Pet follows owner around the room, responds to commands, stats appear in pet info panel.

---

## Phase 10 — Groups & Forums (64 packets)

**Realms:** `groups-forums` (64)
**Modules:** social, game

Exit: Group creation flow completes; group badge appears on member profiles; forum thread can be posted.

---

## Phase 11 — Moderation & Safety (83 packets)

**Realms:** `moderation-safety` (83)
**Modules:** moderation

Implementation note: Ban flow must be fast. In distributed mode: moderation publishes `moderation.ban.issued.<userID>` via NATS; gateway subscribes and closes socket immediately. In all-in-one mode: direct channel notification.

Exit: A staff account bans a user via mod-tool; the banned user's socket is closed within 1 second.

---

## Phase 12 — Games & Entertainment (49 packets)

**Realms:** `games-entertainment` (49)
**Modules:** game (mini-game systems)

Exit: A Freeze game starts, players throw snowballs, tiles freeze, a winner is declared.

---

## Phase 13 — Remaining Realms (120 packets)

**Realms:** `achievements-talents` (24), `quests-campaigns` (33), `notifications-landing` (22), `crafting-recycling` (16), `camera-photos` (15), `other` (10)

Exit: Achievement unlock triggers on qualifying action; notification toast appears; recycler accepts items.

---

## Implementation velocity targets

Assuming a small team (2-3 engineers):

| Phase | Packets | Estimated weeks |
|---|---|---|
| 0 — Foundation | 0 | 3-4 |
| 1 — Connection | 43 | 2 |
| 2 — Identity | 56 | 2 |
| 3 — Room & Movement | 124 | 4-5 |
| 4 — Navigator | 55 | 2 |
| 5 — Social | 30 | 2 |
| 6 — Furniture & WIRED | 99 | 4-5 |
| 7 — Economy | 108 | 4 |
| 8 — Subscriptions | 50 | 2 |
| 9 — Pets | 41 | 2 |
| 10 — Groups | 64 | 3 |
| 11 — Moderation | 83 | 3 |
| 12 — Games | 49 | 3 |
| 13 — Remaining | 120 | 4 |
| **Total** | **922** | **~41 weeks** |

---

## Tracking

Each packet's implementation status is tracked in the GitHub project board. The protogen generator marks each handler stub with:

```go
// Status: stub | in-progress | done
// Phase: 3
```

A CI check fails if any packet reachable from Phase <= N is still `stub` when a Phase N milestone is merged to `main`.
