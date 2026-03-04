# Pathfinding — Algorithm

A* with 3D height-aware movement costs, an octile heuristic, and an optional
diagonal movement mode.

---

## FindPath

```go
func FindPath(l *Layout, from, to Tile, opts Options) []PathStep
```

Returns an ordered `[]PathStep` from `from` to `to`, **excluding** the start
tile. Returns `nil` when no path exists.

| Parameter | Description |
|---|---|
| `l` | Pre-built Layout for the room |
| `from` | Starting tile (position + Z at time of request) |
| `to` | Destination tile |
| `opts` | Per-request behavioural options |

Early exits:
- `from` or `to` is outside `l`'s bounds → `nil`
- Destination tile is `TileBlocked` and `opts.Flying` is false → `nil`

---

## Options

```go
type Options struct {
    AllowDiagonal       bool
    Flying              bool
    WalkthroughEntities bool
}
```

| Field | Default | Effect |
|---|---|---|
| `AllowDiagonal` | `false` | Enables 8-direction movement (4+diagonal). Diagonal steps cost `CostDiagonal` |
| `Flying` | `false` | Entity may traverse `TileBlocked` tiles at flat cost |
| `WalkthroughEntities` | `false` | Entity ignores other entities occupying tiles (used by ghosts, bots) |

---

## Cost Constants

| Constant | Value | Applied when |
|---|---|---|
| `CostFlat` | `1.0` | Cardinal (N/S/E/W) move on same Z level |
| `CostDiagonal` | `1.414` | Diagonal move on same Z level (≈ √2) |
| `CostClimb` | `1.5` | Multiplied by positive `dZ` and added to base cost |
| `CostDescend` | `0.8` | Multiplied by `abs(dZ)` and added to base cost for downward steps |
| `MaxStepUp` | `1.1` | Maximum allowed `dZ` per step upward; exceeding returns cost -1 (blocked) |
| `MaxStepDown` | `2.0` | Maximum allowed `abs(dZ)` per step downward; exceeding returns cost -1 |

---

## Movement Cost Formula

```
moveCost(from, to, diagonal, opts):
    if Flying:
        return CostDiagonal if diagonal else CostFlat

    dZ = to.Z - from.Z
    if dZ > MaxStepUp:  return -1  (blocked — too steep to climb)
    if dZ < -MaxStepDown: return -1 (blocked — drop too large)

    base = CostDiagonal if diagonal else CostFlat

    if dZ > 0:  return base + dZ * CostClimb
    if dZ < 0:  return base + abs(dZ) * CostDescend
    return base
```

A return value of `-1` causes the neighbour to be skipped entirely.

---

## Heuristic

`heuristic3d` uses the **octile** distance formula, which is admissible for
8-directional movement:

```
dx = abs(a.X - b.X)
dy = abs(a.Y - b.Y)
diag = min(dx, dy)
straight = (dx + dy) - 2 * diag
h = CostFlat * straight + CostDiagonal * diag
```

The Z axis is not included in the heuristic; height cost is accounted for in
`moveCost` during actual expansion.

---

## Internal Data Structures

The algorithm uses **flat slices** indexed by `y*width + x` instead of
pointer-based structs to minimise GC pressure during per-tick pathfinding.

| Slice | Type | Purpose |
|---|---|---|
| `gScore` | `[]float32` | Best known cumulative cost to each tile |
| `parent` | `[]int` | Flat index of the predecessor tile for path reconstruction |
| `closed` | `[]bool` | Whether a tile has been settled (expanded) |

The open set is a standard `container/heap` min-heap of `*astarNode` structs,
ordered by `f = g + h`.

```go
type astarNode struct {
    x, y int
    g, f float32
}
```

---

## Path Reconstruction

`reconstruct` walks the `parent` slice backwards from `goalIdx` to the start,
collecting `PathStep` values, then reverses the slice:

```
path = []
ci = goalIdx
while ci != -1:
    append PathStep{X: ci%w, Y: ci/w, Z: Tiles[y][x].Z}
    ci = parent[ci]
reverse(path)
return path  // start tile is the first element and is stripped by the caller
```

The start tile is excluded from the returned slice: the caller's entity is
already standing there.
