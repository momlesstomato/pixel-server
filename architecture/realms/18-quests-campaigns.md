# Realm: Quests & Campaigns

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 160 | **Phase:** 13 (Remaining) | **Packets:** 33 (15 c2s, 18 s2c)
> **Services:** game | **Status:** Not yet implemented

---

## Overview

Quests & Campaigns manages time-limited challenges, seasonal events, community competitions, and quest chains. Unlike achievements (permanent milestones), quests are time-bound and rotate on a schedule. Campaigns are larger seasonal events with themed rewards.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 15 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `quest.list` | _(none)_ | Request available quests |
| -- | `quest.activate` | `questId:int32` | Start tracking a quest |
| -- | `quest.cancel` | `questId:int32` | Abandon active quest |
| -- | `quest.get_daily` | _(none)_ | Get daily quests |
| -- | `quest.complete` | `questId:int32` | Submit quest completion |
| -- | `campaign.list` | _(none)_ | Request active campaigns |
| -- | `campaign.get_progress` | `campaignId:string` | Get progress in campaign |
| -- | `campaign.collect_reward` | `campaignId`, `rewardId` | Claim campaign reward |
| -- | `competition.get_info` | `competitionId:string` | Get competition details |
| -- | `competition.submit` | `competitionId`, `roomId` | Submit room for competition |
| -- | `competition.vote` | `competitionId`, `roomId`, `score` | Vote on competition entry |
| -- | Additional quest/campaign query packets |

### S2C -- 18 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `quest.list` | Available quests with progress |
| -- | `quest.activated` | Quest tracking started |
| -- | `quest.completed` | Quest completion confirmation + reward |
| -- | `quest.progress_update` | Progress increment notification |
| -- | `quest.daily_available` | New daily quests available |
| -- | `campaign.data` | Campaign details with objectives |
| -- | `campaign.progress` | Campaign progress update |
| -- | `campaign.reward_claimed` | Reward claim result |
| -- | `competition.info` | Competition details |
| -- | `competition.results` | Competition voting results |
| -- | Additional notification/status packets |

---

## Implementation Analysis

### Quest Types

| Type | Trigger | Example |
|------|---------|---------|
| RoomVisit | Enter X unique rooms | "Visit 5 different rooms" |
| ChatSend | Send X messages | "Say hello to 3 people" |
| FurniPlace | Place X items | "Decorate your room" |
| FriendMake | Add X friends | "Make a new friend" |
| TradeComplete | Complete X trades | "Complete a trade" |
| GamePlay | Play X game rounds | "Play Freeze" |
| PetCare | Feed/train pet X times | "Take care of your pet" |

### Quest Lifecycle

```
1. Quests rotate on a schedule (daily/weekly)
2. User can track one quest at a time
3. Progress tracked via same event system as achievements
4. On completion: award credits/badges/items, mark as done
5. Completed quests cannot be repeated until next rotation
```

### Campaign System

Campaigns are seasonal events (e.g., Winter Campaign, Summer Campaign):
- Time-limited (2-4 weeks)
- Multiple objectives forming a checklist
- Progressive rewards at milestones
- Some objectives are daily (must be done each day)
- Leaderboards for top participants

### Competition System

Room competitions allow users to submit rooms for community voting:
- Admin creates competition with theme and dates
- Users submit rooms during submission period
- Voting period: users rate rooms 1-5 stars
- Results published at end, winners receive prizes

---

## Caveats & Edge Cases

### 1. Quest Progress Persistence
Quest progress must survive server restarts. Store in PostgreSQL, not just memory.

### 2. Daily Reset Timing
Daily quests reset at midnight server time (UTC). Users in different timezones may perceive unfair timing. Document the reset time clearly.

### 3. Campaign Expiry
When a campaign expires, unclaimed rewards must still be claimable for a grace period (7 days). After that, rewards are lost.

### 4. Competition Voting Fraud
Users could vote-bomb competitors. Mitigate: one vote per user per room, cannot vote on own submission.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Quest rotation** | Manual admin action | Scheduled rotation via cron config |
| **Progress tracking** | Direct DB in handler | Event-driven (shared with achievements) |
| **Campaigns** | Not implemented in most emulators | Full seasonal campaign system |
| **Competitions** | Basic or missing | Room submission + voting + leaderboard |

---

## Dependencies

- All gameplay realms (quests track progress across features)
- **Phase 3 (Room)** for competitions
- **Achievement event system** for progress tracking

---

## Testing Strategy

### Unit Tests
- Quest progress increment and completion check
- Campaign milestone calculation
- Competition voting aggregation

### Integration Tests
- Quest rotation with time simulation
- Campaign reward claim and expiry

### E2E Tests
- Client activates quest, performs action, sees progress, completes quest
