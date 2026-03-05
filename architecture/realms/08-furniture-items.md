# Realm: Furniture & Items

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 80 | **Phase:** 6 (Furniture & WIRED) | **Packets:** 99 (52 c2s, 47 s2c)
> **Services:** game (item interaction, WIRED engine) | **Status:** Not yet implemented

---

## Overview

Furniture & Items is the **largest realm** at 99 packets. It covers placement, movement, removal, interaction, and state management for both floor and wall items, plus specialty systems: dimmers, WIRED logic, mannequins, YouTube displays, dice, multi-state toggles, post-it notes, stack helpers, color wheels, one-way doors, and builders club items. The WIRED sub-system alone is one of the most complex features in the entire protocol.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 6

---

## Packet Inventory

### C2S (Client to Server) -- 52 packets

#### Core Placement & Movement

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 1258 | `furniture.place` | `itemId:int32`, placement data | Place item from inventory |
| 3456 | `furniture.pickup` | `itemId:int32` | Return item to inventory |
| 248 | `furniture.floor_update` | `itemId`, `x`, `y`, `rotation` | Move/rotate floor item |
| 168 | `furniture.wall_update` | `itemId`, `wallPosition` | Move wall item |

#### Interactions

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 99 | `furniture.toggle_multistate` | `itemId:int32`, `state:int32` | Toggle multi-state floor item |
| 210 | `furniture.toggle_wall_multistate` | `itemId:int32`, `state:int32` | Toggle multi-state wall item |
| 3617 | `furniture.toggle_random_state` | `itemId:int32` | Trigger random state change |
| 1990 | `furniture.activate_dice` | `itemId:int32` | Roll a dice |
| 1533 | `furniture.deactivate_dice` | `itemId:int32` | Close/reset dice |
| 2765 | `furniture.click_one_way_door` | `itemId:int32` | Walk through one-way door |
| 2144 | `furniture.click_color_wheel` | `itemId:int32` | Spin color wheel |
| 3839 | `furniture.set_stack_height` | `itemId:int32`, `height:int32` | Set stack helper height |

#### Dimmer (Moodlight)

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2813 | `furniture.get_dimmer_settings` | `itemId:int32` | Request moodlight state |
| 1648 | `furniture.save_dimmer` | `presetId`, `bgOnly`, `color`, `intensity` | Save dimmer preset |
| 2296 | `furniture.toggle_dimmer` | `itemId:int32` | Toggle dimmer on/off |

#### WIRED Logic

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 768 | `furniture.open_wired` | `itemId:int32` | Open WIRED configuration UI |
| 1520 | `furniture.save_wired_trigger` | trigger config data | Save WIRED trigger |
| 3203 | `furniture.save_wired_condition` | condition config data | Save WIRED condition |
| 2281 | `furniture.save_wired_action` | action config data | Save WIRED action |
| 3373 | `furniture.apply_wired_snapshot` | `itemId:int32` | Apply WIRED state snapshot |

#### Post-it Notes

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2248 | `furniture.place_postit` | `itemId`, `wallPosition` | Place post-it on wall |
| 3283 | `furniture.save_postit` | `itemId`, `color`, `text` | Write/edit post-it |

#### Specialty Items

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2880 | `furniture.apply_toner` | `itemId`, `hue`, `saturation`, `lightness` | Apply room toner |
| 2209 | `furniture.save_mannequin_look` | `itemId`, `figure` | Save outfit to mannequin |
| 2850 | `furniture.save_mannequin_name` | `itemId`, `name` | Name mannequin |
| 336 | `furniture.get_youtube_status` | `itemId:int32` | Get YouTube player state |
| 3005 | `furniture.control_youtube` | `itemId`, `action` | Play/pause/seek YouTube |
| 2069 | `furniture.set_youtube_playlist` | `itemId`, `playlistId` | Set YouTube playlist |
| 462 | `furniture.builders_club_place_wall` | `pageId`, `itemId`, `wallPosition` | Builders club wall placement |

