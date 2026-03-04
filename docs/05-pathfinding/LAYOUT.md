# Pathfinding — Layout

The `Layout` type represents a room's tile grid before pathfinding runs.
It is produced once from the static `Model.Heightmap` string and reused for
the lifetime of the room.

---

## TileState (`pkg/pathfinding.TileState`)

Describes the passability of a single tile.

| Constant | Value | Meaning |
|---|---|---|
| `TileOpen` | `0` | Passable floor tile |
| `TileBlocked` | `1` | Wall or otherwise impassable tile |
| `TileSeat` | `2` | Passable, but movement terminates here (sit) |
| `TileBed` | `3` | Passable, but movement terminates here (lay) |

---

## Tile (`pkg/pathfinding.Tile`)

Represents a single cell in the tile grid.

| Field | Type | Description |
|---|---|---|
| `X` | `int16` | Column index, 0-based |
| `Y` | `int16` | Row index, 0-based |
| `Z` | `float32` | Stack height (0.0 = ground floor, increments for each stacking level) |
| `State` | `TileState` | Passability constant |

---

## PathStep (`pkg/pathfinding.PathStep`)

A waypoint returned by `FindPath`. Mirrors the same type in `pkg/room` but
lives in the pathfinding package to keep it dependency-free.

| Field | Type | Description |
|---|---|---|
| `X` | `int16` | Tile column |
| `Y` | `int16` | Tile row |
| `Z` | `float32` | Stack height at this waypoint |

---

## Layout (`pkg/pathfinding.Layout`)

The pre-computed tile grid for a room.

| Field | Type | Description |
|---|---|---|
| `Width` | `int` | Number of tile columns |
| `Height` | `int` | Number of tile rows |
| `Tiles` | `[][]Tile` | Row-major 2D slice: `Tiles[y][x]` |

### `NewLayout(width, height int) *Layout`

Creates an empty Layout with all tiles initialised to `TileOpen` at `Z = 0.0`.

### `InBounds(x, y int) bool`

Returns `true` if `0 <= x < Width && 0 <= y < Height`. Called before any
tile access to prevent out-of-bounds panics in the hot path.

### `At(x, y int) *Tile`

Returns a pointer to the tile at column `x`, row `y`. Panics if out of bounds —
always guard with `InBounds` first.

---

## ParseHeightmap

```go
func ParseHeightmap(hm string) *Layout
```

Converts a multi-line ASCII heightmap string from `Model.Heightmap` into a
fully populated `*Layout`.

### Row Encoding

Each line in the string corresponds to one row (Y axis). Lines are separated
by `'\n'`. Empty lines are skipped. The width is inferred from the first
non-empty row.

### Character Encoding

| Character | Tile state | Z value |
|---|---|---|
| `x` or `X` | `TileBlocked` | `0.0` |
| `0` – `9` | `TileOpen` | `0.0` – `9.0` |
| `a` – `z` | `TileOpen` | `10.0` – `35.0` |
| Any other | `TileBlocked` | `0.0` |

**Example:**

```
00000
0xxx0
0x0x0
0xxx0
00000
```

Produces a 5×5 layout with a blocked inner ring forming a hollow rectangle.

### Edge Cases

- An empty or all-whitespace heightmap returns a `NewLayout(0, 0)`.
- Rows shorter than the first row leave remaining tiles at their default
  `TileOpen, Z=0` state.
