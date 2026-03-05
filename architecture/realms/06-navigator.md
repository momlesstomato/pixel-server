# Realm: Navigator

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 50 | **Phase:** 4 (Navigator) | **Packets:** 55 (37 c2s, 18 s2c)
> **Services:** navigator | **Status:** Not yet implemented

---

## Overview

The Navigator realm handles room discovery, search, favourites, categories, room creation, and all UI state for the navigator window. With 37 c2s packets, it is the most client-request-heavy realm -- reflecting the complexity of the navigator UI's search tabs, saved searches, category modes, and room card displays. This realm is implemented before furniture (Phase 4 before Phase 6) because engineers need it to test room entry during development.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 4

---

## Packet Inventory

### C2S (Client to Server) -- 37 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2110 | `navigator.init` | _(none)_ | Initialize navigator, receive all state |
| 249 | `navigator.search` | `code:string`, `data:string` | Search rooms in a category |
| 637 | `navigator.search_open` | `code:string` | Mark search tab as expanded |
| 1834 | `navigator.search_close` | `code:string` | Mark search tab as collapsed |
| 2226 | `navigator.search_save` | `code:string`, `data:string` | Save a search as bookmark |
| 1954 | `navigator.search_delete` | `id:int32` | Delete saved search |
| 3159 | `navigator.settings_save` | `x:int32`, `y:int32`, `width:int32`, `height:int32`, `openSearches:boolean` | Save navigator window position/size |
| 1202 | `navigator.category_mode` | `category:string`, `mode:int32` | Toggle category list/grid mode |
| 3817 | `navigator.favourite_add` | `roomId:int32` | Add room to favourites |
| 309 | `navigator.favourite_remove` | `roomId:int32` | Remove room from favourites |
| 2230 | `navigator.get_room_info` | `roomId:int32`, `enterRoom:boolean`, `forwardRoom:boolean` | Request room info card |
| 2128 | `navigator.can_create_room` | _(none)_ | Check if user can create rooms |
| 2752 | `navigator.create_room` | `name:string`, `description:string`, `model:string`, `category:int32`, `maxUsers:int32`, `tradeMode:int32` | Create new room |
| 3027 | `navigator.get_flat_cats` | _(none)_ | Request flat (user room) categories |
| 39 | `navigator.my_guild_bases_search` | _(none)_ | Search rooms with user's guilds |
| 172 | `navigator.forward_to_competition_room` | _(none)_ | Navigate to competition room |
| 272 | `navigator.my_room_rights_search` | _(none)_ | Search rooms where user has rights |
| 314 | `navigator.convert_global_room_id` | `globalId:string` | Convert global room ID to local |
| 433 | `navigator.competition_room_search` | `data:string` | Search competition rooms |
| 685 | `navigator.go_to_flat` | `roomId:int32` | Navigate directly to a room |
| 865 | `navigator.forward_to_random_competition_room` | _(none)_ | Random competition room |
| 1002 | `navigator.my_frequent_room_history_search` | _(none)_ | Search recently visited rooms |
| 1229 | `navigator.official_rooms` | _(none)_ | Request official/staff rooms |
| 1450 | `navigator.forward_to_submittable_room` | _(none)_ | Navigate to submittable room |
| 1669 | `navigator.getpopularroomtagsmessage` | _(none)_ | Get popular room tags |
| 1786 | `navigator.promote_room` | room + promotion fields | Create room promotion |
| 1874 | `navigator.edit_room_promotion` | promotion fields | Edit existing promotion |
| 2023 | `navigator.getuserflatcats` | _(none)_ | Get user flat categories (alternate) |
| 2166 | `navigator.myroommessage` | _(none)_ | Get own rooms |
| 2439 | `navigator.flatcreated` | _(none)_ | Notify room created |
| 2608 | `navigator.can_create_room_event` | _(none)_ | Check if user can create events |
| 2762 | `navigator.myroomssearchmessage` | _(none)_ | Search own rooms |
| 2875 | `navigator.updateroomfilter` | filter fields | Update room filter |
| 3376 | `navigator.getroomfiltermessage` | `roomId:int32` | Get room filter config |
| 3582 | `navigator.roomadvertisedisappear` | `roomId:int32` | Dismiss room advertisement |
| 3849 | `navigator.settingsmessage` | _(none)_ | Get navigator settings |
| 3902 | `navigator.roomtextmessage` | text fields | Room text search |

### S2C (Server to Client) -- 18 packets

