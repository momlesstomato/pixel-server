# Furni Interactions: Interactive Furniture Completion Plan

## Overview

This plan covers the missing in-room interactive furniture behavior required to
make the furniture realm production-ready. It closes the gap left intentionally
by the economy and room-basics plans: placed items already exist, but most
interaction state machines do not. The goal is a server-authoritative
implementation that keeps movement, access, room state, and packet emission
inside clear DDD and hexagonal boundaries.

The scope includes:

- Generic floor multistate toggles
- Dice activation and reset
- Teleporters, including cross-room transfer
- Rollers for entities and furniture
- Post-it wall items
- Room dimmers / moodlights
- Present / gift opening
- Stack-helper / stackable furniture height override

The scope explicitly excludes:

- Wired
- Jukebox / sound machine / song disks
- Group furniture
- Guild-specific furniture widgets

The guiding rules for this plan are:

1. Room movement and entity mutation remain server-authoritative.
2. Furniture state changes persist through the furniture repository, not ad-hoc
   runtime maps.
3. Cross-room teleports bypass password and lock checks but still enforce room
   ban restrictions.
4. Rollers and teleporters are single-occupancy per interaction cycle.
5. Edge cases are handled deterministically instead of silently desyncing.

---

## Current Status

Implemented in the current server code:

- Generic floor multistate toggles
- Dice activation and reset
- Teleporters, including same-room and cross-room transfer through persisted
   hidden pairing metadata and room-entry override flow
- Rollers for entities and furniture through the room tick processor and the
   rolling packet surface
- Post-it wall placement, movement, note text, and color updates
- Room dimmer preset query, save, toggle, and object-data persistence
- Present opening through gift metadata and in-room item transformation
- Stack-helper height override with effective floor-height rebroadcast

Focused automated coverage now exists for:

- Existing multistate, dice, and stack-helper behavior
- Post-it wall data save and rebroadcast
- Dimmer preset query and toggle flow
- Cross-room teleporter forwarding
- Roller tick movement for one item and one avatar
- Present opening and transformation

Still intentionally out of scope:

- Wired
- Jukebox / sound machine / song disks
- Group furniture
- Guild-specific furniture widgets

---

## Vendor Cross-Reference Matrix

| Feature | Renderer / protocol expectation | Comet | PlusEMU / Arcturus | Pixel-server target |
|---------|---------------------------------|-------|--------------------|---------------------|
| Multistate toggle | C2S 99 toggles next state | Yes | Yes | **Implemented** generic cycle support |
| Dice | C2S 1990 / 1533, S2C 3431 | Yes | Yes | **Implemented** roll + clear |
| Teleporter | Multi-step animation + room forward | Yes | Yes | **Implemented** same-room and cross-room |
| Roller | S2C rolling bundle, entity and furni sliding | Yes | Yes | **Implemented** bounded cycle engine |
| Post-it | Wall state + note payload | Yes | Partial | **Implemented** wall storage + validation |
| Dimmer | Preset payload + object data string | Yes | Partial | **Implemented** presets + toggle + persistence |
| Gift open | C2S open, S2C gift_opened | Yes | Yes | **Implemented** in-room reveal transform |
| Stack helper | C2S 3839, S2C 2816 | Yes | Partial | **Implemented** runtime and persistence |

---

## Packet Registry

### Client to Server

| ID | Name | Purpose |
|----|------|---------|
| 99 | `furniture.toggle_multistate` | Toggle floor multistate items |
| 168 | `furniture.wall_update` | Reposition wall items |
| 1990 | `furniture.activate_dice` | Start a dice roll |
| 1533 | `furniture.deactivate_dice` | Clear or close dice result |
| 2296 | `furniture.toggle_dimmer` | Toggle room dimmer enabled state |
| 2813 | `furniture.get_dimmer` | Request dimmer presets |
| 1648 | `furniture.save_dimmer` | Save one dimmer preset |
| 3558 | `furniture.open_gift` | Open a present / gift item |
| 3839 | `furniture.set_stack_height` | Set stack-helper override height |

### Server to Client

| ID | Name | Purpose |
|----|------|---------|
| 3431 | `furniture.dice_value` | Broadcast dice result or hidden state |
| 3207 | `room.rolling` | Broadcast roller movement for items and one entity |
| 2816 | `furniture.stack_height` | Confirm stack-helper current height |
| 2710 | `furniture.dimmer_presets` | Send dimmer presets and selection |
| 2009 | `furniture.wall_item_updated` | Broadcast wall item move/update |
| 56 | `furniture.gift_opened` | Reveal opened gift contents |

---

