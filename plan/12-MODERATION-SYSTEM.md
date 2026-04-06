# 12 - Moderation System

## Overview

This plan defines the moderation system for pixel-server, covering both
**room-level actions** (kick, ban, mute scoped to a single room) and
**hotel-level actions** (global kicks, warns, bans, mutes that persist
across all rooms and sessions).

Both action types share a unified database schema with a scope
discriminator column ("room" vs "hotel"). Room actions can be undone by
room owners or staff. Hotel actions are **append-only** -- they can be
deactivated by a moderator but never deleted from the database, ensuring
a permanent audit trail.

**Navigator status**: All 22 navigator packets (11 C2S + 11 S2C) are
100% implemented. No navigator work is required.

---

## Architecture

### Unified Action Schema

A single `moderation_actions` table stores all moderation actions with
a scope discriminator:

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PK |
| scope | VARCHAR(10) | NOT NULL ("room" or "hotel") |
| action_type | VARCHAR(20) | NOT NULL (kick, ban, mute, warn) |
| target_user_id | INTEGER | NOT NULL |
| issuer_id | INTEGER | NOT NULL |
| room_id | INTEGER | NULL (set when scope=room) |
| reason | TEXT | NOT NULL DEFAULT '' |
| duration_minutes | INTEGER | NULL (NULL = permanent) |
| expires_at | TIMESTAMPTZ | NULL (NULL = permanent) |
| active | BOOLEAN | NOT NULL DEFAULT TRUE |
| deactivated_by | INTEGER | NULL |
| deactivated_at | TIMESTAMPTZ | NULL |
| ip_address | VARCHAR(45) | NULL (for hotel IP bans) |
| machine_id | VARCHAR(64) | NULL (for hotel machine bans) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

**Indexes:**
- idx_mod_actions_target ON (target_user_id, scope, active)
- idx_mod_actions_room ON (room_id, active) WHERE scope = 'room'
- idx_mod_actions_expires ON (expires_at) WHERE active = TRUE
- idx_mod_actions_ip ON (ip_address) WHERE ip_address IS NOT NULL

### Key Design Decisions

1. **Unified table** -- Room and hotel actions in one table simplifies
   querying a user's full moderation history and avoids schema duplication.

2. **Soft-deactivation** -- Hotel actions set active=false instead of
   DELETE. The deactivated_by and deactivated_at columns record who
   lifted the action and when, preserving full audit trail.

3. **Room actions deletable** -- Room-scoped actions (kicks, bans) can
   be hard-deleted by room owners or staff, matching existing room_bans
   behavior.

4. **Duration model** -- duration_minutes stores the intended duration,
   expires_at stores the computed expiry timestamp. NULL expires_at means
   permanent. Active checks: active=true AND (expires_at IS NULL OR
   expires_at > NOW()).

5. **Existing room_bans migration** -- The existing room_bans table will
   continue to function alongside moderation_actions. Room ban checks
   query both tables for backwards compatibility. Future migration will
   consolidate.

---

## Vendor Cross-Reference

### Moderation Action Matrix

| Action | PlusEMU | Arcturus | comet-v2 | pixelsv |
|--------|---------|----------|----------|---------|
| Room kick | KickUser packet | RoomUserKickEvent | KickUserMessageEvent | **Room action (transient)** |
| Room ban | room_bans table | room_bans + duration | Room state | **Room action + existing room_bans** |
| Room mute | MuteAllInRoom toggle | cmd_roommute | RoomMuteState enum | **Room action (togglable)** |
| Hotel kick | ModerationKickEvent | ModToolKickEvent | Session termination | **Hotel action (transient)** |
| Hotel warn | ModerationCaution | ModerationCaution | ModerationCaution | **Hotel action (persistent)** |
| Hotel ban | moderation_bans table | DB + config | DB with expiry | **Hotel action (persistent)** |
| Hotel mute | ModerationMuteEvent | mute_end_timestamp | MuteUserMessageEvent | **Hotel action (persistent)** |
| IP ban | bantype=ip | type=ip | type=ip | **Hotel action (ip_address col)** |
| Machine ban | bantype=machine | type=machine | type=machine | **Hotel action (machine_id col)** |

### Our Improvements Over Vendors

1. **Unified audit trail** -- All vendors store bans, mutes, kicks in
   separate systems. We unify in one table with full issuer tracking.

2. **Immutable hotel records** -- PlusEMU uses ephemeral in-memory dicts.
   Arcturus has partial DB. We persist everything permanently.

3. **Deactivation tracking** -- No vendor tracks who lifted a ban and
   when. We record deactivated_by and deactivated_at.

4. **Dual-scope architecture** -- Clean separation of room vs hotel
   actions with shared querying infrastructure.

---

## Packet Registry

### Client-to-Server (moderation)

