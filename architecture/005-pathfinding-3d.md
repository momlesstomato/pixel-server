# 3D Pathfinding

## Current state (legacy emulators)

The existing `RoomLayout.findPath` (Arcturus) performs a 2D A* search on `(x, y)` tile coordinates. Z (stack height) appears only as a hard-coded step-height cut-off:

```java
double height = currentAdj.getStackHeight() - current.getStackHeight();
if (height > MAXIMUM_STEP_HEIGHT) {
    closedList.add(currentAdj);   // reject tile: too high
    openList.remove(currentAdj);
}
```

This means:
- **All tiles at the same height are treated as equivalent.** An entity navigating from z=0 to a staircase that reaches z=3 must pick a particular entry tile, but the cost of climbing vs. staying flat is not reflected in the path.
- **Multi-floor rooms** (e.g. rooms with stacked items creating raised platforms) are not handled: path cost does not increase for climbing, so the "cheapest" path may route through unnecessarily tall obstacles.
- **Flying entities** (effects that suspend gravity) cannot follow a different cost surface.

### Open-list performance problem

The open list is a `LinkedList` with linear `lowestFInOpen` scan:

```java
private RoomTile lowestFInOpen(Collection<RoomTile> openList) {
    RoomTile cheapest = null;
    for (RoomTile tile : openList) {  // O(n) every iteration
        if (cheapest == null || tile.getfCosts() < cheapest.getfCosts())
            cheapest = tile;
    }
    return cheapest;
}
```