| ID | Name | Key Fields | Summary |
|----|------|------------|---------|
| 2690 | `navigator.metadata` | `topLevelContexts[]`, `savedSearches[]` | Navigator initialization data |
| 2523 | `navigator.search_results` | `code`, `data`, `results[]` (room objects with full details) | Search results |
| 1730 | `navigator.search_saved` | `savedSearches[]` | Saved search list |
| 1200 | `navigator.room_info` | full room object | Room info card data |
| 2466 | `navigator.flat_cats` | `categories[]` | Room categories list |
| 1304 | `navigator.settings` | `x`, `y`, `width`, `height`, `openSearches` | Navigator window settings |
| 378 | `navigator.can_create_room` | `result:int32`, `maxRooms:int32` | Room creation eligibility |
| 1304 | `navigator.create_room_result` | `roomId`, `roomName` | Created room confirmation |
| 3049 | `navigator.popular_tags` | `tags[]` | Popular room tags |
| 1840 | `navigator.favourite_changed` | `roomId`, `added:boolean` | Favourite toggle confirmation |
| 2726 | `navigator.room_rating` | `roomId`, `score` | Room rating update |
| 2312 | `navigator.event_categories` | `categories[]` | Event categories list |
| 1577 | `navigator.roomfilterstatus` | filter status | Room filter update |
| 1715 | `navigator.can_create_room_event` | `result:boolean` | Event creation eligibility |
| 1740 | `navigator.collapsed` | `collapsedCategories[]` | Collapsed search tabs |
| 1903 | `navigator.lifted_rooms` | `rooms[]` | Promoted/lifted rooms |
| 2726 | `navigator.room_updated` | room object | Room info updated |
| 3875 | `navigator.guestroomsearchresult` | room search results | Guest room search results |

---

## Architecture Mapping

### Service Ownership

The **navigator module** is a dedicated bounded context that:
- Maintains room metadata cache (room names, owners, user counts, scores).
- Executes search queries against PostgreSQL and Redis.
- Receives room state updates from the game service via NATS.

```
Client ──packet──▶ Gateway ──NATS──▶ Navigator Service
                                          │
                   ◀──NATS(session.output)─┘
                                          │
Game Service ──NATS(navigator.room_updated)──▶ Navigator Service (cache invalidation)
```

### Database Tables

| Table | Columns (Key) | Usage |
|-------|---------------|-------|
| `rooms` | id, owner_id, name, description, model_name, state, users_max, category, score, tags | Room metadata |
| `room_models` | name, door_x, door_y, door_z, door_dir, heightmap, club_only | Room layout templates |
| `navigator_categories` | id, name, min_rank, visible, public | Category definitions |
| `navigator_public_rooms` | room_id, category_id | Featured/public rooms |
| `user_favourites` | user_id, room_id | Favourite room bookmarks |
| `navigator_saved_searches` | user_id, code, data | Saved search bookmarks |
| `room_promotions` | room_id, title, description, category, start_time, end_time | Active promotions |

### Redis Keys

| Key Pattern | Usage |
|-------------|-------|
| `room:users:<roomId>` | Live user count for room |
| `navigator:popular` | Sorted set of rooms by score (cached, TTL: 5min) |
| `navigator:tags:popular` | Sorted set of popular tags |
| `user:favourites:<userId>` | Set of favourite room IDs |

### NATS Subjects

| Subject | Direction | Purpose |
|---------|-----------|---------|
| `navigator.input.<sessionID>` | gateway -> navigator | Incoming navigator packets |
| `session.output.<sessionID>` | navigator -> gateway | Outgoing responses |
| `navigator.room_updated.<roomID>` | game -> navigator | Room user count / state change |
| `navigator.room_created` | navigator -> game | New room created event |

---

## Implementation Analysis

### Search System Architecture

The navigator search is the most complex feature. The `code` field in `navigator.search` determines the search type:

| Code | Meaning | Query Strategy |
|------|---------|----------------|
| `official` | Official/public rooms | `navigator_public_rooms` JOIN `rooms` |
| `popular` | Rooms sorted by current users | Redis sorted set `navigator:popular` |
| `categories` | Rooms in a category | `rooms WHERE category = ?` |
| `my` | User's own rooms | `rooms WHERE owner_id = ?` |
| `favourites` | User's favourite rooms | `user_favourites` JOIN `rooms` |
| `history` | Recently visited | `user_room_history` (last 50 visits) |
| `rights` | Rooms with rights | `room_rights WHERE user_id = ?` JOIN `rooms` |
| `guild` | Guild-associated rooms | `rooms WHERE group_id IN (user's guilds)` |
| `query` | Free-text search | `rooms WHERE LOWER(name) LIKE LOWER(?)` or `tags @> ARRAY[?]` |

The `data` field provides the search query string for `query` type searches.

**Performance strategy:**
- Popular rooms list cached in Redis with 5-minute TTL.
- Free-text search uses PostgreSQL GIN trigram index on `rooms.name`.
- Tag search uses PostgreSQL GIN index on `rooms.tags` array column.
- Results are capped at 100 rooms per search to prevent payload bloat.