#### Additional 20+ packets for exchange items, rental items, group items, present opening, creditfurni, custom data, trophy editing, jukebox, crafting-related, etc.

### S2C (Server to Client) -- 47 packets

#### Item Lists

| ID | Name | Summary |
|----|------|---------|
| 1749 | `furniture.objects_floor` | Full floor item dump on room enter |
| 1369 | `furniture.objects_wall` | Full wall item dump on room enter |
| 2040 | `furniture.floor_item_add` | New floor item placed |
| 1534 | `furniture.floor_item_update` | Floor item moved/rotated |
| 2703 | `furniture.floor_item_remove` | Floor item picked up |
| 1455 | `furniture.wall_item_add` | New wall item placed |
| 473 | `furniture.wall_item_update` | Wall item moved |
| 3375 | `furniture.wall_item_remove` | Wall item picked up |

#### State Updates

| ID | Name | Summary |
|----|------|---------|
| 2547 | `furniture.state_update` | Item state changed (toggle, dice result, etc.) |
| 3431 | `furniture.data_update` | Item extra data changed |
| 2275 | `furniture.slide_object` | Item sliding on roller |
| 1560 | `furniture.dice_value` | Dice roll result |

#### WIRED

| ID | Name | Summary |
|----|------|---------|
| 1830 | `furniture.wired_trigger` | WIRED trigger configuration data |
| 1108 | `furniture.wired_condition` | WIRED condition configuration data |
| 1434 | `furniture.wired_action` | WIRED action configuration data |
| 2130 | `furniture.wired_save_result` | WIRED configuration save result |
| 3049 | `furniture.wired_reward_result` | WIRED reward outcome |

#### Dimmer, Post-it, Mannequin, YouTube, Toner responses + exchange results, jukebox data, group furniture data, rental extension results, etc.

---

## Architecture Mapping

### Service Ownership

Items live within the room worker's ECS world:

```
Room Worker
├── Item Registry (map[int32]*ItemEntity)
├── ECS Components:
│   ├── ItemInteraction (type, state, extraData)
│   ├── Position (x, y, z for floor items)
│   ├── WallPosition (wallX, wallY, localX, localY for wall items)
│   └── Dirty (triggers state broadcast)
├── ItemInteractionSystem (per-tick item logic)
├── RollerSystem (item sliding)
└── WiredSystem (trigger/condition/action evaluation)
```

### Database Tables

| Table | Usage |
|-------|-------|
| `items` | item_id, user_id, room_id, definition_id, x, y, z, rotation, wall_position, extra_data | Core item data |
| `item_definitions` | definition_id, sprite_name, type (s=floor, i=wall), width, height, interaction_type, interaction_count | Item type metadata |
| `items_wired` | item_id, wired_type, trigger_data, condition_data, action_data | WIRED configuration |
| `items_dimmer` | item_id, preset_id, bg_only, color, intensity, enabled | Moodlight presets |
| `items_teleporter_links` | item_id_a, item_id_b | Teleporter pair links |
| `items_limited_edition` | item_id, limited_number, limited_total | Limited edition tracking |

### Interaction Type Registry

| Interaction Type | Behavior | Example Items |
|------------------|----------|---------------|
| `default` | No interaction | Decorations |
| `gate` | Open/close (blocking toggle) | Iron Gate, Teleport Gate |
| `teleport` | Paired teleportation | Teleporter |
| `roller` | Moves entities/items on tick | Rollers |
| `dice` | Random number generation | Dice (1-6) |
| `vendingmachine` | Dispense hand item | Vending machine |
| `onewaygate` | Walk through one direction | One-way gate |
| `mannequin` | Save/load outfit | Mannequin |
| `dimmer` | Room moodlight | Moodlight |
| `postit` | Sticky note on wall | Post-it |
| `stackhelper` | Set stacking height | Stack helper |
| `colorwheel` | Random color selection | Color wheel |
| `trophy` | Display text | Trophy |
| `wired_trigger` | WIRED trigger | Various triggers |
| `wired_condition` | WIRED condition | Various conditions |
| `wired_effect` | WIRED effect/action | Various effects |
| `youtube` | Video display | YouTube TV |
| `toner` | Room color toner | Background toner |
| `jukebox` | Music player | Jukebox |
| `crackable` | Break to get reward | Crackable items |
| `exchange` | Convert to credits | Credit furni |
| `rentable` | Rent-to-own | Rentable space |

