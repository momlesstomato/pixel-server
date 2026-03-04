# Room — Data Models

All persistent domain structs, the repository interface, and the in-memory
implementation for the room realm.

---

## Room (`pkg/room.Room`)

Represents a single guest room. Persisted in PostgreSQL; in-memory during
development and tests.

| Field | Type | Description |
|---|---|---|
| `ID` | `int32` | Unique room identifier (database primary key) |
| `OwnerID` | `int32` | User ID of the room owner |
| `OwnerName` | `string` | Username of the owner (denormalised for display) |
| `Name` | `string` | Public room name shown in the navigator |
| `Description` | `string` | Room description text |
| `ModelName` | `string` | References a static `Model` record |
| `Password` | `string` | Entry password (only checked when `State == "password"`) |
| `State` | `string` | Access mode: `"open"`, `"locked"`, `"password"`, `"invisible"` |
| `UsersNow` | `int32` | Current occupancy count (updated each tick) |
| `UsersMax` | `int32` | Maximum allowed occupancy |
| `Category` | `int32` | Navigator category ID |
| `Score` | `int32` | Aggregate like/vote score |
| `PaperFloor` | `string` | Floor decoration code |
| `PaperWall` | `string` | Wall decoration code |
| `PaperLandscape` | `string` | Landscape decoration code |
| `FloorThickness` | `int32` | Floor tile thickness index |
| `WallThickness` | `int32` | Wall thickness index |
| `WallHeight` | `int32` | Wall height offset |
| `HideWall` | `bool` | Whether walls are hidden |
| `AllowPets` | `bool` | Whether pets may enter |
| `AllowPetsEat` | `bool` | Whether pets may eat placed food items |
| `AllowWalkthrough` | `bool` | Whether avatars may walk through each other |
| `ChatMode` | `int32` | Chat bubble style mode |
| `ChatWeight` | `int32` | Chat bubble weight/size index |
| `ChatSpeed` | `int32` | Chat bubble display speed |
| `ChatHearRange` | `int32` | Chat tile-distance hearing range |
| `ChatProtection` | `int32` | Flood protection setting |
| `TradeMode` | `int32` | Trading permission level |
| `RollerSpeed` | `int32` | Roller furniture movement speed |
| `MuteOption` | `int32` | Who can mute (0=owner, 1=rights) |
| `KickOption` | `int32` | Who can kick (0=owner, 1=rights) |
| `BanOption` | `int32` | Who can ban (0=owner, 1=rights) |
| `Tags` | `string` | Comma-separated searchable tags |
| `Group` | `int32` | Associated group ID (0 = none) |

---

## Model (`pkg/room.Model`)

Stores the static, read-only layout of a room. Loaded once per room type;
never mutated while a room is running.

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Model identifier key (referenced by `Room.ModelName`) |
| `DoorX` | `int32` | Spawn tile X coordinate |
| `DoorY` | `int32` | Spawn tile Y coordinate |
| `DoorZ` | `float64` | Spawn tile Z height (stack layer) |
| `DoorDir` | `int32` | Spawn facing direction (0–7, N=0, NE=1 …) |
| `Heightmap` | `string` | Multi-line ASCII tile map (see PATHFINDING/LAYOUT.MD) |
| `ClubOnly` | `bool` | Whether Habbo Club membership is required to enter |

---

## Sentinel Errors

| Symbol | Package | Meaning |
|---|---|---|
| `room.ErrNotFound` | `pkg/room` | Returned by any repository method when no record matches the query |

---

## Repository Interface (`pkg/room.Repository`)

```go
GetByID(ctx, id int32) (*Room, error)
GetByOwner(ctx, ownerID int32) ([]*Room, error)
Create(ctx, r *Room) error
Update(ctx, r *Room) error
Delete(ctx, id int32) error
GetModel(ctx, name string) (*Model, error)
Search(ctx, query string, limit int) ([]*Room, error)
GetPopular(ctx, limit int) ([]*Room, error)
```

All methods take `context.Context` as the first argument. `ErrNotFound` is
returned (not a DB-level error) when no record matches.

---

## In-Memory Implementation (`pkg/room/memory.RoomRepo`)

Used in unit tests and local development. No persistence across restarts.

**Backing store:** `map[int32]*Room` protected by a `sync.RWMutex`. IDs are
assigned via an atomic counter when `Create` is called with `ID == 0`.

```
Constructor: memory.NewRoomRepo()
```

The type checks `room.Repository` at compile time:
```go
var _ room.Repository = (*RoomRepo)(nil)
```

`GetModel` always returns `room.ErrNotFound` — Model records must be seeded
through a concrete PostgreSQL implementation in production.

`Search` performs a case-insensitive substring match on `Name`.

`GetPopular` returns rooms in arbitrary insertion order, up to `limit`.