## Architecture Split

### Furniture domain

The furniture realm owns:

- Item definition interaction typing
- Item state serialization / validation
- Wall placement string validation
- Gift metadata and unwrap contracts
- Dimmer preset models and serialization

New domain types should remain split by responsibility instead of a monolithic
interaction file. Planned additions:

- `pkg/furniture/domain/item_data.go`
- `pkg/furniture/domain/wall_item.go`
- `pkg/furniture/domain/dimmer.go`
- `pkg/furniture/domain/gift.go`
- `pkg/furniture/domain/interaction.go`

### Furniture application

The application layer owns:

- Validating interaction requests
- Persisting item extra data and wall placement
- Resolving teleporter partners and gift contents
- Producing deterministic interaction results for adapters

New application capabilities:

- Update item extra data
- Update wall item placement
- Toggle multistate with max-state validation
- Roll / clear dice
- Save / load / toggle dimmer presets
- Open gifts into revealed room items
- Set stack-helper height

### Room engine

The room realm remains the owner of:

- Entity lookup and movement
- Tile occupancy and path blocking
- Teleporter transfer sequencing
- Roller cycle progression

Furniture must not directly mutate room entities from outside the room runtime.
When room-side mutation is required, furniture uses narrow room callbacks or
room messages wired at runtime startup.

### Realtime adapters

Realtime handlers remain thin. They:

- Parse packets
- Check room membership / rights
- Call application services
- Broadcast packet composers

They must not embed teleporter, roller, or dimmer state machines inline.

---

## Item State Contracts

### Generic multistate

- Storage: integer string in `item.extra_data`
- Valid range: `0 <= state < interaction_modes_count`
- Empty or malformed value falls back to `0`
- Toggle advances to next state modulo mode count
- Definitions with mode count `<= 1` are treated as no-op

### Dice

- `"0"` = idle / hidden
- `"-1"` = rolling lock
- `"1"` to `"6"` = result
- Concurrent rolls on the same item are rejected
- Deactivate returns to `"0"`

### Teleporter

- Same-room teleporter partner stored as another room item ID
- Cross-room teleporter partner may resolve to a different room item and room ID
- Extra data format must allow partner resolution without magical naming rules
- Unpaired teleporters fail safely and release the user

### Roller

- Rollers use definition interaction type plus room-facing direction
- Runtime cycle speed must be configurable
- Rollers themselves do not persist transient queue state in DB
- Moved item/entity sets are per-cycle runtime state only

### Post-it

- Format: `COLOR MESSAGE`
- Valid colors:
  - `FFFF33`
  - `FF9CFF`
  - `9CCEFF`
  - `9CFF9C`
- Invalid or malformed data normalizes to `FFFF33 `
- Message length must be bounded server-side

### Dimmer

- Item object data string format: `enabled,preset,effect,color,brightness`
- Preset payload is separate from object data
- Only one active dimmer per room is authoritative
- Color whitelist must match client-supported colors

### Gift

- Gift metadata contains the contained definition and product code needed for
   reveal
- Opening transforms the placed present into the contained definition in-room
- The client reveal packet is sent after the transformation succeeds

### Stack helper

- Height stored as decimal string with two digits precision
- Request height uses hundredths of a tile unit
- Height must be clamped to a safe configured range
- Floor update packets must emit effective stack height after override

---

## Behavioral Contracts And Edge Cases

### Teleporter

1. Interaction is single-user per teleporter pair while active.
2. User must be on the teleporter tile or on the tile directly in front before
   the teleport sequence advances.
3. Same-room teleports move to the paired teleporter, then to the exit tile in
   front of the destination teleporter.
4. Cross-room teleports:
   - bypass lock and password checks,
   - bypass entry-key requirements,
   - still reject if the actor is banned from the destination room,
   - fail if the destination room or paired item no longer exists.
5. If the destination exit tile is blocked by a hard obstacle, another entity,
   or unwalkable top furniture, the teleport aborts and the actor is restored to
   the source-side fallback tile.
6. If the actor disconnects during teleport, both source and destination locks
   are released.
7. If either teleporter is moved or picked up mid-sequence, the sequence aborts
   and the actor is released safely.

### Roller

1. Rollers operate on a repeating room cycle configured in furniture runtime
   config.
2. One roller cycle may move multiple floor items and at most one unit move per
   serialized packet, matching the client rolling packet format.
3. Movement direction is the roller's facing direction.
4. Entity rolling is rejected when the destination tile:
   - is out of bounds,
   - is blocked in the room layout,
   - is occupied by a non-moving entity,
   - exceeds max step height,
   - would strand the entity inside invalid furniture geometry.
