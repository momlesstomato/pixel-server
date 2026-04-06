# 07 - Economy System: Furniture, Catalog, Inventory, Subscriptions & Trading

## Overview

The Economy System is the broadest cross-cutting feature set in the
emulator. It encompasses **five tightly coupled sub-realms**:

1. **Furniture & Item Definitions** — the static catalog of all item
   types, interaction metadata, and physical properties.
2. **Catalog & Store** — the storefront that players browse and purchase
   from, with pages, offers, gift wrapping, vouchers, and limited editions.
3. **Inventory** — the player's owned items (furniture, badges, effects,
   bots, pets), with pagination, unseen-item tracking, and clothing.
4. **Subscription & Offers** — Habbo Club membership, targeted offers,
   club gifts, campaign calendars, and kickback/loyalty.
5. **Economy & Trading** — credits, activity-point currencies, the
   Marketplace (player-to-player auction house), and direct peer trading.

**Scope restriction:** This plan covers the "non-tickable" economy — item
ownership, purchasing, trading, and inventory management. It explicitly
**excludes** all in-room furniture behavior: placement/positioning,
ticking, wired logic, interactions (dice, teleporters, rollers), and
room-entity state. Those belong to a future Room & Furniture Interactions
realm.

The pixel-protocol references **129 C2S packets** and **128 S2C packets**
across the five realms covered here. After filtering out in-room
interaction packets (deferred), this plan targets **~82 C2S** and
**~78 S2C** packets.

---

## Vendor Cross-Reference

### Furniture Definition Schema

| Column | comet-v2 | PlusEMU | Arcturus | pixelsv (proposed) |
|--------|----------|---------|----------|--------------------|
| Table name | `furniture` | `furniture` | `items_base` | `item_definitions` |
| id | INT PK | INT PK | INT PK | BIGINT PK |
| item_name | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) |
| public_name | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) | VARCHAR(100) |
| type | ENUM('s','i','e','h','v','r') | ENUM('s','i','e','h','v','r') | ENUM('s','i') | VARCHAR(1) NOT NULL |
| width | INT(1) | INT(1) | INT(1) | SMALLINT DEFAULT 1 |
| length | INT(1) | INT(1) | INT(1) | SMALLINT DEFAULT 1 |
| stack_height | VARCHAR(255) | VARCHAR(255) | DOUBLE | NUMERIC(6,2) DEFAULT 1.0 |
| can_stack | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT true |
| can_sit | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT false |
| is_walkable | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT false |
| sprite_id | INT | INT | INT | INT NOT NULL |
| allow_recycle | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT true |
| allow_trade | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT true |
| allow_marketplace_sell | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT false |
| allow_gift | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT true |
| allow_inventory_stack | ENUM('0','1') | ENUM('0','1') | BOOL | BOOLEAN DEFAULT true |
| interaction_type | VARCHAR(50) | VARCHAR(50) | VARCHAR(50) | VARCHAR(50) DEFAULT 'default' |
| interaction_modes_count | INT(1) | INT(1) | INT(1) | SMALLINT DEFAULT 1 |
| effect_id | INT | INT | INT | INT DEFAULT 0 |
| revision | INT | INT | INT | INT DEFAULT 1 |

### Catalog Schema

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Pages table | `catalog_pages` | `catalog_pages` | `catalog_pages` | `catalog_pages` |
| Items table | `catalog_items` | `catalog_items` | `catalog_items` | `catalog_items` |
| Page hierarchy | `parent_id` INT | `parent_id` INT | `parent_id` INT | `parent_id` BIGINT |
| Multi-currency | credits+pixels+diamonds+seasonal | credits+duckets | credits+points | two generic cost slots with currency_types FK |
| Club-only pages | `club_only` ENUM | `club_only` ENUM | `min_sub` INT | `min_club_level` INT |
| Featured pages | `catalog_featured_pages` | — | `catalog_featured_pages` | `catalog_featured_pages` |
| Clothing table | `catalog_clothing` | `catalog_clothing` | `catalog_clothing` | `catalog_clothing` |
| Voucher table | `vouchers` | `catalog_vouchers` | `vouchers` | `vouchers` |
| Limited tracking | in `catalog_items` cols | in `catalog_items` cols | in `catalog_items` cols | in `catalog_items` cols |
| Gift wrapping | `catalog_gift_wrapping` | — | `gift_wrappings` | `catalog_gift_wrapping` |

### Item Instance Schema

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Table | `items` | `items` | `items` | `items` |
| Owner | `user_id` INT | `user_id` INT | `user_id` INT | `user_id` BIGINT FK |
| Location | `room_id` INT (0=inv) | `room_id` INT (0=inv) | `room_id` INT (0=inv) | `room_id` BIGINT (0=inv) |
| Definition FK | `base_item` INT | `base_item` INT | `item_id` INT | `definition_id` BIGINT FK |
| Extra data | TEXT | TEXT | TEXT | TEXT DEFAULT '' |
| Limited edition | `limited_number`/`limited_stack` | `limited_number`/`limited_stack` | separate table | `limited_number`/`limited_total` |
| Room position | x/y/z/rot/wall_pos | x/y/z/rot/wall_pos | x/y/z/rot/wall_pos | **Deferred to room realm** |

### Currency System

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Credits storage | `users.credits` | `users.credits` | `users.credits` | `user_currencies` table |
| Activity points | `users.activity_points` | `users.activity_points` | dedicated table | `user_currencies` table |
| Point types | duckets/diamonds/seasonal | duckets | configurable types | **Fully extensible — no hardcoded types** |
| Notification | `UserCurrencyComposer` | `CreditBalanceComposer` | `UserCreditsComposer` | `user.credits` (configurable type) + `user.currency` |

### Trading System

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Trade log | `trade_logs` | `logs_client_trade` | `room_trade_log` + `_items` | `trade_logs` + `trade_log_items` |
| Max items per trade | Uncapped | Uncapped | Uncapped | **Configurable (default 20)** |
| Trade lock | Mod sanction | Per-user | `can_trade` + perk | **Permission-based** |
| Atomic swap | Sequential UPDATE | Sequential UPDATE | Loop with verify | **PostgreSQL transaction** |
| Room requirement | Must be in same room | Must be in same room | Must be in same room | **Defer** (no rooms yet) |

### Marketplace

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Offers table | `marketplace_items` | `catalog_marketplace_offers` | `marketplace_items` | `marketplace_offers` |
| Statistics | `catalog_marketplace_data` | `catalog_marketplace_data` | `marketplace_data` | `marketplace_statistics` |
| Commission | 1% | 1% | 1% | **Configurable (default 1%)** |
| Offer expiry | Never | Never | Never | **Configurable (default 48h)** |
| Price stats | avg/sold per sprite | avg/sold per sprite | avg/sold per sprite | avg/sold/min/max per definition |

### Subscription System

| Aspect | comet-v2 | PlusEMU | Arcturus | pixelsv |
|--------|----------|---------|----------|---------|
| Table | `player_subscriptions` | `subscriptions` | `users_subscriptions` | `user_subscriptions` |
| Duration | start/expire timestamps | timestamp_bought/expire | start + duration (sec) | `started_at` + `duration_days` |
| Club gifts | `presents` counter | implicit | `remainingClubGifts` | `club_gifts_claimed` counter |
| Kickback | No | No | `logs_hc_payday` + streak | **Defer** (V2) |
| Types | `habbo_vip` | — | `HABBO_CLUB` | `habbo_club` |

### Our Improvements Over Vendors

1. **Extensible currency system** — vendors hardcode 3-4 currency types.
   We use a `user_currencies` table with a type column, allowing plugins
   to register custom currencies without schema changes.
2. **PostgreSQL transactions for trades** — vendors use sequential UPDATEs
   with application-level verification. We wrap the entire swap in a
   serializable transaction with row-level locking.
3. **Configurable marketplace expiry** — all vendors keep offers forever.
   We add configurable TTL with automatic expiry job.
4. **Normalized item definitions** — vendors duplicate columns across
   furniture/catalog. We normalize with proper FKs.
5. **Permission-based trade control** — instead of hardcoded rank checks,
   trade permission uses the dotted-notation system (`economy.trade`).
6. **Deferred room positioning** — item instances store ownership only.
   Room coordinates are added when the room realm is implemented.
7. **Gift wrapping with audit trail** — track who sent gifts, when, and
   the wrapping configuration.
8. **Marketplace price history** — vendors only store running average. We
   keep min/max and recent transaction window for richer statistics.
   **Validated:** Nitro client supports `dayOffsets[]`, `averagePrices[]`,
   and `soldAmounts[]` arrays in the MarketplaceItemStatsParser (packet
   725). Price history charts are fully renderable client-side.
9. **Marketplace offer expiry** — all vendors keep offers forever. We add
   configurable TTL with automatic expiry job. **Validated:** Nitro
   client supports `timeLeftMinutes` per offer and `offerTime` in the
   config packet (1823). Countdown display is natively supported.
10. **Fully extensible currency model** — no currency type is hardcoded in
    the schema. All currencies (including the main hard-currency) live in
    `user_currencies` keyed by an operator-assigned integer ID defined in
    the `currency_types` registry table. The `CURRENCY_CREDITS_TYPE_ID`
    config variable tells the server which type ID maps to the `user.credits`
    (3475) packet; all others are sent via `user.currency` (2018). Currency
    types are split into two audit categories:
    - **Normal** (e.g., duckets): simple balance tracking, no audit trail.
    - **Trackable** (e.g., diamonds): balance + full `currency_transactions`
      table recording every add/deduct with reason, reference, and timestamp.
    The `trackable` flag lives on `currency_types`. Plugins can register
    custom types at startup without schema changes.

---

## Design Decisions

### Currency Architecture

All currencies — including the main hard-currency — are stored in the
`user_currencies` table keyed by an integer type ID. No currency column
exists on `users`. The `currency_types` table acts as the registry; entries
are seeded at startup and can be extended by plugins without schema changes.

**Rationale:** Vendors hardcode `users.credits` plus 2-3 activity-point
columns, requiring schema migrations to add new currency types. Our model
is identical in access cost (single indexed row lookup by composite PK) but
imposes no schema coupling. Operators and plugins can register any number
of currencies by inserting rows into `currency_types`.

**Protocol mapping:** The Nitro client has two distinct currency packets:
- `user.credits` (3475) — single integer balance, sent as string
- `user.currency` (2018) — array of `[type, amount]` pairs

The `CURRENCY_CREDITS_TYPE_ID` config variable (default `1`) identifies
which `currency_types.id` is serialised into packet 3475. All other enabled
currencies are sent via packet 2018.

**Currency type registry:** Rows in `currency_types` define all known
currencies. The `type` column in `user_currencies` is an integer FK to
`currency_types.id`. Plugins call `RegisterCurrencyType(id, name)` to add
entries. No Go constants hardcode type IDs — the credits type is resolved
at runtime from configuration.

