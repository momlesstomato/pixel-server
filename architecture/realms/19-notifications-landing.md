# Realm: Notifications & Landing

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 200 | **Phase:** 13 (Remaining) | **Packets:** 22 (6 c2s, 16 s2c)
> **Services:** gateway, game | **Status:** Not yet implemented

---

## Overview

Notifications & Landing manages the hotel view (desktop/landing page), news articles, promotional content, info feed, welcome gifts, and the new-user experience (NUX). This is the most S2C-heavy realm (16 server-to-client vs 6 client-to-server), reflecting its push-notification nature.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 6 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `landing.get_promo_articles` | _(none)_ | Request news/promo articles |
| -- | `landing.get_community_goal` | _(none)_ | Get community goal for landing |
| -- | `landing.get_hall_of_fame` | _(none)_ | Get hall of fame data |
| -- | `landing.welcome_gift_claim` | _(none)_ | Claim welcome gift |
| -- | `nux.completed_step` | `stepId:int32` | Mark NUX step complete |
| -- | `notifications.mark_read` | `notificationId:int32` | Dismiss notification |

### S2C -- 16 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `landing.promo_articles` | News articles with images |
| -- | `landing.community_goal` | Community progress bar |
| -- | `landing.hall_of_fame` | Top players list |
| -- | `landing.welcome_gift` | Available welcome gift |
| -- | `nux.status` | NUX progress state |
| -- | `nux.step_data` | NUX step instructions |
| -- | `notification.alert` | Generic notification popup |
| -- | `notification.bubble` | Chat-bubble-style notification |
| -- | `notification.achievement` | Achievement unlock notification |
| -- | `notification.info_feed` | Info feed item |
| -- | `notification.motd` | Message of the day |
| -- | `notification.offer` | Special offer notification |
| -- | `notification.moderation` | Moderation notice |
| -- | `notification.gift_received` | Gift notification |
| -- | `notification.friend_request` | Friend request alert |
| -- | `notification.badge_received` | New badge notification |

---

## Implementation Analysis

### Landing Page Data

On hotel view (desktop_view), the client requests landing page content:
- **Promo articles:** Loaded from `landing_articles` table (admin-managed).
- **Community goals:** Aggregated data from community goal tracking.
- **Hall of fame:** Top players by achievement score.

### Notification System

Notifications are a central dispatch system for all realms:

```go
type NotificationService interface {
    SendAlert(ctx context.Context, userID int32, message string) error
    SendBubble(ctx context.Context, userID int32, icon string, message string) error
    SendAchievement(ctx context.Context, userID int32, badge string) error
}
```

All notifications route through `session.output.<sessionID>` via NATS. For offline users, persistent notifications are stored and delivered on login.

### NUX (New User Experience)

NUX guides new users through first steps:
1. Change your look (link to figure editor)
2. Enter a room (link to navigator)
3. Chat with someone (send a message)
4. Make a friend (send friend request)

Each step is tracked in `user_settings.nux_step`. On completion, a welcome gift is awarded.

### Welcome Gift

Configurable via `landing_welcome_gifts` table:
- Available for first-time users only
- Contains items, credits, or badges
- Claimed via `landing.welcome_gift_claim`
- One-time claim (idempotent)

---

## Caveats

### 1. Notification Deduplication
Multiple systems may trigger the same notification type (e.g., badge received from achievement + catalog). Deduplicate by notification type + context within a 5-second window.

### 2. Promo Article Caching
Landing page articles change infrequently. Cache in Redis with admin-triggered invalidation.

### 3. NUX Persistence
NUX state must survive reconnections. If a user completes step 2 and reconnects, they should resume at step 3.

---

## Dependencies

- **Phase 2 (Identity)** -- user settings for NUX state
- **All other phases** -- notifications triggered by every realm

---

## Testing Strategy

### Unit Tests
- NUX step progression logic
- Welcome gift claim idempotency
- Notification deduplication

### Integration Tests
- Landing page data loading
- Notification delivery to online user
- NUX completion triggers welcome gift