| ID | Name | Fields | Phase |
|----|------|--------|-------|
| 1320 | room.kick_user | userId (int32) | **M1** |
| 1477 | room.ban_user | userId (int32), roomId (int32), banType (string) | **M1** |
| 2582 | mod.kick_user | userId (int32), message (string) | **M1** |
| 1945 | mod.mute_user | userId (int32), message (string), minutes (int32) | **M1** |
| 2766 | mod.ban_user | userId (int32), message (string), banType (int32), cfhTopic (string), duration (int32) | **M1** |
| 1840 | mod.warn_user | userId (int32), message (string) | **M1** |

### Server-to-Client (moderation, existing)

| ID | Name | Fields | Status |
|----|------|--------|--------|
| 1683 | user.banned | message (string) | **Exists** |
| 1890 | session.moderation_caution | message (string), detail (string) | **Exists** |
| 4000 | handshake.disconnect_reason | code (int32) | **Exists** |

---

## SDK Events

### Moderation Events (new)

| Event | Type | Fields |
|-------|------|--------|
| UserKicking | Cancellable | TargetID, IssuerID, RoomID, Scope |
| UserKicked | Non-cancellable | TargetID, IssuerID, RoomID, Scope |
| UserBanning | Cancellable | TargetID, IssuerID, Scope, BanType, Duration, Reason |
| UserBanned | Non-cancellable | TargetID, IssuerID, Scope, BanType, Duration, Reason |
| UserMuting | Cancellable | TargetID, IssuerID, Scope, Duration, Reason |
| UserMuted | Non-cancellable | TargetID, IssuerID, Scope, Duration, Reason |
| UserWarning | Cancellable | TargetID, IssuerID, Message |
| UserWarned | Non-cancellable | TargetID, IssuerID, Message |
| ActionDeactivating | Cancellable | ActionID, DeactivatedBy |
| ActionDeactivated | Non-cancellable | ActionID, DeactivatedBy |

---

## Permission Scopes

### Permission constants (`pkg/moderation/domain/permission.go`)

| Constant | Value | Purpose |
|----------|-------|---------|
| `PermKick` | `moderation.kick` | Hotel-level kick (force disconnect) |
| `PermBan` | `moderation.ban` | Hotel-level ban (account/IP/machine) |
| `PermMute` | `moderation.mute` | Hotel-level mute (chat restriction) |
| `PermWarn` | `moderation.warn` | Send warning/caution to user |
| `PermTradeLock` | `moderation.trade_lock` | Hotel-level trade lock sanction |
| `PermUnban` | `moderation.unban` | Deactivate hotel bans |
| `PermUnmute` | `moderation.unmute` | Deactivate hotel mutes |
| `PermHistory` | `moderation.history` | View user moderation history |
| `PermTool` | `moderation.tool` | Enable moderator tool initialization on login |
| `PermAmbassador` | `role.ambassador` | Identifies ambassador role for alert broadcasts |

### Realtime wiring

The `PermissionChecker` interface is defined in `pkg/moderation/adapter/realtime/runtime.go`
and wired at startup via `modRT.SetPermissionChecker(...)` in `core/cli/serve_routes.go`.

Every hotel-level realtime action checks the issuer's permission before executing:

| Packet handler | Guard permission |
|----------------|-----------------|
| `handleModKick` | `moderation.kick` |
| `handleModBan` | `moderation.ban` |
| `handleModMute` | `moderation.mute` |
| `handleModWarn` | `moderation.warn` |
| `handleTradeLock` | `moderation.trade_lock` |
### Moderator tool initialization (PostAuthHook)

The moderation realm participates in the **PostAuthHook** lifecycle to
send `ModeratorInitPacket` (S2C 2696) immediately after authentication:

1. `pkg/handshake/adapter/realtime/handler.go` defines the `PostAuthHook`
   interface with `OnPostAuth(ctx, connID, userID)`.
2. After `postauth.Run()` completes, `handleAuthPacket` checks whether
   the `UserRuntime` (a `compositeRuntime`) implements `PostAuthHook`.
3. `compositeRuntime.OnPostAuth` iterates its component runtimes and
   delegates to any that implement `PostAuthHook`.
4. `moderationrealtime.Runtime.OnPostAuth` calls `SendModToolInit`,
   which checks `moderation.tool` permission before sending the packet.
5. `TicketPermission` and `ChatlogPermission` fields on the packet are
   resolved via `moderation.history` permission (not hardcoded).
CFH (`handleCallForHelp`) requires no special permission — any authenticated user may submit.
Room-level actions use existing room ownership / rights checks, not permission scopes.

---

## HTTP API

### Endpoints

