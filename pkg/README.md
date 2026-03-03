# pkg

`pkg/` contains reusable libraries shared by multiple services. Code here should be stable, testable, and free of service-specific wiring.

## Package map

| Path | Responsibility |
| --- | --- |
| `pkg/core/*` | Shared infrastructure primitives (codec, config, logging, NATS wrapper, test helpers). |
| `pkg/protocol` | Generated Pixel Protocol packet structs, header constants, decode helpers, and packet name maps. |
| `pkg/pathfinding` | Deterministic 3D pathfinding utilities used by room simulation. |
| `pkg/plugin` | Plugin API and runtime for event hooks/interceptors without breaking room ownership rules. |
| `pkg/user`, `pkg/room`, `pkg/item`, `pkg/social`, `pkg/navigator`, `pkg/catalog`, `pkg/moderation` | Domain types, repository interfaces, and domain-owned subjects. |

## Boundaries

- `pkg/*` must not import `services/*`.
- Domain packages own domain contracts and state transitions; infrastructure is injected by service startup.
- Generated code in `pkg/protocol` is never hand-edited. Regenerate from `vendor/pixel-protocol/spec/protocol.yaml`.
- Room simulation invariants still apply inside shared libraries: one room goroutine owns one ECS world.

## Documentation contract

- Keep package READMEs aligned with code behavior, not planned behavior.
- When adding or changing a package API, update the corresponding README in the same change.
- Use package READMEs to document constraints and integration points that are not obvious from type signatures alone.