**Default seeded types** (IDs chosen to match Nitro's `activityPointType`
values where applicable; the credits type has no wire-format ID requirement):
- `1` = Credits (main hard-currency, sent via packet 3475)
- `0` = Duckets (activity points, `activityPointType=0`)
- `5` = Diamonds (trackable, `activityPointType=5`)
- `105` = Seasonal currency (`activityPointType=105`)

**Performance:** `user_currencies` has a composite primary key
`(user_id, currency_type)`. A balance lookup is a single-row PK seek —
equivalent cost to reading a column from `users`. The session-level
balance cache (see Optimizations) means DB reads only occur on login;
all subsequent balance checks are in-memory.

### Item Definitions vs Instances

**Item definitions** (`item_definitions`) are static metadata loaded once
at startup and cached in memory. They describe what a furniture item IS
(dimensions, interaction type, sprite, trade rules).

**Item instances** (`items`) represent specific owned copies. Each instance
references a definition via FK. The instance stores owner, extra data, and
limited-edition numbering. Room coordinates are **deferred** — columns will
be added when the room realm lands.

### Trade Without Rooms

The original Habbo protocol requires both traders to be in the same room.
Since our Room realm is not yet implemented, **direct trading is deferred**.
However, the trade domain model, validation logic, and Marketplace are
implemented now, ready for room-realm integration.

The **Marketplace** does not require rooms — it is a global auction house
accessible from any state. Marketplace is implemented in full.

### Catalog Page Layouts

The Nitro client supports multiple catalog page layouts:
`default_3x3`, `frontpage`, `spaces_new`, `recycler`, `trophies`,
`pets`, `soundmachine`, `guilds`, `guild_furni`, `club_buy`,
`club_gift`, `vip_buy`, `marketplace`, `marketplace_own_items`,
`info_duckets`, `info_loyalty`, `loyalty_vip`, `bots`, `pets2`,
`pets3`, `default_3x3_color_grouping`, `recent_purchases`, `room_bundle`.

We store the layout name as a string and pass it to the client. The server
does not interpret layout semantics — the client renders accordingly.

### Limited Edition Items

Limited editions have a total print run (`limited_total`) and each instance
gets a serial number (`limited_number`). On purchase, the server atomically
increments the sold counter and assigns the next serial. When all copies
sell out, the offer becomes unavailable.

**Race condition:** Two concurrent buyers requesting the last copy. Handled
via PostgreSQL `UPDATE ... RETURNING` with a `WHERE limited_sells <
limited_total` guard. Only one succeeds.

### Unseen Items

When items are added to inventory (purchase, trade, gift), the client
expects an "unseen items" notification. We track this via an in-memory
set on the session, flushed to the client on next inventory request or
immediately via `user.unseen_items` (2103). Categories:
- 1 = Furniture
- 2 = Rentable
- 3 = Pet
- 4 = Badge
- 5 = Bot
- 6 = Effect
- 7 = Game (deferred)

---

## Sub-Realm 1: Furniture & Item Definitions

### Database Schema

#### Table: `item_definitions`

Static item metadata. Loaded at startup, cached in memory. Seeded from
external data files. Admin-editable via API.

```
item_definitions
├── id                      BIGINT PK AUTO
├── item_name               VARCHAR(100) UNIQUE NOT NULL
├── public_name             VARCHAR(100) NOT NULL DEFAULT ''
├── item_type               VARCHAR(1) NOT NULL DEFAULT 's'
├── width                   SMALLINT NOT NULL DEFAULT 1
├── length                  SMALLINT NOT NULL DEFAULT 1
├── stack_height            NUMERIC(6,2) NOT NULL DEFAULT 1.0
├── can_stack               BOOLEAN NOT NULL DEFAULT true
├── can_sit                 BOOLEAN NOT NULL DEFAULT false
├── is_walkable             BOOLEAN NOT NULL DEFAULT false
├── sprite_id               INT NOT NULL
├── allow_recycle           BOOLEAN NOT NULL DEFAULT true
├── allow_trade             BOOLEAN NOT NULL DEFAULT true
├── allow_marketplace_sell  BOOLEAN NOT NULL DEFAULT false
├── allow_gift              BOOLEAN NOT NULL DEFAULT true
├── allow_inventory_stack   BOOLEAN NOT NULL DEFAULT true
├── interaction_type        VARCHAR(50) NOT NULL DEFAULT 'default'
├── interaction_modes_count SMALLINT NOT NULL DEFAULT 1
├── effect_id               INT NOT NULL DEFAULT 0
├── revision                INT NOT NULL DEFAULT 1
├── created_at              TIMESTAMP
├── updated_at              TIMESTAMP
```

**Item types:**
- `s` = floor/standing item
- `i` = wall item
- `e` = effect
- `h` = handler/special
- `v` = vest/clothing
- `r` = rentable
- `b` = badge (virtual, not stored as item instance)

**Interaction types** (string mapping, subset relevant to non-room):
- `default` — no special behavior
- `gate` — passable when open (room behavior deferred)
- `teleport` — paired teleporter (room behavior deferred)
- `trophy` — displays text
- `postit` — sticky note
- `gift` — wrapped present
- `exchange` — redeemable for credits
- `badge_display` — shows a badge
- `mannequin` — stores outfit
- `clothing` — purchasable clothing
- All room-specific types (`dice`, `roller`, `vendingmachine`, `wired_*`,
  `banzai_*`, `freeze_*`, etc.) are stored but behavior is deferred.

### Item Instance Table

#### Table: `items`

Owned item instances. Room-position columns are deferred.

```
items
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id) INDEX
├── room_id             BIGINT NOT NULL DEFAULT 0 INDEX
├── definition_id       BIGINT NOT NULL (FK → item_definitions.id)
├── extra_data          TEXT NOT NULL DEFAULT ''
├── limited_number      INT NOT NULL DEFAULT 0
├── limited_total       INT NOT NULL DEFAULT 0
├── created_at          TIMESTAMP
├── updated_at          TIMESTAMP
```

**`room_id = 0`** means the item is in the player's inventory.
**`room_id > 0`** means it is placed in a room (position columns deferred).

When the Room realm is implemented, additional columns (`x`, `y`, `z`,
`rot`, `wall_pos`) will be added via migration.

### Hexagonal Package Layout

```
pkg/furniture/
├── domain/
│   ├── definition.go       ← ItemDefinition entity
│   ├── item.go             ← Item instance entity
│   ├── repository.go       ← Repository interface
│   └── errors.go           ← Domain errors
├── application/
│   ├── service.go          ← Definition loading, item CRUD
│   └── exchange.go         ← Credit exchange redemption
├── adapter/
│   ├── httpapi/
│   │   ├── definition.go   ← Admin CRUD for item definitions
│   │   └── item.go         ← Admin item management
│   └── command/
│       └── command.go      ← CLI commands
├── infrastructure/
│   ├── model/
│   │   ├── definition.go   ← GORM model for item_definitions
│   │   └── item.go         ← GORM model for items
│   ├── store/
│   │   └── repository.go   ← PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go   ← Schema migrations
│   └── seed/
│       └── seeds.go        ← Sample item definitions
└── stage.go                ← Module bootstrap
```

---

## Sub-Realm 2: Catalog & Store

### Database Schema

#### Table: `catalog_pages`

Hierarchical page tree. Clients show pages in a tree navigator.

```
catalog_pages
├── id                  BIGINT PK AUTO
├── parent_id           BIGINT NOT NULL DEFAULT -1 (FK → catalog_pages.id, -1 = root)
├── caption             VARCHAR(100) NOT NULL
├── icon_image          INT NOT NULL DEFAULT 1
├── visible             BOOLEAN NOT NULL DEFAULT true
├── enabled             BOOLEAN NOT NULL DEFAULT true
├── min_rank            INT NOT NULL DEFAULT 1
├── min_club_level      INT NOT NULL DEFAULT 0
├── order_num           INT NOT NULL DEFAULT 0
├── page_layout         VARCHAR(50) NOT NULL DEFAULT 'default_3x3'
├── page_headline       TEXT NOT NULL DEFAULT ''
├── page_teaser         TEXT NOT NULL DEFAULT ''
├── page_special        TEXT NOT NULL DEFAULT ''
├── page_text_1         TEXT NOT NULL DEFAULT ''
├── page_text_2         TEXT NOT NULL DEFAULT ''
├── page_text_details   TEXT NOT NULL DEFAULT ''
├── link                VARCHAR(100) NOT NULL DEFAULT ''
├── created_at          TIMESTAMP
├── updated_at          TIMESTAMP
```

#### Table: `catalog_items`

Purchasable offers within a catalog page.

```
catalog_items
├── id                  BIGINT PK AUTO
├── page_id             BIGINT NOT NULL (FK → catalog_pages.id) INDEX
├── item_definition_id  BIGINT NOT NULL (FK → item_definitions.id)
├── catalog_name        VARCHAR(100) NOT NULL DEFAULT ''
├── cost_primary        INT NOT NULL DEFAULT 0
├── cost_primary_type   INT NOT NULL DEFAULT 1 (FK → currency_types.id)
├── cost_secondary      INT NOT NULL DEFAULT 0
├── cost_secondary_type INT NOT NULL DEFAULT 0 (FK → currency_types.id)
├── amount              INT NOT NULL DEFAULT 1
├── limited_total       INT NOT NULL DEFAULT 0
├── limited_sells       INT NOT NULL DEFAULT 0
├── offer_active        BOOLEAN NOT NULL DEFAULT true
├── extra_data          VARCHAR(255) NOT NULL DEFAULT ''
├── badge_id            VARCHAR(10) NOT NULL DEFAULT ''
├── club_only           BOOLEAN NOT NULL DEFAULT false
├── order_num           INT NOT NULL DEFAULT 0
├── created_at          TIMESTAMP
├── updated_at          TIMESTAMP
```

**Multi-currency:** Instead of separate `cost_pixels`, `cost_diamonds`,
`cost_seasonal` columns (vendor pattern), we use two generic cost slots
with FK references to `currency_types`. `cost_primary_type` defaults to
the credits type (`CURRENCY_CREDITS_TYPE_ID`); `cost_secondary_type`
defaults to duckets (`0`). Any enabled currency type can be used in either
slot — no type is assumed by the schema.

#### Table: `catalog_featured_pages`

Featured/promoted catalog entries shown on the storefront.

```
catalog_featured_pages
├── id                  BIGINT PK AUTO
├── caption             VARCHAR(100) NOT NULL
├── image               VARCHAR(255) NOT NULL DEFAULT ''
├── page_link           VARCHAR(100) NOT NULL DEFAULT ''
├── page_id             BIGINT NOT NULL DEFAULT 0
├── enabled             BOOLEAN NOT NULL DEFAULT true
├── created_at          TIMESTAMP
```

