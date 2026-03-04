# User Profile — Packets

Complete reference for all C2S handlers and S2C response packets in the
user-profile realm.

---

## Client → Server (C2S) Handlers

All handlers are registered in `identity.Router` and dispatched by header ID.
Handler source: `services/game/internal/identity/handler.go`.

---

### get_user_info — header 357

*(No fields.)*

**Handler** `HandleGetInfo`: Re-sends `user_info` (1554) for the requesting
session's own user. Useful when the client needs to refresh identity data.

---

### get_user_profile — header 3265

| Field | Type | Description |
|---|---|---|
| `TargetUserID` | `int32` | User whose profile to fetch |

**Handler** `HandleGetProfile`: Looks up the target user by ID. Sends
`user_profile` (2526) with public profile data. If target not found, logs error,
no response.

---

### update_figure — header 2730

| Field | Type | Description |
|---|---|---|
| `Figure` | `string` | New figure string |
| `Gender` | `string` | `"M"` or `"F"` |

**Handler** `HandleUpdateFigure`: Persists via `user.Repository.Update`. Sends
`figure_update` (2429) confirming the new figure and gender.

---

### update_motto — header 2228

| Field | Type | Description |
|---|---|---|
| `Motto` | `string` | New motto text |

**Handler** `HandleUpdateMotto`: Persists via repository. Sends `user_info`
(1554) with updated motto.

---

### settings_volume — header 1367

| Field | Type | Description |
|---|---|---|
| `VolumeSystem` | `int32` | System audio volume |
| `VolumeFurni` | `int32` | Furniture sounds volume |
| `VolumeTrax` | `int32` | Trax music volume |

**Handler** `HandleSettingsVolume`: Persists via `user.Repository.UpdateSettings`.
No response packet.

---

### settings_old_chat — header 1262

| Field | Type | Description |
|---|---|---|
| `OldChat` | `bool` | Classic chat style preference |

**Handler** `HandleSettingsOldChat`: Persists. No response packet.

---

### settings_room_invites — header 1086

| Field | Type | Description |
|---|---|---|
| `BlockRoomInvites` | `bool` | Whether to block room invite notifications |

**Handler** `HandleSettingsRoomInvites`: Persists. No response packet.

---

### set_home_room — header 1740

| Field | Type | Description |
|---|---|---|
| `RoomID` | `int32` | New home room ID |

**Handler** `HandleSetHomeRoom`: Persists. Sends `user_home_room` (2875) with
updated value.

---

### get_ignored — header 3878

*(No fields.)*

**Handler** `HandleGetIgnored`: Fetches full ignore list via
`user.IgnoreRepository.GetIgnored`. Sends `ignored_users` (126).

---

### ignore_user — header 1117

| Field | Type | Description |
|---|---|---|
| `Username` | `string` | Username to add to ignore list |

**Handler** `HandleIgnore`: Looks up user by username; adds ID to ignore list via
`IgnoreRepository.AddIgnore`. Sends updated `ignored_users` (126).  
Silent no-op if username not found.

---

### ignore_user_id — header 3314

| Field | Type | Description |
|---|---|---|
| `TargetUserID` | `int32` | User ID to add to ignore list |

**Handler** `HandleIgnoreByID`: Directly adds ID without username lookup. Sends
`ignored_users` (126).

---

### unignore_user — header 2061

| Field | Type | Description |
|---|---|---|
| `Username` | `string` | Username to remove from ignore list |

**Handler** `HandleUnignore`: Looks up user; removes from ignore list. Sends
`ignored_users` (126).

---

### get_wardrobe — header 2742

*(No fields.)*

**Handler** `HandleGetWardrobe`: Fetches up to 10 saved outfits via
`WardrobeRepository.GetOutfits`. Sends `wardrobe` (3315).

---

### save_wardrobe_outfit — header 800

| Field | Type | Description |
|---|---|---|
| `SlotID` | `int32` | Wardrobe slot index (1-indexed, 1–10) |
| `Figure` | `string` | Figure string to save |
| `Gender` | `string` | `"M"` or `"F"` |

**Handler** `HandleSaveWardrobeOutfit`: Persists via `WardrobeRepository.SaveOutfit`.
Sends `wardrobe` (3315) with updated outfit list.

---

### check_username — header 3950

| Field | Type | Description |
|---|---|---|
| `Username` | `string` | Username to check availability of |

**Handler** `HandleCheckName`: Calls `user.Repository.GetByUsername`. If found,
returns result code 4 (taken). If not found, returns code 0 (available).  
Sends `check_username_result` (563).

Result codes:

| Code | Meaning |
|---|---|
| `0` | Available |
| `4` | Already taken |

---

### change_username — header 2977

| Field | Type | Description |
|---|---|---|
| `Username` | `string` | Desired new username |

**Handler** `HandleChangeName`: Checks `user.AllowNameChange`. If false, no
action. If true, calls `GetByUsername` to check availability, then `Update` to
persist. Sends `user_info` (1554) with new username.

---

### get_email_status — header 2557

*(No fields.)*

**Handler** `HandleGetEmailStatus`: Sends hardcoded `email_status` (612) with
`{HasEmail: true, IsVerified: true}`.

---

### respect_user — header 2694

| Field | Type | Description |
|---|---|---|
| `TargetUserID` | `int32` | User to receive the respect point |

**Handler** `HandleRespect`: Increments `user.RespectPoints` on the target user
via `Repository.Update`. No response packet.

---

## Server → Client (S2C) — Response packets

| Packet | Header | Key fields | Sent by |
|---|---|---|---|
| `user_info` | 1554 | ID, Username, Figure, Gender, Motto, AccountCreated, AllowNameChange, SafetyLocked | Login bundle #1, motto update, figure update |
| `user_permissions` | 3531 | ClubLevel, SecurityLevel, IsAmbassador | Login bundle #2 |
| `user_settings` | 513 | Volumes (system/furni/trax), OldChat, IgnoreRoomInvites, BlockCameraFollow, FriendBarOpen, UIFlags | Login bundle #3 |
| `user_credits` | 869 | Credits as string | Login bundle #4 |
| `user_currency` | 2018 | `[]{type, amount}` array | Login bundle #5 |
| `subscription_status` | 954 | ProductName, DaysRemaining, IsVip, PastVip | Login bundle #6 |
| `user_home_room` | 2875 | HomeRoom, RoomToEnter | Login bundle #7, set_home_room |
| `noobness_level` | 3738 | Level (0/1/2) | Login bundle #8 |
| `user_perks` | 2561 | `[]{code, message, isAllowed}` array | Login bundle #9 |
| `user_profile` | 2526 | ID, Username, Figure, Motto, AccountCreated, FriendsCount, RoomsCount, IsOnline | get_user_profile |
| `figure_update` | 2429 | Figure, Gender | update_figure |
| `ignored_users` | 126 | `[]string` username list | get_ignored, ignore_*, unignore_* |
| `wardrobe` | 3315 | `[]{SlotID, Figure, Gender}` array | get_wardrobe, save_wardrobe_outfit |
| `check_username_result` | 563 | Username, ResultCode | check_username |
| `email_status` | 612 | HasEmail, IsVerified | get_email_status |