For a 64×64 room (4096 tiles) with a complex blocked layout this is O(n²) in open-list size. A trivial fix is a min-heap (Go's `container/heap`). The full fix is JPS.

---

## Design goals for pixel-server

1. **True X-Y-Z cost** – climbing costs more than flat movement; descending may cost less; the path naturally routes around tall obstacles.
2. **Staircase traversal** – consecutive tiles with rising Z are treated as a ramp; the path prefers them over a vertical jump when both reach the same destination.
3. **Flying mode** – an entity with the "flying" flag (effect ID) ignores floor height entirely; cost function becomes flat Euclidean.
4. **Sub-millisecond** – a 64×64 room producing a path across ~100 tiles must complete in < 500 µs on a commodity core.
5. **Deterministic** – given the same heightmap and entity positions, the same path is always produced. No randomness; the system is fully reproducible in tests.

---

## Tile representation

```go
// pkg/pathfinding/tile.go

type Tile struct {
    X, Y  int16
    Z     float32  // stack height (0.0 = floor, 1.0 = one furniture height, etc.)
    State TileState // Open | Blocked | Seat | Bed
}

type TileState uint8

const (
    TileOpen    TileState = iota
    TileBlocked           // wall or occupied
    TileSeat              // passable but terminal (sit)
    TileBed               // passable but terminal (lay)
)
```

Heights are `float32` because Habbo uses fractional heights (0.5, 0.25, 1.5, etc.) produced by stacked furniture. The heightmap is pre-computed when the room loads and invalidated on item placement/removal.

---

## 3D A* algorithm

### Node

```go
type node struct {
    x, y    int16
    z       float32
    g, h, f float32
    parent  *node
}
```

### Cost functions

```go
const (
    CostFlat     = 1.0
    CostDiagonal = 1.414 // √2
    CostClimb    = 1.5   // per unit of positive Δz
    CostDescend  = 0.8   // per unit of negative Δz (easier)
    MaxStepUp    = 1.1   // maximum Δz for a single step (matches Habbo spec)
    MaxStepDown  = 2.0   // can drop further than climbing
)

func moveCost(from, to *Tile, diagonal bool) float32 {
    dz := to.Z - from.Z
    var base float32
    if diagonal {
        base = CostDiagonal
    } else {
        base = CostFlat
    }
    if dz > 0 {
        return base + dz*CostClimb
    }
    return base + abs32(dz)*CostDescend
}
```

### Heuristic

Octile distance extended to 3D (admissible, consistent):

```go
func heuristic(a, b *Tile) float32 {
    dx := abs32(float32(a.x - b.x))
    dy := abs32(float32(a.y - b.y))
    dz := abs32(a.z - b.z)
    diag := min32(dx, dy)
    straight := (dx + dy) - 2*diag
    return CostFlat*straight + CostDiagonal*diag + dz*CostClimb
}
```

### Open set

A binary min-heap keyed by `f` score. Go standard library's `container/heap` with a pre-allocated backing slice (size `width*height`, reset per call via index array). No allocations in the hot loop.

### Closed set

A flat `[]bool` of size `width*height`. Index `y*width + x` is set when a tile is closed. Reset by maintaining a "generation counter" (one `uint32` increment invalidates the whole map without zeroing, using a parallel `[]uint32 closedGen` slice).

---

## Jump Point Search (JPS) for open rooms

For rooms with large open areas (a common case in Habbo), A* with standard neighbor enumeration is suboptimal. JPS prunes symmetric paths by identifying "jump points" — tiles where a turn is forced — and skips straight runs entirely.

JPS is applied **only on flat ground** (constant Z tiles). When a vertical transition is detected (Δz ≠ 0), the algorithm falls back to standard 8-directional A*.

This hybrid approach gives JPS speed on the common open floor case and correctness everywhere else.

```
Flat run (JPS active):   O(k) where k = number of jump points
Staircase / platform:    O(n log n) standard A*
Combined cost estimate:  sub-100 µs for 64×64 rooms in practice
```

---

## Hierarchical Pre-computed Abstraction (HPA*)

For very large rooms (128×128+) or room templates with complex multi-level layouts, a second-level HPA* layer is pre-computed at room load time.

### Cluster graph

The room is divided into `16×16` clusters. Entrance/exit points between adjacent clusters are identified. A cluster-level graph is built with edges weighted by the intra-cluster A* cost between entrance points.

```
Room load:
  1. Partition room into 16×16 clusters.
  2. For each pair of adjacent clusters, find all border tile pairs (x,y)-(x±1,y) or (x,y)-(x,y±1).
  3. Run intra-cluster A* between pairs to compute edge weight.
  4. Store as adjacency list: clusterGraph[clusterID] → []Edge{clusterID, weight}.

Path request:
  1. Use cluster graph to get approximate corridor: [C1, C3, C7, C12, C_goal].
  2. Run tile-level A* only within the clusters on the corridor.
  3. Stitch sub-paths together.
```

For rooms < 32×32 (the majority), HPA* is skipped; plain JPS-A* is used directly.

### Invalidation

On item placement or removal that modifies heights/blockedness:
- Identify affected clusters (usually 1–4).
- Recompute only the border edges for those clusters.
- The intra-cluster sub-graph is rebuilt; the inter-cluster graph is partially updated.

---

## Flying entities

For entities with an active flight effect, a separate cost function is used:

```go
func flyCost(from, to *Tile, diagonal bool) float32 {
    if diagonal {
        return CostDiagonal
    }
    return CostFlat
}
```

All tiles (regardless of `State`) are traversable except permanent `TileBlocked` walls. Maximum step height checks are skipped.

---

## API (`pkg/pathfinding`)

```go
// Layout is the room's pre-computed tile grid.
type Layout struct {
    Width, Height int
    Tiles         [][]Tile
    Cluster       *ClusterGraph  // nil if room is small
}

// Options control pathfinding behaviour per request.
type Options struct {
    AllowDiagonal bool
    Flying        bool
    WalkthroughEntities bool
}

// FindPath returns an ordered slice of PathStep from start to goal,
// or nil if no path exists. PathStep carries X, Y, Z of each tile center.
// Thread-safe (read-only on Layout).
func FindPath(l *Layout, from, to Tile, opts Options) []PathStep
```

`FindPath` is **pure and stateless** with respect to `Layout` (which is read-only after room load). It allocates only the returned `[]PathStep`; all working memory (heap, gen array) is taken from a `sync.Pool` of pre-sized scratch buffers.

---

## Testing strategy

Table-driven unit tests cover:
- Straight horizontal path on flat floor.
- Diagonal path.
- Path around a wall.
- Staircase: rising Z from 0→1→2, verify cost > flat equivalent.
- Flying entity walks through blocked tile.
- No path exists → `FindPath` returns nil.
- Benchmark: 64×64 room, random 20% blocked, path from corner to corner — must complete < 500 µs.

All tests run without a database or network; `Layout` is constructed in-memory from a string heightmap for parity with the existing Habbo room model format.