#### Table: `catalog_clothing`

Clothing sets unlockable by redeeming clothing furniture items.

```
catalog_clothing
├── id                  BIGINT PK AUTO
├── clothing_name       VARCHAR(100) NOT NULL
├── clothing_parts      TEXT NOT NULL DEFAULT ''
```

#### Table: `catalog_gift_wrapping`

Available gift wrap options for the gift purchase flow.

```
catalog_gift_wrapping
├── id                  BIGINT PK AUTO
├── wrapping_type       VARCHAR(10) NOT NULL DEFAULT 'new'
├── sprite_id           INT NOT NULL
├── enabled             BOOLEAN NOT NULL DEFAULT true
```

#### Table: `vouchers`

Redeemable codes for credits, points, badges, or items.

```
vouchers
├── id                      BIGINT PK AUTO
├── code                    VARCHAR(128) UNIQUE NOT NULL
├── reward_type             VARCHAR(20) NOT NULL
├── reward_currency_type    INT NULL (FK → currency_types.id, set when reward_type = 'currency')
├── reward_data             TEXT NOT NULL DEFAULT ''
├── max_uses                INT NOT NULL DEFAULT 1
├── current_uses            INT NOT NULL DEFAULT 0
├── enabled                 BOOLEAN NOT NULL DEFAULT true
├── created_at              TIMESTAMP
├── updated_at              TIMESTAMP
```

**Reward types:** `currency`, `badge`, `furniture`.

When `reward_type = 'currency'`, `reward_currency_type` identifies which
currency type to credit and `reward_data` holds the integer amount as a
string. This replaces the old hardcoded strings `credits`, `duckets`,
`diamonds`, `seasonal` — any registered currency type can be rewarded.

**Redemption guard:** Atomic `UPDATE vouchers SET current_uses =
current_uses + 1 WHERE id = ? AND current_uses < max_uses AND enabled =
true RETURNING id`. If no rows returned, the voucher is exhausted or
invalid. Per-user uniqueness enforced via `voucher_redemptions` table.

#### Table: `voucher_redemptions`

Track which users redeemed which vouchers for one-time-per-user enforcement.

```
voucher_redemptions
├── id                  BIGINT PK AUTO
├── voucher_id          BIGINT NOT NULL (FK → vouchers.id)
├── user_id             BIGINT NOT NULL (FK → users.id)
├── redeemed_at         TIMESTAMP NOT NULL DEFAULT NOW()
├── UNIQUE(voucher_id, user_id)
```

### Catalog Purchase Flow

```
Client sends catalog.purchase (3492)
  │
  ├── Validate page exists, is visible/enabled
  ├── Validate offer exists, is active, on this page
  ├── Validate user has required club level
  ├── Validate user has sufficient credits + points
  ├── If limited: validate remaining stock > 0
  ├── Fire CatalogPurchaseEvent (cancellable plugin event)
  │
  ├── BEGIN TRANSACTION
  │   ├── Deduct cost_primary from user_currencies (currency type = cost_primary_type)
  │   ├── Deduct cost_secondary from user_currencies (currency type = cost_secondary_type)
  │   ├── If limited: UPDATE catalog_items SET limited_sells += 1
  │   │   WHERE limited_sells < limited_total (atomic guard)
  │   ├── Create N item instances in items table
  │   ├── If badge_id present: award badge
  │   └── COMMIT
  │
  ├── Send catalog.purchase_ok (869) to buyer
  ├── Send user.credits (3475) updated balance
  ├── Send user.currency (2018) updated points
  ├── Send user.unseen_items (2103) with new item IDs
  └── Send user.furniture_add (104) per new item
```

### Gift Purchase Flow

```
Client sends catalog.purchase_gift (1411)
  │
  ├── Validate recipient exists
  ├── Validate offer is giftable (item_definitions.allow_gift)
  ├── Validate gift wrapping config
  ├── All purchase validations from above
  ├── Fire GiftPurchaseEvent (cancellable)
  │
  ├── BEGIN TRANSACTION
  │   ├── Deduct cost_primary + cost_secondary from buyer (via user_currencies)
  │   ├── Create item instance owned by RECIPIENT
  │   ├── Set extra_data to gift wrapping metadata
  │   ├── Create gift_log entry (audit trail)
  │   └── COMMIT
  │
  ├── Send catalog.purchase_ok (869) to buyer
  ├── If recipient online:
  │   ├── Send user.unseen_items (2103) to recipient
  │   └── Send user.furniture_add (104) to recipient
  └── Update buyer currency packets
```

### Hexagonal Package Layout

```
pkg/catalog/
├── domain/
│   ├── page.go             ← CatalogPage entity
│   ├── offer.go            ← CatalogOffer entity
│   ├── voucher.go          ← Voucher entity
│   ├── repository.go       ← Repository interface
│   └── errors.go           ← Domain errors
├── application/
│   ├── service.go          ← Page tree, offer lookup
│   ├── purchase.go         ← Purchase use case
│   ├── gift.go             ← Gift purchase use case
│   └── voucher.go          ← Voucher redemption use case
├── adapter/
│   ├── realtime/
│   │   └── runtime.go      ← Packet handler dispatch
│   ├── httpapi/
│   │   ├── page_routes.go  ← Admin page CRUD
│   │   ├── offer_routes.go ← Admin offer CRUD
│   │   ├── voucher_routes.go ← Admin voucher CRUD
│   │   └── openapi.go      ← OpenAPI specs
│   └── command/
│       └── command.go      ← CLI commands
├── infrastructure/
│   ├── model/
│   │   ├── page.go         ← GORM: catalog_pages
│   │   ├── offer.go        ← GORM: catalog_items
│   │   ├── voucher.go      ← GORM: vouchers
│   │   ├── clothing.go     ← GORM: catalog_clothing
│   │   └── wrapping.go     ← GORM: catalog_gift_wrapping
│   ├── store/
│   │   └── repository.go   ← PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go   ← Schema migrations
│   └── seed/
│       └── seeds.go        ← Sample catalog pages/items
└── stage.go                ← Module bootstrap
```

---

## Sub-Realm 3: Inventory

### Database Schema

#### Table: `user_badges` (new)

Badge ownership and slot assignment.

```
user_badges
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id) INDEX
├── badge_code          VARCHAR(50) NOT NULL
├── slot_id             SMALLINT NOT NULL DEFAULT 0
├── created_at          TIMESTAMP
├── UNIQUE(user_id, badge_code)
```

**Slots:** 0 = unequipped, 1-5 = visible equipped positions. Maximum 5
concurrent visible badges.

#### Table: `user_effects`

Avatar effect ownership and activation.

```
user_effects
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id) INDEX
├── effect_id           INT NOT NULL
├── duration            INT NOT NULL DEFAULT 86400
├── quantity            INT NOT NULL DEFAULT 1
├── activated_at        TIMESTAMP NULL
├── is_permanent        BOOLEAN NOT NULL DEFAULT false
├── created_at          TIMESTAMP
├── UNIQUE(user_id, effect_id)
```

**Duration:** seconds remaining. Permanent effects (rank-granted) have
`is_permanent = true` and ignore duration. Activation timestamp is set
when the player first enables the effect; expiry is computed as
`activated_at + duration seconds`.

#### Table: `currency_types` (new)

Registry of all currency types. Seeded at startup; extended by plugins.

```
currency_types
├── id              INT PK (operator-assigned integer; must be unique)
├── name            VARCHAR(50) UNIQUE NOT NULL  (e.g. "credits", "duckets")
├── display_name    VARCHAR(100) NOT NULL DEFAULT ''
├── trackable       BOOLEAN NOT NULL DEFAULT false
├── enabled         BOOLEAN NOT NULL DEFAULT true
├── created_at      TIMESTAMP
```

**`trackable`:** when `true`, every balance change is recorded in
`currency_transactions`. Use for premium currencies that need an audit
trail (e.g. diamonds). Keep `false` for high-frequency normal currencies
(e.g. duckets) to avoid write amplification.

**No Go constants:** type IDs are runtime values read from `currency_types`.
The credits type ID is resolved from `CURRENCY_CREDITS_TYPE_ID` config.

**Default seeded rows:**

| id | name | trackable |
|----|------|-----------|
| 1 | credits | false |
| 0 | duckets | false |
| 5 | diamonds | true |
| 105 | seasonal | false |

#### Table: `user_currencies` (new)

All currency balances for all users. Covers hard-currency and activity
points in a single extensible table.

```
user_currencies
├── user_id             BIGINT NOT NULL (FK → users.id)
├── currency_type       INT NOT NULL (FK → currency_types.id)
├── amount              INT NOT NULL DEFAULT 0
├── PRIMARY KEY(user_id, currency_type)
```

A row is created with `amount = 0` for each enabled currency type when a
new user is registered. Balance is looked up by PK — single indexed seek.

#### Table: `currency_transactions`

Audit log for trackable currencies. Only populated for currency types
marked as `trackable` in the type registry.

```
currency_transactions
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id) INDEX
├── currency_type       INT NOT NULL
├── amount              INT NOT NULL (positive=credit, negative=debit)
├── balance_after       INT NOT NULL
├── reason              VARCHAR(50) NOT NULL
├── reference_type      VARCHAR(50) NOT NULL DEFAULT ''
├── reference_id        BIGINT NOT NULL DEFAULT 0
├── created_at          TIMESTAMP NOT NULL DEFAULT NOW()
```

**Reason values:** `purchase`, `sale`, `admin`, `voucher`, `exchange`,
`marketplace_buy`, `marketplace_sell`, `gift`, `trade`, `subscription`.

**Reference types:** `catalog_offer`, `marketplace_offer`, `voucher`,
`item`, `admin`, `trade_log`.

**Performance:** Partitioned by `created_at` monthly. Indexed on
`(user_id, currency_type, created_at)` for efficient time-range queries.

### Inventory Loading & Pagination

Furniture inventory is loaded on-demand and paginated in fragments.

**Fragment size:** 1000 items per page (matching Arcturus/comet-v2).

```
Client sends user.get_furniture (3150) or user.get_furniture_not_in_room (3500)
  │
  ├── Load all items WHERE user_id = ? AND room_id = 0
  ├── Compute totalFragments = ceil(count / 1000)
  │
  └── For each fragment:
      └── Send user.furniture (994) with:
          ├── totalFragments
          ├── fragmentNumber (0-indexed)
          └── up to 1000 item records
```

### Unseen Items Tracking

When items are added to inventory (purchase, trade, gift, marketplace
buy), the server tracks them as "unseen" in-memory on the connection.

On the next inventory load or on a push notification, the server sends
`user.unseen_items` (2103) listing all new items by category.

The client sends `user.unseen_reset_items` (2343) or
`user.unseen_reset_category` (3493) to acknowledge receipt.

