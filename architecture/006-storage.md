# Storage Strategy

## Overview

pixel-server uses two storage backends:

| Backend | Role |
|---|---|
| **PostgreSQL 16** | All durable state |
| **Redis 7 (Valkey)** | Ephemeral, latency-critical, pub/sub |

No ORM. All SQL is explicit `pgx/v5`. Schema evolves via Atlas HCL migrations.

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

## Complete schema catalogue

Every table and column here is present in at least two of the three reference implementations (Arcturus, Comet-v2, PlusEMU). Gaps from the original emulators are corrected to modern PostgreSQL standards.

### Core identity

```sql
CREATE TABLE users (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username     TEXT    NOT NULL UNIQUE,
    mail         TEXT,
    password     TEXT    NOT NULL,  -- bcrypt hash
    look         TEXT    NOT NULL DEFAULT '',
    gender       CHAR(1) NOT NULL DEFAULT 'M',
    motto        TEXT    NOT NULL DEFAULT '',
    home_room    BIGINT,
    rank         SMALLINT NOT NULL DEFAULT 1,
    online       BOOLEAN NOT NULL DEFAULT FALSE,
    last_online  TIMESTAMPTZ,
    machine_id   TEXT,              -- client fingerprint from handshake
    ip_last      INET,
    ip_reg       INET,
    credits      INT NOT NULL DEFAULT 0,
    activity_pts INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 1:1 with users. Separates mutable settings from identity to keep
-- the users table narrow for auth lookups.
-- Fields confirmed from Arcturus users_settings UPDATE query.
CREATE TABLE user_settings (
    user_id                    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    achievement_score          INT      NOT NULL DEFAULT 0,
    respects_received          INT      NOT NULL DEFAULT 0,
    respects_given             INT      NOT NULL DEFAULT 0,
    daily_respect_points       INT      NOT NULL DEFAULT 0,
    daily_pet_respect_points   INT      NOT NULL DEFAULT 0,
    online_time                INT      NOT NULL DEFAULT 0,  -- seconds accumulated
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

-- Multiple currency types per user (credits=0, pixels/pts=5, seasonal=103, …)
CREATE TABLE user_currency (
    user_id  BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type     SMALLINT NOT NULL,
    amount   INT      NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, type)
);

-- HC / VIP subscriptions
CREATE TABLE user_subscriptions (
    id                BIGINT   GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id           BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_type TEXT     NOT NULL,
    timestamp_start   BIGINT   NOT NULL,
    duration          INT      NOT NULL,   -- seconds
    active            BOOLEAN  NOT NULL DEFAULT TRUE
);

-- Saved figure outfits (wardrobe)
CREATE TABLE user_wardrobe (
    user_id BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot    SMALLINT NOT NULL,
    look    TEXT     NOT NULL,
    gender  CHAR(1)  NOT NULL DEFAULT 'M',
    PRIMARY KEY (user_id, slot)
);

-- Clothing items owned
CREATE TABLE user_clothing (
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    part_id  INT    NOT NULL,
    PRIMARY KEY (user_id, part_id)
);

-- Effects collection
CREATE TABLE user_effects (
    user_id            BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    effect_id          INT     NOT NULL,
    total_duration     INT     NOT NULL DEFAULT 0,
    is_activated       BOOLEAN NOT NULL DEFAULT FALSE,
    activate_timestamp BIGINT  NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, effect_id)
);

-- Ignore list
CREATE TABLE user_ignores (
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, target_id)
);

-- Profile relationship badges (love/like/skull)
CREATE TABLE user_relationships (
    user_id   BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    relation  SMALLINT NOT NULL,
    PRIMARY KEY (user_id, target_id)
);

-- Tags (displayed on profile page)
CREATE TABLE user_tags (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag     TEXT   NOT NULL,
    PRIMARY KEY (user_id, tag)
);

-- Favourite rooms
CREATE TABLE user_favourite_rooms (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, room_id)
);

-- Navigator saved searches
CREATE TABLE user_saved_searches (
    id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label   TEXT   NOT NULL DEFAULT '',
    query   TEXT   NOT NULL DEFAULT ''
);

-- Name change history
CREATE TABLE namechange_log (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users(id),
    old_name  TEXT   NOT NULL,
    new_name  TEXT   NOT NULL,
    ts        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Badges and Achievements

```sql
CREATE TABLE badges (
    id         TEXT PRIMARY KEY,       -- e.g. 'ACH_BasicClub1'
    name       TEXT NOT NULL,
    desc       TEXT NOT NULL DEFAULT '',
    badge_type TEXT NOT NULL DEFAULT 'normal'
);

