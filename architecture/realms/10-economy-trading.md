# Realm: Economy & Trading

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 110 | **Phase:** 7 (Economy) | **Packets:** 54 (28 c2s, 26 s2c)
> **Services:** game (trading within room), catalog (marketplace) | **Status:** Not yet implemented

---

## Overview

The Economy & Trading realm handles user-to-user trading, the marketplace (auction house), currency management, community goals, and credit/point operations. This realm contains some of the most abuse-prone features in the system -- item duplication, currency manipulation, and market manipulation are all risks.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 7

---

## Packet Inventory

### C2S -- 28 packets

#### Currency

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 273 | `user.get_currency` | _(none)_ | Request currency balances |
| 1265 | `room.update_category_trade` | `roomId`, `category`, `tradeMode` | Update room trade setting |

#### User-to-User Trading

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1481 | `trade.open` | `userId:int32` | Initiate trade with user |
| 3863 | `trade.accept` | _(none)_ | Accept current trade state |
| 2341 | `trade.cancel` | _(none)_ | Cancel trade |
| 2551 | `trade.close` | _(none)_ | Close trade window |
| 2760 | `trade.confirm` | _(none)_ | Final confirmation (2nd stage) |
| 3107 | `trade.add_item` | `itemId:int32` | Add single item to offer |
| 1263 | `trade.add_items` | `itemId:int32`, `amount:int32` | Add multiple of same item |
| 3845 | `trade.remove_item` | `itemId:int32` | Remove item from offer |
| 1444 | `trade.unaccept` | _(none)_ | Revoke acceptance |

#### Marketplace

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2597 | `marketplace.get_config` | _(none)_ | Request marketplace settings |
| 3447 | `marketplace.sell_item` | `itemId:int32`, `price:int32` | List item for sale |
| 2105 | `marketplace.get_own_items` | _(none)_ | View own listings |
| 434 | `marketplace.cancel_sale` | `offerId:int32` | Cancel marketplace listing |
| 2407 | `marketplace.search_offers` | `minPrice`, `maxPrice`, `query`, `sortType` | Search marketplace |
| 1603 | `marketplace.buy_offer` | `offerId:int32` | Purchase from marketplace |
| 2650 | `marketplace.redeem_credits` | _(none)_ | Withdraw marketplace earnings |
| 1866 | `marketplace.buy_tokens` | _(none)_ | Purchase marketplace tokens |
| 848 | `marketplace.get_can_sell` | `itemId:int32` | Check if item is sellable |
| 3288 | `marketplace.get_item_stats` | `spriteId:int32` | Get price history |

#### Community Goals

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 90 | `economy.redeem_community_goal_prize` | _(none)_ | Claim community goal reward |
| 1145 | `economy.community_goal_progress` | _(none)_ | Check community goal progress |
| 1343 | `economy.concurrent_users_goal_progress` | _(none)_ | Check concurrent users goal |
| 2167 | `economy.community_goal_hall_of_fame` | _(none)_ | View hall of fame |

### S2C -- 26 packets

#### Trading

