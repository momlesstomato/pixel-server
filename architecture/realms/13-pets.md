# Realm: Pets

> **Position:** 140 | **Phase:** 9 (Pets) | **Packets:** 41 (21 c2s, 20 s2c)
> **Services:** game (PetAI ECS system) | **Status:** Not yet implemented

---

## Overview

The Pets realm manages pet lifecycle: purchasing, naming, placing in rooms, AI behavior (following, commands, mood), breeding, leveling, supplements, and pet packages. Pets are ECS entities with the `PetAI` component that runs within the room worker's tick loop. They use the same movement and position systems as avatars but with autonomous behavior.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 9

---

## Packet Inventory

### C2S -- 21 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| 2647 | `pet.place` | `petId`, `x`, `y` | Place pet from inventory into room |
| 1581 | `pet.pickup` | `petId:int32` | Return pet to inventory |
| 3449 | `pet.move` | `petId`, `x`, `y` | Move pet to tile |
| 2934 | `pet.info` | `petId:int32` | Request pet stats card |
| 1036 | `pet.ride` | `petId:int32` | Mount pet for riding |
| 186 | `pet.remove_saddle` | `petId:int32` | Remove saddle from pet |
| 3379 | `pet.toggle_breeding` | `petId:int32` | Toggle breeding availability |
| 1472 | `pet.toggle_riding` | `petId:int32` | Toggle riding permission |
| 1328 | `pet.use_product` | `petId`, `itemId` | Use food/drink on pet |
| 2161 | `pet.get_commands` | `petId:int32` | Get available commands |
| 3698 | `pet.open_package` | `petId:int32` | Open pet package (mystery pet) |
| 549 | `pet.select` | `petId:int32` | Select pet (focus) |
| 1638 | `pet.breed` | `petId1`, `petId2` | Breed two pets |
| 1521 | `pet.harvest` | `petId:int32` | Harvest from plant pet |
| 2768 | `pet.give_hand_item` | `petId`, `handItemId` | Give item to pet |
| 3202 | `pet.respect` | `petId:int32` | Give respect to pet |
| 1756 | `pet.get_breeds` | `petType:string` | Get available breeds/colors |
| 749 | `pet.supplement` | `petId`, `supplementId` | Apply supplement to pet |
| 2713 | `pet.cancel_breeding` | _(none)_ | Cancel breeding process |
| 3382 | `pet.confirm_breeding` | breeding fields | Confirm breeding |
| 3835 | `pet.compost` | `petId:int32` | Compost plant pet |

### S2C -- 20 packets

| ID | Name | Summary |
|----|------|---------|
| 2788 | `pet.respected` | Pet received respect |
| -- | `pet.info` | Pet stats (level, happiness, energy, hunger, thirst, experience) |
| -- | `pet.commands` | Available commands for pet |
| -- | `pet.breeds` | Available breeds for pet type |
| -- | `pet.level_up` | Pet gained a level |
| -- | `pet.breeding_result` | Breeding outcome |
| -- | `pet.status_update` | Pet mood/state change |
| -- | `pet.supplement_result` | Supplement applied result |
| -- | `pet.package_result` | Package opening result |
| -- | `pet.figure_update` | Pet appearance change |
| -- | `pet.experience_update` | Experience points gained |
| -- | `pet.training_panel` | Training commands UI data |
| -- | `pet.harvest_result` | Harvest from plant pet |
| -- | `pet.name_validation` | Pet name check result |
| -- | plus additional status/breeding/riding packets |

---

## Implementation Analysis

### Pet Entity Model (ECS)

Pets use the same ECS entity system as avatars:

```go
// Pet ECS Components
Position     // X, Y, Z (shared with avatar)
TileRef      // Grid position (shared)
WalkPath     // Movement path (shared)
EntityKind   // Kind = 3 (Pet)
PetAI        // Pet-specific AI state
PetStats     // happiness, energy, hunger, thirst, experience, level
PetBreeding  // breeding state, partner, timer
```

### PetAI System (20 Hz)

The `PetAISystem` runs each tick and manages autonomous behavior:

