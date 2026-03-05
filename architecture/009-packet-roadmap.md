# Packet Implementation Roadmap

## Overview

The `pixel-protocol` spec defines **922 packets** (461 c2s + 461 s2c as of spec v1.0.0) across 21 realms. This document defines the implementation order, grouped into phases that each produce a system that is usable and testable end-to-end even before later phases are complete.

Terminology for this roadmap:

- "Service" means internal module/bounded context inside the single `pixelsv` binary.
- "NATS subject" means internal contract topic unless an external broker adapter is explicitly enabled.

The phases are ordered by **dependency depth**, not by packet count. A phase is not started until all packets it depends on (for a connected session to function) are implemented.

---

## How to read this document

Each phase lists:
- Target realms and packet counts
- The minimal capability unlocked at phase completion
- Implementation notes specific to pixel-server's architecture
- Entry and exit criteria (what needs to pass in CI before the phase is "done")

**"Implemented"** means: packet is decoded in `pkg/protocol`, handler is registered in the appropriate module, handler contains correct business logic (not a TODO stub), at least one happy-path integration test passes.

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
- `tools/protogen` reads `spec/protocol.yaml` and emits stubs for all 922 packets in `pkg/protocol`. The stubs panic ("not implemented"); they compile.  
- `pkg/codec` with `Reader`/`Writer` and round-trip tests for all primitive types.  
- `pkg/bus` thin internal contract bus abstraction.  
- `pkg/storage` generic interfaces (`RowQuerier`, `KeyValueStore`) with in-memory fakes.  
- `pkg/pathfinding` 3D A* with full unit test suite (see [005-pathfinding-3d.md](005-pathfinding-3d.md)).  
- `pkg/ecs` component registration; empty `World`; one system skeleton.  
- Docker Compose: `pixelsv`, `postgres`, `redis`.  
- Atlas migration: `users`, `rooms`, `items`, `bans` tables created.  
- CI: `go build ./...`, `go test ./...`, `go vet ./...`, lint all pass.  

Exit: `go build ./...` green; all tables created; `pkg/pathfinding` 100% test coverage.

---

## Phase 1 — Connection (43 packets)

**Realms:** `handshake-security` (13), `session-connection` (30)  
**Services:** gateway, auth  

**Capability unlocked:** A Nitro client can connect, complete Diffie-Hellman, authenticate via SSO ticket, and maintain a session (keep-alive, ping/pong, disconnect cleanly).

### handshake-security (13 packets)

| ID | Direction | Name | Notes |
|---|---|---|---|
| 4000 | c2s | `handshake.release_version` | Read, validate client version string; reject if unknown |
| 1053 | c2s | `handshake.client_variables` | Read and discard (no validation required by spec) |
| 3110 | c2s | `handshake.init_diffie` | Verify encryption module; send s2c response |
| 773  | c2s | `handshake.complete_diffie` | Compute shared secret; install RC4 on gateway session |
| 2419 | c2s | `security.sso_ticket` | Validate token in Redis; resolve userID |
| 1735 | c2s | `handshake.machine_id` | Store client fingerprint on session |
| 1347 | s2c | `handshake.init_diffie` | RSA-signed DH prime + generator |
| 3885 | s2c | `handshake.complete_diffie` | Server DH public key |
| _remaining_ | s2c | Auth error packets | Disconnect with reason codes |

### session-connection (30 packets)

Key packets:
- `session.ping` / `session.pong` — keep-alive (gateway proxies without hitting the game module)
- `session.latency_measure` (c2s + s2c) — round-trip time tracking
- `session.disconnect` (s2c) — graceful close with reason code
- `availability.status` (s2c) — hotel open/closed flag on login
- `connection.error` (s2c) — general error envelope

Implementation note: Most session-connection packets are handled entirely in the gateway with no external broker round-trip. The gateway handles ping/pong inline; only authenticated session state changes are published to internal topics.

Exit: A Nitro client connects to `gateway` on port 2096, completes handshake, receives the `availability.status` packet, and maintains an idle connection without disconnecting.

---

## Phase 2 — Identity (56 packets)

**Realms:** `user-profile` (56)  
**Services:** auth, game (thin profile load)  

**Capability unlocked:** A logged-in user has a name, figure, motto, credits, activity points, and subscription status. The client receives the `user.authenticated` data composite and loads the hotel view.

