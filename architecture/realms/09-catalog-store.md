# Realm: Catalog & Store

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 90 | **Phase:** 7 (Economy) | **Packets:** 21 (10 c2s, 11 s2c)
> **Services:** catalog | **Status:** Not yet implemented

---

## Overview

The Catalog & Store realm handles store browsing, product display, purchases, gift wrapping, voucher redemption, and bundle discounts. Despite having only 21 packets (the second smallest realm with actual business logic), it is one of the most economically sensitive realms -- every purchase modifies user currency and creates inventory items. Correctness is paramount.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 7

---

## Packet Inventory

### C2S (Client to Server) -- 10 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1195 | `catalog.get_index` | `type:string` | Request catalog page tree |
| 412 | `catalog.get_page` | `pageId:int32`, `offerId:int32`, `type:string` | Request catalog page content |
| 3492 | `catalog.purchase` | `pageId:int32`, `offerId:int32`, `extraData:string`, `amount:int32` | Purchase item(s) |
| 1411 | `catalog.purchase_gift` | `pageId`, `offerId`, `extraData`, `receiverName`, `message`, `spriteId`, `color`, `ribbonId`, `anonymous` | Purchase as gift |
| 339 | `catalog.redeem_voucher` | `code:string` | Redeem discount/credit voucher |
| 418 | `catalog.get_gift_wrapping_config` | _(none)_ | Request gift wrap options |
| 2436 | `catalog.get_gift` | `presentId:int32` | Open received gift |
| 1347 | `catalog.check_giftable` | `offerId:int32` | Check if offer can be gifted |
| 223 | `catalog.bundle_discount_ruleset` | _(none)_ | Request bundle discount rules |
| 2150 | `catalog.mark_new_additions_opened` | _(none)_ | Mark new items as seen |

### S2C (Server to Client) -- 11 packets

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 1032 | `catalog.index` | page tree (id, name, icon, children[], offers count, visible, min_rank) | Catalog page hierarchy |
| 804 | `catalog.page` | `pageId`, `type`, `images[]`, `texts[]`, `offers[]` (each with items, credits cost, points cost, limited stock) | Page content with product details |
| 869 | `catalog.purchase_ok` | `offer`, `items[]`, `creditCost`, `pointsCost` | Purchase successful |
| 1404 | `catalog.purchase_error` | `errorCode:int32` | Purchase failed |
| 3770 | `catalog.purchase_not_allowed` | `errorCode:int32` | Purchase blocked (rank/VIP) |
| 1866 | `catalog.published` | `furniActive:boolean` | Catalog data updated notification |
| 2234 | `catalog.gift_wrapping_config` | colors, ribbons, boxes, wrapping prices | Gift wrap configuration |
| 3336 | `catalog.voucher_ok` | `productName`, `credits`, `points` | Voucher redeemed successfully |
| 714 | `catalog.voucher_error` | `errorCode:string` | Voucher redemption failed |
| 1517 | `catalog.gift_receiver_not_found` | _(none)_ | Gift target user doesn't exist |
| 2347 | `catalog.bundle_discount_ruleset` | discount tiers | Bundle pricing rules |

---

## Architecture Mapping

### Service Ownership

The **catalog module** is a dedicated bounded context:

```
Client ──packet──▶ Gateway ──NATS──▶ Catalog Service
                                          │
                   ◀──NATS(session.output)─┘
                                          │
                   ──NATS(catalog.purchase_completed)──▶ Game Service (inventory)
```

### Database Tables

| Table | Columns (Key) | Usage |
|-------|---------------|-------|
| `catalog_pages` | id, parent_id, caption, icon, visible, min_rank, order, page_type, page_layout | Page hierarchy |
| `catalog_offers` | id, page_id, sprite_name, cost_credits, cost_points, cost_seasonal, amount, limited_total, limited_sold, club_only | Product offers |
| `catalog_offer_items` | offer_id, item_definition_id, amount | Items per offer |
| `vouchers` | code, credits, points, items, redeemed_by, redeemed_at, expires_at | Voucher definitions |
| `gift_wrapping_config` | box_types, ribbon_types, colors, prices | Gift wrap options |