### Hexagonal Package Layout

```
pkg/inventory/
├── domain/
│   ├── badge.go            ← Badge entity
│   ├── effect.go           ← Effect entity
│   ├── currency.go         ← Currency + CurrencyType value objects
│   ├── repository.go       ← Repository interface
│   └── errors.go           ← Domain errors
├── application/
│   ├── service.go          ← Inventory loading, pagination
│   ├── badges.go           ← Badge equip/unequip, award
│   ├── effects.go          ← Effect activate/deactivate
│   └── currency.go         ← Credit/point add/deduct
├── adapter/
│   ├── realtime/
│   │   └── runtime.go      ← Packet handler dispatch
│   ├── httpapi/
│   │   ├── badge_routes.go ← Admin badge management
│   │   ├── currency.go     ← Admin currency management
│   │   └── openapi.go      ← OpenAPI specs
│   └── command/
│       └── command.go      ← CLI commands
├── infrastructure/
│   ├── model/
│   │   ├── badge.go        ← GORM: user_badges
│   │   ├── effect.go       ← GORM: user_effects
│   │   ├── currency_type.go ← GORM: currency_types
│   │   └── currency.go     ← GORM: user_currencies
│   ├── store/
│   │   └── repository.go   ← PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go   ← Schema migrations
│   └── seed/
│       └── seeds.go        ← Default badges/effects
└── stage.go                ← Module bootstrap
```

---

## Sub-Realm 4: Subscription & Offers

### Database Schema

#### Table: `user_subscriptions`

Habbo Club membership tracking.

```
user_subscriptions
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id) INDEX
├── subscription_type   VARCHAR(50) NOT NULL DEFAULT 'habbo_club'
├── started_at          TIMESTAMP NOT NULL
├── duration_days       INT NOT NULL
├── active              BOOLEAN NOT NULL DEFAULT true
├── created_at          TIMESTAMP
├── updated_at          TIMESTAMP
```

**Duration model:** `started_at + duration_days` = expiry time. Duration
is extended by adding to `duration_days`. Active flag is set to false by
the expiry checker when time elapses.

**Expiry checker:** A periodic goroutine (every 60s) queries subscriptions
where `started_at + (duration_days * interval '1 day') < NOW() AND active
= true`, marks them inactive, and triggers the `SubscriptionExpired`
plugin event.

#### Table: `catalog_club_offers`

Available club membership purchase options.

```
catalog_club_offers
├── id                  BIGINT PK AUTO
├── name                VARCHAR(100) NOT NULL
├── days                INT NOT NULL
├── cost_primary        INT NOT NULL DEFAULT 0
├── cost_primary_type   INT NOT NULL DEFAULT 1 (FK → currency_types.id)
├── cost_secondary      INT NOT NULL DEFAULT 0
├── cost_secondary_type INT NOT NULL DEFAULT 0 (FK → currency_types.id)
├── offer_type          VARCHAR(10) NOT NULL DEFAULT 'HC'
├── giftable            BOOLEAN NOT NULL DEFAULT false
├── enabled             BOOLEAN NOT NULL DEFAULT true
├── created_at          TIMESTAMP
```

**Offer types:** `HC` = Habbo Club, `VIP` = VIP Club. Maps to club_level
values in permission groups.

#### Table: `targeted_offers`

Time-limited promotional offers shown to specific players.

```
targeted_offers
├── id                  BIGINT PK AUTO
├── offer_code          VARCHAR(100) NOT NULL
├── title               VARCHAR(255) NOT NULL
├── description         TEXT NOT NULL DEFAULT ''
├── image_url           VARCHAR(255) NOT NULL DEFAULT ''
├── icon_url            VARCHAR(255) NOT NULL DEFAULT ''
├── end_at              TIMESTAMP NOT NULL
├── cost_primary        INT NOT NULL DEFAULT 0
├── cost_primary_type   INT NOT NULL DEFAULT 1 (FK → currency_types.id)
├── cost_secondary      INT NOT NULL DEFAULT 0
├── cost_secondary_type INT NOT NULL DEFAULT 0 (FK → currency_types.id)
├── purchase_limit      INT NOT NULL DEFAULT 1
├── catalog_item_id     BIGINT NULL (FK → catalog_items.id)
├── enabled             BOOLEAN NOT NULL DEFAULT true
├── created_at          TIMESTAMP
```

#### Table: `user_targeted_offer_state`

Tracks per-user targeted offer interaction state.

```
user_targeted_offer_state
├── id                  BIGINT PK AUTO
├── user_id             BIGINT NOT NULL (FK → users.id)
├── offer_id            BIGINT NOT NULL (FK → targeted_offers.id)
├── state               SMALLINT NOT NULL DEFAULT 0
├── purchases           INT NOT NULL DEFAULT 0
├── UNIQUE(user_id, offer_id)
```

### Subscription Status Packet

The `user.subscription` (954) S2C packet is sent on login and on any
subscription state change:

```
Fields:
  subscriptionType  string   "habbo_club"
  daysSinceStart    int32    days since subscription started
  memberPeriods     int32    number of renewal periods
  isVIP             bool     whether VIP tier
  pastClubDays      int32    total historical HC days
  remainingSeconds  int32    seconds until expiry
  daysRemaining     int32    days until expiry (rounded)
```

### Hexagonal Package Layout

```
pkg/subscription/
├── domain/
│   ├── subscription.go     ← Subscription entity
│   ├── club_offer.go       ← ClubOffer entity
│   ├── targeted_offer.go   ← TargetedOffer entity
│   ├── repository.go       ← Repository interface
│   └── errors.go           ← Domain errors
├── application/
│   ├── service.go          ← Subscription status, extension
│   ├── offers.go           ← Club offer lookup, purchase
│   └── targeted.go         ← Targeted offer management
├── adapter/
│   ├── realtime/
│   │   └── runtime.go      ← Packet handler dispatch
│   ├── httpapi/
│   │   ├── sub_routes.go   ← Admin subscription management
│   │   ├── offer_routes.go ← Admin offer CRUD
│   │   └── openapi.go      ← OpenAPI specs
│   └── command/
│       └── command.go      ← CLI commands
├── infrastructure/
│   ├── model/
│   │   ├── subscription.go ← GORM: user_subscriptions
│   │   ├── club_offer.go   ← GORM: catalog_club_offers
│   │   └── targeted.go     ← GORM: targeted_offers + state
│   ├── store/
│   │   └── repository.go   ← PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go   ← Schema migrations
│   └── seed/
│       └── seeds.go        ← Default club offers
└── stage.go                ← Module bootstrap
```

---

## Sub-Realm 5: Economy & Trading

### Database Schema

#### Table: `marketplace_offers`

Player-listed items for sale on the global Marketplace.

```
marketplace_offers
├── id                  BIGINT PK AUTO
├── item_id             BIGINT NOT NULL (FK → items.id)
├── seller_id           BIGINT NOT NULL (FK → users.id) INDEX
├── asking_price        INT NOT NULL
├── state               SMALLINT NOT NULL DEFAULT 1
├── listed_at           TIMESTAMP NOT NULL DEFAULT NOW()
├── sold_at             TIMESTAMP NULL
├── buyer_id            BIGINT NULL (FK → users.id)
├── definition_id       BIGINT NOT NULL (FK → item_definitions.id)
├── sprite_id           INT NOT NULL
├── limited_number      INT NOT NULL DEFAULT 0
├── limited_total       INT NOT NULL DEFAULT 0
├── extra_data          TEXT NOT NULL DEFAULT ''
```

**States:** 1 = OPEN, 2 = SOLD, 3 = EXPIRED, 4 = CANCELLED.

#### Table: `marketplace_statistics`

Running price statistics per item definition.

```
marketplace_statistics
├── definition_id       BIGINT PK (FK → item_definitions.id)
├── sold_count          INT NOT NULL DEFAULT 0
├── avg_price           INT NOT NULL DEFAULT 0
├── updated_at          TIMESTAMP
```

#### Table: `trade_logs`

Audit log of completed trades.

```
trade_logs
├── id                  BIGINT PK AUTO
├── user_one_id         BIGINT NOT NULL (FK → users.id)
├── user_two_id         BIGINT NOT NULL (FK → users.id)
├── traded_at           TIMESTAMP NOT NULL DEFAULT NOW()
```

#### Table: `trade_log_items`

Items exchanged in each trade, linked to the trade log.

```
trade_log_items
├── id                  BIGINT PK AUTO
├── trade_id            BIGINT NOT NULL (FK → trade_logs.id)
├── item_id             BIGINT NOT NULL
├── user_id             BIGINT NOT NULL
├── definition_id       BIGINT NOT NULL
```

### Marketplace Purchase Flow

```
Client sends marketplace.buy_offer (1603)
  │
  ├── Validate offer exists, state = OPEN
  ├── Validate buyer != seller
  ├── Validate buyer has sufficient credits
  ├── Fire MarketplacePurchaseEvent (cancellable)
  │
  ├── BEGIN TRANSACTION
  │   ├── UPDATE offers SET state=SOLD, sold_at=NOW(), buyer_id=?
  │   │   WHERE id=? AND state=OPEN (atomic guard)
  │   ├── Deduct asking_price from buyer (currency type = CURRENCY_CREDITS_TYPE_ID)
  │   ├── Credit seller: asking_price - commission (same currency type)
  │   ├── Transfer item: UPDATE items SET user_id = buyer_id
  │   ├── UPDATE marketplace_statistics (avg_price, sold_count)
  │   └── COMMIT
  │
  ├── Send marketplace.buy_result (2032) to buyer
  ├── Send user.credits (3475) to buyer
  ├── Send user.unseen_items (2103) to buyer
  ├── If seller online: send user.credits (3475) to seller
  └── Update marketplace search cache
```

### Direct Trade Flow (DEFERRED)

Direct trading requires both users to be in the same room. Since rooms are
not yet implemented, the full trade packet flow is **deferred**. The domain
model and validation logic are built now for future activation.

**Trade state machine (for future reference):**
```
IDLE → trade.open (1481)
  → OFFERING: both users add/remove items
  → trade.accept (3863) / trade.unaccept (1444)
  → BOTH_ACCEPTED: trade.confirm (2760)
  → COMPLETED: items swapped atomically
  → trade.completed (1001) or trade.closed (1373)
```

### Credit Exchange (Exchange Furni)

Items with `interaction_type = 'exchange'` can be redeemed for credits.
The `extra_data` field contains the credit value. On redemption:

