# 08 - Remaining Packets: Full Implementation Plan

## Overview

The pixel-protocol defines **922 packets** (463 C2S, 459 S2C) across 16
realms. This plan inventories every realm, states which packets are
implemented, which are next, and which are deferred until a Room realm
or other prerequisite exists.

**Status legend:** DONE = fully handled, STUB = dispatch exists but body
is empty/placeholder, TODO = not yet coded, DEFER = blocked on Room realm
or other missing prerequisite.

---

## Realm Status Summary

| Realm | C2S Total | S2C Total | Implemented | Stubs | TODO (doable) | DEFER |
|-------|-----------|-----------|-------------|-------|---------------|-------|
| Handshake & Security | 8 | 5 | 13 | 0 | 0 | 0 |
| Session & Connection | 10 | 20 | 30 | 0 | 0 | 0 |
| User & Profile | 29 | 27 | ~50 | 0 | ~6 | 0 |
| Messenger & Social | 14 | 16 | 30 | 0 | 0 | 0 |
| Navigator | 37 | 18 | 22 | 0 | 0 | ~33 |
| Room | 46 | 44 | 0 | 0 | 0 | 90 |
| Room Entities | 14 | 20 | 0 | 0 | 0 | 34 |
| Furniture & Items | 52 | 47 | 1 | 3 | ~10 | ~85 |
| Catalog & Store | 10 | 11 | 12 | 0 | 7 | 0 |
| Subscription & Offers | 26 | 24 | 7 | 0 | ~10 | ~32 |
| Economy & Trading | 28 | 26 | 16 | 0 | ~9 | ~29 |
| Inventory | 13 | 20 | 11 | 0 | ~12 | ~10 |
| Groups & Forums | 36 | 28 | 0 | 0 | ~15 | ~49 |
| Pets | 21 | 20 | 0 | 0 | 0 | 41 |
| Achievements & Talents | 10 | 14 | 0 | 0 | ~8 | ~16 |
| Quests & Campaigns | 15 | 18 | 0 | 0 | 0 | 33 |
| Games & Entertainment | 21 | 28 | 0 | 0 | 0 | 49 |
| Moderation & Safety | 43 | 40 | 0 | 0 | ~10 | ~73 |
| Camera & Photos | 8 | 7 | 0 | 0 | 0 | 15 |
| Notifications & Landing | 6 | 16 | 0 | 0 | ~6 | ~16 |
| Crafting & Recycling | 9 | 7 | 0 | 0 | 0 | 16 |
| Other | 7 | 3 | 0 | 0 | 0 | 10 |

---

## Phase 1: Complete Economy Realm (DONE)

### Milestone 1.1 — Catalog Completion (DONE)

Complete the remaining catalog packets. These do not require rooms.

| # | Packet | ID | Dir | Status | Task |
|---|--------|----|-----|--------|------|
| 1 | `catalog.redeem_voucher` | 339 | C2S | TODO | Validate voucher code, apply rewards, send 3336 or 714 |
| 2 | `catalog.voucher_ok` | 3336 | S2C | TODO | Confirm voucher redemption |
| 3 | `catalog.voucher_error` | 714 | S2C | TODO | Report voucher failure |
| 4 | `catalog.check_giftable` | 1347 | C2S | TODO | Check offer giftable flag, send 761 |
| 5 | `catalog.is_offer_giftable` | 761 | S2C | TODO | Return giftable status |
| 6 | `catalog.bundle_discount_ruleset` | 223 | C2S | TODO | Get bundle discount rules |
| 7 | `catalog.bundle_discount_ruleset` | 2347 | S2C | TODO | Return bundle rules |
| 8 | `catalog.gift_receiver_not_found` | 1517 | S2C | TODO | Gift recipient error |
| 9 | `catalog.published` | 1866 | S2C | TODO | Catalog update notification |
| 10 | `catalog.get_gift` | 2436 | C2S | TODO | Request gift delivery details |
| 11 | `catalog.mark_catalog_new_additions_page_opened` | 2150 | C2S | TODO | No-op acknowledgement |

