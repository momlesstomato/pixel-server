# Realm: Moderation & Safety

> **Position:** 180 | **Phase:** 11 (Moderation) | **Packets:** 83 (43 c2s, 40 s2c)
> **Services:** moderation | **Status:** Not yet implemented

---

## Overview

Moderation & Safety is the third-largest realm at 83 packets. It covers the mod-tool UI, call-for-help (CFH) ticket system, sanctions (ban/mute/kick), chat review, guide system, room information for moderators, user chatlogs, and the guardian voting system. This realm is security-critical: all operations require elevated permissions.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 11

---

## Packet Inventory

### C2S -- 43 packets

#### Mod Tool

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 707 | `moderation.request_room_info` | `roomId:int32` | Get room info for mod panel |
| 1391 | `moderation.request_user_chatlog` | `userId:int32` | Get user's chat history |
| 2587 | `moderation.request_room_chatlog` | `roomId:int32` | Get room's chat history |
| 31 | `moderation.preferences` | _(none)_ | Get mod tool preferences |

#### Sanctions

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1392 | `moderation.sanction` | `userId`, `sanctionType`, `reason`, `hours` | Apply sanction |
| 229 | `moderation.sanction_alert` | `userId`, `message` | Send alert to user |
| 1945 | `moderation.sanction_mute` | `userId`, `minutes` | Mute user globally |
| 2582 | `moderation.sanction_kick` | `userId:int32` | Kick from current room |
| 1681 | `moderation.sanction_default` | `userId:int32` | Apply default sanction |
| 1840 | `moderation.alert_event` | `userId`, `message` | Send mod alert |

#### Call-for-Help

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1691 | `moderation.call_for_help` | `message`, `category`, `reportedUserId`, `roomId` | Submit report |
| 2755 | `moderation.call_for_help_selfie` | same + photo data | Report with screenshot |
| 15 | `moderation.pick_issues` | `issueIds:int32[]` | Moderator picks up issues |
| 1572 | `moderation.release_issues` | `issueIds:int32[]` | Release picked issues |
| 2067 | `moderation.close_issues` | `issueIds`, `closeType` | Close issues with resolution |
| 2717 | `moderation.close_default_action` | `issueId:int32` | Close with default action |
| 2746 | `moderation.get_cfh_status` | _(none)_ | Get pending CFH status |
| 211 | `moderation.get_cfh_chatlog` | `issueId:int32` | Get chat context for report |

#### Guide System

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1922 | `guide.on_duty_update` | `onDuty:boolean` | Toggle guide duty status |
| 1424 | `guide.guide_decides` | `accepted:boolean` | Accept/decline help request |
| 234 | `guide.invite_requester` | _(none)_ | Invite requester to guide's room |
| 291 | `guide.requester_cancels` | _(none)_ | Requester cancels request |
| 477 | `guide.feedback` | `rating:int32` | Rate guide session |
| 519 | `guide.is_typing` | `typing:boolean` | Typing indicator in guide chat |
| 887 | `guide.resolved` | _(none)_ | Mark guide session as resolved |
| 1052 | `guide.get_requester_room` | _(none)_ | Get requester's room ID |

#### Chat Review (Guardian)

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2501 | `chat_review.guide_detached` | _(none)_ | Guardian leaves review |
| -- | Additional chat review voting/decision packets |

### S2C -- 40 packets

Key responses:
- `moderation.mod_tool_data` -- mod panel initialization
- `moderation.issue_list` -- open CFH tickets
- `moderation.issue_info` -- ticket details
- `moderation.room_info` -- room data for mod panel
- `moderation.user_chatlog` -- user chat history
- `moderation.room_chatlog` -- room chat history
- `moderation.sanction_status` -- current sanctions on user
- `moderation.cfh_result` -- report submission result
- `guide.session_started` -- guide session established
- `guide.session_message` -- message in guide chat
- `guide.reporting_status` -- guide system status
- `chat_review.session_results` -- chat review voting results

---

## Implementation Analysis

### Permission Model

All moderation actions require rank >= 3 (Staff). Permission checks:

```go
func (h *Handler) requireStaff(session Session) error {
    if session.Rank() < RankStaff {
        return ErrInsufficientPermission
    }
    return nil
}
```

| Rank | Capabilities |
|------|-------------|
| 3 (Staff) | View chatlogs, issue alerts, soft-mute |
| 4 (Senior Staff) | Ban, kick, close CFH tickets |
| 5 (Admin) | All moderation + configure sanctions |
| 6 (Developer) | All + system-level operations |

### Call-for-Help (CFH) System

```
Report Flow:
  1. User submits CFH (moderation.call_for_help)
  2. Insert into moderation_tickets:
     - reporter_id, reported_user_id, room_id, category, message
     - status: OPEN
     - created_at: NOW()
  3. Broadcast to all online moderators: moderation.issue_list
  4. Moderator picks issue (moderation.pick_issues):
     - status: PICKED, handler_id: moderator
  5. Moderator reviews:
     - Get chatlog (moderation.get_cfh_chatlog)
     - Get room info (moderation.request_room_info)
  6. Moderator resolves:
     - Apply sanction if needed
     - Close issue (moderation.close_issues)
     - status: CLOSED, resolution: <action taken>
  7. Reporter notified: moderation.cfh_result

Ticket Categories:
  - Inappropriate behavior
  - Scam/fraud
  - Bullying/harassment
  - Inappropriate content
  - Other
```

