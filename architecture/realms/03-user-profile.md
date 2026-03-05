# Realm: User & Profile

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 30 | **Phase:** 2 (Identity) | **Packets:** 56 (29 c2s, 27 s2c)
> **Services:** game (identity module) | **Status:** 16 handlers implemented

---

## Overview

The User & Profile realm is the identity backbone of the entire system. It covers user data after authentication: figure (appearance), motto, credits, activity points, subscription status, settings, wardrobe, ignore lists, badges, relationships, and name changes. This realm's data is required by virtually every other realm -- room entities need figures, messenger needs usernames, trading needs credit balances.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 2

---

## Packet Inventory

### C2S (Client to Server) -- 29 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 357 | `user.get_info` | _(none)_ | Request own user data (login bundle) |
| 3265 | `user.get_profile` | `userId:int32`, `openProfile:boolean` | View another user's profile card |
| 2249 | `user.get_profile_by_name` | `username:string` | Look up profile by username |
| 2138 | `user.get_relationship_status` | `userId:int32` | Get relationship info for a user |
| 3768 | `user.set_relationship_status` | `userId:int32`, `status:int32` | Set relationship type (heart/smile/skull) |
| 17 | `user.get_tags` | `userId:int32` | Get user tags |
| 2730 | `user.update_figure` | `figure:string`, `gender:string` | Change appearance |
| 2228 | `user.update_motto` | `motto:string` | Change motto |
| 2694 | `user.respect` | `userId:int32` | Give respect to another user |
| 3878 | `user.get_ignored` | _(none)_ | Request ignore list |
| 1117 | `user.ignore` | `username:string` | Ignore a user by name |
| 3314 | `user.ignore_id` | `userId:int32` | Ignore a user by ID |
| 2061 | `user.unignore` | `userId:int32` | Remove user from ignore list |
| 1367 | `user.settings_volume` | `systemVolume:int32`, `furniVolume:int32`, `traxVolume:int32` | Update volume settings |
| 1262 | `user.settings_old_chat` | `oldChat:boolean` | Toggle old-style chat |
| 1086 | `user.settings_room_invites` | `blockInvites:boolean` | Toggle room invite blocking |
| 1752 | `user.effect_enable` | `effectId:int32` | Enable an avatar effect |
| 2742 | `user.get_wardrobe` | _(none)_ | Request saved outfits |
| 800 | `user.save_wardrobe_outfit` | `slotId:int32`, `look:string`, `gender:string` | Save outfit to wardrobe slot |
| 2977 | `user.change_name` | `name:string` | Request username change |
| 3950 | `user.check_name` | `name:string` | Check if username is available |
| 2109 | `user.approve_name` | `name:string` | Confirm name change |
| 1740 | `user.set_home_room` | `roomId:int32` | Set home room |
| 2557 | `user.get_email_status` | _(none)_ | Check email verification status |
| 1904 | `user.get_club_offers` | _(none)_ | Request HC/VIP subscription offers |
| 2661 | `user.get_habbo_club_center_info` | _(none)_ | Request club center dashboard |
| 3285 | `user.get_sanction_status` | _(none)_ | Check moderation sanction status |
| 1671 | `user.get_sound_settings` | _(none)_ | Request sound/music settings |
| 3608 | `user.nux_completed` | _(none)_ | Mark new-user experience as completed |

### S2C (Server to Client) -- 27 packets

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 2725 | `user.object` | `userId`, `username`, `figure`, `gender`, `motto`, `credits`, etc. | Full user data composite |
| 3898 | `user.perks` | perk list | User permission perks |
| 2442 | `user.permissions` | `clubLevel`, `securityLevel`, `ambassador` | Permission levels |
| 3579 | `user.profile` | `userId`, `username`, `figure`, `motto`, `createdDate`, `achievementScore`, `friendCount`, `isOnline`, `isMyFriend`, `isRequestSent`, `groups[]` | Public profile data |
| 2016 | `user.figure_update` | `figure:string`, `gender:string` | Figure changed broadcast |
| 1712 | `user.change_name_update` | `userId:int32`, `newName:string`, `oldName:string` | Name change broadcast |
| 2275 | `user.credits` | `credits:string` | Credit balance update |
| 2018 | `user.currency` | currency array | Activity points by type |
| 1290 | `user.subscription` | `productName`, `daysRemaining`, etc. | Subscription status |
| 2773 | `user.wardrobe` | `outfits[]` | Wardrobe contents |
| 126 | `user.ignored_list` | `usernames[]` | Current ignore list |
| 3006 | `user.ignore_result` | `result:int32` | Ignore operation result |
| 3189 | `user.respect_received` | `userId`, `respectsReceived`, `respectsRemaining` | Respect given notification |
| 3192 | `user.home_room` | `roomId:int32` | Home room set confirmation |
| 3480 | `user.check_name_result` | `resultCode:int32`, `name:string`, `suggestions[]` | Name availability check result |
| 195 | `user.effect_list` | effects array | Available avatar effects |
| 1959 | `user.effect_activated` | `effectId:int32`, `duration:int32` | Effect activated confirmation |
| 3473 | `user.effect_selected` | `effectId:int32` | Effect selected |
| 1889 | `user.email_status` | `email:string`, `verified:boolean` | Email verification status |
| 3554 | `user.tags` | `userId`, `tags[]` | User tags response |
| 2016 | `user.figure_update` | `figure`, `gender` | Appearance changed |
| 3874 | `user.relationship_status` | `userId`, `relationships[]` | Relationship data |
| 2930 | `user.club_center_info` | multiple fields | HC center dashboard data |
| 1290 | `user.habbo_club_subscription` | subscription fields | Subscription details |
| 2285 | `user.sanction_status` | sanction fields | Current moderation sanctions |
| 3738 | `user.nux_status` | `completed:boolean` | NUX completion state |
| 2773 | `user.wardrobe_update` | wardrobe data | Wardrobe save confirmation |

