# pkg/pathfinding

`pkg/pathfinding` implements deterministic room navigation for the game tick loop.

## Current implementation

- Grid-based 3D A* over `Layout`/`Tile`.
- Height-aware movement costs (`CostClimb`, `CostDescend`) and step limits (`MaxStepUp`, `MaxStepDown`).
- Optional diagonals and flying behavior via `Options`.
- Pure computation only (no I/O, no global state).

## Core API

- `NewLayout(width, height int) *Layout`
- `ParseHeightmap(hm string) *Layout`
- `FindPath(l *Layout, from, to Tile, opts Options) []PathStep`

Path contract:
- Returns `nil` when no route exists or endpoints are invalid.
- Excludes the start tile from the returned steps.
- Returns stable output for the same input.

## Heightmap format

- `'x'` / `'X'`: blocked tile
- `'0'..'9'`: Z values `0..9`
- `'a'..'z'`: Z values `10..35`

Rows are newline-separated.

## Usage

```go
layout := pathfinding.ParseHeightmap("000\n0x0\n000")
from := *layout.At(0, 1)
to := *layout.At(2, 1)
steps := pathfinding.FindPath(layout, from, to, pathfinding.Options{
    AllowDiagonal: true,
})
```

## Notes

- This package currently ships 3D A* only; JPS/HPA* are architectural targets, not runtime behavior today.
- Benchmarks live in `pathfinding_test.go` and should be extended when algorithm behavior changes.