### Ban System

```
Ban Types:
  - ALERT: Warning message only (no restriction)
  - MUTE: Cannot chat for duration (minutes)
  - KICK: Removed from current room
  - BAN: Cannot login for duration (hours/permanent)
  - IP_BAN: Blocks IP address (requires admin)
  - MACHINE_BAN: Blocks machine ID (requires admin)

Ban Flow:
  1. Moderator sends moderation.sanction
  2. Moderation service validates:
     - Moderator has sufficient rank
     - Target rank < moderator rank (cannot ban higher rank)
     - Sanction type is valid for moderator's rank
  3. Insert into moderation_bans:
     - user_id, type, reason, duration, moderator_id, created_at, expires_at
  4. Publish ban event to NATS
  5. Gateway subscribes to ban events:
     - If BAN/IP_BAN/MACHINE_BAN: disconnect user within 500ms
     - If MUTE: set mute state on session
     - If KICK: send room leave command

Critical requirement: Ban propagation to gateway must be < 500ms.
Use Redis PUBLISH ban:<userId> for instant notification.
```

### Chat History

Chatlogs are essential for moderation. Store all room chat in a time-partitioned table:

```sql
CREATE TABLE room_chatlogs (
    id          BIGSERIAL,
    room_id     INT NOT NULL,
    user_id     INT NOT NULL,
    username    VARCHAR(25) NOT NULL,
    message     TEXT NOT NULL,
    chat_type   SMALLINT NOT NULL,  -- 0=say, 1=shout, 2=whisper
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Monthly partitions
CREATE TABLE room_chatlogs_2026_01 PARTITION OF room_chatlogs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

**Async batch writer:** Chat logging must not block the room tick. Use the batch writer pattern from [006-storage.md](../006-storage.md) -- buffer messages and flush every second.

### Guide System

The guide system pairs new users with experienced helpers:

```
1. New user requests help
2. System finds available guide (on-duty, in lobby)
3. Guide session created
4. Guide and requester can chat via guide.session_message
5. Guide can invite requester to their room
6. Session resolved by either party
7. Requester rates the guide
```

---

## Caveats & Edge Cases

### 1. Rank Escalation Prevention
A moderator must not be able to sanction users of equal or higher rank. Always check: `target.rank < moderator.rank`. This is missed in some reference emulators.

### 2. Ban Circumvention
IP bans are trivially bypassed (VPN). Machine ID bans are more effective but can be spoofed. Layer both for best coverage. Log ban attempts for pattern detection.

### 3. Chat Log Privacy
Whispers are logged for moderation but should be clearly marked as private. Only rank 4+ should access whisper logs. GDPR considerations: implement log retention policy (90 days default, configurable).

### 4. Concurrent Issue Handling
Two moderators might pick the same CFH ticket simultaneously. Use `UPDATE moderation_tickets SET handler_id = ?, status = 'PICKED' WHERE id = ? AND status = 'OPEN'` -- only one succeeds.

### 5. Ban Duration Edge Cases
- 0 hours = alert only (no ban)
- -1 hours = permanent ban
- Very large hours = effectively permanent, but use explicit permanent flag instead
- Ban expiry check: `WHERE expires_at > NOW() OR expires_at IS NULL`

### 6. Gateway Ban Latency
The 500ms ban propagation requirement means Redis PUB/SUB is preferred over NATS for this path. Gateway subscribes to `ban` channel, immediately closes socket on match.

### 7. Offline Users
Banning an offline user must persist the ban for enforcement on next login. The ban check happens during SSO validation in the auth service.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Ban propagation** | In-process (instant, single-node) | Redis PUB/SUB (< 500ms, multi-node) |
| **Chat logging** | Synchronous DB write per message | Async batch writer (no tick blocking) |
| **Chat storage** | Single table (slow at scale) | Monthly partitions for fast queries |
| **CFH handling** | Basic pick/close | Full workflow with chatlogs and context |
| **Rank checks** | Hardcoded in each handler | Centralized permission middleware |
| **Guide system** | Not implemented in most emulators | Full guide matching + chat + rating |
| **GDPR** | Not considered | Configurable retention + deletion |

---

## Dependencies

- **Phase 2 (Identity)** -- user rank for permission checks
- **Phase 3 (Room)** -- room chatlogs, room info
- **Phase 5 (Social)** -- reporter/reported user context
- **PostgreSQL** -- bans, tickets, chatlogs (partitioned)
- **Redis** -- ban PUB/SUB for instant propagation

---

## Testing Strategy

### Unit Tests
- Rank permission checking (every rank combination)
- Ban duration calculation and expiry
- CFH ticket state machine
- Chat log query construction

### Integration Tests
- Full CFH flow: submit → pick → review → resolve
- Ban apply and expiry against real DB
- Chat log insertion via batch writer
- Concurrent issue pick (only one succeeds)

### E2E Tests
- Moderator bans user, user is disconnected within 1 second
- Moderator views chat history for a room
- User submits CFH, moderator sees it in issue list
