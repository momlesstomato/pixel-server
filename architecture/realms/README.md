# Realm Analysis Index

This directory contains deep-dive planning analysis for each protocol realm.

Terminology note for this directory:

- References to "services" map to **internal modules/bounded contexts** inside the single `pixelsv` binary.
- References to "NATS subjects" map to **internal contract topics** unless explicitly marked as external adapter integration.

These analyses complement the [Packet Implementation Roadmap](../009-packet-roadmap.md) with per-realm architecture planning and reference emulator comparisons.

---

## Implementation Order

Documents are numbered by implementation order, matching the phases in [009-packet-roadmap.md](../009-packet-roadmap.md). Work top-to-bottom; each realm depends on the realms above it.

| # | Realm | Packets | Phase | Module(s) | Doc |
|---|-------|---------|-------|------------|-----|
| 01 | Handshake & Security | 13 (8 c2s, 5 s2c) | 1 — Connection | gateway, auth | [01-handshake-security.md](01-handshake-security.md) |
| 02 | Session & Connection | 30 (10 c2s, 20 s2c) | 1 — Connection | gateway | [02-session-connection.md](02-session-connection.md) |
| 03 | User & Profile | 56 (29 c2s, 27 s2c) | 2 — Identity | game (identity) | [03-user-profile.md](03-user-profile.md) |
| 04 | Room | 90 (46 c2s, 44 s2c) | 3 — Room & Movement | game (room worker) | [04-room.md](04-room.md) |
| 05 | Room Entities | 34 (14 c2s, 20 s2c) | 3 — Room & Movement | game (ECS) | [05-room-entities.md](05-room-entities.md) |
| 06 | Navigator | 55 (37 c2s, 18 s2c) | 4 — Navigator | navigator | [06-navigator.md](06-navigator.md) |
| 07 | Messenger & Social | 30 (14 c2s, 16 s2c) | 5 — Social | social | [07-messenger-social.md](07-messenger-social.md) |
| 08 | Furniture & Items | 99 (52 c2s, 47 s2c) | 6 — Furniture & WIRED | game (items, WIRED) | [08-furniture-items.md](08-furniture-items.md) |
| 09 | Catalog & Store | 21 (10 c2s, 11 s2c) | 7 — Economy | catalog | [09-catalog-store.md](09-catalog-store.md) |
| 10 | Economy & Trading | 54 (28 c2s, 26 s2c) | 7 — Economy | game, catalog | [10-economy-trading.md](10-economy-trading.md) |
| 11 | Inventory | 33 (13 c2s, 20 s2c) | 7 — Economy | game (inventory) | [11-inventory.md](11-inventory.md) |
| 12 | Subscription & Offers | 50 (26 c2s, 24 s2c) | 8 — Subscriptions | catalog, auth | [12-subscription-offers.md](12-subscription-offers.md) |
| 13 | Pets | 41 (21 c2s, 20 s2c) | 9 — Pets | game (PetAI) | [13-pets.md](13-pets.md) |
| 14 | Groups & Forums | 64 (36 c2s, 28 s2c) | 10 — Groups | social, game | [14-groups-forums.md](14-groups-forums.md) |
| 15 | Moderation & Safety | 83 (43 c2s, 40 s2c) | 11 — Moderation | moderation | [15-moderation-safety.md](15-moderation-safety.md) |
| 16 | Games & Entertainment | 49 (21 c2s, 28 s2c) | 12 — Games | game (mini-games) | [16-games-entertainment.md](16-games-entertainment.md) |
| 17 | Achievements & Talents | 24 (10 c2s, 14 s2c) | 13 — Remaining | game | [17-achievements-talents.md](17-achievements-talents.md) |
| 18 | Quests & Campaigns | 33 (15 c2s, 18 s2c) | 13 — Remaining | game | [18-quests-campaigns.md](18-quests-campaigns.md) |
| 19 | Notifications & Landing | 22 (6 c2s, 16 s2c) | 13 — Remaining | gateway, game | [19-notifications-landing.md](19-notifications-landing.md) |
| 20 | Crafting & Recycling | 16 (9 c2s, 7 s2c) | 13 — Remaining | game | [20-crafting-recycling.md](20-crafting-recycling.md) |
| 21 | Camera & Photos | 15 (8 c2s, 7 s2c) | 13 — Remaining | game | [21-camera-photos.md](21-camera-photos.md) |
| 22 | Other | 10 (7 c2s, 3 s2c) | 13 — Remaining | game | [22-other.md](22-other.md) |

---

## Phase Summary

| Phase | Realms | Total Packets | Est. Weeks |
|-------|--------|---------------|------------|
| **1 — Connection** | 01 Handshake, 02 Session | 43 | 2 |
| **2 — Identity** | 03 User & Profile | 56 | 2 |
| **3 — Room & Movement** | 04 Room, 05 Room Entities | 124 | 4-5 |
| **4 — Navigator** | 06 Navigator | 55 | 2 |
| **5 — Social** | 07 Messenger & Social | 30 | 2 |
| **6 — Furniture & WIRED** | 08 Furniture & Items | 99 | 4-5 |
| **7 — Economy** | 09 Catalog, 10 Trading, 11 Inventory | 108 | 4 |
| **8 — Subscriptions** | 12 Subscription & Offers | 50 | 2 |
| **9 — Pets** | 13 Pets | 41 | 2 |
| **10 — Groups** | 14 Groups & Forums | 64 | 3 |
| **11 — Moderation** | 15 Moderation & Safety | 83 | 3 |
| **12 — Games** | 16 Games & Entertainment | 49 | 3 |
| **13 — Remaining** | 17-22 (Achievements, Quests, Notifications, Crafting, Camera, Other) | 120 | 4 |
| | | **922** | **~41 weeks** |

---

## Totals

- **22 realms**, **922 packets** (463 c2s + 459 s2c)
- **13 implementation phases** spanning ~41 weeks for a 2-3 engineer team
- **7 bounded contexts** (gateway, auth, game, social, navigator, catalog, moderation)

## How to Read These Documents

Each realm analysis follows a consistent structure:

1. **Overview** -- realm purpose, packet counts, phase mapping
2. **Packet Inventory** -- complete table of every C2S and S2C packet with IDs, fields, and summaries
3. **Architecture Mapping** -- which module owns the realm, contract topics, database tables
4. **Implementation Analysis** -- detailed approach for pixel-server, reference emulator patterns
5. **Caveats & Edge Cases** -- pitfalls observed in Comet v2, Arcturus, and PlusEMU
6. **Improvements Over Legacy** -- where pixel-server's architecture enables better solutions
7. **Dependencies** -- what must be implemented before this realm
8. **Testing Strategy** -- unit, integration, and e2e test requirements