Key packets:
- `user.authenticated` (s2c) — sends user data after SSO
- `user.figure_update` (c2s + s2c) — appearance change
- `user.motto_update` (c2s + s2c) — motto change
- `user.credits` (s2c) — credit balance
- `user.activity_points` (s2c) — activity / pixel points balance
- `user.subscription` (s2c) — HC/VIP status
- `user.settings` (c2s + s2c) — client preferences (volume, old chat mode, etc.)
- `user.ignore_list` (s2c) — ignored user IDs on login
- `user.wardrobe` (s2c) — saved outfits
- `user.badges` (s2c) — badge collection subset (full inventory in Phase 7)

Implementation note: Profile data is loaded through domain-owned repositories built on top of generic `pkg/storage/postgres` primitives. After `session.authenticated` is received by the game module, it prepares a `user.authenticated` composite and publishes it to `session.output.<sessionID>`.

### Permissions parity plan (Phase 2 entry requirement)

Before Phase 2 is marked complete, permission behavior must be aligned to vendor baselines:

- **Comet v2** references:
	- `vendor/comet-v2/Comet-Server/src/main/java/com/cometproject/server/storage/queries/permissions/PermissionsDao.java`
	- `vendor/comet-v2/Comet-Server/src/main/java/com/cometproject/server/game/rooms/types/components/RightsComponent.java`
	- `vendor/comet-v2/database.sql` tables: `server_permissions_ranks`, `permission_perks`, `permission_commands`, `room_rights`
- **Arcturus** references:
	- `vendor/Arcturus-Community/src/main/java/com/eu/habbo/habbohotel/permissions/PermissionsManager.java`
	- `vendor/Arcturus-Community/src/main/java/com/eu/habbo/habbohotel/permissions/Rank.java`
- **PlusEMU** references:
	- `vendor/PlusEMU/**` patterns around `Permissions.HasRight(...)`, rank checks, and room-rights checks

Planned pixel-server implementation in Phase 2:

1. Introduce identity permission profile builder in `internal/modules/game/identity` (started).
2. Map storage rank/perk state to `user.permissions` packet payload (`clubLevel`, `securityLevel`, ambassador).
3. Reserve room-rights and command-rights enforcement for Phase 3+ while keeping profile permissions deterministic in Phase 2.
4. Add integration tests asserting rank/perk variants produce expected packet values.

Exit: After login, the client renders the hotel view with its username, figure, and credits displayed.

---

## Phase 3 — Room Entry & Movement (124 packets)

**Realms:** `room` (90), `room-entities` (34)  
**Services:** game (room worker, ECS, pathfinding), navigator (room metadata)  

**Capability unlocked:** Players can enter rooms, see each other, walk, chat, and express basic postures (sit, stand, wave, idle dance). The ECS world and 20 Hz tick loop are live.

### room (90 packets) — selected highlights

| Group | Key packets | Notes |
|---|---|---|
| Entry | `room.enter`, `room.open`, `room.doorbell` | Load room from DB, spawn entity |
| Model | `room.heightmap`, `room.relative_map`, `room.open_connection` | Send layout to client |
| Rights | `room.rights_list`, `room.give_rights`, `room.take_rights` | Rights bitmask |
| Settings | `room.settings`, `room.update_settings` | Owner-only config |
| Banning | `room.ban`, `room.unban`, `room.banned_list` | Ban from room, not global |
| Events | `room.event`, `room.event_cancel` | Room event metadata |
| Moderation | `room.kick`, `room.mute_user` | Room-scoped |
| Doorbell | `room.doorbell_accept`, `room.doorbell_reject` | Lock door flow |

### room-entities (34 packets) — selected highlights

| Group | Key packets | Notes |
|---|---|---|
| Entity list | `room_entities.objects`, `room_entities.statuses` | Full entity dump on enter |
| Movement | `room_entities.move` (c2s), `room_entities.update` (s2c) | Walk command; position batch update |
| Chat | `room_entities.chat`, `room_entities.whisper`, `room_entities.shout` | Chat bubble types |
| Typing | `room_entities.typing_start`, `room_entities.typing_stop` | Typing indicator |
| Expression | `room_entities.action` (c2s) — wave, idle dance | Posture change |
| Hand items | `room_entities.carry_object`, `room_entities.drop_hand_item` | Carry drink/food |

### Implementation notes