5. Furniture rolling is rejected when the destination tile has incompatible top
   stack state or when the moved item would overlap blocked geometry.
6. When a roller is removed while an entity or item is on it, that target stays
   in its current tile and the cycle state is cleared.
7. Roller chains must not double-move an entity or item in the same cycle.
8. Rollers leading into the room door tile trigger normal room leave behavior.

### Post-it

1. Wall position strings must be validated before update.
2. Notes may only be edited by an authorized room controller.
3. Empty notes remain valid and preserve the selected color.
4. Invalid wall positions are rejected without mutating state.
5. Moving a post-it preserves its note content.

### Dimmer

1. Only the active room dimmer may answer preset queries.
2. Picking up the active dimmer clears room dimmer state immediately.
3. Saving a preset validates preset index, color, and brightness bounds.
4. Toggling the dimmer updates both the room object data string and the room
   effect event payload.
5. New entrants must receive the current dimmer object data and the preset list.

### Dice

1. Dice cannot be re-rolled while already rolling.
2. Interacting from a distance should first walk the user to the valid use tile.
3. Deactivate clears current result and persists `0`.
4. Rigged results may be supported via explicit runtime hook, not magic
   attributes hidden in unrelated services.

### Gift

1. Only the room owner / controller with pickup rights may open a placed gift.
2. Opening succeeds only when contained item metadata is valid.
3. If transformation fails, the present is not consumed.
4. Opening reveals the contained item in-room; it does not transfer the item to
   inventory.

### Stack helper

1. Only dedicated stack-helper definitions may accept height-change packets.
2. Height changes must rebroadcast floor item update and stack-height confirm.
3. Pathfinding and seat height must use the effective override height.
4. Negative or extreme heights are clamped instead of stored verbatim.

---

## Implementation Breakdown

Phases 1 through 5 are implemented in the current codebase. Phase 6 now has
focused furniture and room unit coverage for the newly added interaction
families; broader end-to-end parity coverage can still be expanded later.

### Phase 1: State mutation foundations

- Add repository methods for updating item extra data and wall placement
- Add domain parsers and validators for interactive item state
- Extend definition interaction types to include `gift`, `postit`, `dimmer`,
  `stack_helper`, and teleporter door variants where needed

### Phase 2: Packet and composer surface

- Add missing C2S dispatch branches for dice, wall update, dimmer, gift, and
  stack height
- Add S2C composers for dice value, room rolling, stack height, wall update,
  dimmer presets, and gift opened

### Phase 3: Application services

- Implement multistate toggle service
- Implement dice roll / clear service
- Implement post-it update / wall move service
- Implement dimmer load / save / toggle service
- Implement gift open service
- Implement stack-helper service

### Phase 4: Room-runtime integration

- Add teleporter transition coordinator
- Add roller cycle processor
- Add room callbacks for:
  - entity lookup by connection and user
  - occupancy checks
  - forced walk / forced reposition
  - entity transfer out / into another room
  - tile effective height queries

### Phase 5: Cross-room teleporter flow

- Source room reserves actor
- Destination room validates ban-only gate
- Destination room loads if absent
- Actor leaves source room cleanly
- Actor enters destination room at teleporter tile
- Actor exits at destination front tile
- Failure returns actor to source fallback

### Phase 6: Tests

- Application tests for every item-state parser and service
- Realtime tests for every packet handler
- Room engine tests for roller and teleporter movement edge cases
- E2E coverage under room / furniture flows for at least:
  - same-room teleporter
  - cross-room teleporter
  - blocked destination teleporter
  - roller entity move
  - roller item move
  - dice roll and clear
  - post-it save and wall move
  - dimmer preset save / toggle
  - gift opening
  - stack-helper height override

---

## Missing Interactive Furniture After This Plan

These remain intentionally out of scope after this plan:

- Wired triggers, conditions, effects, and addon execution
- Sound machine / jukebox / song disk inventory and playback
- Group furniture rights and badge-bound behavior
- Guild / forum furniture widgets
- Other niche widget-driven interactives not covered by the packet set above

---

## Success Criteria

The plan is complete when:

1. All packet handlers listed above are wired and tested.
2. Teleporters support same-room and cross-room transfer with ban-only target
   gating.
3. Rollers move entities and furniture deterministically and broadcast the
   rolling packet shape expected by Nitro.
4. Post-it, dimmer, gift, dice, and stack-helper state persist correctly.
5. The room engine remains authoritative for movement and collision.
6. Wired, jukebox, groups, and songs remain the only explicitly missing major
   interactive furniture families.