**Caveats:**
- Voucher redemption needs a `vouchers` table with code, reward type,
  reward amount, uses/max-uses, expiry. Schema exists in plan 07.
- Bundle discount is client-cosmetic only — no server logic needed beyond
  sending the ruleset packet.

### Milestone 1.2 — Inventory Full Implementation (DONE)

Complete inventory management beyond furniture and currency.

| # | Packet | ID | Dir | Status | Task |
|---|--------|----|-----|--------|------|
| 1 | `user.get_badges` | 2769 | C2S | STUB | Send full badge inventory (717) |
| 2 | `user.badges` | 717 | S2C | TODO | Encode badge list + slot assignments |
| 3 | `user.get_current_badges` | 2091 | C2S | TODO | Handle, send 1087 |
| 4 | `user.current_badges` | 1087 | S2C | TODO | Encode equipped badge slots |
| 5 | `user.update_badges` | 644 | C2S | TODO | Save badge slot assignments |
| 6 | `user.badge_received` | 2493 | S2C | TODO | New badge notification |
| 7 | `user.unseen_reset_items` | 2343 | C2S | TODO | Mark items as seen |
| 8 | `user.unseen_reset_category` | 3493 | C2S | TODO | Mark category as seen |
| 9 | `user.effects` | 340 | S2C | TODO | Encode effect inventory |
| 10 | `user.effect_activate` | 2959 | C2S | STUB | Activate effect, send 1959 |
| 11 | `user.effect_activated` | 1959 | S2C | TODO | Confirm effect activation |
| 12 | `user.effect_selected` | 3473 | S2C | TODO | Selected effect confirmation |
| 13 | `user.effect_added` | 2867 | S2C | TODO | New effect notification |
| 14 | `user.effect_removed` | 2228 | S2C | TODO | Effect expiry notification |
| 15 | `user.get_furniture_not_in_room` | 3500 | C2S | TODO | Same as 3150 variant |
| 16 | `user.furniture_add` | 104 | S2C | TODO | Single item inventory add |
| 17 | `user.furniture_remove` | 159 | S2C | TODO | Single item inventory remove |
| 18 | `user.furniture_refresh` | 3151 | S2C | TODO | Invalidation signal |
| 19 | `user.clothing` | 1450 | S2C | TODO | Unlocked clothing sets |
| 20 | `user.clothing_redeem` | 3374 | C2S | TODO | Redeem clothing from furniture |
| 21 | `user.get_group_badges` | 21 | C2S | DEFER | Requires groups |
| 22 | `user.get_group_memberships` | 367 | C2S | DEFER | Requires groups |

**Caveats:**
- Badge storage needs a `user_badges` table and a `badge_slots` table (or combined).
- Effects need a `user_effects` table with duration/activation tracking.
- The `3500` (get_furniture_not_in_room) handler can reuse the `3150`
  handler logic since we have no room concept yet.

### Milestone 1.3 — Marketplace Implementation (DONE)

Fill in the marketplace stubs with real logic.

| # | Packet | ID | Dir | Status | Task |
|---|--------|----|-----|--------|------|
| 1 | `marketplace.get_config` | 2597 | C2S | STUB | Return config (1823) |
| 2 | `marketplace.config` | 1823 | S2C | TODO | Encode marketplace settings |
| 3 | `marketplace.sell_item` | 3447 | C2S | TODO | List item for sale |
| 4 | `marketplace.item_posted` | 1359 | S2C | TODO | Confirm listing |
| 5 | `marketplace.can_sell` | 54 | S2C | TODO | Sell permission check |
| 6 | `marketplace.get_can_sell` | 848 | C2S | TODO | Request sell permission |
| 7 | `marketplace.search_offers` | 2407 | C2S | STUB | Search, send 680 |
| 8 | `marketplace.items_searched` | 680 | S2C | TODO | Encode search results |
| 9 | `marketplace.buy_offer` | 1603 | C2S | TODO | Purchase listing |
| 10 | `marketplace.buy_result` | 2032 | S2C | TODO | Purchase result |
| 11 | `marketplace.get_own_items` | 2105 | C2S | STUB | Own listings, send 3884 |
| 12 | `marketplace.own_items` | 3884 | S2C | TODO | Encode own listings |
| 13 | `marketplace.cancel_sale` | 434 | C2S | STUB | Cancel listing, send 3264 |
| 14 | `marketplace.cancel_sale_result` | 3264 | S2C | TODO | Confirm cancellation |
| 15 | `marketplace.get_item_stats` | 3288 | C2S | STUB | Price history, send 725 |
| 16 | `marketplace.item_stats` | 725 | S2C | TODO | Encode price statistics |
| 17 | `marketplace.redeem_credits` | 2650 | C2S | TODO | Cash out sales revenue |
| 18 | `marketplace.buy_tokens` | 1866 | C2S | DEFER | Token purchase (optional) |