- On `room.enter`, the room worker is created (or woken from idle) and the ECS entity for the entering player is spawned.
- `room_entities.move` triggers 3D A* path computation (see [005-pathfinding-3d.md](005-pathfinding-3d.md)); `WalkPath` component is updated.
- Each tick, `MovementSystem` advances the path; `BroadcastSystem` collects dirty entities and publishes `room_entities.update` to all session outputs in the room.
- Chat goes through `ChatCooldownSystem` (rate limiting) before being broadcast.

Exit: Two logged-in clients can enter the same room, see each other's avatars, walk to arbitrary tiles, and chat.

---

## Phase 4 — Navigator (55 packets)

**Realms:** `navigator` (55)  
**Services:** navigator  

**Capability unlocked:** Players can browse categories, search rooms by name, create new rooms, add/remove favourites, and view promoted rooms. Navigator is implemented before furniture because engineers need it to enter rooms for testing during Phases 3 and beyond — without it, room entry requires hard-coded room IDs.

Key packets:
- `navigator.search` (c2s + s2c) — search results
- `navigator.categories` (s2c) — flat/tree category list
- `navigator.room_info` (s2c) — room metadata card
- `navigator.create_flat` (c2s) — create user room
- `navigator.favourites_add`, `navigator.favourites_remove` (c2s)
- `navigator.home_room` (c2s) — set home room
- `navigator.popular_rooms` (s2c) — score-sorted listing
- `navigator.promoted_rooms` (s2c) — advertised rooms

Exit: Navigator window works end-to-end; rooms appear in search results; room creation redirects client to the new room.

---

## Phase 5 — Social & Messenger (30 packets)

**Realms:** `messenger-social` (30)  
**Services:** social  

**Capability unlocked:** Friend lists load, friend requests are sent/accepted/declined, private messages are sent and received in real-time, and room invitations work.

Key packets:
- `messenger.friends` (s2c) — friend list on login
- `messenger.request` (c2s + s2c) — add friend flow
- `messenger.accept`, `messenger.decline` (c2s)
- `messenger.remove` (c2s + s2c)
- `messenger.send_message` (c2s) — private message
- `messenger.message_received` (s2c) — delivery
- `messenger.invite` (c2s + s2c) — room invitation
- `messenger.user_search` (c2s + s2c) — find user by name

Exit: Friend list loads; messages sent from one client arrive in real-time on a second client's messenger window; room invitations send the receiver into the correct room.

---

## Phase 6 — Furniture & Items (99 packets)

**Realms:** `furniture-items` (99)  
**Services:** game (item interaction, ICycleable, WIRED), inventory (item ownership)  

**Capability unlocked:** Floor and wall items can be placed, moved, removed, and interacted with. Rollers, teleporters, gates, crackables, and dimmer operate correctly. WIRED logic engine is live.

Key packet groups:
- `furniture.objects_floor` / `furniture.objects_wall` (s2c) — item dump on room enter
- `furniture.place`, `furniture.move`, `furniture.remove` (c2s + s2c) — placement
- `furniture.update_state` (s2c) — interaction outcome broadcast
- `furniture.interact_floor` / `furniture.interact_wall` (c2s) — trigger interaction
- `furniture.roller_result` (s2c) — roller movement broadcast
- `furniture.teleport` (c2s + s2c sequence) — teleporter flow
- `furniture.dimmer` (c2s + s2c) — room dimmer state
- `furniture.wired_condition`, `furniture.wired_effect`, `furniture.wired_trigger` (c2s + s2c) — WIRED configuration
- `furniture.mannequin`, `furniture.youtube`, `furniture.decoration` — specialty items

Implementation note: `ItemInteractionSystem` replaces the `ICycleable` pattern. Each interaction type is modelled as an ECS component set. WIRED conditions and effects are represented as a directed graph stored per room, evaluated by `WiredSystem` each tick.

Exit: A room with a full furniture layout loads correctly; players can interact with a gate, teleporter, roller, and basic WIRED trigger/effect pair.

---

## Phase 7 — Economy, Catalog & Inventory (108 packets)

**Realms:** `economy-trading` (54), `catalog-store` (21), `inventory` (33)  
**Services:** catalog, inventory, game (trading session within room)  

**Capability unlocked:** Catalog browsing and purchasing, item inventory management, user-to-user trading, marketplace, and credit display.

### catalog-store (21)
- `catalog.page` (c2s + s2c) — browse pages
- `catalog.offer` (s2c) — product listing
- `catalog.purchase` (c2s) — buy item; triggers `catalog.purchase_completed` contract event
- `catalog.purchase_ok` / `catalog.purchase_failed` (s2c)
- `catalog.gift_wrap` (c2s + s2c) — gift flow
- `catalog.voucher_redeem` (c2s + s2c) — discount voucher