```
PetAISystem tick:
  For each pet entity with PetAI component:
    1. Decay stats:
       - Energy: -1 every 600 ticks (30 seconds)
       - Happiness: -1 every 1200 ticks (60 seconds)
       - Hunger: +1 every 900 ticks (45 seconds)
       - Thirst: +1 every 900 ticks (45 seconds)
    2. Evaluate behavior based on stats:
       - If energy < 10: sleep (lay posture)
       - If hunger > 80: wander to food bowl
       - If happiness < 20: sad gesture
    3. If following owner:
       - Compute path to 1 tile adjacent to owner
       - Walk toward owner (uses same PathSystem)
    4. If idle (no command, not following):
       - Random wander every 10-30 seconds
       - Random gesture every 20-60 seconds
    5. Process pending commands:
       - Evaluate command success based on level + happiness
       - Execute command action
```

### Pet Commands

| Command | Level Req | Action |
|---------|-----------|--------|
| Free | 0 | Stop following, wander |
| Sit | 1 | Sit posture |
| Down | 2 | Lay posture |
| Here | 3 | Come to owner's tile |
| Speak | 4 | Show speech bubble |
| Play dead | 5 | Lay + close eyes |
| Beg | 6 | Beg gesture |
| Stand | 7 | Stand posture |
| Follow | 8 | Follow owner continuously |
| Jump | 9 | Jump animation |
| Play | 10+ | Play animation |

Command success probability: `base_chance + (happiness * 0.5) + (level * 2)`, capped at 95%.

### Pet Breeding

```
1. Two pets of same type in same room
2. Both must have breeding enabled
3. Owner sends pet.breed with both pet IDs
4. Validate:
   - Same type, both Level 4+
   - Not siblings (different parent IDs)
   - Breeding cooldown elapsed (24 hours)
5. Breeding animation (5 second timer)
6. Generate offspring:
   - Random breed variant from parent pool
   - Random name suggestion
   - Level 1, full stats
7. Add to owner's inventory
8. Send pet.breeding_result
```

### Pet Stats and Leveling

```
Level-up formula:
  Required XP per level = level * 100

XP sources:
  - Command success: +10 XP
  - Pet respect: +20 XP
  - Pet food: +5 XP
  - Time in room: +1 XP per minute

Maximum level: 20
Maximum stats: 100 each (happiness, energy, hunger, thirst)
```

---

## Caveats & Edge Cases

### 1. Pet Death/Starvation
Reference emulators don't implement pet death. Pets at 0 energy simply sleep and refuse commands. pixel-server should follow this pattern (no pet death) but implement visual indicators of poor health.

### 2. Pet Name Validation
Same rules as usernames: 3-15 chars, alphanumeric only, word filter check. Pet names must be validated server-side even though the client has its own filter.

### 3. Plant Pets (Monster Plants)
Plant pets have unique mechanics: they don't move, they grow over time, they can be watered and harvested. The PetAI system must handle plant type differently (no movement, growth ticks instead).

### 4. Pet Riding
Rideable pets (horses) allow the owner to mount them. While mounted, the owner's position is locked to the pet's position. Walk commands move the pet, not the avatar.

### 5. Pet Room Limits
Maximum pets per room: 10 (configurable). Combined with bots, the total NPC entity limit prevents server overload.

### 6. Pet Persistence
Pet stats must be saved to the database on room unload or periodic flush. Don't save every tick -- batch at 60-second intervals or on room idle.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **AI system** | Timer-based callbacks | ECS PetAI component with deterministic ticks |
| **Movement** | Separate pathfinding | Shared 3D A* with avatar system |
| **Stats decay** | Variable timing | Fixed tick rate (deterministic) |
| **Breeding** | Basic random | Configurable genetics with parent pool |
| **Command system** | Hardcoded per type | Configurable command sets via data |

---

## Dependencies

- **Phase 3 (Room)** -- room worker, ECS world for pet entities
- **Phase 6 (Furniture)** -- pet food bowls, pet houses as item interactions
- **Phase 7 (Inventory, Catalog)** -- pet acquisition and storage

---

## Testing Strategy

### Unit Tests
- Pet stat decay rates (deterministic tick-based)
- Command success probability calculation
- Level-up XP thresholds
- Breeding validation rules
- Plant pet growth mechanics

### Integration Tests
- Pet place/pickup lifecycle
- Stats persistence across room unload/reload
- Breeding flow with offspring creation

### E2E Tests
- Client places pet, sees it walk around
- Client issues commands, pet responds
- Two pets breed, offspring appears in inventory
