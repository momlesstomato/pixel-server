# HC Club

This document explains how Pixel Server currently configures the Habbo Club
shop, monthly HC gifts, and HC payday timing.

## Overview

HC data is split across three concerns:

| Concern | Storage | Purpose |
|---------|---------|---------|
| HC shop shell | `catalog_pages` | Provides the `vip_buy` and `club_gifts` catalog page shells |
| HC offers | `catalog_club_offers` | Defines the purchasable membership offers sent in `2405` |
| HC gifts | `subscription_club_gifts` | Defines the monthly redeemable furniture gifts sent in `619` |
| HC payday config | `subscription_payday_config` | Defines the recurring payday cadence and reward formula |
| Per-user HC progress | `subscription_benefits` | Tracks next payday time, current cycle spend, streak, and claimed gifts |

## HC Shop

The HC shop uses a split payload model:

- `catalog.page` (`804`) provides the `vip_buy` page shell.
- `catalog.club_offers` (`2405`) provides the actual HC offer entries.
- `catalog.direct_sms_club_buy` (`195`) is also emitted with an empty payload
  so Nitro receives the optional direct-buy availability event expected by some
  club-buy flows.

The seeded page shell currently lives in `pkg/catalog/infrastructure/seed/catalog_pages.go`
as `HC Shop`, uses two localization images, and is backfilled to `vip_buy` for
modern Nitro compatibility.

## Configuring HC Gifts

Monthly HC gifts are currently configured from the database table
`subscription_club_gifts`.

Relevant columns:

| Column | Meaning |
|--------|---------|
| `name` | Client-visible gift name and selector key |
| `item_definition_id` | Furniture definition delivered on claim |
| `extra_data` | Optional custom item data |
| `days_required` | HC age required before this gift becomes selectable |
| `vip_only` | Restrict the gift to VIP offers only |
| `enabled` | Include or exclude the gift from `619` |
| `order_num` | Display order inside the selector |

Example SQL:

```sql
INSERT INTO subscription_club_gifts
    (name, item_definition_id, extra_data, days_required, vip_only, enabled, order_num)
VALUES
    ('HC Amber Lamp', 250, '', 31, false, true, 10);
```

Operational notes:

- `item_definition_id` must reference an existing furniture definition.
- The client preview sprite is resolved from the linked furniture definition.
- Gift availability is derived from HC age in days and the number of gifts the
  player has already claimed from `subscription_benefits.club_gifts_claimed`.
- The bootstrap defaults are seeded in
  `pkg/subscription/infrastructure/seed/club_gifts.go`.

## Configuring HC Payday Timing

HC payday timing and rewards are configured through the singleton row in
`subscription_payday_config` and can be managed through HTTP.

Available fields:

| Field | Meaning |
|-------|---------|
| `interval_days` | Number of days between paydays |
| `kickback_percentage` | Percentage of cycle spend returned on payday |
| `flat_credits` | Fixed credit reward added every payday |
| `minimum_credits_spent` | Minimum spend required before kickback applies |
| `streak_bonus_credits` | Extra reward after consecutive successful paydays |

HTTP endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/v1/subscriptions/payday/config` | Read the active payday config |
| `PATCH` | `/api/v1/subscriptions/payday/config` | Update the active payday config |
| `GET` | `/api/v1/subscriptions/user/{userId}/payday` | Inspect one user's current payday state |
| `POST` | `/api/v1/subscriptions/user/{userId}/payday/trigger` | Force or execute one payday |

Example request:

```http
PATCH /api/v1/subscriptions/payday/config
X-API-Key: <api-key>
Content-Type: application/json

{
  "interval_days": 31,
  "kickback_percentage": 10,
  "flat_credits": 5,
  "minimum_credits_spent": 25,
  "streak_bonus_credits": 2
}
```

Operational notes:

- Payday timing is derived from `subscription_benefits.next_payday_at`, not
  subscription expiry.
- Successful catalog credit purchases feed the current payday cycle through the
  subscription purchase observer.
- A default config row is seeded in
  `pkg/subscription/infrastructure/seed/club_gifts.go`.

## SDK Events

Plugins can hook the HC flow through the subscription event package:

| Event | Cancellable | Purpose |
|-------|-------------|---------|
| `PaydayTriggering` | Yes | Before one payday reward is granted |
| `PaydayTriggered` | No | After one payday reward is granted |
| `ClubGiftClaiming` | Yes | Before one HC gift is delivered |
| `ClubGiftClaimed` | No | After one HC gift is delivered |

These events live under `sdk/events/subscription/`.