**Performance Bottleneck:** Marketplace search with sorting/filtering can
be slow on large datasets. Solution: PostgreSQL GIN index on item names,
server-side result pagination (limit 100 per page), Redis cache for popular
searches with 60s TTL.

**Caveats:**
- Marketplace commission rate must be configurable.
- Offer expiry requires a background goroutine (like messenger purge).
- Price statistics require a `marketplace_statistics` table with daily
  aggregation.

### Milestone 1.4 — Subscription & Offers Core (DONE)

| # | Packet | ID | Dir | Status | Task |
|---|--------|----|-----|--------|------|
| 1 | `user.get_subscription` | 3166 | C2S | DONE | Already handling |
| 2 | `user.subscription` | 954 | S2C | DONE | Already responding |
| 3 | `catalog.get_club_offers` | 3285 | C2S | STUB | Send real offers (2405) |
| 4 | `catalog.club_offers` | 2405 | S2C | TODO | Encode available offers |
| 5 | `catalog.get_club_gift_info` | 487 | C2S | TODO | Gift eligibility |
| 6 | `catalog.club_gift_info` | 619 | S2C | TODO | Gift options |
| 7 | `catalog.select_club_gift` | 2276 | C2S | TODO | Claim gift |
| 8 | `catalog.club_gift_selected` | 659 | S2C | TODO | Confirm claimed |
| 9 | `catalog.get_hc_extend_offer` | 2462 | C2S | TODO | Extension offer |
| 10 | `catalog.club_extend_offer` | 3964 | S2C | TODO | Extension details |
| 11 | `catalog.get_product_offer` | 2594 | C2S | TODO | Product details |
| 12 | `catalog.product_offer` | 3388 | S2C | TODO | Product response |
| 13 | `user.get_kickback_info` | 869 | C2S | TODO | HC kickback |
| 14 | `user.kickback_info` | 3277 | S2C | TODO | Kickback response |

**Deferred subscription packets** (targeted offers, campaigns, NUX, etc.):

| Packet | ID | Dir | Reason |
|--------|----|-----|--------|
| `offer.get_targeted` | 596, 2487 | C2S | Targeted offer engine |
| `offer.targeted` | 119 | S2C | Targeted offer engine |
| `offer.purchase_targeted` | 1826 | C2S | Targeted offer engine |
| `offer.set_targeted_state` | 2041 | C2S | Targeted offer engine |
| `offer.targeted_viewed` | 3483 | C2S | Targeted offer engine |
| `offer.targeted_not_found` | 1237 | S2C | Targeted offer engine |
| `calendar.*` | various | both | Campaign calendar system |
| `subscription.bonus_rare_info` | 957/1533 | both | Bonus rare system |
| `subscription.start_campaign` | 1697 | C2S | Campaign engine |
| `catalog.get_limited_offer_next` | 410 | C2S | Limited offer scheduler |
| `catalog.limited_offer_appearing_next` | 44 | S2C | Limited offer scheduler |
| `catalog.get_basic_extend_offer` | 603 | C2S | Basic membership |
| `catalog.get_direct_club_buy` | 801 | C2S | SMS purchase |
| `catalog.builders_club_*` | various | both | Builders Club (legacy) |
| `catalog.get_page_expiration` | 742/2668 | both | Page expiry |
| `catalog.get_earliest_expiry` | 3135/2515 | both | Page expiry |
| `catalog.not_enough_balance` | 3914 | S2C | Balance check |
| `subscription.new_user_experience_*` | 3575/3639 | S2C | NUX gifts |

