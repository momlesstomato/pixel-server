# Pathfinding — Overview

`pkg/pathfinding` provides a dependency-free, 3D-aware A* implementation used
by the game service to route avatars, bots, and pets across a room's tile grid.

---

## What Is Built

| Capability | Status |
|---|---|
| Tile grid (`Layout`, `Tile`, `TileState`) | ✅ Complete |
| `ParseHeightmap` — ASCII map to Layout | ✅ Complete |
| A* with 4 or 8 directional movement | ✅ Complete |
| 3D height-cost model (climb / descend) | ✅ Complete |
| Flying entity support (traverses all tiles) | ✅ Complete |
| Entity walkthrough option | ✅ Complete |
| JPS (Jump Point Search) optimisation | 🔄 Planned |
| HPA* hierarchical abstraction | 🔄 Planned |

---

## Package Layout

```
pkg/pathfinding/
├── tile.go      — TileState enum, Tile struct, PathStep
├── layout.go    — Layout struct, NewLayout, ParseHeightmap, InBounds, At
└── astar.go     — FindPath, Options, cost constants, heuristic, heap internals
```

---

## Key Design Properties

- **Pure computation.** No I/O, no context, no external dependencies.
  Algorithms use flat slice arithmetic instead of pointer-heavy data structures
  for cache efficiency.
- **3D height awareness.** Every tile has a Z (stack height). Movement cost
  increases for climbing and decreases slightly for descending. A step that
  exceeds `MaxStepUp` or `MaxStepDown` is blocked.
- **Flying mode.** Flying entities (flagged via `Options.Flying`) treat all
  tiles as passable and pay only base directional cost.
- **Single-call API.** `FindPath` returns a fully reconstructed path or `nil`.
  The caller (game service) assigns the result to the entity's `WalkPath`
  component.

---

## Further Reading

| Page | Contents |
|---|---|
| [LAYOUT.MD](LAYOUT.MD) | TileState, Tile, Layout struct; ParseHeightmap char encoding |
| [ALGORITHM.MD](ALGORITHM.MD) | FindPath, Options, cost model, heuristic, A* internals |
| [INTEGRATION.MD](INTEGRATION.MD) | How the room goroutine calls FindPath and feeds results to ECS |