---

## Architecture Mapping

### Service Ownership

The `game` service's `internal/identity` module owns this realm. After authentication, the game service:
1. Receives `session.authenticated` event via NATS.
2. Loads user data from PostgreSQL via a domain-owned repository built on `pkg/storage/postgres` primitives.
3. Builds a "login bundle" of S2C packets: `user.object`, `user.perks`, `user.permissions`, `user.credits`, `user.currency`, `user.subscription`, `user.wardrobe`, `user.ignored_list`, `user.effect_list`.
4. Publishes the bundle to `session.output.<sessionID>` via NATS.

### Database Tables

| Table | Columns (Key) | Usage |
|-------|---------------|-------|
| `users` | id, username, figure, gender, motto, credits, pixels, points, rank, home_room | Core user data |
| `user_settings` | user_id, volume_*, old_chat, block_invites, chat_color, focus_preference | Client preferences |
| `user_wardrobe` | user_id, slot_id, look, gender | Saved outfits (max 10 slots) |
| `user_badges` | user_id, code, slot, acquired | Badge collection + equipped slots |
| `user_ignores` | user_id, ignored_user_id | Ignore list |
| `user_effects` | user_id, effect_id, duration, activated | Avatar effects |
| `user_relationships` | user_id, target_user_id, type | Heart/smile/skull relationships |
| `user_tags` | user_id, tag | User-defined tags |
| `server_permissions_ranks` | rank_id, permissions | Rank-based permission mapping |
| `permission_perks` | rank_id, perk_name | Per-rank feature toggles |

### NATS Subjects

| Subject | Direction | Purpose |
|---------|-----------|---------|
| `session.authenticated` | auth -> game | Triggers login bundle build |
| `session.output.<sessionID>` | game -> gateway | Delivers login bundle + all responses |

---

## Implementation Analysis

### Login Bundle Construction

The login bundle is the most critical part of Phase 2. It must be assembled atomically and efficiently:

```go
func (s *Service) BuildLoginBundle(ctx context.Context, userID int32, rank int32) ([]*codec.Frame, error) {
    // Parallel data loading (fan-out)
    user, settings, badges, wardrobe, ignores, effects := parallelLoad(ctx, userID)

    // Build packet sequence
    frames := []*codec.Frame{
        encode(UserObjectOut{...user fields...}),
        encode(UserPerksOut{...perks from rank...}),
        encode(UserPermissionsOut{clubLevel, securityLevel, ambassador}),
        encode(UserCreditsOut{credits}),
        encode(UserCurrencyOut{pixels, points}),
        encode(UserSubscriptionOut{...}),
        encode(UserEffectListOut{effects}),
        encode(UserIgnoredListOut{ignores}),
        encode(UserWardrobeOut{wardrobe}),
        encode(SessionFirstLoginOfDayOut{}) // if applicable
    }
    return frames, nil
}
```

**Performance consideration:** The login bundle loads from 6+ tables. Use parallel queries (goroutines with errgroup) to minimize latency. Target: < 50ms total.

### Permission System

The permission system maps user ranks to capabilities. Based on reference emulators (Comet v2, Arcturus):

```
Rank 1: Normal user
Rank 2: VIP user (HC subscriber)
Rank 3: Staff (basic moderation)
Rank 4: Senior Staff
Rank 5: Administrator
Rank 6: Developer
```

Each rank has associated perks (`server_permissions_ranks`) and commands (`permission_commands`). The `user.permissions` (2442) packet sends:
- `clubLevel` -- 0 (none), 1 (HC), 2 (VIP)
- `securityLevel` -- numeric rank
- `ambassador` -- boolean flag

### Figure Validation

