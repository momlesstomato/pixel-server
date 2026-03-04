# User Profile — Overview

The **user-profile** realm (Phase 2) handles everything related to a player's
personal identity — who they are, what they look like, their settings and
preferences, and their social list management.

All logic lives in `services/game/internal/identity/` and is wired into the
game service's NATS `Listener`.

---

## Documents in this realm

| File | What it covers |
|---|---|
| [LOGIN-BUNDLE.MD](LOGIN-BUNDLE.MD) | The 9-packet login burst sent right after authentication |
| [PACKETS.MD](PACKETS.MD) | All 17 C2S handler packets and every S2C response packet |
| [DATA-MODELS.MD](DATA-MODELS.MD) | User, Settings, Badge, Wardrobe, Ignore models + repository interfaces |
| [PLUGIN-HOOKS.MD](PLUGIN-HOOKS.MD) | Plugin events, interceptors, realm relations, permissions, config |

---

## Timeline

```
game service receives session.authenticated
        │
        ▼
Listener.handleAuthenticated
  ├── create natsSession
  ├── add to sessions map
  ├── emit event.PlayerJoined
  └── identity.SendLoginBundle (9 packets → client)

later, client sends any user-profile C2S packet (e.g. update_figure)
        │
        ▼
gateway routes to room.input.<sessionID>
        │
        ▼
Listener.handlePacket
  ├── interceptor.RunBefore
  ├── emit event.PacketIn
  ├── identityRouter.Dispatch (matches header ID → handler func)
  ├── handler encodes + sends S2C packet via natsSession.Send
  └── interceptor.RunAfter
```

---

## Module layout

```
services/game/internal/identity/
├── service.go      — Service struct: business logic (build login bundle, updates)
├── handler.go      — Handler struct: decode C2S → call service → encode S2C
├── router.go       — Router struct: dispatches by header ID to handler func
├── login_bundle.go — SendLoginBundle + BuildLoginBundle helpers
├── role.go         — BuildRoleProfile helper (club level, security, ambassador)
└── identity_test.go— 505-line comprehensive test suite
```

---

## Package structure

`identityRouter.Dispatch(ctx, sess, headerID, payload)` tries each registered
handler in header-ID order. Returns `(handled bool, err error)`.

Unhandled packets fall through to the game listener's "unhandled" debug log path
and are then passed to `interceptor.RunAfter`.