```
Client sends furniture.item_exchange_redeem (3115)
  │
  ├── Validate item exists in user inventory
  ├── Validate interaction_type == 'exchange'
  ├── Parse credit value from item name (e.g., 'CF_50' → 50 credits)
  │
  ├── BEGIN TRANSACTION
  │   ├── DELETE item from items table
  │   ├── ADD amount to user_currencies (currency type = CURRENCY_CREDITS_TYPE_ID)
  │   └── COMMIT
  │
  ├── Send user.credits (3475)
  ├── Send user.furniture_remove (159)
  └── Fire CreditExchangeEvent
```

### Hexagonal Package Layout

```
pkg/economy/
├── domain/
│   ├── marketplace.go      ← MarketplaceOffer entity
│   ├── trade.go            ← Trade session entity (for future)
│   ├── repository.go       ← Repository interface
│   └── errors.go           ← Domain errors
├── application/
│   ├── service.go          ← Marketplace listing, search
│   ├── purchase.go         ← Marketplace purchase use case
│   └── exchange.go         ← Credit exchange use case
├── adapter/
│   ├── realtime/
│   │   └── runtime.go      ← Packet handler dispatch
│   ├── httpapi/
│   │   ├── marketplace.go  ← Admin marketplace management
│   │   └── openapi.go      ← OpenAPI specs
│   └── command/
│       └── command.go      ← CLI commands
├── infrastructure/
│   ├── model/
│   │   ├── marketplace.go  ← GORM: marketplace_offers
│   │   ├── statistics.go   ← GORM: marketplace_statistics
│   │   └── trade_log.go    ← GORM: trade_logs + items
│   ├── store/
│   │   └── repository.go   ← PostgreSQL repository
│   ├── migration/
│   │   └── migrations.go   ← Schema migrations
│   └── seed/
│       └── seeds.go        ← Sample marketplace data
└── stage.go                ← Module bootstrap
```

---

## Packet Registry

### Sub-Realm 1: Furniture & Items (Non-Room Subset)

Only the packets relevant to inventory-level operations. All room-placement
and interaction packets are deferred.

#### Client-to-Server

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3115 | `furniture.item_exchange_redeem` | itemId (int32) | **M2** |
| 3898 | `furniture.furniture_aliases` | (empty) | **M1** |
| 3558 | `furniture.present_open_present` | itemId (int32) | **M3** |

#### Server-to-Client

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 1723 | `furniture.aliases` | count, [name, alias]* | **M1** |
| 56 | `furniture.gift_opened` | itemType, spriteId, productCode, … | **M3** |
| 377 | `furniture.limited_sold_out` | (empty) | **M2** |

### Sub-Realm 2: Catalog & Store

#### Client-to-Server (10 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 1195 | `catalog.get_index` | mode (string) | **M1** |
| 412 | `catalog.get_page` | pageId, offerId, mode | **M1** |
| 3492 | `catalog.purchase` | pageId, offerId, extraData, amount | **M1** |
| 1411 | `catalog.purchase_gift` | pageId, offerId, extraData, receiverName, wrapping… | **M2** |
| 339 | `catalog.redeem_voucher` | code (string) | **M2** |
| 418 | `catalog.get_gift_wrapping_config` | (empty) | **M2** |
| 1347 | `catalog.check_giftable` | offerId (int32) | **M2** |
| 223 | `catalog.bundle_discount_ruleset` | (empty) | **DEFER** |
| 2150 | `catalog.mark_catalog_new_additions_page_opened` | (empty) | **DEFER** |
| 2436 | `catalog.get_gift` | (empty) | **DEFER** |

#### Server-to-Client (11 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 1032 | `catalog.index` | root node tree | **M1** |
| 804 | `catalog.page` | pageId, mode, items, layout data | **M1** |
| 869 | `catalog.purchase_ok` | offer data | **M1** |
| 1404 | `catalog.purchase_error` | errorCode (int32) | **M1** |
| 3770 | `catalog.purchase_not_allowed` | errorCode | **M1** |
| 3336 | `catalog.voucher_ok` | productName, description | **M2** |
| 714 | `catalog.voucher_error` | errorCode (int32) | **M2** |
| 1517 | `catalog.gift_receiver_not_found` | (empty) | **M2** |
| 2234 | `catalog.gift_wrapping_config` | configData | **M2** |
| 1866 | `catalog.published` | furniAddons (bool) | **M3** |
| 2347 | `catalog.bundle_discount_ruleset` | data | **DEFER** |

### Sub-Realm 3: Inventory

#### Client-to-Server (13 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3150 | `user.get_furniture` | (empty) | **M1** |
| 3500 | `user.get_furniture_not_in_room` | (empty) | **M1** |
| 2769 | `user.get_badges` | (empty) | **M1** |
| 2091 | `user.get_current_badges` | userId (int32) | **M1** |
| 644 | `user.update_badges` | badgeSlots (array) | **M1** |
| 2343 | `user.unseen_reset_items` | category, itemIds | **M1** |
| 3493 | `user.unseen_reset_category` | category (int32) | **M1** |
| 2959 | `user.effect_activate` | effectId (int32) | **M2** |
| 3374 | `user.clothing_redeem` | itemId (int32) | **M3** |
| 367 | `user.get_group_memberships` | (empty) | **DEFER** |
| 21 | `user.get_group_badges` | (empty) | **DEFER** |
| 3095 | `user.get_pets` | (empty) | **DEFER** |
| 3848 | `user.get_bots` | (empty) | **DEFER** |

#### Server-to-Client (20 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 994 | `user.furniture` | totalFragments, fragmentNum, items… | **M1** |
| 104 | `user.furniture_add` | itemData | **M1** |
| 159 | `user.furniture_remove` | itemId (int32) | **M1** |
| 3151 | `user.furniture_refresh` | (empty) | **M1** |
| 717 | `user.badges` | badgeCodes, equippedSlots | **M1** |
| 2493 | `user.badge_received` | badgeId, badgeCode | **M1** |
| 1087 | `user.current_badges` | userId, slotData | **M1** |
| 2103 | `user.unseen_items` | categories, itemIds | **M1** |
| 340 | `user.effects` | effectList | **M2** |
| 2867 | `user.effect_added` | effectData | **M2** |
| 2228 | `user.effect_removed` | effectId | **M2** |
| 1959 | `user.effect_activated` | effectData | **M2** |
| 3473 | `user.effect_selected` | effectId | **M2** |
| 1450 | `user.clothing` | clothingParts | **M3** |
| 3475 | `user.credits` | credits (string) | **M1** |
| 2018 | `user.currency` | currencies array | **M1** |
| 2275 | `user.activity_point_notification` | amount, change, type | **M1** |
| 3086 | `user.bots` | botList | **DEFER** |
| 3522 | `user.pets` | petList | **DEFER** |
| 2101 | `user.pet_added` | petData | **DEFER** |

### Sub-Realm 4: Subscription & Offers

#### Client-to-Server (subset, 12 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3166 | `user.get_subscription` | productName (string) | **M1** |
| 3285 | `catalog.get_club_offers` | (empty) | **M1** |
| 2462 | `catalog.get_hc_extend_offer` | (empty) | **M2** |
| 603 | `catalog.get_basic_extend_offer` | (empty) | **M2** |
| 2276 | `catalog.select_club_gift` | offerId (int32) | **M2** |
| 487 | `catalog.get_club_gift_info` | (empty) | **M2** |
| 869 | `user.get_kickback_info` | (empty) | **DEFER** |
| 2487 | `offer.get_targeted` | (empty) | **M3** |
| 596 | `offer.get_next_targeted` | (empty) | **M3** |
| 1826 | `offer.purchase_targeted` | offerId, count | **M3** |
| 2041 | `offer.set_targeted_state` | offerId, state | **M3** |
| 2257 | `calendar.open_door` | day (int32) | **DEFER** |

#### Server-to-Client (subset, 12 packets)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 954 | `user.subscription` | subscriptionType, days, isVIP… | **M1** |
| 2405 | `catalog.club_offers` | offers array | **M1** |
| 3964 | `catalog.club_extend_offer` | offerData | **M2** |
| 619 | `catalog.club_gift_info` | giftData | **M2** |
| 659 | `catalog.club_gift_selected` | productData | **M2** |
| 3277 | `user.kickback_info` | kickbackData | **DEFER** |
| 119 | `offer.targeted` | offerData | **M3** |
| 1237 | `offer.targeted_not_found` | (empty) | **M3** |
| 3914 | `catalog.not_enough_balance` | credits, points, pointsType | **M1** |
| 2188 | `catalog.club_gift_notification` | count | **M2** |
| 1452 | `catalog.builders_club_subscription` | data | **DEFER** |
| 195 | `catalog.direct_sms_club_buy` | data | **DEFER** |

### Sub-Realm 5: Economy & Trading

#### Client-to-Server (subset, non-room)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 273 | `user.get_currency` | (empty) | **M1** |
| 2597 | `marketplace.get_config` | (empty) | **M2** |
| 2407 | `marketplace.search_offers` | minPrice, maxPrice, searchQuery, sortMode | **M2** |
| 2105 | `marketplace.get_own_items` | (empty) | **M2** |
| 1603 | `marketplace.buy_offer` | offerId (int32) | **M2** |
| 434 | `marketplace.cancel_sale` | offerId (int32) | **M2** |
| 3447 | `marketplace.sell_item` | itemId, askingPrice | **M2** |
| 3288 | `marketplace.get_item_stats` | definitionId | **M2** |
| 848 | `marketplace.get_can_sell` | (empty) | **M2** |
| 2650 | `marketplace.redeem_credits` | (empty) | **M2** |
| 1481 | `trade.open` | userId (int32) | **DEFER** |
| 3107 | `trade.add_item` | itemId (int32) | **DEFER** |
| 1263 | `trade.add_items` | count, itemIds | **DEFER** |
| 3845 | `trade.remove_item` | itemId (int32) | **DEFER** |
| 3863 | `trade.accept` | (empty) | **DEFER** |
| 1444 | `trade.unaccept` | (empty) | **DEFER** |
| 2760 | `trade.confirm` | (empty) | **DEFER** |
| 2341 | `trade.cancel` | (empty) | **DEFER** |
| 2551 | `trade.close` | (empty) | **DEFER** |

#### Server-to-Client (subset, non-room)