`user.update_figure` (2730) must validate the figure string format before saving. Habbo figures follow the pattern:
```
hr-{hairId}-{colorId}.hd-{headId}-{colorId}.ch-{chestId}-{colorId}...
```

Validation requirements:
1. String must match the figure part pattern.
2. Each part ID must exist in the furniture definitions (loaded from `figuredata.xml`).
3. Gender must be `"M"` or `"F"`.
4. HC-exclusive items require active subscription.

**Caveat from Arcturus:** Figure validation is commonly bypassed in emulators, allowing invalid figures that crash client rendering. pixel-server should enforce validation strictly.

### Name Change Flow

The name change is a three-step process:
1. `user.check_name` (3950) -- Client sends desired name, server validates and returns suggestions if taken.
2. `user.change_name` (2977) -- Client submits chosen name.
3. `user.approve_name` (2109) -- Client confirms the change.

**Validation rules:**
- Length: 3-15 characters
- Characters: alphanumeric + limited special chars
- Not taken by another user
- Not on the word filter blacklist
- Cooldown: one name change per 30 days

**Edge case from Comet v2:** Name changes must update all references (room owner names, group member names, friend list entries). This is the most expensive part -- use NATS events to notify social and navigator services.

### Respect System

`user.respect` (2694) gives respect points. Constraints:
- Each user gets 3 respect points per day to give.
- Cannot respect yourself.
- Cannot respect the same user more than once per day.
- Respect count is tracked on the receiver's profile.

Track daily respect allocation in Redis with TTL at midnight reset.

---

## Caveats & Edge Cases

### 1. Login Bundle Ordering
The Nitro client expects packets in a specific order. `user.object` must arrive before `user.permissions`. `user.credits` must arrive before the navigator opens. Test the exact ordering against the Nitro client.

### 2. Currency Overflow
Credits and activity points are `int32`. Maximum is 2,147,483,647. Never allow operations that would overflow. Check before adding credits from catalog purchases, trades, or admin commands.

### 3. Wardrobe Slot Limits
Standard users get 5 wardrobe slots; HC users get 10. The `user.save_wardrobe_outfit` handler must enforce this limit based on subscription status.

### 4. Ignore List Size
No hard limit in the protocol, but practical limit of ~100 entries. Loading a massive ignore list on every login is expensive. Consider pagination or lazy loading.

### 5. Relationship Types
Relationships use integer types:
- 0: None
- 1: Heart (love)
- 2: Smile (like)
- 3: Skull (enemy)

Only one relationship per target user. Setting a new type replaces the old one. Setting type 0 removes the relationship.

### 6. Settings Persistence Race
If a user rapidly toggles settings (e.g., old chat on/off), multiple `user.settings_old_chat` packets arrive in quick succession. Use last-write-wins semantics. Debounce database writes to avoid excessive UPDATE queries.

### 7. Profile Privacy
`user.get_profile` (3265) returns public data. However, some fields should be hidden based on the requester's relationship:
- Friend count: visible to all
- Online status: visible to all
- Last login: visible to friends only (configurable)
- Groups: visible to all

### 8. NUX (New User Experience)
`user.nux_completed` (3608) marks the tutorial as done. This flag must persist to avoid showing the NUX on every login. Store as boolean in `user_settings`.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Login bundle** | Sequential DB queries (200-500ms) | Parallel fan-out queries (< 50ms target) |
| **Figure validation** | Skipped or minimal | Full validation against figuredata definitions |
| **Name changes** | Direct DB update, stale references | Event-driven propagation to all services |
| **Permission mapping** | Hardcoded rank checks | Config-driven rank/perk tables |
| **Respect tracking** | DB counter per click | Redis daily allocation with TTL |
| **Settings persistence** | Immediate DB write per change | Debounced batch writes |
| **Profile loading** | Full user row every time | Read-through Redis cache with invalidation |

---

## Dependencies

- **Phase 1 (Connection)** -- user must be authenticated before identity packets are meaningful
- **pkg/user** -- domain models (User, Settings, Badge, WardrobeOutfit, IgnoredUser)
- **pkg/user/memory** -- in-memory fakes for unit testing
- **PostgreSQL** -- user data, settings, wardrobe, badges, ignores, relationships
- **Redis** -- respect daily allocation, profile cache, online status

---

## Testing Strategy

### Unit Tests
- Login bundle builder (mock repositories, verify packet sequence)
- Figure string validation (valid/invalid cases)
- Name change validation (length, characters, blacklist)
- Respect daily limit enforcement
- Permission mapping from rank to packet fields

### Integration Tests
- Full login bundle against real PostgreSQL (testcontainers)
- Figure update persists and loads correctly
- Name change updates all references
- Wardrobe save/load round-trip
- Ignore list CRUD operations

### E2E Tests
- Client logs in and receives correct figure, motto, credits in hotel view
- Client changes figure and another client in the same room sees the update
- Name change flow completes and name appears correctly everywhere
