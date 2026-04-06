# 12 - Moderation System Breakdown

## Overview

This document analyzes the moderation systems across all four vendor
emulators (PlusEMU, Arcturus, comet-v2, and the Nitro renderer) to
inform future moderation implementation in pixel-server. Moderation is
a cross-cutting concern that touches rooms, chat, users, trading, and
the global session lifecycle.

**This document is reference-only.** Implementation of the moderation
system is deferred to a dedicated milestone after room completion.

---

## Vendor Architecture Comparison

### Core Moderation Components

| Component | PlusEMU (C#) | Arcturus (Java) | comet-v2 (Java) |
|-----------|-------------|-----------------|-----------------|
| Entry point | `ModerationManager` singleton | `ModToolManager` | Packet event handlers |
| Ticket system | `ModerationTicket` in-memory dict | `ModToolIssue` | `IHelpTicket` with state machine |
| Ban storage | `Dictionary<string, ModerationBan>` cached | DB-driven + config | DB-driven with expiry |
| Chat evidence | `reportedChats` list per ticket | `ModToolChatLog` per room | `chat_messages` text in ticket |
| Permission model | Manager-level rank check | `acc_supporttool` + granular `cmd_*` | `mod_tool` minimum rank |
| Sanction escalation | Tracked via `user_info` counters | Probation timestamps + auto-escalation | Preset action templates |
| Cache strategy | In-memory with RCON reload | Server-start load + config-driven | Database queries per action |

---

## 1. Ban System

### Ban Types Across Vendors

| Ban Type | PlusEMU | Arcturus | comet-v2 | pixelsv (proposed) |
|----------|---------|----------|----------|-------------------|
| Account ban | `bantype=user` | `type=account` | `type=user` | **Account ban** |
| IP ban | `bantype=ip` | `type=ip` | `type=ip` | **IP ban** |
| Machine ban | `bantype=machine` | `type=machine` | `type=machine` | **Machine ban** |
| Super ban (all) | — | `type=super` (account+IP+machine) | — | **Super ban** |
| Room ban | Room-level `room_bans` | Room-level `room_bans` | Room state columns | **Existing `room_bans` table** |

### Ban Duration Models

| Vendor | Duration Approach |
|--------|-------------------|
| PlusEMU | Unix timestamp expiry (`expire` column, 0 = permanent) |
| Arcturus | Preset durations: 18 hours, 7 days, 30 days, 100 years (permanent) |
| comet-v2 | Flexible hours/days with `expire` timestamp |
| pixelsv | **Expiry timestamp (NULL = permanent), same as room_bans pattern** |

### Ban Database Schema

| Column | PlusEMU | Arcturus | comet-v2 | pixelsv (proposed) |
|--------|---------|----------|----------|-------------------|
| id | INT PK | INT PK | INT PK | BIGSERIAL PK |
| type | ENUM(user,ip,machine) | ENUM(account,ip,machine,super) | ENUM(user,ip,machine) | VARCHAR(20) NOT NULL |
| target | `value` VARCHAR | user_id + ip + machine_id cols | `data` VARCHAR | `target_value` VARCHAR |
| user_id | — | INT | — | INT (nullable) |
| reason | TEXT | TEXT | TEXT | TEXT NOT NULL |
| issuer_id | — | INT | `added_by` INT | INT NOT NULL |
| expires_at | `expire` UNIX INT | TIMESTAMP | `expire` TIMESTAMP | TIMESTAMPTZ (NULL = perm) |
| created_at | — | `timestamp` | — | TIMESTAMPTZ DEFAULT NOW() |

---

## 2. Mute System

### Mute Scope Comparison

| Scope | PlusEMU | Arcturus | comet-v2 |
|-------|---------|----------|----------|
| Room-wide mute | Room owner toggle → all chat blocked | `cmd_roommute` permission | `RoomMuteState` enum (NONE/RIGHTS) |
| Per-user room mute | — | `room_mutes` table with duration | Room-level mute tracking |
| Global user mute | `ModerationMuteEvent` | `users_settings.mute_end_timestamp` | `MuteUserMessageEvent` |
| Word-triggered mute | — | `wordfilter.mute` column (seconds) | — |
| Flood mute | Automatic (configurable) | Automatic + progressive | Automatic (750ms cooldown) |

### Mute Duration Models

| Vendor | Approach |
|--------|----------|
| PlusEMU | Mute events set timestamp, checked per-message |
| Arcturus | `mute_end_timestamp` on user settings, word filter adds seconds |
| comet-v2 | `MuteUserMessageEvent` with timed duration |
| pixelsv | **Existing flood control (3/3s → 10s). Global mute deferred.** |

---

## 3. Kick System

### Kick Types

| Type | PlusEMU | Arcturus | comet-v2 |
|------|---------|----------|----------|
| Room kick (owner) | `KickUser` packet (1320) | `RoomUserKickEvent` | `KickUserMessageEvent` (3838) |
| Room kick (mod) | `ModerationKickEvent` (2582) | `ModToolKickEvent` | Higher rank override |
| Hotel kick | Forced disconnect | `ModToolKickEvent` from hotel | Session termination |
| Kick all | — | `cmd_kickall` permission | — |

### Kick Protection

| Rule | PlusEMU | Arcturus | comet-v2 |
|------|---------|----------|----------|
| Can't kick owner | Yes | Yes | Yes |
| Can't kick mods | Rank check | Permission check | Rank check |
| Self-kick | Ignored | Ignored | Ignored |

---

## 4. Ticket / Call-For-Help System

### Ticket Lifecycle

```
Player reports → Ticket OPEN → Staff picks up → IN_PROGRESS → Action taken → CLOSED
                                                  ↓
                                            Marked ABUSIVE (false report)
                                            Marked INVALID (insufficient info)
```

### Ticket Data Model Comparison

| Field | PlusEMU | Arcturus | comet-v2 |
|-------|---------|----------|----------|
| id | INT | INT | INT |
| type | INT (category) | — | — |
| category | INT | — | `category_id` INT |
| sender_id | INT | INT | `submitter_id` INT |
| reported_id | INT | INT | `reported_id` INT |
| moderator_id | INT | INT | `moderator_id` INT |
| room_id | INT | INT | `room_id` INT |
| message | STRING | STRING | STRING |
| state | ENUM(1=unassigned,2=assigned,3=closed) | — | ENUM(OPEN,IN_PROGRESS,CLOSED,INVALID,ABUSIVE) |
| priority | INT | — | INT |
| chat_evidence | `reportedChats` list | `ModToolChatLog` entries | `chat_messages` TEXT |
| created_at | TIMESTAMP | TIMESTAMP | `timestamp_opened` |
| closed_at | — | — | `timestamp_closed` |

### Ticket Storage

| Vendor | Strategy |
|--------|----------|
| PlusEMU | In-memory `ConcurrentDictionary<int, ModerationTicket>`, lost on restart |
| Arcturus | Memory-based `ModToolIssue` objects |
| comet-v2 | Database-persisted `moderation_help_tickets` table (survives restarts) |
| pixelsv | **Database-persisted (follow comet-v2 pattern)** |

---

## 5. Moderation Presets & Action Templates

### Preset System

All three vendors use pre-configured moderation action templates that
staff can apply with a single click. These define the severity and
consequences of each moderation action.

| Aspect | PlusEMU | Arcturus | comet-v2 |
|--------|---------|----------|----------|
| User presets | `moderation_presets` (type=user) | — | `moderation_presets` |
| Room presets | `moderation_presets` (type=room) | — | `moderation_presets` |
| Action categories | `moderation_preset_action_categories` | — | `moderation_action_categories` |
| Action outcomes | `moderation_preset_action_messages` | Sanction escalation | `moderation_actions` |
| Category examples | PII, Sexually Explicit, Bullying | — | PII, Sexually Explicit, Scam |

### Action Outcome Fields (comet-v2)

| Field | Type | Description |
|-------|------|-------------|
| mute_hours | INT | Duration of chat mute |
| ban_hours | INT | Duration of account ban |
| avatar_ban_hours | INT | Duration of avatar reset |
| trade_lock_hours | INT | Duration of trading restriction |

---

## 6. Chat Log & Evidence System

### Chat Log Approaches

| Aspect | PlusEMU | Arcturus | comet-v2 |
|--------|---------|----------|----------|
| Storage | Per-ticket list in memory | `ModToolChatLog` objects | `chatlogs` / `chat_messages` in DB |
| Query scope | Current ticket only | Per-room, time-windowed | Per-room, filterable |
| Persistence | Lost on restart | Memory (during session) | Database-backed |
| Moderator access | `GetModeratorTicketChatlogsEvent` | `GetModeratorRoomChatlogEvent` | Via ticket DAO |
| Room visit log | — | `ModToolRoomVisit` tracking | — |

### pixelsv Approach

Our `room_chat_logs` table (from plan 11) serves dual purpose:
1. **Moderation evidence** — moderators query by room + time range
2. **Historical record** — CLI export for audit trails

This eliminates the vendor-common pattern of ephemeral chat storage that
is lost on server restart.

---

## 7. Word Filter System

### Word Filter Comparison

| Feature | PlusEMU | Arcturus | comet-v2 |
|---------|---------|----------|----------|
| Table | — | `wordfilter` | — |
| Scope | Global | Per-word with room reporting flag | Global |
| Auto-mute | — | `mute` column (seconds per word) | — |
| Auto-report | — | `report` column (flag CFH) | — |
| Replacement | Asterisks | Configurable | Asterisks |
| Per-room filter | — | — | — |
| Bad bubble filter | — | `commands.cmd_chatcolor.banned_numbers` | — |

### pixelsv Approach (Deferred)

- Global word filter with per-word mute duration
- Optional per-room word list (room owner managed)
- Plugin-extensible filter pipeline

---

## 8. Ambassador System

### Ambassador Role

| Feature | PlusEMU | Arcturus | comet-v2 |
|---------|---------|----------|----------|
| Alert packet | `AmbassadorAlertEvent` | Permission-gated | — |
| Permission | Rank-based | `acc_ambassador` permission | — |
| Capabilities | Send warnings to users | Warn, minor moderation | — |
| Distinction | Separate from full moderators | Separate permission node | Not implemented |

### pixelsv Approach (Deferred)

Already has `AmbassadorPermission` config field in
`core/permission/config.go`. Implementation deferred to moderation
milestone.

---

## 9. Trade Lock System

### Trading Restrictions

| Feature | PlusEMU | Arcturus | comet-v2 |
|---------|---------|----------|----------|
| Trade lock | `user_info.trading_locked` counter | `ModToolSanctionTradeLockEvent` | `trade_lock_hours` in actions |
| Duration | Permanent until lifted | Probation-based | Hours-based |
| Scope | Per-user | Per-user with escalation | Per-user |

---

## 10. Sanction Escalation

### Escalation Models

| Vendor | Model |
|--------|-------|
| PlusEMU | Counters in `user_info`: cfhs, cfhs_abusive, cautions, bans, trading_locked |
| Arcturus | Probation system: `ModToolSanctionItem` with timestamps, auto-escalation |
| comet-v2 | Preset-driven: action templates define severity per category |

### Arcturus Probation Detail

Arcturus implements the most sophisticated escalation:
1. First offense → Warning
2. Within probation window → Mute
3. Repeated within probation → Temporary ban
4. Continued violations → Permanent ban

Probation timestamps decay over time, allowing de-escalation for
reformed behavior.

---

## 11. Moderator Tool Packets

### Key Packet IDs (Vendor Comparison)

| Action | PlusEMU | comet-v2 | Nitro IDs |
|--------|---------|----------|-----------|
| Moderator init | — | — | C2S: 2 |
| Get room chatlog | 3605 | 2804 | C2S: 3 |
| Get user chatlog | — | — | C2S: 3 |
| Get user info | — | — | C2S: 3 |
| Mod kick | 2582 | 794 | C2S varies |
| Mod mute | 1945 | 3861 | C2S varies |
| Mod ban | 2766 | 265 | C2S varies |
| Mod caution | 1840 | 2849 | C2S varies |
| Mod trade lock | 3742 | — | C2S varies |
| Get banned users | 2267 | 2652 | C2S: 2652 |
| Unban user | 992 | 451 | C2S: 3842 |
| Toggle mute tool | 3637 | 36 | C2S varies |

---

## 12. Database Tables Summary (pixelsv Proposal)

### Core Moderation Tables (Future)

| Table | Purpose |
|-------|---------|
| `moderation_bans` | Global bans (account, IP, machine, super) |
| `moderation_tickets` | Player reports / call-for-help |
| `moderation_presets` | Staff action message templates |
| `moderation_actions` | Action outcome definitions (mute/ban/lock hours) |
| `moderation_action_categories` | Grouping for action types |
| `user_moderation_history` | Per-user sanction log for escalation |

### Existing Tables (Already Implemented)

| Table | Purpose |
|-------|---------|
| `room_bans` | Per-room user bans with expiry |
| `room_chat_logs` | Chat history (plan 11, dual-purpose) |

---

## 13. Implementation Priority (Future Milestone)

### Phase 1 — Core Moderation

- Global ban system (account, IP, machine)
- Moderator tool init packet
- Room chatlog query (moderator access to `room_chat_logs`)
- Kick from hotel

### Phase 2 — Ticket System

- Call-for-help submit + withdraw
- Ticket assignment + resolution
- Chat evidence capture
- Moderation presets

### Phase 3 — Advanced

- Sanction escalation with probation
- Word filter pipeline
- Trade lock
- Ambassador alerts
- Per-room word filter management
