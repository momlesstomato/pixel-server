# Realm: Achievements & Talents

> **Position:** 150 | **Phase:** 13 (Remaining) | **Packets:** 24 (10 c2s, 14 s2c)
> **Services:** game | **Status:** Not yet implemented

---

## Overview

Achievements & Talents tracks player progress through achievement milestones, talent tracks (quest-like progression paths), badge point limits, and game achievements. This is a system that hooks into virtually every other realm: rooms, furniture, social, games, and catalog all trigger achievement progress events.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 10 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 219 | `achievement.list` | _(none)_ | Request all achievements with progress |
| 1371 | `achievement.get_badge_point_limits` | _(none)_ | Get badge point thresholds |
| 3077 | `achievement.request_badge` | `requestCode:string` | Claim promotional badge |
| 1364 | `achievement.check_badge_request` | `requestCode:string` | Check if badge code is valid |
| 359 | `achievement.resolution_open` | _(none)_ | Open achievement resolution UI |
| 2127 | `talent.get_level` | _(none)_ | Get talent track progress |
| 3144 | `achievement.reset_resolution` | _(none)_ | Reset achievement resolution |
| 196 | `talent.helper_track` | _(none)_ | Get helper talent track |
| 389 | `games.get_user_achievements` | _(none)_ | Get game-specific achievements |
| 2399 | `games.get_achievements` | _(none)_ | Get all game achievements |

### S2C -- 14 packets

| ID | Name | Summary |
|----|------|---------|
| 305 | `achievement.list` | Full achievement list with progress |
| 2501 | `achievement.badge_point_limits` | Point thresholds per badge level |
| 2998 | `achievement.badge_request_fulfilled` | Promotional badge claim result |
| 66 | `achievement.resolutions` | Achievement resolution data |
| 740 | `achievement.resolution_completed` | Resolution completed notification |
| 638 | `talent.level_up` | Talent track level advancement |
| 2107 | `talent.track` | Talent track full data |
| 1878 | `achievement.notification` | Achievement progress notification popup |
| 1797 | `achievement.unlocked` | Achievement fully unlocked |
| 3743 | `achievement.score_updated` | Achievement score changed |
| -- | `games.achievements` | Game achievement list |
| -- | `games.user_achievements` | User's game achievement progress |
| -- | Additional achievement-related response packets |

---

## Implementation Analysis

### Achievement System Architecture

```
Event-Driven Achievement Tracking:
  Every game action publishes an achievement event:
    "achievement.progress" {
      userId:    int32
      category:  string    // e.g., "RoomEntry", "ChatSent", "ItemPlaced"
      increment: int32     // how much to add
    }

Achievement Engine (background worker):
  1. Consume achievement.progress events
  2. Load user's current progress for category
  3. Increment progress counter
  4. Check against thresholds for each level
  5. If threshold reached:
     a. Unlock next level badge
     b. Add achievement score
     c. Send achievement.notification to user
     d. Send achievement.unlocked if fully completed
```

### Achievement Categories (from reference emulators)

| Category | Trigger | Levels |
|----------|---------|--------|
| RoomEntry | Enter rooms | 1, 5, 25, 100, 500 |
| ChatSent | Send chat messages | 1, 50, 250, 1000, 5000 |
| FriendsMade | Accept friend requests | 1, 5, 20, 50, 100 |
| ItemsPlaced | Place furniture | 1, 10, 100, 500 |
| TradingDone | Complete trades | 1, 10, 50 |
| RespectGiven | Give respect | 1, 10, 100, 500 |
| PetsTrained | Train pets | 1, 10, 50, 100 |
| ForumPosts | Post in forums | 1, 10, 50, 200 |
| GamesWon | Win mini-games | 1, 5, 25, 100 |

### Talent Tracks

Talent tracks are guided progression paths (like quest chains):

```
Helper Track:
  Level 1: Complete tutorial → Reward: 500 credits
  Level 2: Make 5 friends → Reward: Badge "Helper L2"
  Level 3: Enter 10 rooms → Reward: Effect "Star"
  Level 4: Place 5 items → Reward: Badge "Helper L4"

Citizen Track:
  Level 1: Change figure → Reward: 200 credits
  Level 2: Set motto → Reward: Badge "Citizen L2"
  Level 3: Create room → Reward: 500 credits
```

Talent tracks are database-configurable with ordered levels, requirements, and rewards.

### Database Tables

| Table | Usage |
|-------|-------|
| `achievements` | id, category, name, levels[](threshold, badge_code, points) | Achievement definitions |
| `user_achievements` | user_id, achievement_id, progress, level, updated_at | User progress |
| `talent_tracks` | id, name, levels[](requirements, rewards) | Talent track definitions |
| `user_talent_progress` | user_id, track_id, level, completed | User talent progress |

---

## Caveats & Edge Cases

### 1. Event Deduplication
Achievement progress events can fire multiple times for the same action (e.g., NATS at-least-once delivery). Progress incrementing must be idempotent or use deduplication keys.

### 2. Achievement Score Consistency
The total achievement score displayed on profiles is a denormalized sum. It must be recalculated from individual achievement levels, not incrementally (to self-heal from bugs).

### 3. Retroactive Achievement Credit
When new achievements are added, existing users should receive credit for past actions if the system tracks historical counts. This requires careful backfill logic.

### 4. Badge Point Limits
Some badges require a minimum achievement score to earn. The `achievement.get_badge_point_limits` response defines these thresholds.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Tracking** | Direct DB update in handler | Event-driven async worker |
| **Configuration** | Hardcoded categories/thresholds | Database-configurable |
| **Talent tracks** | Not implemented in most emulators | Full implementation with rewards |
| **Score consistency** | Incremental (drift risk) | Recalculated from source |

---

## Dependencies

- All previous phases (achievements track progress across every feature)
- **PostgreSQL** -- achievement definitions and progress

---

## Testing Strategy

### Unit Tests
- Progress threshold checking
- Achievement score calculation
- Talent track level advancement

### Integration Tests
- Achievement event processing end-to-end
- Badge award on threshold reach
- Talent track completion with rewards
