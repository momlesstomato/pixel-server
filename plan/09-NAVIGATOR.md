# 09 - Navigator Realm

## Overview

The Navigator realm owns room browsing, creation, favourites, saved searches,
and flat categories. It is the primary discovery mechanism for rooms, allowing
users to search, filter, and organise the room listing.

The pixel-protocol references **11 C2S packets** and **11 S2C packets** for
this realm. All four vendor emulators implement the core navigator flow.
PlusEMU provides the most complete implementation including saved searches
and event categories.

---

## Vendor Cross-Reference

### Navigator Feature Matrix

| Feature | PlusEMU (C#) | Arcturus (Java) | comet-v2 (Java) | pixelsv (impl) |
|---------|-------------|-----------------|-----------------|----------------|
| Navigator tabs | Hardcoded 7 | DB-driven | Hardcoded | **Hardcoded 7** |
| Room categories | DB-driven | DB-driven | DB-driven | **DB-driven** |
| Saved searches | Yes | Yes | No | **Yes** |
| Room creation limit | Configurable | Hardcoded 25 | Hardcoded 25 | **Hardcoded 25** |
| Favourite limit | Hardcoded 30 | Hardcoded 30 | Hardcoded 30 | **Hardcoded 30** |
| Room search | LIKE on name | LIKE on name+owner | LIKE on name | **ILIKE name+tags** |
| Room tags | Comma-separated | JSON array | List | **Comma-separated** |
| Trade mode | Int 0-2 | Int 0-2 | Int 0-2 | **Int 0-2** |
| Collapsed persistence | DB column | Redis | None | **Noop (deferred)** |
| Settings persistence | DB column | DB column | None | **Noop (deferred)** |
| Event categories | Yes | Yes | No | **Deferred** |

### Database Schema Comparison

| Aspect | PlusEMU | Arcturus | comet-v2 | pixelsv |
|--------|---------|----------|----------|---------|
| Rooms table | `rooms` | `rooms` | `rooms` | `rooms` |
| Categories table | `navigator_categories` | `navigator_categories` | `navigator_flatcats` | `navigator_categories` |
| Favourites table | `user_favorites` | `users_favorites` | `room_favorites` | `navigator_favourites` |
| Saved searches | `navigator_searches` | `navigator_searches` | None | `navigator_saved_searches` |
| Room PK | `id` auto | `id` auto | `id` auto | `id` auto |

### Our Improvements Over Vendors

1. **ILIKE search on name and tags** — vendors only search by room name.
   We also search tag content for better discoverability.
2. **Proper pagination** — vendors return all rooms. We return paginated
   results with offset and limit for scalability.
3. **Plugin event system** — room creation, deletion, and favourite mutations
   fire cancellable SDK events for plugin interception.
4. **Full REST API parity** — every navigator operation is available via both
   realtime packets and HTTP REST endpoints with OpenAPI documentation.
5. **CLI admin commands** — category listing and room inspection available
   via Cobra CLI for operational tooling.

---

## Packet Registry

### Client-to-Server (11 packets)

| ID | Name | Fields | Status |
|----|------|--------|--------|
| 2110 | `navigator.init` | (empty) | **Done** |
| 249 | `navigator.search` | searchCode (string), filter (string) | **Done** |
| 2230 | `navigator.get_guest_room` | roomId (int32), forward (int32), enter (int32) | **Done** |
| 3027 | `navigator.get_flat_categories` | (empty) | **Done** |
| 2128 | `navigator.can_create_room` | (empty) | **Done** |
| 2752 | `navigator.create_room` | name, desc, state, catId, maxUsers, tradeMode, tags[] | **Done** |
| 3817 | `navigator.add_favourite` | roomId (int32) | **Done** |
| 309 | `navigator.remove_favourite` | roomId (int32) | **Done** |
| 2226 | `navigator.save_search` | searchCode (string), filter (string) | **Done** |
| 1954 | `navigator.delete_search` | searchId (int32) | **Done** |
| 3159 | `navigator.save_settings` | x, y, w, h, hidden, mode | **Noop** |

### Server-to-Client (11 packets)

| ID | Name | Fields | Status |
|----|------|--------|--------|
| 3052 | `navigator.metadata` | topLevelContexts[] | **Done** |
| 2690 | `navigator.search_results` | searchCode, filter, blocks[] | **Done** |
| 1543 | `navigator.collapsed` | categories[] | **Done** |
| 3984 | `navigator.saved_searches` | searches[] | **Done** |
| 518 | `navigator.settings` | x, y, w, h, hidden, mode | **Done** |
| 687 | `navigator.guest_room_data` | room data + extended flags | **Done** |
| 1562 | `navigator.flat_categories` | categories[] | **Done** |
| 378 | `navigator.can_create_room` | resultCode, maxRooms | **Done** |
| 1304 | `navigator.room_created` | roomId, name | **Done** |
| 2524 | `navigator.favourite_changed` | roomId, added | **Done** |
| 151 | `navigator.favourites_list` | maxFavourites, roomIds[] | **Done** |

---

## Architecture

### Package Layout

```
pkg/navigator/
  domain/
    errors.go          - 6 domain errors
    navigator.go       - Category, SavedSearch structs
    room.go            - Room, Favourite structs
    repository.go      - 16-method Repository interface, RoomFilter, RoomPatch
  application/
    service.go         - Service struct, category CRUD
    room_service.go    - Room CRUD with event firing
    search_service.go  - Search + favourite methods with event firing
    tests/             - 4 test files (service, search, event, stub)
  infrastructure/
    model/
      navigator.go     - GORM Category, SavedSearch models
      room.go          - GORM Room, Favourite models
    store/
      repository.go    - Store struct, mappers, tag helpers
      category_store.go - Category persistence
      room_store.go    - Room persistence with ILIKE search
      search_store.go  - SavedSearch + Favourite persistence
    migration/
      migration.go     - 4 migration steps
  packet/
    constants.go       - 11 C2S + 11 S2C packet IDs
    navigator_packet.go - Metadata, Collapsed, Settings, SavedSearches
    search_packet.go   - SearchResults, GuestRoomData, EncodeRoomData
    room_packet.go     - FlatCategories, CanCreateRoom, RoomCreated, Favourites
    tests/             - 2 test files (navigator_test, room_test)
  adapter/
    realtime/
      runtime.go       - Runtime struct, NewRuntime, sendPacket
      dispatch.go      - Handle() dispatch, init/search/category handlers
      dispatch_rooms.go - Room/favourite handlers
    httpapi/
      contracts.go     - Service interface
      routes.go        - REST route registration
      openapi.go       - OpenAPI path definitions
    command/
      command.go       - CLI command factory
      actions.go       - Category list, room get actions
```

### SDK Events

| Event | Type | Trigger |
|-------|------|---------|
| `RoomCreating` | Cancellable | Before room persistence |
| `RoomCreated` | After | After room persistence |
| `RoomDeleting` | Cancellable | Before room deletion |
| `RoomDeleted` | After | After room deletion |
| `FavouriteAdding` | Cancellable | Before favourite creation |
| `FavouriteAdded` | After | After favourite creation |

### Database Migrations

| ID | Name | Tables |
|----|------|--------|
| 20260325_13 | NavigatorCategories | `navigator_categories` |
| 20260325_14 | Rooms | `rooms` |
| 20260325_15 | SavedSearches | `navigator_saved_searches` |
| 20260325_16 | Favourites | `navigator_favourites` |

---

## Testing

### Unit Tests (12 tests)

- `TestNewServiceRejectsNilRepository` — constructor validation
- `TestServiceCategoryCRUD` — category create, find, list, delete
- `TestServiceRoomCRUD` — room create, find, list, update, delete
- `TestServiceSavedSearchCRUD` — saved search create, list, delete
- `TestServiceFavouriteCRUD` — favourite add, list, remove
- `TestServiceFavouriteLimitReached` — max favourite enforcement
- `TestRoomCreateFiresEvents` — room creation event dispatch
- `TestRoomCreateCancelledByPlugin` — room creation cancellation
- `TestRoomDeleteFiresEvents` — room deletion event dispatch
- `TestRoomDeleteCancelledByPlugin` — room deletion cancellation
- `TestFavouriteAddFiresEvents` — favourite add event dispatch
- `TestFavouriteAddCancelledByPlugin` — favourite add cancellation

### Packet Tests (10 tests)

- `TestNavigatorMetaDataPacketEncode` — metadata serialization
- `TestNavigatorCollapsedPacketEncode` — collapsed serialization
- `TestNavigatorSettingsPacketEncode` — settings serialization
- `TestNavigatorSavedSearchesPacketEncode` — saved searches serialization
- `TestNavigatorMetaDataPacketEncodeEmpty` — empty metadata
- `TestFlatCategoriesPacketEncode` — flat categories serialization
- `TestCanCreateRoomResponsePacketEncode` — can create room response
- `TestRoomCreatedPacketEncode` — room created serialization
- `TestFavouriteChangedPacketEncode` — favourite changed serialization
- `TestFavouritesListPacketEncode` — favourites list serialization

### E2E Tests (5 tests)

- `Test11NavigatorCategoryFlow` — full category lifecycle
- `Test11NavigatorRoomFlow` — full room lifecycle with update
- `Test11NavigatorFavouriteFlow` — favourite add and remove
- `Test11NavigatorSavedSearchFlow` — saved search lifecycle
- `Test11NavigatorRoomSearchFilter` — ILIKE search (PostgreSQL-only, skipped on SQLite)

---

## Deferred Items

- **Navigator settings persistence** — save_settings handler is a noop. Requires
  a user_navigator_settings table in a future migration.
- **Event categories** — PlusEMU and Arcturus support "event" type categories.
  Deferred until room events are implemented.
- **Collapsed category persistence** — currently returns empty list on init.
  Requires per-user preference storage.
- **Room enter/forward flow** — GetGuestRoom sends room data but does not
  teleport the user. Requires room instance management (Milestone 3.x).
- **Room password/doorbell** — state "password" and "locked" are stored but
  not enforced at realtime level.