| Method | Path | Purpose | Permission |
|--------|------|---------|------------|
| GET | /api/v1/moderation/actions | List actions (filterable) | moderation.history |
| GET | /api/v1/moderation/actions/:id | Get single action detail | moderation.history |
| GET | /api/v1/moderation/users/:userId/actions | User action history | moderation.history |
| POST | /api/v1/moderation/actions | Create hotel action | moderation.{type} |
| PATCH | /api/v1/moderation/actions/:id/deactivate | Deactivate action | moderation.un{type} |
| GET | /api/v1/moderation/users/:userId/active | Check active restrictions | moderation.history |

### Query Parameters (GET /actions)

| Param | Type | Description |
|-------|------|-------------|
| scope | string | Filter by "room" or "hotel" |
| action_type | string | Filter by kick/ban/mute/warn |
| target_user_id | int | Filter by target user |
| active | bool | Filter active/inactive |
| page | int | Pagination offset |
| limit | int | Page size (max 100) |

---

## CLI Commands

### moderation subcommand tree

| Command | Description |
|---------|-------------|
| moderation list | List actions with filters (--scope, --type, --user-id, --active) |
| moderation ban | Create hotel ban (--user-id, --reason, --duration, --ip, --machine-id) |
| moderation unban | Deactivate a ban (--action-id) |
| moderation history | Show user moderation history (--user-id) |

---

## Integration Points

### Authentication Flow (hotel ban check)

On login, the authentication/handshake flow must check for active hotel
bans before completing the session. If an active ban exists:
1. Send UserBannedPacket (1683) with the ban reason
2. Send DisconnectReason (4000) with code 1 (just banned) or 10 (still banned)
3. Close the WebSocket connection

### Chat Service (hotel mute check)

Before processing any chat message, the chat service must check for
active hotel mutes. If muted:
1. Suppress the message
2. Optionally send a ModerationCaution (1890) informing the user

### Room Entry (room ban check)

The existing room ban check in handleOpenFlat continues to work.
Additionally, room-scoped moderation_actions with type=ban are checked.

---

## Implementation Scope

### Package Structure

| Package | Files | Purpose |
|---------|-------|---------|
| pkg/moderation/domain | action.go, repository.go, errors.go | Domain types + interfaces |
| pkg/moderation/application | service.go | Business logic |
| pkg/moderation/infrastructure/model | action.go | GORM model |
| pkg/moderation/infrastructure/store | action_store.go | Repository impl |
| pkg/moderation/infrastructure/migration | migration.go | DB migration |
| pkg/moderation/adapter/httpapi | contracts.go, routes.go, openapi.go | REST API |
| pkg/moderation/adapter/command | command.go, actions.go | CLI |
| pkg/moderation/adapter/realtime | dispatch.go, runtime.go | Packet handlers |
| pkg/moderation/packet | constants.go, packets.go | Packet types |
| sdk/events/moderation | 10 event files | SDK events |

### Implementation Checklist

- [x] Domain types (Action, ActionScope, ActionType, repository interface)
- [x] GORM model + migration
- [x] Store implementation
- [x] Application service (Create, Deactivate, List, CheckActive)
- [x] SDK events (5 before/after pairs)
- [x] Packet decoders (room kick/ban, mod kick/mute/ban/warn)
- [x] Realtime dispatch + handlers
- [x] HTTP API routes + OpenAPI
- [x] CLI commands
- [x] Wire into serve_services.go and serve_routes.go
- [x] Tests for all layers
- [x] Integrate hotel ban check in authentication flow
- [x] Integrate hotel mute check in chat service

### Phase 2 (Completed)

- [x] Ticket / call-for-help system (CFH packets, TicketService, HTTP + CLI)
- [x] Moderation presets / action templates (PresetService, HTTP + CLI)
- [x] Sanction escalation with probation model (Escalate in Service)
- [x] Word filter pipeline (global + per-room) (WordFilterService, ChatService integration)
- [x] Trade lock (TypeTradeLock, SanctionTradeLockPacket, HasActiveTradeLock)
- [x] Ambassador alerts (AlertAmbassadors, realtime dispatch after kick/ban)
- [x] Moderator tool initialization packet (ModeratorInitPacket S2C 2696)
- [x] Room visit tracking (VisitService, realtime recording on room entry)

---

## Status

### Implemented

- [x] Plan document created (12-MODERATION-SYSTEM.md)
- [x] Existing room_bans infrastructure (room_bans table, room ban packets)
- [x] UserBannedPacket (1683) exists
- [x] ModerationCautionPacket (1890) exists
- [x] Room kick/ban packet constants defined (1320, 1477)
- [x] Chat history (room_chat_logs) for moderation evidence
- [x] Domain types and infrastructure
- [x] Service layer
- [x] Packet implementations
- [x] Realtime handlers
- [x] HTTP API
- [x] CLI commands
- [x] SDK events
- [x] Tests
- [x] Wired into serve_services.go, serve_routes.go, registry.go
- [x] Hotel mute check integrated into chat service

### In Progress

None -- all items are complete.
