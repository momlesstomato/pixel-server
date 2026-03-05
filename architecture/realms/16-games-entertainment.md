# Realm: Games & Entertainment

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 170 | **Phase:** 12 (Games) | **Packets:** 49 (21 c2s, 28 s2c)
> **Services:** game (mini-game ECS systems) | **Status:** Not yet implemented

---

## Overview

Games & Entertainment manages in-room mini-games (Freeze, Battle Banzai, BattleBall), the game center, leaderboards, polls, quizzes, and voting systems. Mini-games are the most complex ECS sub-systems: they layer team management, game-specific physics (projectiles, tile claiming), scoring, timers, and powerups on top of the standard room entity model.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 12

---

## Packet Inventory

### C2S -- 21 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `game.join_team` | `teamId:int32` | Join a game team (red/blue/green/yellow) |
| -- | `game.leave_team` | _(none)_ | Leave current team |
| -- | `game.start` | _(none)_ | Trigger game start (host only) |
| -- | `game.stop` | _(none)_ | Stop game (host/timer) |
| -- | `game.throw_snowball` | `x:int32`, `y:int32` | Throw snowball (Freeze) |
| -- | `game.activate_tile` | `itemId:int32` | Step on Banzai tile |
| -- | `game_center.join_queue` | `gameType:int32` | Join game center queue |
| -- | `game_center.leave_queue` | _(none)_ | Leave queue |
| -- | `leaderboard.get` | `gameType`, `period` | Get leaderboard |
| -- | `poll.answer` | `pollId`, `questionId`, `answers[]` | Submit poll answer |
| -- | `poll.reject` | `pollId:int32` | Dismiss poll |
| -- | Additional game interaction and query packets |

### S2C -- 28 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `game.started` | Game has begun |
| -- | `game.finished` | Game over, show results |
| -- | `game.team_scores` | Live score update for all teams |
| -- | `game.player_status` | Player freeze/shield/lives update |
| -- | `game.tile_update` | Banzai tile color changed |
| -- | `game.timer_update` | Game timer tick |
| -- | `game.powerup_spawned` | Powerup appeared |
| -- | `game_center.game_list` | Available games |
| -- | `game_center.queue_status` | Queue position/match found |
| -- | `leaderboard.data` | Leaderboard rankings |
| -- | `poll.question` | Poll presented to user |
| -- | `poll.results` | Poll results summary |
| -- | Additional game state, animation, and notification packets |

---

## Implementation Analysis

### Mini-Game Architecture

Mini-games run as optional ECS sub-systems within the room worker:

```
Room Worker
├── Standard ECS Systems (always active)
│   ├── MovementSystem
│   ├── ChatCooldownSystem
│   └── BroadcastSystem
└── Game Systems (active when game running)
    ├── GameManagerSystem (state machine: idle → starting → playing → finished)
    ├── FreezeSystem OR BanzaiSystem OR BattleBallSystem
    ├── GameTimerSystem (countdown)
    ├── GameScoreSystem (team scoring)
    └── GamePowerupSystem (powerup spawning/collection)
```

### Freeze Game

The most complex mini-game:

```
Game State:
  - Teams: 2-4 (red, blue, green, yellow)
  - Players per team: 1-5
  - Timer: configurable (60-300 seconds)
  - Lives per player: 3

Tick logic:
  1. Process thrown snowballs:
     - Snowball travels 1 tile per tick
     - On hit: freeze target for 5 seconds, -1 life
     - Explosion pattern: cross (normal) or diamond (mega)
  2. Update freeze timers:
     - Frozen players cannot move
     - After freeze duration: unfreeze
  3. Check powerup pickups:
     - Shield: immunity for 10 seconds
     - Mega ball: larger explosion radius
     - Extra life: +1 life
  4. Check game over:
     - Timer expired → team with most lives wins
     - All opposing players eliminated → team wins

Scoring:
  - Freeze opponent: +5 points
  - Get frozen: -5 points
  - Pickup powerup: +2 points
```

**Key insight from Comet v2:** Freeze balls have a `ticksUntilExplode` counter and travel along directional vectors. The explosion pattern checks tiles in cardinal (and optionally diagonal) directions up to the ball's range.