### Milestone 1.5 — Trading (DEFERRED)

All trading packets are **deferred** until the Room realm exists, since
trading requires both parties to be in the same room.

| Packet | ID | Dir | Note |
|--------|----|-----|------|
| `trade.open` | 1481 | C2S | Room-dependent |
| `trade.add_item` | 3107 | C2S | Room-dependent |
| `trade.add_items` | 1263 | C2S | Room-dependent |
| `trade.remove_item` | 3845 | C2S | Room-dependent |
| `trade.accept` | 3863 | C2S | Room-dependent |
| `trade.unaccept` | 1444 | C2S | Room-dependent |
| `trade.confirm` | 2760 | C2S | Room-dependent |
| `trade.cancel` | 2341 | C2S | Room-dependent |
| `trade.close` | 2551 | C2S | Room-dependent |
| `trade.opened` | 2505 | S2C | Room-dependent |
| `trade.list_item` | 2024 | S2C | Room-dependent |
| `trade.accepted` | 2568 | S2C | Room-dependent |
| `trade.confirmation` | 2720 | S2C | Room-dependent |
| `trade.completed` | 1001 | S2C | Room-dependent |
| `trade.closed` | 1373 | S2C | Room-dependent |
| `trade.open_failed` | 217 | S2C | Room-dependent |
| `trade.other_not_allowed` | 1254 | S2C | Room-dependent |
| `trade.you_not_allowed` | 3058 | S2C | Room-dependent |
| `trade.not_open` | 3128 | S2C | Room-dependent |
| `trade.no_such_item` | 2873 | S2C | Room-dependent |

---

## Phase 2: Navigator (DONE — No Room Loading)

The navigator has been implemented without room loading. Search
results return room metadata from the database; the full "enter room"
flow needs the Room realm. See plan/09-NAVIGATOR.md for details.

### Milestone 2.1 — Navigator Init & Settings (DONE)

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `navigator.init` | 2110 | C2S | Init navigator, send 3052 + 3984 + 1543 + 518 |
| 2 | `navigator.metadata` | 3052 | S2C | Search context structure |
| 3 | `navigator.saved_searches` | 3984 | S2C | User saved searches |
| 4 | `navigator.collapsed` | 1543 | S2C | Collapsed category tabs |
| 5 | `navigator.settings` | 518 | S2C | Window layout preferences |
| 6 | `navigator.settings_save` | 3159 | C2S | Persist layout preferences |
| 7 | `navigator.search_save` | 2226 | C2S | Save search bookmark |
| 8 | `navigator.search_delete` | 1954 | C2S | Delete search bookmark |
| 9 | `navigator.search_open` | 637 | C2S | Mark tab expanded |
| 10 | `navigator.search_close` | 1834 | C2S | Mark tab collapsed |
| 11 | `navigator.category_mode` | 1202 | C2S | Set display mode |

### Milestone 2.2 — Navigator Search & Room Info (DONE)

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `navigator.search` | 249 | C2S | Search rooms by category |
| 2 | `navigator.search_result` | 2690 | S2C | Room search results |
| 3 | `navigator.get_room_info` | 2230 | C2S | Detailed room info |
| 4 | `navigator.room_info` | 687 | S2C | Room metadata |
| 5 | `navigator.get_flat_cats` | 3027 | C2S | Room categories |
| 6 | `navigator.flat_cats` | 1562 | S2C | Category list |
| 7 | `navigator.can_create_room` | 2128 | C2S | Room creation check |
| 8 | `navigator.can_create_room` | 378 | S2C | Creation permission |
| 9 | `navigator.create_room` | 2752 | C2S | Create a room |
| 10 | `navigator.room_created` | 1304 | S2C | Room created confirmation |

### Milestone 2.3 — Navigator Favourites (DONE)

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `navigator.favourite_add` | 3817 | C2S | Add favourite room |
| 2 | `navigator.favourite_remove` | 309 | C2S | Remove favourite |
| 3 | `navigator.favourite_changed` | 2524 | S2C | Favourite change notification |
| 4 | `navigator.favourites` | 151 | S2C | Full favourites list |

