# Storage Strategy

## Overview

pixelsv uses two storage backends:

| Backend | Role |
|---|---|
| **PostgreSQL 16** | All durable state |
| **Redis 7 (Valkey)** | Ephemeral, latency-critical, pub/sub, ECS snapshots |

No ORM. All SQL is explicit `pgx/v5`. Schema evolves via Atlas HCL migrations.

## Role-Aware Storage Access

Not every role needs every storage backend:

| Role | PostgreSQL | Redis | Reason |
|---|---|---|---|
| `gateway` | No | Yes | Session store, ban hot-path |
| `game` | Yes | Yes | Room/item persistence, ECS snapshots, presence |
| `auth` | Yes | Yes | User lookup, ticket store |
| `social` | Yes | Yes | Friends, messages, online count |
| `navigator` | Yes | Yes | Room search, navigator cache |
| `catalog` | Yes | No | Catalog pages, economy |
| `moderation` | Yes | Yes | Bans, tickets, ban cache |
| `api` | Yes | Yes | Full admin access |
| `jobs` | Yes | Yes | Partition maintenance, leaderboard refresh |
| `all` | Yes | Yes | Everything |

Each role process creates connections only to the backends it needs.

---

## Why PostgreSQL over MySQL

The legacy emulators use MySQL/MariaDB. PostgreSQL 16 is preferred because:

- **`LISTEN/NOTIFY`** — push-model room config reload without polling.
- **`COPY` protocol** — bulk item grants via pgx's `CopyFrom` API.
- **`jsonb`** — flexible item extra-data without schema churn.
- **Range partitioning** — auto-partition high-write log tables by month.
- **Full-text search** with `tsvector` on room names/descriptions; no Elasticsearch.
- **Row-level security** — future hosted/multi-tenant deployability.
- No JVM, no MySQL connector quirks, standard `pgsql` tooling.

---

## Storage Ports and Adapters

Storage follows hexagonal architecture:

- **Ports** (interfaces): Defined in `pkg/storage/interfaces/` for generic primitives, and in each realm's `domain/` for domain-specific repositories.
- **Adapters** (implementations): `pkg/storage/postgres/` and `pkg/storage/redis/` for generic adapters; `internal/<realm>/adapters/postgres/` for domain-specific repository implementations.

```go
// pkg/storage/interfaces/querier.go
type RowQuerier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// internal/game/domain/repository.go
type RoomRepository interface {
    GetByID(ctx context.Context, id int64) (*Room, error)
    Create(ctx context.Context, r *Room) error
}

// internal/game/adapters/postgres/room_repository.go
type roomRepo struct {
    pool *pgxpool.Pool
}
func (r *roomRepo) GetByID(ctx context.Context, id int64) (*Room, error) {
    // explicit pgx query
}
```

Domain packages define repository interfaces. Adapter packages implement them. The wiring happens in `cmd/pixelsv`.

---

## Complete schema catalogue

Every table and column here is present in at least two of the three reference implementations (Arcturus, Comet-v2, PlusEMU). Gaps from the original emulators are corrected to modern PostgreSQL standards.

### Core identity