### Battle Banzai

```
Game State:
  - Grid of Banzai tiles (3 states per team color)
  - Teams claim tiles by stepping on them
  - Locked tiles (3 consecutive steps) cannot be reclaimed
  - Puck mechanics: kick puck to claim tiles in path

Tick logic:
  1. Check player positions on Banzai tiles:
     - First step: tile turns team color (state 1)
     - Second step: tile solidifies (state 2)
     - Third step: tile locks (state 3)
  2. Puck movement:
     - Puck moves when kicked by player
     - Claims all tiles in path
     - Bounces off walls
  3. Score = number of locked tiles per team
  4. Timer-based (180 seconds default)
```

### Game Center (Lobby)

The game center is a matchmaking lobby outside of rooms:

```
1. Player enters game center
2. Joins queue for desired game type
3. Matchmaking system:
   - Wait for minimum players (2-8)
   - Wait max 60 seconds
   - Create game room instance
4. Players teleported to game room
5. Game plays out
6. Players returned to previous room on finish
```

### Leaderboard System

```
Leaderboard data:
  - Per game type (Freeze, Banzai, etc.)
  - Per period (daily, weekly, monthly, all-time)
  - Top 20 entries
  - Cached in Redis, refreshed on game completion

Redis keys:
  leaderboard:<gameType>:daily    → sorted set (score, userId)
  leaderboard:<gameType>:weekly   → sorted set
  leaderboard:<gameType>:monthly  → sorted set
  leaderboard:<gameType>:alltime  → sorted set
```

### Poll System

Polls are admin-created questionnaires shown to room visitors:

```
1. Admin creates poll (via admin panel or command)
2. Poll attached to room
3. When user enters room: send poll.question
4. User answers or dismisses
5. Answers stored for analytics
6. Results viewable by admin
```

---

## Caveats & Edge Cases

### 1. Game Furniture Requirements
Mini-games require specific furniture items to function:
- Freeze: freeze tiles, freeze gates (team entry), freeze timer, freeze exit
- Banzai: banzai tiles, banzai gates, banzai scoreboard, banzai puck
- Missing items = game cannot start. Validate furniture setup before allowing game.start.

### 2. Player Disconnection During Game
If a player disconnects mid-game, they should be treated as eliminated. Their team continues with remaining players. Score is preserved.

### 3. Team Balance
Allow uneven teams but warn the host. Some emulators force even teams -- pixel-server should make this configurable per room.

### 4. Game Timer Precision
Game timers must be tick-based (not wall-clock) for determinism. At 20 Hz, a 180-second game = 3600 ticks. Display updates sent every 20 ticks (1 second).

### 5. Concurrent Games Per Room
Only one game can run per room at a time. Attempting to start a second game while one is active must fail.

### 6. Snowball Physics
Snowballs in Freeze travel in a straight line. They can be blocked by non-freeze tiles or walls. The explosion radius check must respect tile boundaries.

---

## Improvements Over Legacy

| Area | Legacy | pixel-server |
|------|--------|-------------|
| **Game state** | Static fields on Room class | ECS components (composable, testable) |
| **Timer** | java.util.Timer (non-deterministic) | Fixed tick count (deterministic) |
| **Scoring** | Global mutable map | ECS component per team |
| **Matchmaking** | Not implemented | Game center queue with timeout |
| **Leaderboards** | Database query per view | Redis sorted sets with periodic refresh |

---

## Dependencies

- **Phase 3 (Room)** -- room worker, ECS world
- **Phase 6 (Furniture)** -- game tile items
- **Achievement system** -- game wins trigger achievements

---

## Testing Strategy

### Unit Tests
- Freeze ball trajectory calculation
- Banzai tile claiming state machine
- Score aggregation per team
- Timer tick countdown accuracy
- Powerup spawn probability

### Integration Tests
- Full Freeze game lifecycle (start → play → finish → scores)
- Banzai tile persistence across ticks
- Leaderboard update on game completion

### E2E Tests
- Two clients join teams, game starts, one team wins, scores displayed