| ID | Name | Fields | Priority |
|----|------|--------|----------|
| 3475 | `user.credits` | creditBalance (string) | **M1** |
| 2018 | `user.currency` | count, [type, amount]* | **M1** |
| 1823 | `marketplace.config` | isEnabled, commission, offerMinPrice, … | **M2** |
| 680 | `marketplace.items_searched` | offers, totalOffers | **M2** |
| 3884 | `marketplace.own_items` | offers, credits | **M2** |
| 2032 | `marketplace.buy_result` | result, offerId, … | **M2** |
| 3264 | `marketplace.cancel_sale_result` | offerId, success | **M2** |
| 1359 | `marketplace.item_posted` | result | **M2** |
| 725 | `marketplace.item_stats` | avgPrice, offerCount, history | **M2** |
| 54 | `marketplace.can_sell` | result, maxCredits, description | **M2** |
| 2505 | `trade.opened` | user1Id, user2Id, items | **DEFER** |
| 2024 | `trade.list_item` | items update | **DEFER** |
| 2568 | `trade.accepted` | userId, accepted | **DEFER** |
| 2720 | `trade.confirmation` | (empty) | **DEFER** |
| 1001 | `trade.completed` | (empty) | **DEFER** |
| 1373 | `trade.closed` | userId, reason | **DEFER** |
| 217 | `trade.open_failed` | reason, username | **DEFER** |
| 3058 | `trade.you_not_allowed` | (empty) | **DEFER** |
| 1254 | `trade.other_not_allowed` | (empty) | **DEFER** |
| 2873 | `trade.no_such_item` | (empty) | **DEFER** |
| 3128 | `trade.not_open` | (empty) | **DEFER** |

---

## API & CLI Endpoints

### REST API Endpoints

All behind API key middleware.

#### Furniture & Definitions

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/definitions` | List item definitions (paginated) | **M1** |
| `GET` | `/api/definitions/{id}` | Get single item definition | **M1** |
| `POST` | `/api/definitions` | Create item definition | **M1** |
| `PATCH` | `/api/definitions/{id}` | Update item definition | **M1** |
| `DELETE` | `/api/definitions/{id}` | Delete item definition | **M1** |
| `GET` | `/api/users/{id}/items` | List user's inventory items | **M1** |
| `POST` | `/api/users/{id}/items` | Admin-grant item to user | **M1** |
| `DELETE` | `/api/items/{id}` | Admin-remove item | **M1** |

#### Catalog

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/catalog/pages` | List catalog pages (tree) | **M1** |
| `GET` | `/api/catalog/pages/{id}` | Get page with offers | **M1** |
| `POST` | `/api/catalog/pages` | Create catalog page | **M1** |
| `PATCH` | `/api/catalog/pages/{id}` | Update catalog page | **M1** |
| `DELETE` | `/api/catalog/pages/{id}` | Delete catalog page | **M1** |
| `GET` | `/api/catalog/offers` | List catalog offers | **M1** |
| `POST` | `/api/catalog/offers` | Create catalog offer | **M1** |
| `PATCH` | `/api/catalog/offers/{id}` | Update catalog offer | **M1** |
| `DELETE` | `/api/catalog/offers/{id}` | Delete catalog offer | **M1** |
| `POST` | `/api/vouchers` | Create voucher | **M2** |
| `GET` | `/api/vouchers` | List vouchers | **M2** |
| `DELETE` | `/api/vouchers/{id}` | Delete voucher | **M2** |

#### Inventory

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/users/{id}/badges` | List user badges | **M1** |
| `POST` | `/api/users/{id}/badges` | Admin-grant badge | **M1** |
| `DELETE` | `/api/users/{id}/badges/{code}` | Admin-revoke badge | **M1** |
| `GET` | `/api/users/{id}/effects` | List user effects | **M2** |
| `POST` | `/api/users/{id}/effects` | Admin-grant effect | **M2** |
| `GET` | `/api/users/{id}/currency` | Get user balances | **M1** |
| `PATCH` | `/api/users/{id}/currency` | Admin-set currency | **M1** |

#### Subscription

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/users/{id}/subscription` | Get subscription state | **M3** |
| `POST` | `/api/users/{id}/subscription` | Admin-grant subscription | **M3** |
| `DELETE` | `/api/users/{id}/subscription` | Admin-revoke subscription | **M3** |
| `GET` | `/api/catalog/club-offers` | List club offers | **M3** |
| `POST` | `/api/catalog/club-offers` | Create club offer | **M3** |

#### Marketplace

| Method | Path | Description | Milestone |
|--------|------|-------------|-----------|
| `GET` | `/api/marketplace/offers` | List active offers | **M2** |
| `GET` | `/api/marketplace/statistics` | Get price statistics | **M2** |
| `DELETE` | `/api/marketplace/offers/{id}` | Admin-cancel offer | **M2** |

### CLI Commands

Mirror API 1:1:

| Command | Description | Milestone |
|---------|-------------|-----------|
| `pixelsv definition list` | List item definitions | **M1** |
| `pixelsv definition get <id>` | Get definition | **M1** |
| `pixelsv definition create --name x --type s` | Create definition | **M1** |
| `pixelsv catalog page list` | List catalog pages | **M1** |
| `pixelsv catalog page create --caption x` | Create page | **M1** |
| `pixelsv catalog offer list --page <id>` | List offers | **M1** |
| `pixelsv catalog offer create --page <id>` | Create offer | **M1** |
| `pixelsv item list --user <id>` | List user items | **M1** |
| `pixelsv item grant <userId> <definitionId>` | Grant item | **M1** |
| `pixelsv item remove <itemId>` | Remove item | **M1** |
| `pixelsv badge grant <userId> <code>` | Grant badge | **M1** |
| `pixelsv badge revoke <userId> <code>` | Revoke badge | **M1** |
| `pixelsv currency get <userId>` | Get balances | **M1** |
| `pixelsv currency set <userId> --type <id> --amount N` | Set currency balance | **M1** |
| `pixelsv voucher create --code X --type credits` | Create voucher | **M2** |
| `pixelsv subscription grant <userId> --days N` | Grant HC | **M3** |
| `pixelsv subscription revoke <userId>` | Revoke HC | **M3** |
| `pixelsv marketplace list` | List offers | **M2** |
| `pixelsv marketplace cancel <offerId>` | Cancel offer | **M2** |

---

## Plugin Events

| Event | Cancellable | Fields | Milestone |
|-------|-------------|--------|-----------|
| `CatalogPurchase` | **Yes** | ConnID, UserID, OfferId, PageId, Amount | **M1** |
| `CatalogGiftPurchase` | **Yes** | ConnID, BuyerID, ReceiverID, OfferId | **M2** |
| `VoucherRedeemed` | **Yes** | ConnID, UserID, VoucherCode, RewardType | **M2** |
| `CurrencyChanged` | No | UserID, CurrencyTypeID, OldAmount, NewAmount | **M1** |
| `BadgeAwarded` | **Yes** | UserID, BadgeCode | **M1** |
| `BadgeRevoked` | No | UserID, BadgeCode | **M1** |
| `BadgeSlotsChanged` | No | UserID, OldSlots, NewSlots | **M1** |
| `EffectActivated` | **Yes** | ConnID, UserID, EffectID | **M2** |
| `ItemExchangeRedeemed` | **Yes** | ConnID, UserID, ItemID, Credits | **M2** |
| `GiftOpened` | No | ConnID, UserID, ItemID, SenderID | **M3** |
| `MarketplaceItemListed` | **Yes** | UserID, ItemID, AskingPrice | **M2** |
| `MarketplaceItemPurchased` | **Yes** | BuyerID, SellerID, OfferID, Price | **M2** |
| `MarketplaceItemCancelled` | No | UserID, OfferID | **M2** |
| `SubscriptionCreated` | No | UserID, Type, DurationDays | **M3** |
| `SubscriptionExpired` | No | UserID, Type | **M3** |
| `SubscriptionExtended` | No | UserID, Type, AddedDays | **M3** |
| `ClubGiftClaimed` | **Yes** | ConnID, UserID, GiftID | **M3** |
| `TargetedOfferPurchased` | **Yes** | ConnID, UserID, OfferID | **M3** |
| `TradeCompleted` | No | User1ID, User2ID, TradeLogID | **DEFER** |
| `TradeOpened` | **Yes** | User1ID, User2ID | **DEFER** |
| `TradeCancelled` | No | UserID, OtherUserID, Reason | **DEFER** |

---

## Configuration

| Variable | Default | Description | Milestone |
|----------|---------|-------------|-----------|
| `CATALOG_PURCHASE_COOLDOWN_MS` | 500 | Min interval between purchases | **M1** |
| `CATALOG_MAX_PURCHASE_AMOUNT` | 100 | Max items per single purchase | **M1** |
| `CATALOG_GIFT_ENABLED` | true | Enable gift purchasing | **M2** |
| `MARKETPLACE_ENABLED` | true | Enable the Marketplace | **M2** |
| `MARKETPLACE_COMMISSION_PCT` | 1 | Commission percentage (0-100) | **M2** |
| `MARKETPLACE_MIN_PRICE` | 1 | Minimum listing price | **M2** |
| `MARKETPLACE_MAX_PRICE` | 999999 | Maximum listing price | **M2** |
| `MARKETPLACE_OFFER_EXPIRY_HOURS` | 48 | Hours before auto-expiry | **M2** |
| `MARKETPLACE_MAX_ACTIVE_OFFERS` | 30 | Max concurrent offers per user | **M2** |
| `INVENTORY_FRAGMENT_SIZE` | 1000 | Items per inventory fragment | **M1** |
| `INVENTORY_MAX_ITEMS` | 5000 | Max inventory items per user | **M1** |
| `BADGE_MAX_SLOTS` | 5 | Max visible badge slots | **M1** |
| `SUBSCRIPTION_CHECK_INTERVAL_SEC` | 60 | Expiry check frequency | **M3** |
| `SUBSCRIPTION_DEFAULT_TYPE` | habbo_club | Default subscription type | **M3** |
| `TRADE_ENABLED` | true | Enable direct trading | **DEFER** |
| `TRADE_MAX_ITEMS_PER_USER` | 20 | Max items per trade side | **DEFER** |
| `TRADE_REQUIRE_PERMISSION` | true | Require economy.trade permission | **DEFER** |
| `CURRENCY_CREDITS_TYPE_ID` | 1 | Currency type ID sent via `user.credits` (3475) | **M1** |
| `CURRENCY_INITIAL_AMOUNTS` | `1:0,0:0` | Comma-separated `type_id:amount` pairs granted to new users | **M1** |

---

## Database Migrations

### Migration Order

Registered in `core/postgres/migrations/registry.go`, continuing from
existing migration 04:

| ID | Migration | Tables | Milestone |
|----|-----------|--------|-----------|
| 05 | `05_item_definitions.go` | `item_definitions` | **M1** |
| 06 | `06_items.go` | `items` | **M1** |
| 07 | `07_user_currencies.go` | `currency_types`, `user_currencies` | **M1** |
| 08 | `08_user_badges.go` | `user_badges` | **M1** |
| 09 | `09_catalog_pages.go` | `catalog_pages` | **M1** |
| 10 | `10_catalog_items.go` | `catalog_items`, `catalog_featured_pages` | **M1** |
| 11 | `11_catalog_extras.go` | `catalog_clothing`, `catalog_gift_wrapping` | **M2** |
| 12 | `12_vouchers.go` | `vouchers`, `voucher_redemptions` | **M2** |
| 13 | `13_user_effects.go` | `user_effects` | **M2** |
| 14 | `14_marketplace.go` | `marketplace_offers`, `marketplace_statistics` | **M2** |
| 15 | `15_user_subscriptions.go` | `user_subscriptions`, `catalog_club_offers` | **M3** |
| 16 | `16_targeted_offers.go` | `targeted_offers`, `user_targeted_offer_state` | **M3** |
| 17 | `17_trade_logs.go` | `trade_logs`, `trade_log_items` | **DEFER** |