CREATE TABLE user_badges (
    user_id  BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id TEXT     NOT NULL REFERENCES badges(id),
    slot     SMALLINT NOT NULL DEFAULT 0,   -- 0=not worn, 1-6=slot position
    PRIMARY KEY (user_id, badge_id)
);

CREATE TABLE achievements (
    id               INT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name             TEXT NOT NULL UNIQUE,
    category         TEXT NOT NULL DEFAULT '',
    levels           INT  NOT NULL DEFAULT 1,
    points_per_level INT  NOT NULL DEFAULT 0,
    reward_badge_id  TEXT REFERENCES badges(id)
);

CREATE TABLE achievement_levels (
    achievement_id   INT      NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    level            SMALLINT NOT NULL,
    goal_progress    INT      NOT NULL,
    reward_credits   INT      NOT NULL DEFAULT 0,
    reward_points    INT      NOT NULL DEFAULT 0,
    PRIMARY KEY (achievement_id, level)
);

CREATE TABLE user_achievements (
    user_id        BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id INT      NOT NULL REFERENCES achievements(id),
    progress       INT      NOT NULL DEFAULT 0,
    level          SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, achievement_id)
);

-- Talent tracks (citizenship + helpers programme)
CREATE TABLE talent_levels (
    track    TEXT     NOT NULL,   -- 'citizenship', 'helpers'
    level    SMALLINT NOT NULL,
    badge_id TEXT     REFERENCES badges(id),
    PRIMARY KEY (track, level)
);
```

### Rooms and Items

```sql
CREATE TABLE room_models (
    id        TEXT     PRIMARY KEY,
    heightmap TEXT     NOT NULL,
    door_x    SMALLINT NOT NULL DEFAULT 0,
    door_y    SMALLINT NOT NULL DEFAULT 0,
    door_dir  SMALLINT NOT NULL DEFAULT 2
);

