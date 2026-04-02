# Navigator

## Overview

The navigator realm is the in-game room browser. It exposes flat categories,
search results, favourites, and saved searches to the client. The server
handles real-time navigation initialisation, room search, room creation,
favourite management, and saved search persistence. Every management action
is mirrored through both an HTTP REST API and a CLI command set.

The realm lives under `pkg/navigator/` and follows the project's standard
hexagonal layout:

```
pkg/navigator/
  domain/             – entities, repository interface, domain errors
  application/        – Service (categories, rooms, search, favourites)
  infrastructure/
    model/            – GORM ORM models
    store/            – PostgreSQL repository implementation
    migration/        – up/down schema migrations
  adapter/
    realtime/         – WebSocket packet handler (Runtime + dispatch)
    httpapi/          – REST route handlers and OpenAPI specs
    command/          – Cobra CLI subcommands
  packet/             – packet-ID constants and encoder types
```

## Architecture

The navigator service is wired during `serve` startup inside
`core/cli/serve_economy.go`. Dependency injection is explicit: the HTTP module,
WebSocket runtime, and CLI tree each receive the same application service
instance.

```
navigatorapp.Service ←── navigatorstore.Repository ←── PostgreSQL
      │
      ├── adapter/realtime.Runtime  (WebSocket)
      ├── adapter/httpapi           (REST)
      └── adapter/command           (CLI)
```

## Domain Model

### Category

A `Category` represents one navigable section in the room browser.

| Field | Type | Description |
|-------|------|-------------|
| ID | int | Stable identifier |
| Caption | string | Display name |
| Visible | bool | Client visibility |
| OrderNum | int | Display sort position |
| IconImage | int | Client icon index |
| CategoryType | string | Classification key |

### Room

A `Room` represents one user-created room.

| Field | Type | Description |
|-------|------|-------------|
| ID | int | Stable identifier |
| OwnerID | int | Creator user ID |
| OwnerName | string | Creator display name |
| Name | string | Room display name |
| Description | string | Room description |
| State | string | Access state (open/locked/password) |
| CategoryID | int | Navigator category reference |
| MaxUsers | int | Room capacity |
| CurrentUsers | int | Active occupants |
| Score | int | Star rating |
| Tags | []string | Searchable tags |
| TradeMode | int | Trade policy code |

### SavedSearch

A `SavedSearch` stores one per-user saved navigator search filter.

### Favourite

A `Favourite` links one user to one room they have favourited. The per-user
limit is hardcoded at 30.

## SDK Events

All room and favourite mutations fire SDK events following the cancellable
before/after pattern:

| Before Event | After Event | Trigger |
|-------------|-------------|---------|
| `RoomCreating` | `RoomCreated` | Room persistence |
| `RoomDeleting` | `RoomDeleted` | Room deletion |
| `FavouriteAdding` | `FavouriteAdded` | Favourite creation |

Events live under `sdk/events/navigator/` with one event per file.