### Seed Data

| ID | Seed | Data | Milestone |
|----|------|------|-----------|
| 05 | `05_item_definitions.go` | 20 sample furniture definitions (bed, chair, table, trophy, postit, exchange_50, exchange_100, etc.) | **M1** |
| 06 | `06_catalog_pages.go` | Root page + 3 sample pages (Furni, Rare Items, VIP) | **M1** |
| 07 | `07_catalog_items.go` | 10 sample catalog offers (1 per sample furniture, varying prices) | **M1** |
| 08 | `08_test_items.go` | 5 items granted to test user #1 | **M1** |
| 09 | `09_test_badges.go` | 3 sample badges for test user #1 | **M1** |
| 10 | `10_test_currency.go` | 5000 units of credits type + 1000 duckets for test users (via user_currencies) | **M1** |
| 11 | `11_vouchers.go` | 2 sample vouchers (1 credits, 1 badge) | **M2** |
| 12 | `12_club_offers.go` | 3 club membership offers (1mo, 3mo, 6mo) | **M3** |
| 13 | `13_marketplace_test.go` | 3 sample marketplace offers | **M2** |

---

## Edge Cases & Caveats

### Purchase Race Conditions

**Double purchase:** Client sends two rapid `catalog.purchase` packets.
The purchase cooldown (`CATALOG_PURCHASE_COOLDOWN_MS`) drops the second
if within window. For limited items, the PostgreSQL `UPDATE WHERE
limited_sells < limited_total` guard prevents overselling. The
application also rate-limits per connection.

**Concurrent limited purchase:** Two buyers buy the last copy
simultaneously. The `UPDATE ... WHERE limited_sells < limited_total
RETURNING *` ensures only one succeeds. The loser receives
`catalog.purchase_error` (1404).

### Currency Underflow

Currency balances are validated **before** the transaction. The transaction
uses `UPDATE user_currencies SET amount = amount - ? WHERE user_id = ?
AND currency_type = ? AND amount >= ?` to prevent negative balances
atomically. If no rows are updated the purchase fails with
`catalog.not_enough_balance` (3914). Both cost slots are validated and
deducted in a single transaction.

### Inventory Overflow

When `INVENTORY_MAX_ITEMS` is reached, purchases fail with error code.
Items from trades, gifts, and Marketplace are also rejected. The client
shows a notification.

### Gift to Offline User

Gifts to offline users are valid. The item is created in the recipient's
inventory. On their next login, the `user.unseen_items` packet includes
the gift. No online notification is required.

### Marketplace Offer Expiry

A background goroutine runs every 5 minutes:
```sql
UPDATE marketplace_offers
SET state = 3  -- EXPIRED
WHERE state = 1 AND listed_at + ? * interval '1 hour' < NOW()
RETURNING item_id, seller_id
```
Expired items are returned to the seller's inventory. If the seller is
online, `user.furniture_add` (104) is sent. Otherwise, items appear on
their next inventory load.

### Marketplace Sold Item Retrieval

When a seller's item sells, the credits minus commission are held in
virtual escrow. The seller must call `marketplace.redeem_credits` (2650)
to collect accumulated sales revenue. This matches all vendor
implementations.

### Voucher Code Brute Force

Rate-limit `catalog.redeem_voucher` to 1 attempt per 3 seconds per
connection. Log all failed attempts with the connection's ray ID.
After 10 consecutive failures, temporarily block voucher redemption
for that session.

### Badge Slot Validation

When the client sends `user.update_badges` (644), validate:
- Each badge code is actually owned by the user
- No duplicate badge codes in different slots
- Slot IDs are 1-5 only
- Maximum 5 badges equipped simultaneously

### Effect Duration Tracking

Effects with finite duration begin counting down only when activated.
An unactivated effect with `duration = 86400` stays in inventory
indefinitely until the user enables it. Once `activated_at` is set,
expiry = `activated_at + duration`. A background job checks expired
effects every 60 seconds and sends `user.effect_removed` (2228).

### Subscription Overlap

Purchasing HC when already subscribed extends the existing subscription
by adding to `duration_days`. The `started_at` timestamp is NOT reset.
This means remaining time compounds rather than replaces.

### Credit Exchange Naming Convention

Exchange furniture items follow the pattern `CF_<value>` (e.g., `CF_1`,
`CF_5`, `CF_10`, `CF_20`, `CF_50`, `CF_100`). The credit value is
parsed from the item definition's `item_name` field, not from `extra_data`.

### Post-Auth Currency Burst

On login, the post-auth burst must include:
1. `user.credits` (3475) — balance of currency type `CURRENCY_CREDITS_TYPE_ID`
2. `user.currency` (2018) — all other enabled currency type balances
3. `user.subscription` (954) — HC status

These are sent after the existing user.info/permissions/perks burst.

---

## What Gets Deferred

| Feature | Reason | Depends On |
|---------|--------|------------|
| Direct trading (all trade.* packets) | Requires room realm (same-room check) | Room realm |
| Furniture placement/movement | Room coordinates | Room realm |
| Furniture interactions (dice, teleport, roller, etc.) | Room tick system | Room realm |
| Wired system (all wired packets) | Room automation | Room realm |
| Pet inventory display | Pet entities | Pet realm |
| Bot inventory display | Bot entities | Bot realm |
| Builders Club | Legacy feature | None |
| Campaign calendar | Low priority | None |
| HC kickback/payday | Complex loyalty system | V2 |
| Bundle discount ruleset | Low priority | None |
| Community goals | Social feature | None |
| Crafting/recycling | Complex economy feature | V2 |
| Direct SMS club buy | External integration | None |
| Room-based furniture packets (52 C2S, 47 S2C) | Room realm | Room realm |

---

## Optimizations

### Item Definition Cache

All item definitions are loaded at startup into an in-memory map
(`map[int64]*ItemDefinition`). The cache is invalidated when an admin
modifies definitions via API. A Redis pub/sub channel broadcasts cache
invalidation across instances.

### Catalog Page Tree Cache

The catalog page tree is loaded at startup and cached. Full-tree
serialization is pre-computed and stored in Redis with a version key.
When an admin modifies pages, the version is bumped and all instances
rebuild their cache on next request.

### Inventory Fragment Cache

Furniture inventory fragments are computed on first request and cached
in-memory on the session. Cache is invalidated when items are added or
removed (purchase, trade, gift, marketplace).

### Currency Balance Fast Path

Credit balance is kept in-memory on the session after first load.
Write-through to PostgreSQL on every change. Read from session memory
for packets that include credit balance, avoiding a DB round-trip.

### Marketplace Search Index

Marketplace offers are indexed by `definition_id` and `asking_price`
in PostgreSQL. For high-traffic servers, a Redis sorted set per
definition ID stores active offer prices for O(log N) range queries.

---

## Implementation Roadmap

### Milestone 1: Foundations — Items, Catalog Browsing, Inventory & Currency

Core data model, catalog browsing, basic inventory, and currency system.

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 1 | Create `item_definitions` table + model + migration (05) | — | NOT STARTED |
| 2 | Create `ItemDefinition` domain entity + repository interface | 1 | NOT STARTED |
| 3 | Create `item_definitions` GORM model + PostgreSQL repository | 1, 2 | NOT STARTED |
| 4 | Create `items` table + model + migration (06) | 1 | NOT STARTED |
| 5 | Create `Item` instance domain entity | 4 | NOT STARTED |
| 6 | Create `items` GORM model + extend repository | 4, 5 | NOT STARTED |
| 7 | Create `currency_types` + `user_currencies` tables + migration (07); seed default types (credits, duckets, diamonds, seasonal) | — | NOT STARTED |
| 8 | Create `CurrencyType` + `Currency` domain value objects + repository methods | 7 | NOT STARTED |
| 9 | Create `user_badges` table + model + migration (08) | — | NOT STARTED |
| 10 | Create `Badge` domain entity + repository methods | 9 | NOT STARTED |
| 11 | Create `catalog_pages` table + model + migration (09) | — | NOT STARTED |
| 12 | Create `CatalogPage` domain entity + repository | 11 | NOT STARTED |
| 13 | Create `catalog_items` + `catalog_featured_pages` tables + migration (10) | 11 | NOT STARTED |
| 14 | Create `CatalogOffer` domain entity + repository | 13 | NOT STARTED |
| 15 | Seed: 20 sample item definitions (05) | 1 | NOT STARTED |
| 16 | Seed: root + 3 sample catalog pages (06) | 11 | NOT STARTED |
| 17 | Seed: 10 sample catalog offers (07) | 13, 15 | NOT STARTED |
| 18 | Seed: test items, badges, currencies for test users (08-10) | 4, 9, 7 | NOT STARTED |
| 19 | `furniture.aliases` S2C packet (1723) | 2 | NOT STARTED |
| 20 | `furniture.furniture_aliases` C2S handler (3898) | 19 | NOT STARTED |
| 21 | `catalog.get_index` C2S → `catalog.index` S2C (1195→1032) | 12 | NOT STARTED |
| 22 | `catalog.get_page` C2S → `catalog.page` S2C (412→804) | 14 | NOT STARTED |
| 23 | `catalog.purchase` C2S → `catalog.purchase_ok` S2C (3492→869) | 14, 8 | NOT STARTED |
| 24 | `catalog.purchase_error` (1404) + `purchase_not_allowed` (3770) | 23 | NOT STARTED |
| 25 | `user.get_currency` C2S → `user.credits` + `user.currency` S2C | 8 | NOT STARTED |
| 26 | `user.activity_point_notification` S2C (2275) on currency change | 25 | NOT STARTED |
| 27 | `user.get_furniture` / `user.get_furniture_not_in_room` → `user.furniture` fragment S2C | 6 | NOT STARTED |
| 28 | `user.furniture_add` (104) / `user.furniture_remove` (159) / `user.furniture_refresh` (3151) | 27 | NOT STARTED |
| 29 | `user.unseen_items` (2103) + `unseen_reset_items`/`unseen_reset_category` handlers | 28 | NOT STARTED |
| 30 | `user.get_badges` → `user.badges` (717) + `user.badge_received` (2493) | 10 | NOT STARTED |
| 31 | `user.get_current_badges` → `user.current_badges` (1087) | 30 | NOT STARTED |
| 32 | `user.update_badges` C2S handler (644) | 30 | NOT STARTED |
| 33 | Wire credits + currency + subscription into post-auth burst | 25 | NOT STARTED |
| 34 | `catalog.not_enough_balance` S2C (3914) | 23 | NOT STARTED |
| 35 | Admin API: item definitions CRUD, user items, badges, currency | 3, 6, 10, 8 | NOT STARTED |
| 36 | Admin CLI: definitions, items, badges, currency | 35 | NOT STARTED |
| 37 | Plugin events: CatalogPurchase, CurrencyChanged, CreditsChanged, BadgeAwarded, BadgeRevoked, BadgeSlotsChanged | 23, 25, 30 | NOT STARTED |
| 38 | OpenAPI specs for all M1 endpoints | 35 | NOT STARTED |
| 39 | Unit + integration tests for M1 | all M1 | NOT STARTED |
| 40 | E2E tests: `e2e/11_economy/11_economy_test.go` | all M1 | NOT STARTED |

