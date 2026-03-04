# Pathfinding — Integration with the Room Service

How `FindPath` is connected to the ECS room simulation.

---

## Call Site

Walk requests are received as C2S packets inside the room goroutine. The
handler resolves the entity's current tile from the `TileRef` component and
calls `FindPath` synchronously:

```go
from := pathfinding.Tile{X: tileRef.X, Y: tileRef.Y, Z: position.Z}
to   := pathfinding.Tile{X: targetX, Y: targetY}
opts := pathfinding.Options{
    AllowDiagonal:       true,
    Flying:              false,
    WalkthroughEntities: room.AllowWalkthrough,
}
steps := pathfinding.FindPath(layout, from, to, opts)
```

The `*pathfinding.Layout` is derived from the room's `Model.Heightmap` via
`ParseHeightmap` when the room goroutine boots. It is reused for the lifetime
of the room.

---

## Result Handling

| `FindPath` return value | Handler action |
|---|---|
| `nil` (no path) | Walk request is silently ignored; entity stays in place |
| `[]PathStep` | Assigned to the entity's `WalkPath.Steps`; `Cursor` is reset to `0` |

A nil result is the only silent failure path. No error packet is sent to the
client — clients re-request walks as needed.

---

## ECS Consumption

Each call to `MovementSystem(rw)` during the game loop advances every entity
with a non-empty `WalkPath` by exactly one `PathStep`. After all steps are
consumed (`Cursor == len(Steps)`), the entity is considered stopped.

```
tick N:   WalkPath.Cursor = 0  →  move to Steps[0]
tick N+1: WalkPath.Cursor = 1  →  move to Steps[1]
...
tick N+k: WalkPath.Cursor = k  →  HasSteps() = false → entity stopped
```

The pathfinding package never interacts with Ark directly. The boundary is:
- **Pathfinding package** produces a `[]PathStep`.
- **Room goroutine** writes it to the `WalkPath` component.
- **MovementSystem** reads and advances `WalkPath` each tick.

---

## Layout Lifecycle

```
RoomBoot
  └─ Load Model.Heightmap from repository
  └─ layout = ParseHeightmap(model.Heightmap)
  └─ (layout is stored in room goroutine local state — never shared)

Per walk-request (inside room goroutine)
  └─ FindPath(layout, from, to, opts)

RoomShutdown
  └─ layout goes out of scope (GC'd)
```

---

## Performance Notes

- `ParseHeightmap` allocates a `[][]Tile` once per room at boot — not per
  request.
- `FindPath` allocates three flat slices (`[]float32`, `[]int`, `[]bool`) of
  size `width * height` per call. For a standard 32×32 room this is
  ~12 KB of heap per walk request. A future optimisation can pool these slices.
- Pathfinding runs on the room goroutine, not a worker pool. Because it is
  pure computation (no I/O) and rooms are small, this is acceptable at 20 Hz.

---

## Realm Relations

| Realm | Dependency |
|---|---|
| **ROOM** | Provides the `Model.Heightmap` and the `WalkPath` / `TileRef` / `Position` ECS components |
| **SESSION** | Walk C2S packets arrive via the session's NATS channel into the room goroutine |

---

## Plugin Hooks

There are no dedicated pathfinding plugin events. The `event.PlayerWalk` event
(described in ROOM/PLUGIN-HOOKS.MD) is the appropriate hook for plugins that
need to react to walk assignments.