### NATS Subjects

| Subject | Direction | Purpose |
|---------|-----------|---------|
| `catalog.input.<sessionID>` | gateway -> catalog | Incoming catalog packets |
| `session.output.<sessionID>` | catalog -> gateway | Outgoing responses |
| `catalog.purchase_completed` | catalog -> game | Item creation in inventory |
| `catalog.published` | catalog -> gateway | Catalog data refreshed (broadcast) |

---

## Implementation Analysis

### Purchase Flow (Critical Path)

The purchase flow is the most economically sensitive operation:

```
1. Client sends catalog.purchase (3492) with pageId, offerId, amount
2. Catalog service validates:
   a. Page exists, visible, user rank >= min_rank
   b. Offer exists on page, club_only check
   c. Amount >= 1 and <= configurable max (default: 100)
   d. Limited stock check (if limited_total > 0):
      - SELECT limited_sold FROM catalog_offers WHERE id = ? FOR UPDATE
      - If limited_sold + amount > limited_total → error
   e. Currency check:
      - Total cost = (cost_credits * amount, cost_points * amount)
      - Apply bundle discount rules if applicable
      - User has sufficient credits, points, seasonal
3. Deduct currency (atomic):
   - UPDATE users SET credits = credits - ?, pixels = pixels - ? WHERE id = ? AND credits >= ? AND pixels >= ?
   - If affected rows = 0 → insufficient funds (race-safe)
4. Create items:
   - For each item in offer × amount:
     - INSERT INTO items (user_id, definition_id, extra_data, ...)
   - If limited: UPDATE catalog_offers SET limited_sold = limited_sold + ?
5. Publish catalog.purchase_completed via NATS:
   - Game service updates user's inventory in memory
   - Sends inventory.unseen_items and inventory.item_add to client
6. Send catalog.purchase_ok (869) to buyer
7. If gift: send to receiver instead (see gift flow)
```

**Atomicity is critical.** The currency deduction and item creation must be in the same database transaction. If item creation fails after currency is deducted, the user loses credits with no items.

### Bundle Discount System

Bulk purchases get discounted:

```
catalog.bundle_discount_ruleset:
  tiers:
    - quantity: 2, creditDiscount: 0%, pointsDiscount: 5%
    - quantity: 5, creditDiscount: 10%, pointsDiscount: 10%
    - quantity: 10, creditDiscount: 20%, pointsDiscount: 20%
    - quantity: 40, creditDiscount: 40%, pointsDiscount: 40%
    - quantity: 99, creditDiscount: 50%, pointsDiscount: 50%
```

The discount applies to the full amount, not incrementally. Calculate: `finalCost = baseCost * amount * (1 - discountPercent)`.

### Gift Flow

`catalog.purchase_gift` (1411) extends the purchase with gift wrapping:

```
1. Validate everything in normal purchase flow
2. Additional validation:
   a. Offer is giftable (check_giftable)
   b. Receiver username exists
   c. Receiver is not the sender (optional rule)
3. Purchase item as normal but set user_id to receiver
4. Create gift wrapper item:
   - extra_data contains: message, anonymous flag, original sprite
   - Gift wrapping type determined by spriteId, color, ribbonId
5. Add gift to receiver's inventory
6. Send notification to receiver if online
```

**Caveat from PlusEMU:** Exchange items (credit furni) in gifts are auto-redeemed by some emulators. pixel-server should NOT auto-redeem -- let the receiver open the gift and decide.

### Voucher Redemption

```
1. Client sends catalog.redeem_voucher (339) with code
2. SELECT * FROM vouchers WHERE code = ? AND redeemed_by IS NULL AND (expires_at IS NULL OR expires_at > NOW())
3. If not found → catalog.voucher_error ("invalid_code")
4. If found:
   a. UPDATE vouchers SET redeemed_by = ?, redeemed_at = NOW()
   b. Add credits/points to user
   c. If voucher includes items: create items
   d. Send catalog.voucher_ok with details
```

