# Realm: Inventory

> **Position:** 120 | **Phase:** 7 (Economy) | **Packets:** 33 (13 c2s, 20 s2c)
> **Services:** game (inventory module) | **Status:** Not yet implemented

---

## Overview

The Inventory realm manages the user's item collection: furniture, badges, avatar effects, bots, pets, and clothing. It handles item listing, badge equipping, effect activation, unseen item tracking, and inventory refresh notifications. Inventory is the bridge between the catalog (acquisition), rooms (placement), and trading (transfer).

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 7

---

## Packet Inventory

### C2S -- 13 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 3150 | `inventory.get_furniture` | _(none)_ | Request furniture inventory |
| 3500 | `inventory.get_furniture_not_in_room` | _(none)_ | Furniture not placed in rooms |
| 2091 | `inventory.get_current_badges` | `userId:int32` | Get another user's equipped badges |
| 2769 | `inventory.get_badges` | _(none)_ | Request full badge collection |
| 644 | `inventory.update_badges` | equipped badges array | Update equipped badge slots |
| 2959 | `inventory.effect_activate` | `effectId:int32` | Activate avatar effect |
| 3493 | `inventory.unseen_reset_category` | `category:int32` | Mark unseen category as seen |
| 2343 | `inventory.unseen_reset_items` | `category:int32`, `itemIds:int32[]` | Mark specific items as seen |
| 21 | `inventory.get_group_badges` | _(none)_ | Get group badge collection |
| 367 | `inventory.get_group_memberships` | _(none)_ | Get group memberships |
| 3848 | `inventory.get_bots` | _(none)_ | Request bot inventory |
| 3095 | `inventory.get_pets` | _(none)_ | Request pet inventory |
| 3374 | `inventory.clothing_redeem` | `itemId:int32` | Redeem clothing item |

### S2C -- 20 packets

| ID | Name | Summary |
|----|------|---------|
| 994 | `inventory.furniture` | Full furniture list (paginated) |
| 104 | `inventory.furniture_add` | New item added |
| 3151 | `inventory.furniture_refresh` | Inventory changed, re-fetch |
| 159 | `inventory.furniture_remove` | Item removed from inventory |
| 1087 | `inventory.current_badges` | Equipped badges for a user |
| 717 | `inventory.badges` | Full badge collection |
| 2493 | `inventory.badge_received` | New badge acquired |
| 2103 | `inventory.unseen_items` | Newly acquired items highlight |
| 2867 | `inventory.effect_added` | New effect available |
| 2228 | `inventory.effect_removed` | Effect expired/removed |
| 1959 | `inventory.effect_activated` | Effect activation confirmed |
| 3473 | `inventory.effect_selected` | Effect selected for display |
| 340 | `inventory.effects` | Full effect list |
| 1450 | `inventory.clothing` | Clothing collection |
| 3086 | `inventory.bots` | Bot inventory |
| 1352 | `inventory.bot_added` | Bot added to inventory |
| 233 | `inventory.bot_removed` | Bot placed/removed |
| 3522 | `inventory.pets` | Pet inventory |
| 2101 | `inventory.pet_added` | Pet added to inventory |
| 3253 | `inventory.pet_removed` | Pet placed/removed |

---

## Implementation Analysis

### Furniture Inventory Loading

The furniture inventory can be very large (thousands of items). Loading strategy:

```
1. On inventory.get_furniture (3150):
   a. Query items WHERE user_id = ? AND room_id IS NULL
   b. Paginate: first page of 1000 items
   c. Send inventory.furniture (994) with items + hasMore flag
2. Client requests subsequent pages if needed
3. Cache item list in session memory for quick access during trades
```

**Performance:** Users with 10,000+ items cause expensive queries. Strategies:
- Lazy loading: only fetch on demand, not on login.
- Redis cache: `inventory:<userId>:items` as hash map.
- Pagination: 1000 items per page, client fetches as user scrolls.

### Unseen Items Tracking

"Unseen" items are newly acquired items highlighted in the inventory UI:

```
Categories:
  1 = Furniture
  2 = Rentables
  3 = Pets
  4 = Badges
  5 = Bots
  6 = Effects
  7 = Clothing

On item acquisition:
  1. Add to unseen list: Redis SET unseen:<userId>:<category> item_ids
  2. Send inventory.unseen_items (2103) with category + item IDs

On category opened:
  1. Client sends inventory.unseen_reset_category (3493)
  2. Clear Redis SET unseen:<userId>:<category>
```

### Badge System

Badges are collectible achievements with 5 equippable slots:

```
Equip flow:
  1. Client sends inventory.update_badges (644) with 5 slots
  2. Validate all badge codes exist in user's collection
  3. Update user_badges SET slot = 0 WHERE user_id = ? (unequip all)
  4. For each equipped: UPDATE user_badges SET slot = ? WHERE user_id = ? AND code = ?
  5. Send inventory.current_badges (1087) to user
  6. If in room: broadcast figure update to room entities
```

### Effect System

Avatar effects are time-limited visual enhancements:

```
Activation:
  1. Client sends inventory.effect_activate (2959) with effectId
  2. Validate effect exists and is not expired
  3. Start duration countdown (effects have limited use time)
  4. Send inventory.effect_activated (1959)
  5. If in room: broadcast room_entities.effect to all users

Expiry:
  - Effects with duration decrement while active
  - When duration = 0: remove effect, send inventory.effect_removed
  - Some effects are permanent (duration = -1)
```

---

## Caveats & Edge Cases

### 1. Inventory Size Limits
No hard protocol limit, but practical limits are needed:
- Default max items: 5000 (configurable)
- HC users: 10,000
- Beyond limit: warn but don't prevent (items from trades/gifts must always be accepted)

### 2. Concurrent Inventory Modifications
Trading, catalog purchases, room pickup, and marketplace all modify inventory simultaneously. All operations must check item ownership at transaction time, not just at request time.

### 3. Badge Slot Uniqueness
A badge can only be in one slot. If the client sends the same badge in slots 1 and 3, reject with error.

### 4. Effect Duration Tracking
Effect duration should only decrement while the user is online. Offline time doesn't count. Track `activated_at` + `remaining_seconds` instead of `expires_at`.

### 5. Clothing Redeem Safety
`inventory.clothing_redeem` (3374) converts a clothing item into unlocked figure parts. This is irreversible -- the item is consumed. Validate the item's definition is a clothing type before consuming.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Loading** | Full inventory on login | Lazy + paginated on demand |
| **Unseen tracking** | Database column per item | Redis set per category |
| **Effect duration** | Wall clock (counts offline) | Active time only |
| **Concurrency** | Application locks | Database transactions |

---

## Dependencies

- **Phase 2 (Identity)** -- badge display on profile
- **Phase 6 (Furniture)** -- item definitions
- **Phase 7 (Catalog, Trading)** -- item acquisition and transfer

---

## Testing Strategy

### Unit Tests
- Badge slot validation (uniqueness, existence)
- Effect duration calculation
- Unseen category tracking

### Integration Tests
- Full inventory CRUD against PostgreSQL
- Paginated loading with 5000+ items
- Concurrent inventory modifications (trade + pickup)

### E2E Tests
- Client opens inventory, sees all items with correct counts
- Client equips badges, other users see them on profile