### Milestone 2: Marketplace, Vouchers, Effects & Gifting

Marketplace, voucher system, effects, gift flow, and credit exchange.

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 41 | Create `catalog_clothing` + `catalog_gift_wrapping` tables + migration (11) | M1 | NOT STARTED |
| 42 | Create `vouchers` + `voucher_redemptions` tables + migration (12) | M1 | NOT STARTED |
| 43 | Create `Voucher` domain entity + repository | 42 | NOT STARTED |
| 44 | Create `user_effects` table + model + migration (13) | M1 | NOT STARTED |
| 45 | Create `Effect` domain entity + repository | 44 | NOT STARTED |
| 46 | Create `marketplace_offers` + `marketplace_statistics` tables + migration (14) | M1 | NOT STARTED |
| 47 | Create `MarketplaceOffer` domain entity + repository | 46 | NOT STARTED |
| 48 | Seed: sample vouchers (11), marketplace offers (13) | 42, 46 | NOT STARTED |
| 49 | `catalog.redeem_voucher` C2S → `catalog.voucher_ok`/`voucher_error` S2C | 43 | NOT STARTED |
| 50 | `catalog.purchase_gift` C2S → gift flow + `gift_receiver_not_found` | M1.23, 41 | NOT STARTED |
| 51 | `catalog.get_gift_wrapping_config` C2S → `catalog.gift_wrapping_config` S2C | 41 | NOT STARTED |
| 52 | `catalog.check_giftable` C2S → `catalog.is_offer_giftable` S2C | M1.14 | NOT STARTED |
| 53 | `user.effects` S2C (340) on init + `user.effect_activate` C2S (2959) | 45 | NOT STARTED |
| 54 | `user.effect_added`/`effect_removed`/`effect_activated`/`effect_selected` S2C | 53 | NOT STARTED |
| 55 | Effect expiry background job | 45 | NOT STARTED |
| 56 | `furniture.item_exchange_redeem` C2S (3115) → credit exchange flow | M1.6, M1.8 | NOT STARTED |
| 57 | `furniture.present_open_present` C2S (3558) → `furniture.gift_opened` (56) S2C | M1.6 | NOT STARTED |
| 58 | `furniture.limited_sold_out` S2C (377) on limited item exhaustion | M1.23 | NOT STARTED |
| 59 | `marketplace.get_config` C2S → `marketplace.config` S2C | 47 | NOT STARTED |
| 60 | `marketplace.search_offers` C2S → `marketplace.items_searched` S2C | 47 | NOT STARTED |
| 61 | `marketplace.get_own_items` C2S → `marketplace.own_items` S2C | 47 | NOT STARTED |
| 62 | `marketplace.sell_item` C2S → `marketplace.item_posted` S2C | 47, M1.6 | NOT STARTED |
| 63 | `marketplace.buy_offer` C2S → `marketplace.buy_result` S2C | 47, M1.8 | NOT STARTED |
| 64 | `marketplace.cancel_sale` C2S → `marketplace.cancel_sale_result` S2C | 47 | NOT STARTED |
| 65 | `marketplace.get_item_stats` C2S → `marketplace.item_stats` S2C | 47 | NOT STARTED |
| 66 | `marketplace.get_can_sell` C2S → `marketplace.can_sell` S2C | 47 | NOT STARTED |
| 67 | `marketplace.redeem_credits` C2S → credit flush + `user.credits` update | 47 | NOT STARTED |
| 68 | Marketplace offer expiry background job | 47 | NOT STARTED |
| 69 | `user.clothing_redeem` C2S (3374) → `user.clothing` S2C (1450) | 41 | NOT STARTED |
| 70 | Admin API: vouchers, marketplace, effects | 43, 47, 45 | NOT STARTED |
| 71 | Admin CLI: vouchers, marketplace | 70 | NOT STARTED |
| 72 | Plugin events: VoucherRedeemed, MarketplaceItemListed/Purchased/Cancelled, EffectActivated, ItemExchangeRedeemed, GiftOpened | M2 tasks | NOT STARTED |
| 73 | OpenAPI specs for all M2 endpoints | 70 | NOT STARTED |
| 74 | Unit + integration tests for M2 | all M2 | NOT STARTED |
| 75 | E2E tests: marketplace, vouchers, gifting, effects | all M2 | NOT STARTED |

### Milestone 3: Subscriptions, Club Offers & Targeted Offers

| # | Task | Depends On | Status |
|---|------|------------|--------|
| 76 | Create `user_subscriptions` + `catalog_club_offers` tables + migration (15) | M2 | DONE |
| 77 | Create `Subscription` domain entity + repository | 76 | DONE |
| 78 | Create `ClubOffer` domain entity | 76 | DONE |
| 79 | Create `targeted_offers` + `user_targeted_offer_state` tables + migration (16) | M2 | NOT STARTED |
| 80 | Create `TargetedOffer` domain entity + repository | 79 | NOT STARTED |
| 81 | Seed: 3 club membership offers (12) | 76 | DONE |
| 82 | `user.get_subscription` C2S → `user.subscription` S2C (954) | 77 | DONE |
| 83 | Wire `user.subscription` into post-auth burst | 82 | DONE |
| 84 | `catalog.get_club_offers` C2S → `catalog.club_offers` S2C (encoding matches Arcturus vendor reference) | 78 | DONE |
| 85 | Club purchase flow (extend duration, update club_level) | 78, M1.8 | NOT STARTED |
| 86 | `catalog.get_hc_extend_offer` / `get_basic_extend_offer` C2S | 78 | NOT STARTED |
| 87 | `catalog.get_club_gift_info` C2S → `catalog.club_gift_info` S2C | 77 | DONE |
| 88 | `catalog.select_club_gift` C2S → `catalog.club_gift_selected` S2C | 87 | NOT STARTED |
| 89 | `catalog.club_gift_notification` S2C (2188) on login | 87 | NOT STARTED |
| 90 | Subscription expiry checker background job | 77 | NOT STARTED |
| 91 | `offer.get_targeted` / `get_next_targeted` C2S → `offer.targeted` S2C | 80 | NOT STARTED |
| 92 | `offer.purchase_targeted` C2S flow | 80, M1.8 | NOT STARTED |
| 93 | `offer.set_targeted_state` C2S handler | 80 | NOT STARTED |
| 94 | Admin API: subscription, club offers, targeted offers | 77, 78, 80 | DONE (subscription + club offers) |
| 95 | Admin CLI: subscription, club offers | 94 | DONE (subscription + club offers) |
| 96 | Plugin events: SubscriptionCreated/Expired/Extended, ClubGiftClaimed, TargetedOfferPurchased | M3 tasks | NOT STARTED |
| 97 | OpenAPI specs for all M3 endpoints | 94 | DONE (subscription + club offers) |
| 98 | Unit + integration tests for M3 | all M3 | DONE (subscription + club offers) |
| 99 | E2E tests: subscription flow, club offers, targeted offers | all M3 | NOT STARTED |

---

## Caveats & Technical Notes

### Package Distribution (AGENTS.md Compliance)

The economy system spans 5 packages under `pkg/`:
- `pkg/furniture/` — item definitions + instances (non-room)
- `pkg/catalog/` — catalog pages, offers, vouchers, gift wrapping
- `pkg/inventory/` — badges, effects, currencies, inventory loading
- `pkg/subscription/` — HC subscriptions, club offers, targeted offers
- `pkg/economy/` — marketplace, trading (deferred), credit exchange

Each package follows hexagonal layout with domain/application/adapter/
infrastructure layers. Each package registers its own migrations and
seeds. No cross-realm catch-all files.

### File Size Compliance

All source files must stay under 150 lines (excluding comments). Large
services are split into focused files per use case. Test files exceeding
150 lines use the internal `tests/` folder convention.

### Separation from Room Realm

The `items` table has `room_id` for future room placement, but no
position columns (`x`, `y`, `z`, `rot`, `wall_pos`) until the room realm
is implemented. This prevents premature coupling. The item instance
entity in `pkg/furniture/domain/item.go` has no room-position fields.

### Migration and Seed ID Coordination

Migrations continue from existing ID 04 (user_respects). Economy realm
uses IDs 05-17. Each migration file follows the existing pattern with
explicit up/down support.

Seeds continue from existing ID 04 (test_user_settings). Economy realm
uses IDs 05-13. Seeds provide minimal bootstrap data for testing.

### Event Placement

Events are placed in domain-scoped SDK folders following AGENTS.md:
- `sdk/events/economy/` — catalog, marketplace, currency events
- `sdk/events/inventory/` — badge, effect, inventory events
- `sdk/events/subscription/` — subscription lifecycle events

### E2E Test Organization

Economy E2E tests go under `e2e/11_economy/`:
- `e2e/11_economy/11_economy_test.go` — core catalog + inventory flow
- `e2e/11_economy/11_marketplace_test.go` — marketplace flow
- `e2e/11_economy/11_subscription_test.go` — subscription flow

### .env.example Updates

All new configuration variables must be added to `.env.example`:
```
# Economy
CATALOG_PURCHASE_COOLDOWN_MS=500
CATALOG_MAX_PURCHASE_AMOUNT=100
CATALOG_GIFT_ENABLED=true
MARKETPLACE_ENABLED=true
MARKETPLACE_COMMISSION_PCT=1
MARKETPLACE_MIN_PRICE=1
MARKETPLACE_MAX_PRICE=999999
MARKETPLACE_OFFER_EXPIRY_HOURS=48
MARKETPLACE_MAX_ACTIVE_OFFERS=30
INVENTORY_FRAGMENT_SIZE=1000
INVENTORY_MAX_ITEMS=5000
BADGE_MAX_SLOTS=5
SUBSCRIPTION_CHECK_INTERVAL_SEC=60
SUBSCRIPTION_DEFAULT_TYPE=habbo_club
CURRENCY_CREDITS_TYPE_ID=1
CURRENCY_INITIAL_AMOUNTS=1:0,0:0
```