**Race condition:** Two clients redeeming the same voucher simultaneously. Use `FOR UPDATE` row lock or `UPDATE ... WHERE redeemed_by IS NULL` with affected-rows check.

### Limited Edition Items

Limited items have tracked serial numbers:

```
limited_total: 100 (total ever available)
limited_sold: 73 (already sold)

On purchase:
  limited_number = limited_sold + 1  (this buyer's serial)
  UPDATE catalog_offers SET limited_sold = limited_sold + 1
  INSERT INTO items_limited_edition (item_id, limited_number, limited_total)
```

**Display in catalog:** Show "73/100 sold" to create urgency.

---

## Caveats & Edge Cases

### 1. Race Condition on Limited Items
Two users buying the last limited item simultaneously. Solution: `UPDATE catalog_offers SET limited_sold = limited_sold + 1 WHERE id = ? AND limited_sold < limited_total RETURNING limited_sold`. If no rows returned, item is sold out.

### 2. Catalog Page Caching
The catalog tree rarely changes. Cache the full tree in Redis with a manual invalidation mechanism. On catalog update (admin action), publish `catalog.published` (1866) to force client refresh.

### 3. Currency Type Handling
Offers can cost credits, activity points, seasonal points, or a combination. Each currency type must be deducted independently. The `cost_seasonal` field has a season type enum (0=duckets, 5=diamonds, 103=loyalty points, etc.).

### 4. Pet Purchase Special Flow
Pet offers create a pet entity, not a furniture item. The purchase handler must route to pet creation logic based on the item definition's interaction type.

### 5. Club-Only Pages
Pages with `club_only=true` should be visible but not purchasable by non-HC users. The page response includes the club restriction flag; the client grays out the buy button. Server-side must still validate on purchase.

### 6. Offer Amount Limits
Some offers should be limited to 1 per purchase (e.g., trophies with custom text, mannequins). The item definition's interaction type determines the max amount.

### 7. Gift to Offline User
Gifts to offline users must be stored in the database. When the receiver logs in, the inventory loader includes gift items. The gift notification (minimail or alert) should also be stored.

### 8. Catalog Refresh After Purchase
After a purchase, the client does NOT automatically re-fetch the catalog page. If the purchase depleted a limited item, other clients won't see the update until they refresh. Consider pushing `catalog.published` after limited item purchases.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Purchase atomicity** | Separate currency deduct + item create (unsafe) | Single PostgreSQL transaction |
| **Limited items** | Application-level lock (single-server only) | Database-level `FOR UPDATE` (multi-instance safe) |
| **Catalog caching** | In-memory, stale across servers | Redis cache with NATS-driven invalidation |
| **Bundle discounts** | Hardcoded tiers | Database-configurable discount rulesets |
| **Voucher safety** | No race protection | Row-level locking with `FOR UPDATE` |
| **Purchase routing** | Giant switch statement | Handler registry by interaction type |
| **Currency types** | Mixed handling | Unified multi-currency transaction |

---

## Dependencies

- **Phase 2 (Identity)** -- user credits and currency for purchase validation
- **Phase 6 (Furniture)** -- item definitions for offer display
- **Phase 7 (Inventory)** -- item creation in user inventory
- **pkg/catalog** -- domain models (Page, Offer, Voucher)
- **PostgreSQL** -- catalog tables, items table, vouchers

---

## Testing Strategy

### Unit Tests
- Bundle discount calculation for all tiers
- Purchase validation (rank, club, stock, currency)
- Voucher validation (expired, redeemed, valid)
- Gift wrapping data serialization
- Limited item serial number assignment

### Integration Tests
- Full purchase flow against real PostgreSQL
- Concurrent limited item purchase (race condition)
- Voucher redemption with race protection
- Gift purchase and delivery to offline user
- Catalog tree loading and caching

### E2E Tests
- Client browses catalog, selects item, purchases, sees it in inventory
- Client buys gift, receiver sees it on login
- Client redeems voucher, credits appear
- Two clients race for last limited item, only one succeeds