### Room Creation Flow

`navigator.create_room` (2752):

1. **Validate** -- name length (3-25 chars), description length (0-128), model exists, category valid, maxUsers valid (10-50 for normal, 75-100 for HC).
2. **Check limit** -- `navigator.can_create_room` checks if user hasn't exceeded max rooms (default: 25, HC: 50).
3. **Create** -- INSERT into `rooms` table with default state ("open"), assign owner.
4. **Respond** -- Send `navigator.create_room_result` with new room ID.
5. **Event** -- Publish `navigator.room_created` to NATS for game service to prepare room worker.
6. **Redirect** -- Client automatically enters the new room.

### Favourite System

Simple toggle with Redis caching:
- `navigator.favourite_add` (3817): INSERT `user_favourites`, add to Redis set.
- `navigator.favourite_remove` (309): DELETE from `user_favourites`, remove from Redis set.
- Max favourites: 30 (configurable).
- Favourites are loaded into Redis on first access and invalidated on change.

### Room Info Card

`navigator.get_room_info` (2230) returns a rich room object with:
- Room metadata (name, description, owner, category, tags)
- Current user count and max capacity
- Room state (open/locked/password/invisible)
- Group association (if any)
- Room promotion (if active)
- User's relationship to the room (owner, has rights, is favourite)

This is also the entry point for room join: if `enterRoom=true`, the navigator service forwards the request to the game service.

---

## Caveats & Edge Cases

### 1. Stale User Counts
Room user counts in search results can be stale if the navigator service cache is not frequently updated. Use Redis `room:users:<roomId>` with frequent updates from game service (every 5 seconds via NATS).

### 2. Invisible Rooms
Rooms with state "invisible" should not appear in any search results except for the owner, users with rights, and moderators. This filter must be applied consistently across all search code paths.

### 3. Search Result Ordering
Different search types have different sort orders:
- `popular`: sort by `users_now` DESC
- `categories`: sort by `score` DESC, then `users_now` DESC
- `my`: sort by `name` ASC
- `favourites`: sort by insertion order (user's preference)
- `query`: sort by relevance (trigram similarity), then `users_now` DESC

### 4. Room Model Validation
When creating rooms, the `model` name must reference a valid `room_models` entry. Custom room models (user-created heightmaps) are a Phase 3 feature -- in Phase 4, only predefined models are allowed.

### 5. Navigator Window State
The navigator saves UI state (window position, size, collapsed tabs) per user. These are purely cosmetic and should be persisted lazily -- batch writes, not per-interaction.

### 6. Room Promotion Expiry
Promotions have an `end_time`. The navigator must filter out expired promotions and not display them. A background job should clean up expired promotions periodically (every 15 minutes).

### 7. Category Rank Restrictions
Some navigator categories have `min_rank` requirements. The navigator service must check the user's rank when filtering categories. Categories with `visible=false` are hidden from the UI but accessible via direct search.

### 8. Concurrent Room Creation
Two users creating rooms simultaneously could exceed the room limit if the check and create are not atomic. Use a PostgreSQL advisory lock or `INSERT ... SELECT ... WHERE (SELECT COUNT(*) FROM rooms WHERE owner_id = ?) < max_rooms`.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Search** | In-memory room list scan | PostgreSQL GIN indexes + Redis cache |
| **Popular rooms** | Computed on every request | Cached sorted set, 5-min TTL |
| **User counts** | Stale (updated on room join/leave only) | Real-time Redis counters via NATS |
| **Room creation** | Synchronous DB + redirect | Async event-driven with NATS |
| **Favourites** | DB query on every navigator open | Redis-cached set with write-through |
| **Scalability** | Single-process bottleneck | Independent navigator service |

---

## Dependencies

- **Phase 2 (Identity)** -- user rank for category access, user ID for ownership
- **Phase 3 (Room)** -- room entry mechanics (navigator triggers room join)
- **pkg/navigator** -- domain models (Category, SavedSearch, RoomPromotion)
- **PostgreSQL** -- rooms, categories, favourites, promotions tables
- **Redis** -- user counts, popular rooms cache, favourite sets

---

## Testing Strategy

### Unit Tests
- Search query routing (code -> query strategy)
- Room creation validation (name, model, limits)
- Favourite add/remove logic
- Category filtering by rank
- Invisible room exclusion from results

### Integration Tests
- Full search flow against real PostgreSQL (testcontainers)
- Room creation with limit enforcement
- Favourite persistence and Redis cache consistency
- Popular rooms sorted set accuracy
- Room promotion expiry filtering

### E2E Tests
- Client opens navigator, sees categories and popular rooms
- Client searches by room name, gets correct results
- Client creates room, navigator shows it, client enters it
- Client adds/removes favourites, sees updates across sessions