### inventory (33)
- `inventory.items` (s2c) — item list on login (paginated)
- `inventory.unseen_items` (s2c) — newly acquired items highlight
- `inventory.badges` (s2c) — badge collection
- `inventory.badge_equip`, `inventory.badge_unequip` (c2s)

### economy-trading (54)
- `trading.open`, `trading.close` (c2s + s2c) — trade session
- `trading.offer`, `trading.accept`, `trading.confirm` (c2s + s2c)
- `trading.update` (s2c) — live trade state broadcast
- `trading.marketplace_place`, `trading.marketplace_buy` (c2s + s2c)

Exit: Player can buy a furniture item, see it in inventory, place it in a room, and trade it to another player.

---

## Phase 8 — Subscriptions & Offers (50 packets)

**Realms:** `subscription-offers` (50)  
**Services:** catalog, auth (subscription state on user)  

Mostly read/display packets. Key: HC subscription status update, targeted offers, builders-club management.

Exit: HC subscription purchase flow completes; HC badge appears on user profile.

---

## Phase 9 — Pets (41 packets)

**Realms:** `pets` (41)  
**Services:** game (PetAI ECS system)  

**Capability unlocked:** Pets can be placed from inventory, follow their owner, respond to commands, and gain XP/happiness.

Key packets:
- `pets.place` (c2s + s2c) — place pet from inventory
- `pets.respect` (c2s) — give respect
- `pets.info` (s2c) — pet stats card
- `pets.move` (s2c) — pet walk broadcast (same as entity update with KindPet)
- `pets.chat` (c2s) — issue command; triggers `PetAI` command evaluation

Implementation note: Pets use the same ECS entity model as avatars but with the `PetAI` component. `PetAISystem` evaluates happiness decay, energy, and follow logic each tick.

Exit: Pet places, follows owner around the room, responds to "sit"/"stand" commands, and its stats appear in the pet info panel.

---

## Phase 10 — Groups & Forums (64 packets)

**Realms:** `groups-forums` (64)  
**Services:** social (groups), game (group badge display in rooms)  

**Capability unlocked:** Players can create/join/leave groups, manage member lists, assign group home rooms, and participate in group forums.

Exit: Group creation flow completes; group badge appears on member profiles; group forum thread can be posted and read.

---

## Phase 11 — Moderation & Safety (83 packets)

**Realms:** `moderation-safety` (83)  
**Services:** moderation  

**Capability unlocked:** Mod-tool opens; staff can issue bans, mutes, view chat history, handle reports (call-for-help), and use guardian system.

Implementation note: The ban flow is synchronous-critical: `moderation.ban_issued` must reach the gateway within 500 ms. Use Redis `PUBLISH ban:<userID>` from the moderation module; gateway subscribes and closes socket immediately.

Exit: A staff account bans a user via mod-tool; the banned user's socket is closed within 1 second; the ban persists after a new connection attempt.

---

## Phase 12 — Games & Entertainment (49 packets)

**Realms:** `games-entertainment` (49)  
**Services:** game (mini-game systems: Freeze, Banzai, BattleBall, Snowstorm)  

**Capability unlocked:** Room-embedded mini-games operate.

Exit: A Freeze game starts, players throw snowballs, tiles freeze, a winner is declared.

---

## Phase 13 — Remaining Realms (120 packets)

**Realms:**
- `achievements-talents` (24)
- `quests-campaigns` (33)
- `notifications-landing` (22)
- `crafting-recycling` (16)
- `camera-photos` (15)
- `other` (10)

Implemented last because they depend on many earlier systems (items, rooms, users) and are purely additive.

Exit: Achievement unlock triggers on qualifying action; notification toast appears; recycler accepts items; camera captures room snapshot.

---

## Implementation velocity targets

Assuming a small team (2–3 engineers):

| Phase | Packets | Estimated weeks |
|---|---|---|
| 0 — Foundation | 0 | 3–4 |
| 1 — Connection | 43 | 2 |
| 2 — Identity | 56 | 2 |
| 3 — Room & Movement | 124 | 4–5 |
| 4 — Navigator | 55 | 2 |
| 5 — Social | 30 | 2 |
| 6 — Furniture & WIRED | 99 | 4–5 |
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

A CI check fails if any packet reachable from Phase ≤ N is still `stub` when a Phase N milestone is merged to `main`.
