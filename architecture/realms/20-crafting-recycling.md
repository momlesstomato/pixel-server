# Realm: Crafting & Recycling

> **Position:** 210 | **Phase:** 13 (Remaining) | **Packets:** 16 (9 c2s, 7 s2c)
> **Services:** game | **Status:** Not yet implemented

---

## Overview

Crafting & Recycling manages the Ecotron (item recycler), crafting recipes, and composting. Users feed items into the Ecotron to receive random rewards, or follow crafting recipes to combine specific items into new ones. This is a small, self-contained realm with well-defined input/output mechanics.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 9 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `recycler.open` | _(none)_ | Open recycler UI |
| -- | `recycler.recycle` | `itemIds:int32[]` | Submit items for recycling |
| -- | `recycler.get_status` | _(none)_ | Check recycler result |
| -- | `crafting.get_recipes` | _(none)_ | Get available crafting recipes |
| -- | `crafting.get_ingredients` | `recipeId:int32` | Get required ingredients |
| -- | `crafting.craft` | `recipeId:int32`, `itemIds:int32[]` | Craft item from recipe |
| -- | `crafting.get_recipes_available` | `itemIds:int32[]` | Check which recipes match these items |
| -- | `crafting.compost` | `itemIds:int32[]` | Compost items |
| -- | `crafting.cancel` | _(none)_ | Cancel pending craft |

### S2C -- 7 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `recycler.status` | Recycler result (reward item) |
| -- | `crafting.recipes` | Available recipe list |
| -- | `crafting.recipe_ingredients` | Required items for recipe |
| -- | `crafting.result` | Crafting outcome |
| -- | `crafting.recipes_available` | Matching recipes for items |
| -- | `crafting.compost_result` | Composting result |
| -- | `crafting.pending` | Pending craft status |

---

## Implementation Analysis

### Ecotron Recycler

```
Flow:
  1. User opens recycler → recycler.open
  2. User selects 5+ items from inventory → recycler.recycle
  3. Validate:
     a. All items exist in user's inventory
     b. All items are recyclable (definition flag)
     c. Minimum 5 items
  4. Remove items from inventory
  5. Random reward from reward pool:
     - Common (60%): basic furniture
     - Uncommon (25%): mid-tier furniture
     - Rare (10%): rare furniture
     - Super rare (5%): exclusive items
  6. Add reward to inventory
  7. Send recycler.status with reward
```

### Crafting Recipes

```
Recipe structure:
  recipe_id: 1
  result_item: "gold_chair"
  ingredients:
    - definition_id: 100, amount: 3  (3x wooden chair)
    - definition_id: 200, amount: 1  (1x gold paint)

Craft flow:
  1. User selects items matching a recipe
  2. Server validates exact match (correct definitions and amounts)
  3. Remove ingredient items
  4. Create result item
  5. Send crafting.result
```

### Database Tables

| Table | Usage |
|-------|-------|
| `crafting_recipes` | id, result_definition_id, result_amount | Recipe definitions |
| `crafting_ingredients` | recipe_id, definition_id, amount | Required ingredients |
| `recycler_rewards` | reward_tier, definition_id, weight | Recycler reward pool |

---

## Caveats

### 1. Item Consumption Atomicity
Crafting and recycling both consume items. The consumption and creation must be in a single transaction to prevent item loss.

### 2. Recipe Matching
When checking available recipes for a set of items, the server must handle partial matches. An item set might match multiple recipes -- show all valid options.

### 3. Recycler Randomness
Use server-side randomness only. The reward tier should use weighted random selection from `recycler_rewards`.

---

## Dependencies

- **Phase 7 (Inventory)** -- item consumption and creation
- **PostgreSQL** -- recipes, rewards tables

---

## Testing Strategy

### Unit Tests
- Recipe matching algorithm
- Recycler reward tier selection
- Item validation (recyclable flag)

### Integration Tests
- Full recycle flow against real DB
- Crafting with valid/invalid ingredient sets

### E2E Tests
- Client recycles items, sees reward in inventory