---

## Implementation Analysis

### Item Placement Pipeline

```
1. Client sends furniture.place (1258)
2. Room worker validates:
   a. User has rights (level >= 1) or is owner
   b. Item exists in user's inventory
   c. Target tile is valid:
      - Within room bounds
      - Not blocked by wall/void
      - Stacking allowed (check z-height compatibility)
      - Item dimensions fit at rotation
   d. Room item limit not exceeded (configurable, default: 1500)
3. Remove item from inventory (NATS event to inventory service)
4. Create ECS entity with:
   a. Position(x, y, z)
   b. ItemInteraction(type, state)
   c. TileRef(x, y)
5. Send furniture.floor_item_add (2040) to all users in room
6. Persist to database (async batch writer)
```

### Roller System

Rollers are the most performance-sensitive item type. Every tick:

```
RollerSystem (20 Hz):
  For each roller entity:
    Get entities/items on roller tile
    For each entity on roller:
      Calculate destination tile (roller direction)
      If destination is walkable:
        Start slide animation
        Update entity Position
        Broadcast room_entities.slide_object (2275)
    For each item on roller:
      Calculate destination tile
      Stack height at destination
      Slide item
      Broadcast furniture.slide_object
```

**Performance caveat:** A room with 100 rollers each moving 3 items = 300 position calculations per tick. Must be optimized with spatial indexing.

### WIRED System

WIRED is a visual programming system with three component types:

#### Triggers (Events)
| Type | Fires When |
|------|------------|
| User walks on tile | Entity position matches trigger tile |
| User says keyword | Chat contains configured keyword |
| Periodically | Timer interval elapsed |
| State changes | Item state toggled |
| Score achieved | Team reaches target score |
| User enters room | Entity spawned |
| Collision | Two entities on same tile |

#### Conditions (Filters)
| Type | Checks |
|------|--------|
| User count | Room has N+ users |
| Item state | Specific item in specific state |
| Time elapsed | N seconds since last trigger |
| Team membership | User on specified team |
| Has furni | User carrying hand item |
| User has badge | Badge check |

#### Effects (Actions)
| Type | Does |
|------|------|
| Move items | Slide items to new position |
| Toggle state | Change item state |
| Teleport user | Move user to specific tile |
| Show message | Send chat bubble |
| Give reward | Award badge/item |
| Reset timer | Reset periodic trigger |
| Match position | Move items to match snapshot |
| Give score | Add points to team |

**WIRED Evaluation Pipeline:**
```
1. Trigger fires (event detected)
2. Load linked conditions for trigger
3. Evaluate ALL conditions (AND logic)
4. If all conditions pass:
   a. Load linked effects
   b. Execute effects in order
   c. Apply cooldown to trigger
5. Broadcast state changes
```

**WIRED Storage:** Each WIRED item stores its configuration as JSON in `items_wired`. The configuration includes:
- Selected items (by ID)
- Parameters (keyword, delay, team ID, etc.)
- Snapshot data (item positions for "match position" effect)

### Teleporter System

Teleporters work in pairs linked by `items_teleporter_links`:

```
1. User walks onto teleporter A
2. System looks up linked teleporter B
3. Validate B exists and is in a valid room
4. Close door animation on A (state = "1")
5. If same room: instant position change
6. If different room: trigger room change for user
7. Open door animation on B (state = "2")
8. User appears at B's position
9. Reset both states after delay
```

**Cross-room teleportation** is the hardest case: the user must leave room A and enter room B atomically. This requires coordination between two room workers via NATS.

---

## Caveats & Edge Cases

