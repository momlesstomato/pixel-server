# User Identity

## Overview

The user identity subsystem manages the core profile, figure, motto, and home
room for each player. Identity fields are persisted in PostgreSQL and delivered
to the client during the post-authentication burst via binary packets.

## Domain Model

The `User` aggregate holds all identity fields:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | Primary key (auto-increment) |
| `Username` | string | Unique, 3–64 characters |
| `Figure` | string | Avatar appearance string (1–255 chars) |
| `Gender` | string | `M` or `F` |
| `Motto` | string | Player motto (max 127 chars) |
| `RealName` | string | Optional real name (max 64 chars) |
| `RespectsReceived` | int | Total respects received |
| `HomeRoomID` | int | Default room (-1 = none) |
| `CanChangeName` | bool | Whether name change is allowed |
| `NoobnessLevel` | int | Account maturity (0–2, default 2) |
| `SafetyLocked` | bool | Safety lock status |
| `GroupID` | int | Legacy primary group (default 1) |

## Post-Authentication Burst

After successful SSO authentication, the server sends the following packets:

| Packet | ID | Contents |
|--------|----|----------|
| `user.info` | 2725 | Full identity: ID, username, figure, gender, motto, real name, respects, remaining respects, last access |
| `user.permissions` | 411 | Club level, security level, ambassador flag |
| `user.perks` | 2586 | Client perk grants (USE_GUIDE_TOOL, TRADE, CAMERA, etc.) |
| `user.noobness_level` | 3738 | Account maturity level |
| `user.settings` | 513 | Volume, chat, camera preferences |

## Packets

### user.info (S2C 2725)

| Field | Type | Description |
|-------|------|-------------|
| `UserID` | int32 | User identifier |
| `Username` | string | Display name |
| `Figure` | string | Avatar appearance |
| `Gender` | string | M or F |
| `Motto` | string | Player motto |
| `RealName` | string | Real name |
| `DirectMail` | bool | Direct mail enabled |
| `RespectsReceived` | int32 | Total respects |
| `RespectsRemaining` | int32 | Daily user respects left |
| `RespectsPetRemaining` | int32 | Daily pet respects left |
| `StreamPublishingAllowed` | bool | Stream publish flag |
| `LastAccessDate` | string | Formatted last login |
| `CanChangeName` | bool | Name change allowed |
| `SafetyLocked` | bool | Safety lock active |

### user.get_info (C2S 357)

Empty packet. Client requests a fresh `user.info` response.

### user.update_motto (C2S 2228)

| Field | Type | Description |
|-------|------|-------------|
| `Motto` | string | New motto text (max 127 chars) |

Fires a `MottoChanged` event (cancellable). Persists only if the event is not
cancelled.

### user.update_figure (C2S 2730)

| Field | Type | Description |
|-------|------|-------------|
| `Gender` | string | M or F |
| `Figure` | string | New appearance string (1–255 chars) |

Fires a `FigureChanged` event (cancellable). Responds with `user.figure`
(S2C 2429) containing the updated figure and gender.

### user.set_home_room (C2S 1740)

| Field | Type | Description |
|-------|------|-------------|
| `RoomID` | int32 | Target home room ID (-1 = clear) |

Responds with `user.home_room` (S2C 2875) containing `HomeRoomID` and
`RoomIDToEnter`.

## REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/{id}` | Get full user profile |
| `PATCH` | `/api/v1/users/{id}` | Update profile fields |

**PATCH body (all optional):**

| Field | Type | Constraint |
|-------|------|------------|
| `figure` | string | 1–255 chars |
| `gender` | string | M or F |
| `motto` | string | max 127 chars |
| `home_room_id` | int | >= -1 |

## CLI

```bash
pixelsv user get 1              # Get user profile (JSON)
pixelsv user update 1 \
  --motto "hello" \
  --figure "hr-890-45" \
  --gender F \
  --home-room-id 42             # Update profile fields
```

## Plugin Events

| Event | Cancellable | Fields |
|-------|-------------|--------|
| `FigureChanged` | Yes | ConnID, UserID, OldFigure, NewFigure, Gender |
| `MottoChanged` | Yes | ConnID, UserID, OldMotto, NewMotto |

When cancelled, the mutation is rolled back and the response packet is not
sent.

## Settings

### user.settings (S2C 513)

| Field | Type | Default |
|-------|------|---------|
| `VolumeSystem` | int32 | 100 |
| `VolumeFurni` | int32 | 100 |
| `VolumeTrax` | int32 | 100 |
| `OldChat` | bool | false |
| `RoomInvites` | bool | true |
| `CameraFollow` | bool | true |
| `Flags` | int32 | 0 |
| `ChatType` | int32 | 0 |

Settings changes are debounced with a 2-second coalesce window to avoid write
storms from rapid UI toggles.

### Settings Packets (C2S)

| Packet | ID | Fields |
|--------|----|--------|
| `user.settings_volume` | 1367 | VolumeSystem, VolumeFurni, VolumeTrax |
| `user.settings_room_invites` | configurable | Enabled (bool) |
| `user.settings_old_chat` | configurable | Enabled (bool) |

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/{id}/settings` | Get user settings |
| `PATCH` | `/api/v1/users/{id}/settings` | Update settings (partial) |

**PATCH body (all optional):**

| Field | Type | Constraint |
|-------|------|------------|
| `volume_system` | int | 0–100 |
| `volume_furni` | int | 0–100 |
| `volume_trax` | int | 0–100 |
| `old_chat` | bool | — |
| `room_invites` | bool | — |
| `camera_follow` | bool | — |
| `flags` | int | >= 0 |
| `chat_type` | int | 0 or 1 |
