# Navigator Protocol

## Initialisation Flow

When the client opens the navigator, it sends `navigator.init` (2110). The
server responds with four packets in sequence:

1. **NavigatorMetaData** (3052) — seven default tab entries
2. **NavigatorCollapsed** (1543) — collapsed category list (empty by default)
3. **NavigatorSettings** (518) — window position and dimensions
4. **NavigatorSavedSearches** (3984) — user's saved search entries

## Room Search

The client sends `navigator.search` (249) with a search code and optional
filter string. The server queries rooms matching the filter (ILIKE on name
and tags), groups results into search result blocks, and responds with
`NavigatorSearchResults` (2690).

## Room Creation

1. Client sends `navigator.can_create_room` (2128)
2. Server checks owned room count against MaxRoomsPerm (25)
3. Server responds with `CanCreateRoomResponse` (378)
4. Client sends `navigator.create_room` (2752) with room parameters
5. Server fires `RoomCreating` event (cancellable)
6. Server persists room and fires `RoomCreated` event
7. Server responds with `RoomCreated` (1304)

## Favourite Management

- `navigator.add_favourite` (3817) — adds room to favourites
- `navigator.remove_favourite` (309) — removes room from favourites
- Both send `FavouriteChanged` (2524) and `FavouritesList` (151) on success
- Maximum 30 favourites per user enforced server-side

## Saved Searches

- `navigator.save_search` (2226) — persists a search
- `navigator.delete_search` (1954) — removes a saved search
- Both update the client by resending `NavigatorSavedSearches` (3984)

## Packet ID Reference

### Client-to-Server

| ID | Name |
|----|------|
| 2110 | `navigator.init` |
| 249 | `navigator.search` |
| 2230 | `navigator.get_guest_room` |
| 3027 | `navigator.get_flat_categories` |
| 2128 | `navigator.can_create_room` |
| 2752 | `navigator.create_room` |
| 3817 | `navigator.add_favourite` |
| 309 | `navigator.remove_favourite` |
| 2226 | `navigator.save_search` |
| 1954 | `navigator.delete_search` |
| 3159 | `navigator.save_settings` |

### Server-to-Client

| ID | Name |
|----|------|
| 3052 | `navigator.metadata` |
| 2690 | `navigator.search_results` |
| 1543 | `navigator.collapsed` |
| 3984 | `navigator.saved_searches` |
| 518 | `navigator.settings` |
| 687 | `navigator.guest_room_data` |
| 1562 | `navigator.flat_categories` |
| 378 | `navigator.can_create_room` |
| 1304 | `navigator.room_created` |
| 2524 | `navigator.favourite_changed` |
| 151 | `navigator.favourites_list` |