CREATE TABLE rooms (
    id               BIGINT   GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id         BIGINT   REFERENCES users(id),
    name             TEXT     NOT NULL,
    description      TEXT     NOT NULL DEFAULT '',
    model            TEXT     NOT NULL REFERENCES room_models(id),
    password         TEXT     NOT NULL DEFAULT '',
    state            SMALLINT NOT NULL DEFAULT 0,  -- 0=open 1=locked 2=password 3=invisible
    category         INT,
    max_users        SMALLINT NOT NULL DEFAULT 25,
    score            INT      NOT NULL DEFAULT 0,
    allow_pets       BOOLEAN  NOT NULL DEFAULT TRUE,
    allow_pets_eat   BOOLEAN  NOT NULL DEFAULT FALSE,
    allow_walkthrough BOOLEAN NOT NULL DEFAULT FALSE,
    hide_wall        BOOLEAN  NOT NULL DEFAULT FALSE,
    floor_style      SMALLINT NOT NULL DEFAULT 0,
    wall_style       SMALLINT NOT NULL DEFAULT 0,
    landscape        REAL     NOT NULL DEFAULT 0,
    wall_height      SMALLINT NOT NULL DEFAULT -1,
    chat_mode        SMALLINT NOT NULL DEFAULT 0,
    chat_size        SMALLINT NOT NULL DEFAULT 0,
    chat_speed       SMALLINT NOT NULL DEFAULT 0,
    chat_distance    SMALLINT NOT NULL DEFAULT 0,
    chat_protection  SMALLINT NOT NULL DEFAULT 0,
    who_can_mute     SMALLINT NOT NULL DEFAULT 0,
    who_can_kick     SMALLINT NOT NULL DEFAULT 0,
    who_can_ban      SMALLINT NOT NULL DEFAULT 0,
    trade_state      SMALLINT NOT NULL DEFAULT 2,
    push_pull_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    name_tsv         TSVECTOR GENERATED ALWAYS AS
                     (to_tsvector('english', name || ' ' || description)) STORED,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON rooms USING GIN (name_tsv);
CREATE INDEX ON rooms (owner_id);
CREATE INDEX ON rooms (category) WHERE state != 3;
CREATE INDEX ON rooms (score DESC) WHERE state = 0;

CREATE TABLE room_rights (
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE room_bans (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ban_expires BIGINT NOT NULL DEFAULT 0,
    reason      TEXT   NOT NULL DEFAULT ''
);
CREATE INDEX ON room_bans (room_id);

CREATE TABLE room_mutes (
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE room_word_filter (
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    word    TEXT   NOT NULL,
    PRIMARY KEY (room_id, word)
);

CREATE TABLE room_promotions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id         BIGINT NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    title           TEXT   NOT NULL DEFAULT '',
    description     TEXT   NOT NULL DEFAULT '',
    end_timestamp   BIGINT NOT NULL DEFAULT 0,
    category        INT    NOT NULL DEFAULT 0,
    score           INT    NOT NULL DEFAULT 0
);

-- Base furniture definitions
CREATE TABLE furniture (
    id                  INT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    item_name           TEXT     NOT NULL UNIQUE,
    type                CHAR(1)  NOT NULL DEFAULT 'i',   -- 'i'=floor 's'=wall
    width               SMALLINT NOT NULL DEFAULT 1,
    length              SMALLINT NOT NULL DEFAULT 1,
    stack_height        REAL     NOT NULL DEFAULT 1.0,
    can_stack           BOOLEAN  NOT NULL DEFAULT TRUE,
    can_sit             BOOLEAN  NOT NULL DEFAULT FALSE,
    is_walkable         BOOLEAN  NOT NULL DEFAULT FALSE,
    allow_recycle       BOOLEAN  NOT NULL DEFAULT TRUE,
    allow_trade         BOOLEAN  NOT NULL DEFAULT TRUE,
    allow_marketplace   BOOLEAN  NOT NULL DEFAULT TRUE,
    interaction_type    TEXT     NOT NULL DEFAULT 'default',
    interaction_modes   SMALLINT NOT NULL DEFAULT 1,
    vending_ids         TEXT     NOT NULL DEFAULT '',
    sprite_id           INT      NOT NULL DEFAULT 0,
    effect_id_male      INT      NOT NULL DEFAULT 0,
    effect_id_female    INT      NOT NULL DEFAULT 0
);

-- Placed item instances (in-room or in inventory; room_id NULL = inventory)
CREATE TABLE items (
    id            BIGINT   GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    base_item     INT      NOT NULL REFERENCES furniture(id),
    owner_id      BIGINT   REFERENCES users(id),
    room_id       BIGINT   REFERENCES rooms(id),
    x             SMALLINT,
    y             SMALLINT,
    z             REAL,
    rot           SMALLINT NOT NULL DEFAULT 0,
    extra_data    TEXT     NOT NULL DEFAULT '',
    limited_id    INT,
    limited_total INT
);
CREATE INDEX ON items (room_id)  WHERE room_id IS NOT NULL;
CREATE INDEX ON items (owner_id) WHERE owner_id IS NOT NULL;

-- Teleporter link pairs
CREATE TABLE items_teles (
    item_id    BIGINT PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
    target_id  BIGINT REFERENCES items(id)
);

-- Moodlight states
CREATE TABLE items_moodlight (
    item_id        BIGINT   PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
    enabled        BOOLEAN  NOT NULL DEFAULT FALSE,
    current_preset SMALLINT NOT NULL DEFAULT 1,
    preset_1       TEXT     NOT NULL DEFAULT '1,#000000,255,false',
    preset_2       TEXT     NOT NULL DEFAULT '2,#000000,255,false',
    preset_3       TEXT     NOT NULL DEFAULT '3,#000000,255,false'
);

-- Limited edition stack counters
CREATE TABLE items_limited_edition (
    base_item     INT PRIMARY KEY REFERENCES furniture(id),
    limited_sold  INT NOT NULL DEFAULT 0,
    limited_total INT NOT NULL DEFAULT 0
);

-- Wired prize box rewards
CREATE TABLE items_wired_rewards (
    item_id     BIGINT   NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    reward_type SMALLINT NOT NULL,
    reward_data TEXT     NOT NULL DEFAULT ''
);

-- Crackable item reward tables
CREATE TABLE items_crackable_rewards (
    item_id     BIGINT   NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    base_item   INT      NOT NULL,
    probability SMALLINT NOT NULL DEFAULT 1
);
```

### Bots

```sql
CREATE TABLE bots (
    id               BIGINT   GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id         BIGINT   REFERENCES users(id),
    room_id          BIGINT   REFERENCES rooms(id),
    name             TEXT     NOT NULL,
    motto            TEXT     NOT NULL DEFAULT '',
    look             TEXT     NOT NULL DEFAULT '',
    gender           CHAR(1)  NOT NULL DEFAULT 'M',
    x                SMALLINT NOT NULL DEFAULT 0,
    y                SMALLINT NOT NULL DEFAULT 0,
    z                REAL     NOT NULL DEFAULT 0,
    rot              SMALLINT NOT NULL DEFAULT 0,
    dance            SMALLINT NOT NULL DEFAULT 0,
    automatic_chat   BOOLEAN  NOT NULL DEFAULT FALSE,
    chat_delay       INT      NOT NULL DEFAULT 10,
    chat_random      BOOLEAN  NOT NULL DEFAULT FALSE,
    walking_enabled  BOOLEAN  NOT NULL DEFAULT FALSE,
    follow_enabled   BOOLEAN  NOT NULL DEFAULT FALSE
);

CREATE TABLE bot_chat (
    id       BIGINT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bot_id   BIGINT  NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    message  TEXT    NOT NULL,
    is_shout BOOLEAN NOT NULL DEFAULT FALSE
);
```

### Catalog and Economy

```sql
CREATE TABLE catalog_pages (
    id          INT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id   INT      NOT NULL DEFAULT -1,
    caption     TEXT     NOT NULL,
    page_layout TEXT     NOT NULL DEFAULT 'default_3x3',
    icon_image  TEXT     NOT NULL DEFAULT '',
    visible     BOOLEAN  NOT NULL DEFAULT TRUE,
    enabled     BOOLEAN  NOT NULL DEFAULT TRUE,
    min_rank    SMALLINT NOT NULL DEFAULT 1,
    order_num   INT      NOT NULL DEFAULT 0
);

CREATE TABLE catalog_items (
    id           INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    page_id      INT     NOT NULL REFERENCES catalog_pages(id),
    item_ids     TEXT    NOT NULL,
    catalog_name TEXT    NOT NULL,
    cost_credits INT     NOT NULL DEFAULT 0,
    cost_points  INT     NOT NULL DEFAULT 0,
    points_type  SMALLINT NOT NULL DEFAULT 0,
    amount       INT     NOT NULL DEFAULT 1,
    limited_stacks INT   NOT NULL DEFAULT 0,
    have_offer   BOOLEAN NOT NULL DEFAULT FALSE,
    offer_id     INT     NOT NULL DEFAULT -1,
    badge_id     TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE catalog_featured_pages (
    id           INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    caption      TEXT    NOT NULL,
    image        TEXT    NOT NULL DEFAULT '',
    type         SMALLINT NOT NULL DEFAULT 1,
    page_id      INT     REFERENCES catalog_pages(id),
    page_name    TEXT    NOT NULL DEFAULT '',
    product_name TEXT    NOT NULL DEFAULT '',
    position     INT     NOT NULL DEFAULT 0,
    visible      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE catalog_club_offers (
    id            INT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type          TEXT     NOT NULL DEFAULT 'HABBO_CLUB',
    cost_credits  INT      NOT NULL DEFAULT 0,
    cost_points   INT      NOT NULL DEFAULT 0,
    points_type   SMALLINT NOT NULL DEFAULT 5,
    duration      INT      NOT NULL DEFAULT 31,
    duration_type TEXT     NOT NULL DEFAULT 'd'
);

CREATE TABLE catalog_target_offers (
    id                INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    identifier        TEXT    NOT NULL UNIQUE,
    product_code      TEXT    NOT NULL DEFAULT '',
    cost_credits      INT     NOT NULL DEFAULT 0,
    cost_activity_pts INT     NOT NULL DEFAULT 0,
    activity_pts_type SMALLINT NOT NULL DEFAULT 0,
    purchase_limit    INT     NOT NULL DEFAULT 1,
    expiry_timestamp  BIGINT  NOT NULL DEFAULT 0,
    active            BOOLEAN NOT NULL DEFAULT TRUE,
    title             TEXT    NOT NULL DEFAULT '',
    description       TEXT    NOT NULL DEFAULT '',
    image_url         TEXT    NOT NULL DEFAULT '',
    item_ids          TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE catalog_gift_wrapping (
    id        INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    box_id    INT NOT NULL,
    ribbon_id INT NOT NULL,
    image     TEXT NOT NULL DEFAULT '',
    sprite_id INT NOT NULL DEFAULT 0
);

CREATE TABLE vouchers (
    code         TEXT     PRIMARY KEY,
    type         SMALLINT NOT NULL DEFAULT 0,
    value        INT      NOT NULL DEFAULT 0,
    uses         INT      NOT NULL DEFAULT 1,
    current_uses INT      NOT NULL DEFAULT 0
);

CREATE TABLE marketplace_items (
    id           BIGINT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    item_id      BIGINT  NOT NULL UNIQUE REFERENCES items(id) ON DELETE CASCADE,
    seller_id    BIGINT  NOT NULL REFERENCES users(id),
    asking_price INT     NOT NULL DEFAULT 0,
    listed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON marketplace_items (listed_at DESC) WHERE expires_at > NOW();
```

### Messenger (Social)

```sql
CREATE TABLE messenger_friendships (
    user_one_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_two_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friends_since TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_one_id, user_two_id),
    CHECK (user_one_id < user_two_id)
);
CREATE INDEX ON messenger_friendships (user_two_id);

CREATE TABLE messenger_requests (
    from_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (from_id, to_id)
);

-- Persistent offline messages.
-- Delivered as a batch when the recipient next logs in, then marked read.
-- Hard-deleted after confirmed delivery (cleanup job runs nightly).
CREATE TABLE messenger_messages (
    id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    from_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message  TEXT   NOT NULL,
    sent_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read     BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX ON messenger_messages (to_id) WHERE read = FALSE;
```

### Groups and Forums

```sql
CREATE TABLE groups (
    id          INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    badge       TEXT    NOT NULL DEFAULT '',
    owner_id    BIGINT  NOT NULL REFERENCES users(id),
    home_room   BIGINT  REFERENCES rooms(id),
    state       SMALLINT NOT NULL DEFAULT 0,
    colour_1    INT     NOT NULL DEFAULT 0,
    colour_2    INT     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_memberships (
    group_id  INT     NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id   BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rank      SMALLINT NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE group_requests (
    group_id INT    NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE group_forum_threads (
    id         BIGINT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    group_id   INT     NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    author_id  BIGINT  NOT NULL REFERENCES users(id),
    subject    TEXT    NOT NULL,
    pinned     BOOLEAN NOT NULL DEFAULT FALSE,
    locked     BOOLEAN NOT NULL DEFAULT FALSE,
    deleted    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON group_forum_threads (group_id, created_at DESC);

CREATE TABLE group_forum_posts (
    id         BIGINT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    thread_id  BIGINT  NOT NULL REFERENCES group_forum_threads(id) ON DELETE CASCADE,
    author_id  BIGINT  NOT NULL REFERENCES users(id),
    body       TEXT    NOT NULL,
    deleted    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Trade Logs

```sql
CREATE TABLE trade_logs (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users(id),
    target_id BIGINT NOT NULL REFERENCES users(id),
    room_id   BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE trade_log_items (
    trade_id  BIGINT   NOT NULL REFERENCES trade_logs(id) ON DELETE CASCADE,
    item_id   BIGINT   NOT NULL,
    base_item INT      NOT NULL,
    direction SMALLINT NOT NULL DEFAULT 0   -- 0=offered by user_id, 1=by target
);
```

### Audit and Moderation

All log tables are **partitioned by month**. Monthly partitions are pre-created by a maintenance job.

```sql
-- Room chat log (for mod-tool chatlog view)
CREATE TABLE chat_log (
    id      BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    room_id BIGINT NOT NULL,
    message TEXT   NOT NULL,
    type    SMALLINT NOT NULL DEFAULT 0,   -- 0=say 1=shout 2=whisper
    ts      TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (ts);
CREATE INDEX ON chat_log (room_id, ts DESC);
CREATE INDEX ON chat_log (user_id, ts DESC);

-- Messenger log (for mod-tool private chat view)
CREATE TABLE messenger_chat_log (
    id      BIGINT GENERATED ALWAYS AS IDENTITY,
    from_id BIGINT NOT NULL,
    to_id   BIGINT NOT NULL,
    message TEXT   NOT NULL,
    ts      TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (ts);
CREATE INDEX ON messenger_chat_log (from_id, ts DESC);
CREATE INDEX ON messenger_chat_log (to_id,   ts DESC);

-- Session / IP login journal
CREATE TABLE session_log (
    id         BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id    BIGINT NOT NULL,
    ip         INET   NOT NULL,
    machine_id TEXT,
    ts         TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (ts);
CREATE INDEX ON session_log (user_id, ts DESC);
CREATE INDEX ON session_log (ip,      ts DESC);

-- Room visit tracking (mod-tool + analytics)
CREATE TABLE room_visit_log (
    id         BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id    BIGINT NOT NULL,
    room_id    BIGINT NOT NULL,
    entered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    exited_at  TIMESTAMPTZ
) PARTITION BY RANGE (entered_at);
CREATE INDEX ON room_visit_log (user_id, entered_at DESC);
CREATE INDEX ON room_visit_log (room_id, entered_at DESC);

-- Staff command audit log
CREATE TABLE command_log (
    id      BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    room_id BIGINT,
    command TEXT   NOT NULL,
    args    TEXT   NOT NULL DEFAULT '',
    ts      TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (ts);

-- Economy / currency event log
CREATE TABLE economy_log (
    id          BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id     BIGINT   NOT NULL,
    change_type TEXT     NOT NULL,   -- 'purchase','trade','gift','hc_payday','refund'
    currency    SMALLINT NOT NULL DEFAULT 0,
    delta       INT      NOT NULL,
    reference   TEXT,
    ts          TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (ts);

-- Global account / IP / machine bans
CREATE TABLE bans (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT  REFERENCES users(id),
    ip         INET,
    machine_id TEXT,
    ban_type   TEXT    NOT NULL DEFAULT 'account',
    reason     TEXT    NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    issued_by  BIGINT  REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON bans (user_id) WHERE expires_at IS NULL OR expires_at > NOW();
CREATE INDEX ON bans (ip)      WHERE ip IS NOT NULL AND (expires_at IS NULL OR expires_at > NOW());

-- Mod-tool ticket queue (call for help)
CREATE TABLE moderation_tickets (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    state        TEXT   NOT NULL DEFAULT 'open',
    submitter_id BIGINT NOT NULL REFERENCES users(id),
    reported_id  BIGINT NOT NULL REFERENCES users(id),
    moderator_id BIGINT REFERENCES users(id),
    category_id  INT    NOT NULL DEFAULT 0,
    message      TEXT   NOT NULL DEFAULT '',
    room_id      BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON moderation_tickets (state) WHERE state = 'open';

CREATE TABLE moderation_actions (
    id               INT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id      INT,
    name             TEXT NOT NULL,
    message          TEXT NOT NULL DEFAULT '',
    description      TEXT NOT NULL DEFAULT '',
    ban_hours        INT  NOT NULL DEFAULT 0,
    avatar_ban_hours INT  NOT NULL DEFAULT 0,
    mute_hours       INT  NOT NULL DEFAULT 0,
    trade_lock_hours INT  NOT NULL DEFAULT 0
);
```

### Miscellaneous Features

```sql
-- Camera photos
CREATE TABLE photos (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id   BIGINT NOT NULL REFERENCES users(id),
    room_id    BIGINT NOT NULL,
    image_url  TEXT   NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Quest definitions and player progress
CREATE TABLE quests (
    id            INT  GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category      TEXT NOT NULL,
    series        INT  NOT NULL DEFAULT 1,
    name          TEXT NOT NULL,
    goal_type     INT  NOT NULL DEFAULT 0,
    goal_data     INT  NOT NULL DEFAULT 0,
    reward_type   INT  NOT NULL DEFAULT 0,
    reward_amount INT  NOT NULL DEFAULT 0,
    start_data    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE user_quest_progress (
    user_id   BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quest_id  INT     NOT NULL REFERENCES quests(id),
    progress  INT     NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, quest_id)
);

-- Seasonal activity calendar
CREATE TABLE activity_calendar (
    id          INT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    year        SMALLINT NOT NULL,
    day         SMALLINT NOT NULL,
    reward_type TEXT     NOT NULL DEFAULT 'badge',
    reward_data TEXT     NOT NULL DEFAULT ''
);

CREATE TABLE user_calendar_claims (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_id     INT    NOT NULL REFERENCES activity_calendar(id),
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, day_id)
);

-- Crafting / recycler recipes
CREATE TABLE crafting_recipes (
    id          INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    result_item INT     NOT NULL REFERENCES furniture(id),
    enabled     BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE crafting_ingredients (
    recipe_id    INT      NOT NULL REFERENCES crafting_recipes(id) ON DELETE CASCADE,
    furniture_id INT      NOT NULL REFERENCES furniture(id),
    amount       SMALLINT NOT NULL DEFAULT 1
);

CREATE TABLE user_crafting_recipes (
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipe_id INT    NOT NULL REFERENCES crafting_recipes(id),
    PRIMARY KEY (user_id, recipe_id)
);

-- Server-side configuration (replaces config file for hot-reload)
CREATE TABLE server_config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

-- i18n / localisation strings
CREATE TABLE server_locale (
    key      TEXT NOT NULL,
    language TEXT NOT NULL DEFAULT 'en',
    value    TEXT NOT NULL,
    PRIMARY KEY (key, language)
);

-- Global chat / hotel word filter
CREATE TABLE wordfilter (
    id          INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    word        TEXT    NOT NULL UNIQUE,
    is_banned   BOOLEAN NOT NULL DEFAULT TRUE,
    replacement TEXT    NOT NULL DEFAULT '****'
);

-- Polls
CREATE TABLE polls (
    id       INT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    question TEXT     NOT NULL,
    end_at   TIMESTAMPTZ,
    min_rank SMALLINT NOT NULL DEFAULT 1
);

CREATE TABLE poll_answers (
    id          INT    GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    poll_id     INT    NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    answer      TEXT   NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Async write pattern for log tables

All partitioned log tables are written via an **asynchronous batch writer** in each service — the simulation goroutine never blocks on a log INSERT:

```go
type AsyncLogWriter struct {
    pool  *pgxpool.Pool
    queue chan LogEntry
}

func (w *AsyncLogWriter) Write(e LogEntry) {
    select {
    case w.queue <- e:
    default:
        metrics.LogDropped.Inc()  // non-critical; monitored via Prometheus
    }
}

func (w *AsyncLogWriter) flushLoop() {
    ticker := time.NewTicker(500 * time.Millisecond)
    buf := make([]LogEntry, 0, 256)
    for {
        select {
        case e := <-w.queue:
            buf = append(buf, e)
            if len(buf) >= 256 {
                w.flush(buf)
                buf = buf[:0]
            }
        case <-ticker.C:
            if len(buf) > 0 {
                w.flush(buf)
                buf = buf[:0]
            }
        }
    }
}

func (w *AsyncLogWriter) flush(entries []LogEntry) {
    // pgx CopyFrom into the appropriate table
}
```

This directly fixes the biggest performance problem in the legacy emulators: `INSERT` inside `Room#cycle()` blocking the simulation thread.

---

## Offline messenger message delivery

On login, `social-svc` delivers pending messages and clears them in one round-trip:

```go
rows, _ := pool.Query(ctx,
    `SELECT id, from_id, message, sent_at
     FROM messenger_messages
     WHERE to_id = $1 AND read = FALSE
     ORDER BY sent_at`,
    userID)
// ... publish s2c packets per message ...
pool.Exec(ctx,
    `UPDATE messenger_messages SET read = TRUE
     WHERE to_id = $1 AND read = FALSE`,
    userID)
```

A nightly maintenance job hard-deletes `messenger_messages` older than 30 days with `read = TRUE`.

---

## Redis usage patterns

### Session store
```
HSET session:<UUIDv7>  userID <int64>  roomID <int64>  gameNode <pod>  encKey <hex>
EXPIRE session:<UUIDv7> 3600
```

### Room presence
```
SADD  room:presence:<roomID>  <sessionID>
SCARD room:presence:<roomID>  → fast headcount for hotel view
```

### Ban hot-path cache
```
SET  ban:user:<userID>    "1"   EX <seconds-until-expiry>
SET  ban:ip:<ip>          "1"   EX <seconds-until-expiry>
SET  ban:machine:<id>     "1"   EX <seconds-until-expiry>
```
Checked by the gateway on every new connection before a DB round-trip.  
On ban issue: write DB row → set Redis key → `PUBLISH ban.kick:<userID>`.

### Rate limiting (chat, purchases, friend requests)
```lua
-- Sliding window Lua script
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
local count = redis.call('ZCARD', key)
if count < limit then
    redis.call('ZADD', key, now, now)
    redis.call('EXPIRE', key, window)
    return 1
end
return 0
```

### Online count
```
INCR online:total      -- on session.authenticated
DECR online:total      -- on session.disconnected
```

### Navigator room cache (sorted set)
```
ZADD navigator:<category>  <score>  <roomID>
HSET roominfo:<roomID>     name <n>  users <u>  owner <o>  ...
```
Invalidated via `navigator.room_updated` NATS events.

### Leaderboards
```
ZADD leaderboard:rooms:score  <score>  <roomID>   -- top rooms
ZADD leaderboard:users:credits <credits> <userID>  -- top users
```
Refreshed from DB every 5 minutes by `navigator-svc`.

---

## Migration workflow

Schema source: `pkg/storage/migrations/schema.hcl` (Atlas HCL).  
Monthly log partitions: pre-created by `services/maintenance` — a lightweight cron service that runs `CREATE TABLE chatlog_YYYY_MM PARTITION OF chat_log FOR VALUES FROM (...) TO (...)` for the next 2 months.

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
| Room heightmaps | Loaded into `game-svc` in-memory; invalidated on item change |
