# User Profile — Login Bundle

The login bundle is a sequence of 9 server-to-client packets sent immediately
after a player authenticates. Together they populate every part of the Nitro
client's UI that needs to show personal data before the player can enter a room.

---

## Trigger

`identity.SendLoginBundle` is called by `Listener.handleAuthenticated` in the
game service, immediately after:
1. The session is registered in the sessions map
2. The `player.joined` plugin event is emitted

---

## The 9-packet sequence (in order)

| # | Packet name | Header ID | Purpose |
|---|---|---|---|
| 1 | `user_info` | 1554 | Full identity: username, figure, motto, credits, flags |
| 2 | `user_permissions` | 3531 | Club level, security level, ambassador flag |
| 3 | `user_settings` | 513 | Audio volumes, chat prefs, UI flags |
| 4 | `user_credits` | 869 | Credit balance |
| 5 | `user_currency` | 2018 | Activity point balances (currently empty array) |
| 6 | `subscription_status` | 954 | Club subscription details |
| 7 | `user_home_room` | 2875 | Home room ID, room to enter on login |
| 8 | `noobness_level` | 3738 | Account maturity classification |
| 9 | `user_perks` | 2561 | Feature perk list |

On first send failure the bundle aborts and logs the error. Packets already sent
remain delivered.

---

## Packet field reference

### 1 — user_info (1554)

| Field | Source |
|---|---|
| `ID` | `user.ID` |
| `Username` | `user.Username` |
| `Figure` | `user.Figure` |
| `Gender` | `user.Gender` (`"M"` / `"F"`) |
| `Motto` | `user.Motto` |
| `AccountCreated` | `user.AccountCreated` (Unix timestamp) |
| `NoobnessBool` | `noobness > 0` |
| `AllowNameChange` | `user.AllowNameChange` |
| `SafetyLocked` | `user.SafetyLocked` |

### 2 — user_permissions (3531)

| Field | Source |
|---|---|
| `ClubLevel` | `roleProfile.ClubLevel` (0=none, 1=HC, 2=VIP) |
| `SecurityLevel` | `roleProfile.SecurityLevel` (rank, min 1) |
| `IsAmbassador` | `roleProfile.IsAmbassador` |

### 3 — user_settings (513)

| Field | Source |
|---|---|
| `VolumeSystem` | `settings.VolumeSystem` |
| `VolumeFurni` | `settings.VolumeFurni` |
| `VolumeTrax` | `settings.VolumeTrax` |
| `OldChat` | `settings.OldChat` |
| `IgnoreRoomInvites` | `settings.IgnoreRoomInvites` |
| `BlockCameraFollow` | `settings.BlockCameraFollow` |
| `FriendBarOpen` | `settings.FriendBarOpen` |
| `UIFlags` | `settings.UIFlags` |

### 4 — user_credits (869)

| Field | Value |
|---|---|
| `Credits` | `strconv.Itoa(int(user.Credits))` — credits as a string |

### 5 — user_currency (2018)

Array of `{type int32, amount int32}`. Currently always an empty array.

### 6 — subscription_status (954)

| Field | Value |
|---|---|
| `ProductName` | `"club_habbo"` |
| `DaysRemaining` | `0` |
| `MemberSince` | `0` |
| `IsVip` | `false` |
| `PastVip` | `false` |
| `MinutesLeft` | `0` |

Placeholder values — subscription persistence is not yet implemented.

### 7 — user_home_room (2875)

| Field | Source |
|---|---|
| `HomeRoom` | `user.HomeRoom` |
| `RoomToEnter` | `user.HomeRoom` |

### 8 — noobness_level (3738)

Calculated from `user.AccountCreated` by `BuildNoobnessLevel`:

| Account age | Level | Meaning |
|---|---|---|
| < 3 days | `2` | Very new |
| 3–6 days | `1` | New |
| ≥ 7 days | `0` | Veteran |
| Zero time value | `0` | Default |

### 9 — user_perks (2561)

Array of `{code string, message string, isAllowed bool}`. Currently always an
empty array — perk persistence is not yet implemented.

---

## Role profile logic

`BuildRoleProfile(rank int32, perks []string)` derives the role from user data:

| Input condition | Result |
|---|---|
| Perks contains `"club_vip"` | `ClubLevel = 2` |
| Perks contains `"club_hc"` (no VIP) | `ClubLevel = 1` |
| Neither | `ClubLevel = 0` |
| `rank` value | `SecurityLevel = max(rank, 1)` |
| Perks contains `"ambassador"` | `IsAmbassador = true` |

---

## Error paths

| Failure | Behaviour |
|---|---|
| `user.Repository.GetByID` returns `ErrNotFound` | Bundle aborted; error logged, session stays open |
| Settings not found | `nil` settings passed to `BuildLoginBundle`; default zero-values used |
| `natsSession.Send` fails | Error logged with packet header ID; bundle aborts at that packet |
