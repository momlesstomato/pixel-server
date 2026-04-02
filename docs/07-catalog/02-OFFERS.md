# Offers, Products, Pricing, and Vouchers

## Offers

A `CatalogOffer` is one purchasable entry within a catalog page. It maps a
furniture definition to a price and optional constraints.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | Stable offer identifier |
| `PageID` | int | Owning catalog page |
| `ItemDefinitionID` | int | Foreign key to the furniture item definition |
| `CatalogName` | string | Display name resolved from `item_definitions.public_name` at read time — **not stored in the database** (see [Name Resolution](#offer-name-resolution)) |
| `CostCredits` | int | Credits price; zero for activity-point-only offers |
| `CostActivityPoints` | int | Activity-point price; zero when no activity-point charge |
| `ActivityPointType` | int | Registered activity-point currency type ID (see [Currency Types](#currency-types)) |
| `Amount` | int | Number of items delivered per single purchase |
| `LimitedTotal` | int | Total print run for limited editions; `0` means unlimited |
| `LimitedSells` | int | Running count of limited edition sold units |
| `OfferActive` | bool | Whether the offer is currently purchasable |
| `ExtraData` | string | Item-specific custom data (colour preset, text, etc.) |
| `BadgeID` | string | Optional badge code awarded alongside the item |
| `ClubOnly` | bool | Requires an active club subscription to purchase |
| `OrderNum` | int | Display sort position within the page |

### Domain helpers

```go
// IsLimited reports whether this offer is a limited edition.
func (o CatalogOffer) IsLimited() bool { return o.LimitedTotal > 0 }

// HasStock reports whether limited stock remains.
func (o CatalogOffer) HasStock() bool
```

---

## Offer Name Resolution

Three name concepts exist across the furniture and catalog data models:

| Field | Table | Purpose |
|-------|-------|---------|
| `item_name` | `item_definitions` | Internal unique key for server-side logic (interaction dispatch, pattern matching). Never sent to the client. |
| `public_name` | `item_definitions` | Canonical display name — used in inventory, room view, marketplace, and as the offer display name. |
| ~~`catalog_name`~~ | ~~`catalog_items`~~ | **Removed.** Previously an optional per-offer alias; the column no longer exists in the schema. |

### How the name is resolved

Every offer query JOINs `item_definitions` and selects `public_name` directly
as the effective display name:

```sql
SELECT ci.*, id.public_name AS effective_name
FROM catalog_items ci
LEFT JOIN item_definitions id ON id.id = ci.item_definition_id
```

The `CatalogName` field on the domain entity is always populated from this
JOIN result. It is not present in any write path (create / update) and is
not stored in the database.

---

## Product Types

Each offer contains one or more **products** — the actual item records that
are written to the wire inside an offer entry. The product type code
determines the binary layout of the product record.

| Type code | Meaning | Wire fields |
|-----------|---------|-------------|
| `"s"` | Floor furniture (roomitemtypes) | `spriteId` (int32), `extraParam` (string), `amount` (int32), `isLimited` (bool), and if limited: `seriesSize` (int32) + `remaining` (int32) |
| `"i"` | Wall furniture (wallitemtypes) | Same structure as `"s"` |
| `"e"` | Avatar effect | Same structure as `"s"` |
| `"b"` | Badge | `extraParam` (string = badge code) only — no spriteId, amount, or limited fields |

When an offer includes both a badge and a regular furniture item, two product
records are written: the badge record first (`"b"`), then the item record
(`"s"` or `"i"`).

### Special case: Deal / Room bundle

When an offer's item definition has the interaction type `Deal` or `Roomdeal`,
the products list is populated from the deal's sub-items rather than from
a single `ItemDefinition`. Each sub-item inside the deal uses the type code
matching its own furniture type.

## Currency Types

The Habbo wire protocol encodes offer pricing with two **independent** price
components; neither is mandatory — both can be zero simultaneously for free
offers, or only one can carry a charge:

- **Credits (`CostCredits`)** — maps to the wire `priceCredits` field. Set to
  zero for offers that have no credit cost.
- **Activity Points (`CostActivityPoints`)** — maps to the wire
  `priceActivityPoints` field. `ActivityPointType` identifies which registered
  currency the points belong to.

### Activity-Point Type Registry

Activity-point type IDs are stored in the `currency_types` database table
(managed by the inventory realm). The server validates `ActivityPointType`
against this table when `CostActivityPoints > 0`; unknown or disabled type IDs
are rejected.

The three Habbo-standard types are seeded on first run:

| ID | Name | Display Name |
|-----|------|--------------|
| `0` | `duckets` | Duckets |
| `5` | `diamonds` | Diamonds |
| `105` | `seasonal` | Seasonal Points |

Operators may insert additional rows into `currency_types` at any time to
introduce custom activity-point currencies — no code change is required.
The convenience constants `CurrencyDuckets`, `CurrencyDiamonds`, and
`CurrencySeasonal` in `pkg/inventory/domain/currency.go` are named aliases for
these three seeded IDs and are not a closed list.

## Limited Editions

An offer is a limited edition when `LimitedTotal > 0`. The server tracks sold
units via `LimitedSells`. When `LimitedSells >= LimitedTotal` the offer is
logically out of stock (`HasStock()` returns false) and further purchase
attempts are rejected.

The wire encoding for a limited product record includes both the series size
(`LimitedTotal`) and the number of units still available
(`LimitedTotal - LimitedSells`). The client renders an availability counter
from these two values.

```
products: [
  { type: "i", spriteId: 42, extraParam: "", amount: 1,
    isLimited: true, seriesSize: 100, remaining: 7 }
]
```

## Gift Flow

A user may purchase an offer as a gift for another user. The `catalog.purchase_gift`
packet (c2s 1411) extends the normal purchase with:

| Field | Description |
|-------|-------------|
| `receiverName` | Target username |
| `giftMessage` | Personal message attached to the gift |
| `spriteId` | Wrapping paper sprite |
| `boxId` | Gift box style |
| `ribbonId` | Ribbon style |
| `showMyFace` | Include sender's avatar face on the gift |

The available gift wrapping options are requested separately via
`catalog.get_gift_wrapping_config` (c2s 418) → `catalog.gift_wrapping_config`
(s2c 2234) before the gift dialog opens.

A giftable offer must have `OfferActive = true`. The `giftable` flag in the
wire offer record is derived from the offer's active state and item type — not
all item types support gift wrapping (e.g. pets and bot presets cannot be
gifted).

## Vouchers

A `Voucher` is a one-time redeemable code that delivers a reward when
submitted by the player.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | Stable voucher identifier |
| `Code` | string | Unique redeemable code string |
| `RewardType` | string | Reward category: `"currency"`, `"badge"`, or `"furniture"` |
| `RewardCurrencyType` | *int | Currency type code when `RewardType` is `"currency"` |
| `RewardData` | string | Reward-specific payload (amount, badge code, item definition ID) |
| `MaxUses` | int | Total allowed redemptions across all users |
| `CurrentUses` | int | Running count of completed redemptions |
| `Enabled` | bool | Whether the voucher accepts new redemptions |

### Redemption audit

Every redemption is recorded in a `VoucherRedemption` row:

| Field | Type | Description |
|-------|------|-------------|
| `VoucherID` | int | Redeemed voucher |
| `UserID` | int | Redeeming user |
| `RedeemedAt` | time.Time | Timestamp of redemption |

### Redemption flow

1. Player sends `catalog.redeem_voucher` (c2s 339) with the code string.
2. Server looks up the voucher by code.
3. Domain guards are evaluated:
   - `ErrVoucherNotFound` — code does not exist
   - `ErrVoucherDisabled` — `Enabled = false`
   - `ErrVoucherExhausted` — `CurrentUses >= MaxUses`
   - `ErrVoucherAlreadyRedeemed` — a `VoucherRedemption` row already exists for this user + voucher pair
4. `CurrentUses` is incremented and a `VoucherRedemption` row is inserted
   atomically.
5. The reward is delivered based on `RewardType`:
   - `"currency"` — balance credited to `RewardCurrencyType` using the amount parsed from `RewardData`
   - `"badge"` — `RewardData` holds a badge code; the badge is granted to the player
   - `"furniture"` — `RewardData` holds a furniture definition ID; the item is added to inventory
6. On success: `catalog.voucher_ok` (s2c 3336).
   On any domain error: `catalog.voucher_error` (s2c 714).

### Domain errors

| Error | Meaning |
|-------|---------|
| `ErrVoucherNotFound` | Code does not exist in the database |
| `ErrVoucherDisabled` | Voucher exists but is administratively disabled |
| `ErrVoucherExhausted` | Maximum redemption count has been reached |
| `ErrVoucherAlreadyRedeemed` | This user already redeemed this voucher |

## Domain Errors Reference

| Error | Trigger |
|-------|---------|
| `ErrPageNotFound` | Page ID does not exist |
| `ErrOfferNotFound` | Offer ID does not exist |
| `ErrOfferInactive` | Offer exists but `OfferActive = false` |
| `ErrPageDisabled` | Page `Enabled = false` |
| `ErrInsufficientRank` | `MinPermission` check failed |
| `ErrClubRequired` | `ClubOnly = true` and player has no club subscription |
| `ErrRecipientNotFound` | Gift target username does not exist |
| `ErrPurchaseCooldown` | Purchase rate-limit triggered |
