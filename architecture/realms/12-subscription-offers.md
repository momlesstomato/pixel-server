# Realm: Subscription & Offers

> **Position:** 100 | **Phase:** 8 (Subscriptions) | **Packets:** 50 (26 c2s, 24 s2c)
> **Services:** catalog, auth | **Status:** Not yet implemented

---

## Overview

The Subscription & Offers realm manages HC (Habbo Club) subscriptions, VIP membership, targeted marketing offers, builders club, seasonal calendars, membership extensions, and promotional content. At 50 packets it is larger than expected for a "read-heavy" system -- this is because the Nitro client has extensive UI for subscription management, targeted offers, seasonal events, and loyalty programs.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 8

---

## Packet Inventory

### C2S -- 26 packets

**Subscription Management:** `club.get_offers`, `club.subscribe`, `club.extend`, `club.cancel`, `club.get_center_info`, `club.get_gift_options`

**Targeted Offers:** `offers.get_targeted`, `offers.accept`, `offers.dismiss`, `offers.get_seasonal`, `offers.purchase_seasonal`

**Builders Club:** `builders.get_status`, `builders.purchase`, `builders.get_offers`

**Calendar/Seasonal:** `calendar.open`, `calendar.claim_day`, `calendar.get_campaigns`

**Loyalty/Rewards:** `loyalty.get_points`, `loyalty.claim_reward`

**Miscellaneous:** Additional subscription query and preference packets.

### S2C -- 24 packets

**Subscription Status:** `club.subscription_status` (days remaining, is VIP, streak, paused), `club.offers_list`, `club.gift_options`, `club.center_info` (dashboard with stats)

**Targeted Offers:** `offers.targeted_offer` (banner with price, items, timer), `offers.seasonal_calendar` (grid of daily rewards)

**Builders Club:** `builders.status` (active, seconds remaining, furni limit)

**Calendar:** `calendar.data` (days, claimed, rewards per day)

**Notifications:** `offers.expiring_soon`, `club.expired`, `club.loyalty_info`

---

## Architecture Mapping

### Service Ownership

Split between **catalog service** (purchase flow) and **auth service** (status computation):

- **Auth service** computes subscription status on login and includes it in the login bundle.
- **Catalog service** handles subscription purchases and extensions.
- Subscription state persists in `user_subscriptions` table.

### Database Tables

| Table | Columns | Usage |
|-------|---------|-------|
| `user_subscriptions` | user_id, product_name, start_date, end_date, is_vip, paused, streak_days | Active subscription |
| `targeted_offers` | id, title, description, image, items[], cost_credits, cost_points, start_time, end_time, target_segment | Marketing offers |
| `seasonal_calendar` | campaign_id, day, reward_type, reward_data | Calendar daily rewards |
| `user_calendar_claims` | user_id, campaign_id, day, claimed_at | Claimed calendar days |
| `builders_club` | user_id, active, seconds_remaining, furni_limit, expires_at | Builders club status |

---

## Implementation Analysis

### Subscription Lifecycle

```
Subscribe:
  1. Client purchases HC/VIP via catalog
  2. Catalog service creates/extends subscription
  3. Update user_subscriptions (start_date, end_date, is_vip)
  4. Send club.subscription_status to client
  5. Publish subscription.changed event to NATS
  6. Game service updates user perks (wardrobe slots, figure items)

Expiry:
  1. Background job checks expired subscriptions (every 5 minutes)
  2. For each expired: mark inactive, clear VIP perks
  3. If user online: send club.expired packet
  4. Publish subscription.expired event

Streak:
  - streak_days increments for consecutive monthly renewals
  - streak_days resets to 0 if subscription lapses > 3 days
  - Streak unlocks loyalty rewards at milestones
```

### Targeted Offers

Targeted offers are server-pushed marketing:
1. On login, query `targeted_offers WHERE target_segment matches user AND active`.
2. Send `offers.targeted_offer` with banner, timer, and discounted price.
3. User can accept (purchase) or dismiss.
4. Track impressions and conversions in analytics.

**Segmentation rules:** Based on user attributes like account age, last purchase date, subscription status, credits balance, login frequency.

### Seasonal Calendar

A time-limited grid of daily rewards:
1. On `calendar.open`, send calendar grid with rewards and claimed status.
2. User can claim today's reward if not already claimed.
3. Reward types: credits, items, badges, effects.
4. Some days require consecutive logins (streak requirement).

---

## Caveats & Edge Cases

### 1. Subscription Extension vs Renewal
Extending adds days to `end_date`. Renewing after expiry sets a new `start_date`. Both must handle the streak counter differently.

### 2. VIP Downgrade
VIP users have exclusive items and features. When VIP expires but HC continues, VIP-exclusive features must be revoked (wardrobe slots reduced from 10 to 5, VIP figure items hidden).

### 3. Timezone Handling
Calendar claim windows should be server-timezone based (UTC). A user claiming at 23:59 UTC and again at 00:01 UTC should get two different days.

### 4. Builders Club Overlap
Builders club is separate from HC/VIP. A user can have both simultaneously. Furniture placement limits stack (club limit + normal limit).

### 5. Offer Timing
Targeted offers have expiry timestamps. The client shows a countdown. If the offer expires while the user is viewing it, the purchase must still fail server-side.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Subscription state** | Checked on every action | Cached in session, events on change |
| **Targeted offers** | Static config file | Database-driven with segmentation |
| **Calendar** | Hardcoded rewards | Database-configurable per campaign |
| **Streak tracking** | Not implemented in most emulators | Automatic with 3-day grace period |

---

## Dependencies

- **Phase 2 (Identity)** -- subscription status in login bundle
- **Phase 7 (Catalog)** -- purchase mechanics
- **PostgreSQL** -- subscriptions, offers, calendar tables
- **Redis** -- subscription status cache

---

## Testing Strategy

### Unit Tests
- Subscription duration calculation (extension, renewal, expiry)
- Streak tracking (consecutive, lapsed, grace period)
- Calendar claim eligibility (per-day, streak requirement)
- VIP perk toggling on status change

### Integration Tests
- Subscription purchase and status update flow
- Calendar claim persistence and duplicate prevention
- Targeted offer filtering by user segment

### E2E Tests
- Client purchases HC subscription, sees updated perks
- Client opens calendar, claims today's reward