```sql
CREATE TABLE users (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username     TEXT    NOT NULL UNIQUE,
    mail         TEXT,
    password     TEXT    NOT NULL,
    look         TEXT    NOT NULL DEFAULT '',
    gender       CHAR(1) NOT NULL DEFAULT 'M',
    motto        TEXT    NOT NULL DEFAULT '',
    home_room    BIGINT,
    rank         SMALLINT NOT NULL DEFAULT 1,
    online       BOOLEAN NOT NULL DEFAULT FALSE,
    last_online  TIMESTAMPTZ,
    machine_id   TEXT,
    ip_last      INET,
    ip_reg       INET,
    credits      INT NOT NULL DEFAULT 0,
    activity_pts INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_settings (
    user_id                    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    achievement_score          INT      NOT NULL DEFAULT 0,
    respects_received          INT      NOT NULL DEFAULT 0,
    respects_given             INT      NOT NULL DEFAULT 0,
    daily_respect_points       INT      NOT NULL DEFAULT 0,
    daily_pet_respect_points   INT      NOT NULL DEFAULT 0,
    online_time                INT      NOT NULL DEFAULT 0,
    hof_points                 INT      NOT NULL DEFAULT 0,
    mute_end_timestamp         BIGINT   NOT NULL DEFAULT 0,
    guild_id                   INT      NOT NULL DEFAULT 0,
    login_streak               INT      NOT NULL DEFAULT 0,
    volume_system              SMALLINT NOT NULL DEFAULT 100,
    volume_furni               SMALLINT NOT NULL DEFAULT 100,
    volume_trax                SMALLINT NOT NULL DEFAULT 100,
    chat_color                 SMALLINT NOT NULL DEFAULT 0,
    block_following            BOOLEAN  NOT NULL DEFAULT FALSE,
    block_friendrequests       BOOLEAN  NOT NULL DEFAULT FALSE,
    block_roominvites          BOOLEAN  NOT NULL DEFAULT FALSE,
    block_camera_follow        BOOLEAN  NOT NULL DEFAULT FALSE,
    block_alerts               BOOLEAN  NOT NULL DEFAULT FALSE,
    old_chat                   BOOLEAN  NOT NULL DEFAULT FALSE,
    ignore_bots                BOOLEAN  NOT NULL DEFAULT FALSE,
    ignore_pets                BOOLEAN  NOT NULL DEFAULT FALSE,
    allow_name_change          BOOLEAN  NOT NULL DEFAULT FALSE,
    can_trade                  BOOLEAN  NOT NULL DEFAULT TRUE,
    perk_trade                 BOOLEAN  NOT NULL DEFAULT TRUE,
    nux                        BOOLEAN  NOT NULL DEFAULT TRUE,
    has_default_saved_searches BOOLEAN  NOT NULL DEFAULT FALSE,
    max_friends                SMALLINT NOT NULL DEFAULT 300,
    max_rooms                  SMALLINT NOT NULL DEFAULT 25,
    rent_space_id              INT      NOT NULL DEFAULT 0,
    rent_space_endtime         BIGINT   NOT NULL DEFAULT 0,
    club_expire_timestamp      BIGINT   NOT NULL DEFAULT 0,
    last_hc_payday             BIGINT   NOT NULL DEFAULT 0,
    hc_gifts_claimed           INT      NOT NULL DEFAULT 0,
    forums_post_count          INT      NOT NULL DEFAULT 0,
    talent_citizenship_level   SMALLINT NOT NULL DEFAULT 0,
    talent_helpers_level       SMALLINT NOT NULL DEFAULT 0,
    ui_flags                   INT      NOT NULL DEFAULT 0
);

CREATE TABLE user_currency (
    user_id  BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type     SMALLINT NOT NULL,
    amount   INT      NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, type)
);

CREATE TABLE user_subscriptions (
    id                BIGINT   GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id           BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_type TEXT     NOT NULL,
    timestamp_start   BIGINT   NOT NULL,
    duration          INT      NOT NULL,
    active            BOOLEAN  NOT NULL DEFAULT TRUE
);

CREATE TABLE user_wardrobe (
    user_id BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot    SMALLINT NOT NULL,
    look    TEXT     NOT NULL,
    gender  CHAR(1)  NOT NULL DEFAULT 'M',
    PRIMARY KEY (user_id, slot)
);

CREATE TABLE user_clothing (
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    part_id  INT    NOT NULL,
    PRIMARY KEY (user_id, part_id)
);

CREATE TABLE user_effects (
    user_id            BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    effect_id          INT     NOT NULL,
    total_duration     INT     NOT NULL DEFAULT 0,
    is_activated       BOOLEAN NOT NULL DEFAULT FALSE,
    activate_timestamp BIGINT  NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, effect_id)
);

CREATE TABLE user_ignores (
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, target_id)
);

CREATE TABLE user_relationships (
    user_id   BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    relation  SMALLINT NOT NULL,
    PRIMARY KEY (user_id, target_id)
);

CREATE TABLE user_tags (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag     TEXT   NOT NULL,
    PRIMARY KEY (user_id, tag)
);

CREATE TABLE user_favourite_rooms (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, room_id)
);

CREATE TABLE user_saved_searches (
    id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label   TEXT   NOT NULL DEFAULT '',
    query   TEXT   NOT NULL DEFAULT ''
);

CREATE TABLE namechange_log (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users(id),
    old_name  TEXT   NOT NULL,
    new_name  TEXT   NOT NULL,
    ts        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Remaining schema sections (badges, rooms, items, catalog, messenger, groups, trade, audit, misc) are preserved unchanged — see git history for full schema catalogue. The schema is authoritative and identical regardless of deployment mode.

---

## Async write pattern for log tables

All partitioned log tables are written via an **asynchronous batch writer** — the simulation goroutine never blocks on a log INSERT:

```go
type AsyncLogWriter struct {
    pool  *pgxpool.Pool
    queue chan LogEntry
}

func (w *AsyncLogWriter) Write(e LogEntry) {
    select {
    case w.queue <- e:
    default:
        metrics.LogDropped.Inc()
    }
}
```

The `AsyncLogWriter` is instantiated per role process and shared across realm modules within that process.

---

## Redis usage patterns

### Session store
```
HSET session:<UUIDv7>  userID <int64>  roomID <int64>  instanceID <pod>
EXPIRE session:<UUIDv7> 3600
```

### Room presence
```
SADD  room:presence:<roomID>  <sessionID>
SCARD room:presence:<roomID>
```

### Room ECS snapshots (crash recovery)
```
SET room:snapshot:<roomID> <ark-serde JSON>
```

### Room ownership registry (distributed mode)
```
HSET room:owner <roomID> <instanceID>
```

### Ban hot-path cache
```
SET  ban:user:<userID>    "1"   EX <seconds-until-expiry>
SET  ban:ip:<ip>          "1"   EX <seconds-until-expiry>
SET  ban:machine:<id>     "1"   EX <seconds-until-expiry>
```

### Rate limiting
Sliding window Lua script (unchanged).

### Online count
```
INCR online:total      -- on session.authenticated
DECR online:total      -- on session.disconnected
```

### Navigator room cache
```
ZADD navigator:<category>  <score>  <roomID>
HSET roominfo:<roomID>     name <n>  users <u>  owner <o>
```

### Leaderboards
```
ZADD leaderboard:rooms:score  <score>  <roomID>
ZADD leaderboard:users:credits <credits> <userID>
```
Refreshed from DB every 5 minutes by the `jobs` role scheduler.

---

## Migration workflow

Schema source: `pkg/storage/migrations/schema.hcl` (Atlas HCL).
Monthly log partitions: pre-created by `pixelsv jobs` maintenance role.

CI checks: `atlas schema diff` must produce an empty diff against the committed schema file.

---

## What NOT to store in Redis

| Data | Reason to keep in PostgreSQL |
|---|---|
| Inventory items | Too large; item loss on Redis eviction is unacceptable |
| Badge collection | Low-write; Redis TTL would cause loss |
| Achievement progress | Correctness-critical; must survive restarts |
| Authoritative ban record | Redis is a cache only; primary in PostgreSQL |
| Friendship graph | Read-heavy but must be consistent; query + 60 s cache |
| Room heightmaps | Loaded into the game realm in-memory; invalidated on item change |