**Deferred navigator packets** (room search variations, room-dependent):
- All `*_search` variants (my_rooms, friends_rooms, popular, frequent,
  history, guild, highest_score, room_text): require room population data.
- `navigator.go_to_flat` (685), `navigator.visit_user` (2970): require
  room entry flow.
- Competition room packets: require competition system.
- `navigator.lifted` (3104), `navigator.event_categories` (3244): require
  room events.

---

## Phase 3: Groups & Forums (Partial)

Groups can be partially implemented without rooms. Group creation,
membership management, badge design, and forums are non-room-dependent.
Room-based group features (displaying group in room, guild furniture)
are deferred.

### Milestone 3.1 — Group CRUD

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `group.get_create_options` | 798 | C2S | Badge parts + room bases |
| 2 | `group.create_options` | 2159 | S2C | Create options response |
| 3 | `group.buy` | 230 | C2S | Create group |
| 4 | `group.purchased` | 2808 | S2C | Group created |
| 5 | `group.get_info` | 2991 | C2S | Get group info |
| 6 | `group.info` | 1702 | S2C | Group info response |
| 7 | `group.get_settings` | 1004 | C2S | Get editable settings |
| 8 | `group.settings` | 3965 | S2C | Settings response |
| 9 | `group.save_information` | 3137 | C2S | Update name/description |
| 10 | `group.save_colors` | 1764 | C2S | Update badge colors |
| 11 | `group.save_badge` | 1991 | C2S | Update badge design |
| 12 | `group.save_preferences` | 3435 | C2S | Update member prefs |
| 13 | `group.details_changed` | 1459 | S2C | Change notification |
| 14 | `group.delete` | 1134 | C2S | Delete group |
| 15 | `group.deactivated` | 3129 | S2C | Deactivation notification |

### Milestone 3.2 — Group Membership

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `group.get_members` | 312 | C2S | Paginated members list |
| 2 | `group.members` | 1200 | S2C | Members response |
| 3 | `group.request` | 998 | C2S | Join request |
| 4 | `group.accept_request` | 3386 | C2S | Accept request |
| 5 | `group.decline_request` | 1894 | C2S | Decline request |
| 6 | `group.approve_all_requests` | 882 | C2S | Accept all requests |
| 7 | `group.remove_member` | 593 | C2S | Remove member |
| 8 | `group.remove_member_confirm` | 3593 | C2S | Confirm removal |
| 9 | `group.member_remove_confirm` | 1876 | S2C | Furniture impact |
| 10 | `group.admin_add` | 2894 | C2S | Promote to admin |
| 11 | `group.admin_remove` | 722 | C2S | Demote admin |
| 12 | `group.unblock_member` | 2864 | C2S | Unblock member |
| 13 | `group.favorite` | 3549 | C2S | Set favourite group |
| 14 | `group.unfavorite` | 1820 | C2S | Remove favourite |
| 15 | `group.favorite_update` | 3403 | S2C | Favourite changed |
| 16 | `group.badges` | 2402 | S2C | Group badge codes |
| 17 | `group.get_badge_parts` | 813 | C2S | Badge part catalog |
| 18 | `group.badge_parts` | 2238 | S2C | Badge parts response |
| 19 | `group.list` | 420 | S2C | Group list for user |
| 20 | `group.members_refresh` | 2445 | S2C | Members refresh signal |

