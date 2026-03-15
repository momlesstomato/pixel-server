# Client Perks

## Overview

Client perks are feature flags sent to the Habbo client after authentication.
Each perk maps to a permission string. When the user holds the permission in
any of their assigned groups, the perk is granted. Perks are delivered as
`user.perks` (S2C 2586) during the post-authentication burst and re-sent in
real time when group assignments change.

## Perk Packet (S2C 2586)

The packet contains an array of perk entries:

| Field | Type | Description |
|-------|------|-------------|
| `count` | int32 | Number of entries |
| Per entry: | | |
| `Code` | string | Perk identifier |
| `ErrorMessage` | string | Empty when allowed |
| `IsAllowed` | bool | Whether the user holds the perm |

## Perk Registry

| Code | Permission | Description |
|------|------------|-------------|
| `USE_GUIDE_TOOL` | `perk.guide` | Access Habbo Guide tool |
| `GIVE_GUIDE_TOURS` | `perk.guide.tours` | Give guided tours |
| `JUDGE_CHAT_REVIEWS` | `perk.chat_reviews` | Judge reported chat messages |
| `VOTE_IN_COMPETITIONS` | `perk.competitions` | Vote in competitions |
| `CALL_ON_HELPERS` | `perk.helpers` | Request helper assistance |
| `CITIZEN` | `perk.citizen` | Full citizen access |
| `TRADE` | `perk.trade` | Trade items with other users |
| `HEIGHTMAP_EDITOR_BETA` | `perk.heightmap_editor` | Room heightmap editor |
| `BUILDER_AT_WORK` | `perk.builder` | Builder tools |
| `NAVIGATOR_ROOM_THUMBNAIL_CAMERA` | `perk.room_thumbnail` | Room thumbnail camera |
| `CAMERA` | `perk.camera` | Photo camera feature |
| `MOUSE_ZOOM` | `perk.mouse_zoom` | Mouse zoom in rooms |
| `NAVIGATOR_PHASE_TWO` | `perk.navigator_v2` | Navigator v2 features |
| `SAFE_CHAT` | `perk.safe_chat` | Safe chat mode |
| `HABBO_CLUB_OFFER_BETA` | `perk.club_offer` | Club offer display |

## Resolution

Perks are resolved by `ResolvePerks(access)` after access resolution. The
function iterates the perk registry and checks each permission against the
user's merged grant set. The result is a `[]PerkGrant` array with `IsAllowed`
set accordingly.

## Default Group Grants

With the seeded default groups:

| Group | Perks Granted |
|-------|---------------|
| `default` | SAFE_CHAT, CALL_ON_HELPERS, CITIZEN |
| `vip` | All perks (via `perk.*`) |
| `moderator` | All perks (via `perk.*`) |
| `admin` | All perks (via `*`) |

## Plugin API

Plugins check permissions via `server.Permissions().HasPermission(userID, perm)`.
The same hierarchical resolver is used, meaning `perk.*` grants all
`perk.`-prefixed permissions through the standard wildcard mechanism.
