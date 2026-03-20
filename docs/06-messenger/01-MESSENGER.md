# Messenger

## Overview

The messenger realm handles private messaging, friend lists, friend requests,
and social interactions. It follows the same hexagonal pattern as all other
realms: binary packets for real-time interaction, REST endpoints for
administration, and CLI commands for operator tooling.

## Initialization

When a client sends the messenger init packet the server:

1. Resolves the user's effective friend limit (see [Permissions](#permissions)).
2. Sends the `MessengerInit` packet with the resolved limit, normal limit, and
   VIP limit.
3. Fetches and atomically deletes pending offline messages.
4. Sends the friend list in fragments, flagging each friend with
   `persistedMessage: true` when that friend has offline messages waiting.
5. Delivers each offline message as a `NewConsoleMessage` packet with
   `secondsSinceSent` set to the elapsed time since the original send.
6. Notifies all online friends of the user's online status.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `messenger.init` | 2781 | C2S | (empty) |
| `messenger.init` | 2781 | S2C | UserFriendLimit, NormalLimit, ExtendedLimit, Categories |
| `messenger.friends` | — | S2C | TotalFragments, FragmentNumber, Friends[] |
| `messenger.get_friends` | — | C2S | (empty) |
| `messenger.get_requests` | — | C2S | (empty) |

## Private Messaging

### Send flow

1. Message is trimmed and truncated to 255 characters.
2. Flood control is checked. Violations accumulate per connection; after
   `MESSENGER_FLOOD_VIOLATIONS` rapid sends the connection is muted for
   `MESSENGER_FLOOD_MUTE_SECONDS` seconds.
3. Plugin event `PrivateMessageSent` fires (cancellable).
4. Friendship is verified — non-friends receive `ErrNotFriends`.
5. Message is written to `messenger_message_log` for auditing.
6. If the recipient is online the message is published via the broadcaster.
   If offline it is also written to `offline_messages` for later delivery.

### Packets

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `messenger.send_msg` | 3567 | C2S | UserID, Message |
| `messenger.new_message` | 1587 | S2C | SenderID, Message, SecondsSinceSent |
| `messenger.message_error` | 3359 | S2C | ErrorCode, UserID |

### Error codes

| Code | Meaning |
|------|---------|
| 0 | Not friends |
| 1 | Sender muted (flood) |
| 2 | Generic error |

## Friend Requests

| Packet | ID | Direction | Fields |
|--------|----|-----------|--------|
| `messenger.request_buddy` | — | C2S | Username |
| `messenger.buddy_requests` | — | S2C | Requests[] |
| `messenger.accept_buddy` | — | C2S | FromUserID |
| `messenger.decline_buddy` | — | C2S | FromUserID |
| `messenger.decline_all_buddies` | — | C2S | (empty) |
| `messenger.buddy_update` | — | S2C | Entries[] |

When two users each send a request to the other, the second request
auto-accepts and immediately creates the friendship without a pending row.

## Friend Limits

Friend limits are enforced at request-accept time on both sides. The effective
limit is resolved from the user's permissions in priority order:

| Permission | Effective limit |
|------------|----------------|
| `messenger.friends.unlimited` | No limit (0) |
| `messenger.friends.extended` | `MESSENGER_MAX_FRIENDS_VIP` |
| (default) | `MESSENGER_MAX_FRIENDS` |

The resolved limit is also sent as `UserFriendLimit` in the init packet so the
client can enforce it locally.

## Permissions

| Permission | Effect |
|------------|--------|
| `messenger.flood.bypass` | Skips flood rate limiting entirely |
| `messenger.friends.extended` | Raises friend cap to VIP limit |
| `messenger.friends.unlimited` | Removes friend cap |

## Message Log

Every successfully sent private message is written to `messenger_message_log`
for security and audit purposes. Log rows are independent of offline messages —
they are never deleted on delivery. The retention period is configured via
`MESSENGER_MESSAGE_LOG_TTL_DAYS` (default 30 days).

### Database

Table `messenger_message_log`:

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `from_user_id` | int | Sender |
| `to_user_id` | int | Recipient |
| `message` | varchar(255) | Content |
| `sent_at` | timestamptz | Send timestamp (indexed) |

## Offline Messages

Messages sent to offline users are stored in `offline_messages` and delivered
atomically on the next messenger init. The retention period is configured via
`MESSENGER_OFFLINE_MSG_TTL_DAYS` (default 30 days).

### Database

Table `offline_messages`:

| Column | Type | Description |
|--------|------|-------------|
| `id` | serial | Primary key |
| `from_user_id` | int | Sender |
| `to_user_id` | int | Recipient (indexed) |
| `message` | varchar(255) | Content |
| `sent_at` | timestamptz | Original send timestamp |

## Purge Job

A background ticker purges expired rows from both `offline_messages` and
`messenger_message_log` on a configurable interval. The ticker starts
automatically at server startup and stops when the server context is cancelled.

| Variable | Default | Description |
|----------|---------|-------------|
| `MESSENGER_OFFLINE_MSG_TTL_DAYS` | 30 | Offline message retention in days |
| `MESSENGER_MESSAGE_LOG_TTL_DAYS` | 30 | Message log retention in days |
| `MESSENGER_PURGE_INTERVAL_SECONDS` | 3600 | How often the purge job runs |

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `MESSENGER_MAX_FRIENDS` | 200 | Default friend list capacity |
| `MESSENGER_MAX_FRIENDS_VIP` | 500 | VIP friend list capacity |
| `MESSENGER_FLOOD_COOLDOWN_MS` | 750 | Minimum ms between messages |
| `MESSENGER_FLOOD_VIOLATIONS` | 4 | Violations before mute |
| `MESSENGER_FLOOD_MUTE_SECONDS` | 20 | Mute duration after flood |
| `MESSENGER_OFFLINE_MSG_TTL_DAYS` | 30 | Offline message retention |
| `MESSENGER_MESSAGE_LOG_TTL_DAYS` | 30 | Message log retention |
| `MESSENGER_PURGE_INTERVAL_SECONDS` | 3600 | Purge job interval |
| `MESSENGER_FRAGMENT_SIZE` | 750 | Friends per list fragment packet |