### Milestone 3.3 — Forums (Partial)

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `forum.get_list` | 436 | C2S | Forum list |
| 2 | `forum.list` | 3001 | S2C | Forum list response |
| 3 | `forum.get_stats` | 3149 | C2S | Forum statistics |
| 4 | `forum.stats` | 3011 | S2C | Forum stats response |
| 5 | `forum.get_threads` | 873 | C2S | Thread list |
| 6 | `forum.threads` | 1073 | S2C | Threads response |
| 7 | `forum.get_messages` | 232 | C2S | Message list |
| 8 | `forum.messages` | 509 | S2C | Messages response |
| 9 | `forum.post_message` | 3529 | C2S | Post a message |
| 10 | `forum.post` | 2049 | S2C | Post response |
| 11 | `forum.moderate_message` | 286 | C2S | Moderate message |
| 12 | `forum.moderate_thread` | 1397 | C2S | Moderate thread |
| 13 | `forum.update_settings` | 2214 | C2S | Update forum perms |
| 14 | `forum.get_unread_count` | 2908 | C2S | Unread count |
| 15 | `forum.unread_count` | 2379 | S2C | Unread response |

---

## Phase 4: Achievements & Notifications

### Milestone 4.1 — Achievement System

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `user.achievement_list` | 219 | C2S | Request achievements |
| 2 | `user.achievement_list` | 305 | S2C | Achievement list |
| 3 | `user.achievement_progressed` | 2107 | S2C | Progress notification |
| 4 | `user.achievement_notification` | 806 | S2C | Achievement unlocked |
| 5 | `user.user_achievement_score` | 1968 | S2C | Total score |
| 6 | `user.get_badge_point_limits` | 1371 | C2S | Badge point limits |
| 7 | `user.badge_point_limits` | 2501 | S2C | Limits response |
| 8 | `user.request_badge` | 3077 | C2S | Request promo badge |
| 9 | `user.badge_request_fulfilled` | 2998 | S2C | Badge request result |
| 10 | `user.check_badge_request` | 1364 | C2S | Check badge fulfilled |

### Milestone 4.2 — Notifications & Landing

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `subscription.promo_articles` | 1827 | C2S | Get promo articles |
| 2 | `subscription.promo_articles` | 286 | S2C | Promo articles response |
| 3 | `user.activity_point_notification` | 2275 | S2C | Currency change notification |
| 4 | `messenger.friend_notification` | 3082 | S2C | Friend state notification |
| 5 | `catalog.club_gift_notification` | 2188 | S2C | Club gift available |
| 6 | `session.notification_list` | 1992 | S2C | Notification list |

---

## Phase 5: Room Realm (Major Feature — Future)

The Room realm is the largest unimplemented feature. It unlocks trading,
navigator room entry, room entities, furniture interactions, pets, and
most moderation tools. **Estimated packet count: ~250+.**

### Sub-phases:
1. **Room lifecycle** — create, enter, exit, delete, settings
2. **Room model** — heightmap, blocked tiles, floor plan editor
3. **Room entities** — avatars, units, walking, chat, effects, typing
4. **Room rights** — owner, rights list, kick, ban, mute
5. **Furniture interactions** — place, pickup, move, wired, dice, dimmer
6. **Pets** — place, info, commands, breeding (all require room entities)
7. **Trading** — open, item management, confirm flow (requires room)
8. **Room search integration** — navigator search with room population

---

## Phase 6: Moderation & Safety (Partial)

Some moderation tools can work without rooms (user info lookup, sanction
management). Most require room context.

### Milestone 6.1 — Core Moderation (No Room Required)

| # | Packet | ID | Dir | Task |
|---|--------|----|-----|------|
| 1 | `moderation.mod_tool_user_info` | 3295 | C2S | User info lookup |
| 2 | `moderation.moderation_user_info` | 2866 | S2C | User info response |
| 3 | `moderation.moderation_tool` | 2696 | S2C | Tool initialization |
| 4 | `moderation.call_for_help` | 1691 | C2S | Submit CFH report |
| 5 | `moderation.cfh_topics` | 325 | S2C | CFH topic list |
| 6 | `moderation.get_cfh_status` | 2746 | C2S | CFH status check |
| 7 | `moderation.cfh_result_message` | 3635 | S2C | CFH result |
| 8 | `moderation.moderator_message` | 2030 | S2C | Admin broadcast |
| 9 | `moderation.moderator_action_result` | 2335 | S2C | Mod action result |
| 10 | `moderation.user_sanction_status` | 3679 | S2C | Sanction status |

---

## Dependency Graph