### 1. Item Stacking Z-Height
Items can be stacked on top of each other. The Z-height of a placed item is:
```
z = tile_height + sum(heights of items below)
```
Each item definition has a `stackHeight` attribute. The stack helper item overrides this calculation for custom heights.

### 2. Rotation and Collision
Floor items have different collision masks at different rotations. A 2x1 item rotated 90 degrees becomes 1x2. Collision checking must use the rotated dimensions.

### 3. Wall Item Positioning
Wall items use a special position format: `":w=X,Y l=LocalX,LocalY"` where X,Y is the wall tile and LocalX,LocalY is the offset on the wall face. Validation must ensure the position is on a valid wall segment.

### 4. WIRED Infinite Loops
A WIRED effect can trigger another WIRED trigger, creating infinite loops. Implement:
- Maximum execution depth: 10 levels.
- Cooldown per trigger: minimum 0.5 seconds between fires.
- Per-tick WIRED budget: maximum 50 effect executions per tick.

### 5. Roller Chain Conflicts
Two rollers facing each other create deadlock. Items slide back and forth infinitely. Detect and break cycles by prioritizing the roller with the lower item ID.

### 6. Item Ownership Tracking
When items are placed in rooms, they remain owned by the user but are "in room". If the room owner changes or the room is deleted, items must be returned to owners' inventories. Track `user_id` separately from `room_id` in the `items` table.

### 7. Limited Edition Items
Limited items have a unique number (`limited_number/limited_total`). They cannot be stacked, traded as bundles, or duplicated. The limited status must be preserved through all item operations.

### 8. Dice Fairness
Dice results must be generated server-side (not client-provided). Use cryptographic randomness (`crypto/rand`) for fairness.

### 9. YouTube Display Security
YouTube playlist/video IDs from clients must be validated against an allowlist or YouTube API. Arbitrary URLs could be used for phishing or inappropriate content.

### 10. Item Interaction Rate Limiting
Rapid item toggling (e.g., spam-clicking a gate) creates excessive database writes and broadcasts. Rate-limit interactions to 2 per second per item.

---

## Improvements Over Legacy Emulators

| Area | Legacy Pattern | pixel-server Improvement |
|------|---------------|-------------------------|
| **Item storage** | Immediate DB write per interaction | Async batch writer (no tick blocking) |
| **WIRED evaluation** | Recursive function calls | Iterative with depth/budget limits |
| **Roller system** | Per-roller iteration (slow) | Spatial-indexed batch processing |
| **Interaction types** | `ICycleable` interface + inheritance | ECS component composition |
| **Teleporters** | Same-room only or hacky cross-room | NATS-coordinated cross-room teleportation |
| **State broadcasting** | Per-interaction packet | Batched with Dirty flag in tick |
| **Collision detection** | Full room scan per placement | Spatial hash with O(1) tile lookup |
| **WIRED storage** | Custom binary format | JSON in `items_wired` column |

---

## Dependencies

- **Phase 3 (Room)** -- room worker, ECS world, entity system
- **Phase 7 (Inventory)** -- item ownership, add/remove from inventory
- **pkg/item** -- domain models (Item, ItemDefinition, WiredConfig)
- **pkg/room** -- ECS components, spatial indexing
- **PostgreSQL** -- items, definitions, WIRED config, teleporter links

---

## Testing Strategy

### Unit Tests
- Item stacking height calculation
- Rotation collision mask transformation
- WIRED condition evaluation (each condition type)
- WIRED effect execution (each effect type)
- WIRED loop detection and budget enforcement
- Roller direction computation and chain detection
- Dice random distribution (chi-squared test)
- Wall position format parsing and validation

### Integration Tests
- Full place/move/pickup cycle against real DB
- WIRED trigger -> condition -> effect pipeline
- Teleporter linking and cross-room teleportation
- Roller sliding with entity and item movement
- Item state persistence after room unload/reload

### E2E Tests
- Client places item, second client sees it appear
- Client toggles gate, movement is blocked/unblocked
- Client rolls dice, all users see same result
- WIRED trigger fires on user walk, effect toggles gate
- Teleporter moves user between rooms