| ID | Name | Summary |
|----|------|---------|
| 2505 | `trade.opened` | Trade session started |
| 1001 | `trade.update` | Trade state update (both sides' items) |
| 2720 | `trade.accepted` | Partner accepted |
| 1723 | `trade.confirmed` | Trade completed successfully |
| 1950 | `trade.closed` | Trade closed/cancelled |
| 2843 | `trade.failed` | Trade failed (item missing, etc.) |
| 1628 | `trade.not_allowed` | Trading not allowed in this room |
| 3210 | `trade.already_open` | Already in a trade |

#### Marketplace

| ID | Name | Summary |
|----|------|---------|
| 1823 | `marketplace.config` | Marketplace settings (commission, min/max price) |
| 3290 | `marketplace.own_items` | User's active listings |
| 1973 | `marketplace.offers` | Search results |
| 2032 | `marketplace.buy_result` | Purchase result |
| 3510 | `marketplace.cancel_result` | Cancellation result |
| 1286 | `marketplace.item_stats` | Price history chart data |
| 2306 | `marketplace.sell_result` | Listing creation result |
| 1167 | `marketplace.redeem_result` | Credit withdrawal result |

#### Currency updates, community goal results, etc.

---

## Architecture Mapping

### Trading Ownership

Trading happens within a room worker (two users must be in the same room):

```
Room Worker
├── TradeManager
│   ├── ActiveTrades map[tradeID]*TradeSession
│   └── TradeSession
│       ├── User1 (items offered, accepted flag)
│       ├── User2 (items offered, accepted flag)
│       └── State (open, accepted, confirmed, completed)
```

### Marketplace Ownership

Marketplace is handled by the catalog service (shared economy):

```
Catalog Service
├── MarketplaceModule
│   ├── List item (seller side)
│   ├── Search (buyer side)
│   ├── Buy (transaction)
│   └── Redeem credits (withdrawal)
```

### Database Tables

| Table | Usage |
|-------|-------|
| `trade_log` | trade_id, user1_id, user2_id, timestamp, items_json | Audit trail |
| `marketplace_offers` | id, seller_id, item_id, item_definition_id, price, listed_at, sold_at, buyer_id, status | Active listings |
| `marketplace_earnings` | user_id, pending_credits | Unclaimed marketplace earnings |
| `marketplace_stats` | definition_id, avg_price, min_price, max_price, last_30_days[] | Price history |

---

## Implementation Analysis

### Trading Protocol (Two-Stage Acceptance)

The trading protocol uses a two-stage confirmation to prevent mistakes:

```
Stage 1: Offer Building
  1. trade.open (1481) → create TradeSession
  2. Both users add/remove items
  3. After each change: broadcast trade.update (1001) with both sides
  4. When items change, BOTH accept flags reset

Stage 2: First Accept
  5. Both users send trade.accept (3863)
  6. Both accept flags set → show confirmation dialog
  7. If either changes items → back to Stage 1

Stage 3: Final Confirm
  8. Both users send trade.confirm (2760)
  9. Server validates ALL items still exist in inventories
  10. Atomic transfer:
      a. BEGIN transaction
      b. For each item in User1's offer:
         - Verify item.user_id = user1_id AND item.room_id IS NULL
         - UPDATE items SET user_id = user2_id
      c. For each item in User2's offer:
         - Verify item.user_id = user2_id AND item.room_id IS NULL
         - UPDATE items SET user_id = user1_id
      d. INSERT INTO trade_log
      e. COMMIT
  11. Send trade.confirmed (1723) to both users
  12. Update both inventories via NATS events
```

**Critical safety checks:**
- Items must be in inventory (not placed in room).
- Items must not be in another active trade.
- Items must not have been deleted since being offered.
- Items must not be untradeable (definition flag).

### Marketplace Implementation

```
Listing:
  1. Validate item is marketable (definition.is_tradeable)
  2. Validate price is within bounds (config: min 1, max 10,000,000)
  3. Remove item from user's inventory
  4. INSERT into marketplace_offers (status = 'active')
  5. Deduct listing fee if configured

Search:
  1. Query marketplace_offers WHERE status = 'active'
  2. Apply filters (price range, name, category)
  3. Sort by price/date
  4. Paginate results (max 100 per page)

Purchase:
  1. SELECT ... FROM marketplace_offers WHERE id = ? AND status = 'active' FOR UPDATE
  2. Validate buyer has sufficient credits
  3. Deduct credits from buyer
  4. Add credits to seller's marketplace_earnings (minus commission)
  5. Transfer item to buyer's inventory
  6. Update offer status to 'sold'
  7. If seller online: send notification

Redeem:
  1. SELECT pending_credits FROM marketplace_earnings WHERE user_id = ?
  2. Add to user's credits
  3. Reset pending_credits to 0
```

**Marketplace commission:** Configurable percentage (default: 1 credit per 100 credits, minimum 1 credit). Applied on sale completion, not on listing.

---

## Caveats & Edge Cases

### 1. Item Duplication in Trades
The most critical bug in any emulator. If the item verification and transfer are not in the same transaction, items can be duplicated. **Always use a single PostgreSQL transaction with row-level locks.**

### 2. Trade While Items Placed
Items placed in rooms cannot be traded. The handler must verify `room_id IS NULL` for every offered item. Reference emulators (Comet v2) check this but some skip it for wall items.

### 3. Simultaneous Trades
A user should only be in one trade at a time. The TradeManager must check for existing trades before allowing `trade.open`. Attempting to open a second trade returns `trade.already_open`.

### 4. Trade Disconnect
If a user disconnects during a trade, the trade must be cancelled immediately. All offered items remain with their original owners. No items should be in a "limbo" state.

### 5. Marketplace Price Manipulation
Users could list items at extremely low prices (sniping bait) or extremely high prices (laundering). Implement:
- Minimum price: 1 credit
- Maximum price: configurable (default: 10,000,000)
- Price history tracking for anomaly detection
- Rate limit: max 5 listings per minute

### 6. Exchange Item Trading
Exchange items (credit furni: "CF_1", "CF_5", etc.) can be traded. On confirmation, they could be auto-redeemed to credits or transferred as items. pixel-server should transfer as items (no auto-redeem in trade).

### 7. Trade with Banned/Muted Users
Trade-banned users cannot initiate trades. The `tradeMode` room setting also controls who can trade:
- 0: No trading
- 1: Only rights holders
- 2: Everyone

### 8. Marketplace Expiry
Listings should expire after a configurable period (default: 48 hours). Expired items return to seller's inventory automatically. Background job runs every 15 minutes.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Trade atomicity** | Separate transfers (duplication risk) | Single PostgreSQL transaction |
| **Item verification** | Pre-trade only | Verify at every stage |
| **Trade logging** | Optional/missing | Mandatory audit trail |
| **Marketplace** | In-memory listings | PostgreSQL with `FOR UPDATE` locks |
| **Price history** | Not tracked | Rolling 30-day stats |
| **Marketplace earnings** | Immediate credit | Pending balance + explicit redeem |

---

## Dependencies

- **Phase 3 (Room)** -- users must be in same room to trade
- **Phase 6 (Furniture)** -- item definitions and ownership
- **Phase 7 (Inventory)** -- item add/remove
- **PostgreSQL** -- trades, marketplace, earnings

---

## Testing Strategy

### Unit Tests
- Trade state machine (all transitions)
- Accept flag reset on item change
- Marketplace commission calculation
- Price validation (min/max bounds)

### Integration Tests
- Full trade flow with item transfer verification
- Concurrent trades on same item (one must fail)
- Marketplace listing, purchase, and redeem
- Marketplace expiry and return to inventory

### E2E Tests
- Two clients complete a trade, both see correct inventories
- Client lists item on marketplace, another client buys it
- Attempt to trade items placed in room (must fail)