```
Phase 1 (Economy)         Phase 2 (Navigator)        Phase 3 (Groups)
  │                         │                           │
  ├── 1.1 Catalog ✓ (now)   ├── 2.1 Nav Init (now)     ├── 3.1 Group CRUD (now)
  ├── 1.2 Inventory (now)   ├── 2.2 Search (now)       ├── 3.2 Membership (now)
  ├── 1.3 Marketplace (now) ├── 2.3 Favourites (now)   └── 3.3 Forums (now)
  ├── 1.4 Subscription      │
  └── 1.5 Trading ──────────┘──────> Phase 5 (Room) ←──── Pets, Furniture
                                        │                  Interactions
                                        ▼
                              Phase 6 (Moderation Full)
                              Phase 7 (Games, Camera, Crafting)
```

---

## Performance Bottleneck Analysis

### 1. Furniture Definition Cache
**Problem:** Loading ~5000+ item definitions on every request.
**Solution:** Load once at startup into in-memory map. Invalidate only on
admin API changes. Already implemented in furniture application service.

### 2. Inventory Pagination
**Problem:** Users with 10,000+ items cause large packet payloads.
**Solution:** Fragment the `user.furniture` (994) response. Already
implemented — `FurniListPacket` supports `TotalFragments`/`FragmentIndex`.
Keep fragment size ≤ 500 items.

### 3. Marketplace Search
**Problem:** Full-text search across all active offers.
**Solution:** PostgreSQL `tsvector` column with GIN index on item names.
Cache popular searches in Redis with 60s TTL. Paginate results (max 100
per page). Background goroutine expires offers.

### 4. Badge Inventory
**Problem:** Users with 500+ badges require large payloads.
**Solution:** Client expects all badges in one packet (717). No pagination
in protocol. Use in-memory cache keyed by user ID, invalidated on
badge add/remove. Consider gzip compression at WebSocket frame level.

### 5. Navigator Search Categories
**Problem:** Multiple search tabs send parallel requests.
**Solution:** Each search category is a separate C2S/S2C pair. Server
processes concurrently via goroutines. Cache room counts in Redis
(30s TTL). Use database materialized views for "popular rooms" ranking.

### 6. Group Forums
**Problem:** Thread pagination with message counts.
**Solution:** Use PostgreSQL window functions for thread list with message
counts. Index on `(group_id, created_at)` for thread ordering. Paginate
both threads and messages (20 per page).

---

## Implementation Priority Order

1. **Phase 1.1** — Catalog completion (vouchers, giftable check)
2. **Phase 1.2** — Inventory (badges, effects, unseen tracking)
3. **Phase 1.3** — Marketplace (fill stubs with real logic)
4. **Phase 1.4** — Subscription (club offers, gifts, kickback)
5. **Phase 2.1** — Navigator init and settings
6. **Phase 2.2** — Navigator search (requires rooms table)
7. **Phase 3.1** — Group CRUD
8. **Phase 3.2** — Group membership
9. **Phase 4.1** — Achievements
10. **Phase 4.2** — Notifications
11. **Phase 3.3** — Forums
12. **Phase 6.1** — Core moderation
13. **Phase 5** — Room realm (unlocks everything else)

---

## Packets Explicitly Deferred

### Requires Room Realm (250+ packets across):
- Room: 46 C2S + 44 S2C = 90 packets
- Room Entities: 14 C2S + 20 S2C = 34 packets
- Furniture interactions: ~42 C2S + ~37 S2C ≈ 79 packets
- Pets: 21 C2S + 20 S2C = 41 packets
- Trading: 9 C2S + 11 S2C = 20 packets

### System Not Planned:
- Games & Entertainment: 49 packets (third-party game integration)
- Camera & Photos: 15 packets (image storage system)
- Crafting & Recycling: 16 packets (item crafting engine)
- Other: 10 packets (phone verification, FAQ)
- Quests & Campaigns: 33 packets (quest engine)

### Feature-Gated:
- Targeted Offers: ~6 packets (targeting engine)
- Campaign Calendar: ~6 packets (event calendar)
- Builders Club: ~4 packets (legacy feature)
- NUX Gifts: ~2 packets (new user flow)
