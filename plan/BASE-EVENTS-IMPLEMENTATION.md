# Base Events Implementation

This document summarizes the homogeneous cancellable event system delivered across all mutation endpoints.

## Pattern

Every state mutation follows the same flow:

1. Fire a cancellable **before** event (embeds `sdk.BaseCancellable`).
2. Check `Cancelled()` — abort with error if true.
3. Perform the mutation.
4. Fire a non-cancellable **after** event (embeds `sdk.BaseEvent`).

Services expose `SetEventFirer(func(sdk.Event))` wired at startup in `core/cli/serve.go`.

## SDK Events

| Domain | File | Type | Cancellable |
|---|---|---|---|
| catalog | `page_creating.go` | `PageCreating` | yes |
| catalog | `page_created.go` | `PageCreated` | no |
| catalog | `offer_creating.go` | `OfferCreating` | yes |
| catalog | `offer_created.go` | `OfferCreated` | no |
| catalog | `voucher_redeeming.go` | `VoucherRedeeming` | yes |
| catalog | `voucher_redeemed.go` | `VoucherRedeemed` | no |
| furniture | `definition_creating.go` | `DefinitionCreating` | yes |
| furniture | `definition_created.go` | `DefinitionCreated` | no |
| furniture | `definition_deleting.go` | `DefinitionDeleting` | yes |
| furniture | `definition_deleted.go` | `DefinitionDeleted` | no |
| inventory | `badge_awarding.go` | `BadgeAwarding` | yes |
| inventory | `badge_awarded.go` | `BadgeAwarded` | no |
| inventory | `badge_revoking.go` | `BadgeRevoking` | yes |
| inventory | `badge_revoked.go` | `BadgeRevoked` | no |
| inventory | `credits_updating.go` | `CreditsUpdating` | yes |
| inventory | `credits_updated.go` | `CreditsUpdated` | no |
| inventory | `currency_updating.go` | `CurrencyUpdating` | yes |
| inventory | `currency_updated.go` | `CurrencyUpdated` | no |
| authentication | `sso_generated.go` | `SSOGenerated` | no |
| user | `profile_updating.go` | `ProfileUpdating` | yes |
| user | `settings_updating.go` | `SettingsUpdating` | yes |
| messenger | `friend_added.go` | `FriendAdded` | no |
| messenger | `friend_removed.go` | `FriendRemoved` | yes |
| management | `session_kicked.go` | `SessionKicked` | no |
| management | `hotel_closing.go` | `HotelClosing` | yes |
| management | `hotel_reopening.go` | `HotelReopening` | yes |

All event files live under `sdk/events/<domain>/`, one type per file.

## Service Modifications

| Service | File | Events Fired |
|---|---|---|
| Catalog | `pkg/catalog/application/service.go` | PageCreating, PageCreated |
| Catalog | `pkg/catalog/application/offer_service.go` | OfferCreating, OfferCreated |
| Catalog | `pkg/catalog/application/voucher_service.go` | VoucherRedeeming, VoucherRedeemed |
| Furniture | `pkg/furniture/application/service.go` | DefinitionCreating, DefinitionCreated, DefinitionDeleting, DefinitionDeleted |
| Inventory | `pkg/inventory/application/service.go` | CreditsUpdating, CreditsUpdated, CurrencyUpdating, CurrencyUpdated |
| Inventory | `pkg/inventory/application/badge_service.go` | BadgeAwarding, BadgeAwarded, BadgeRevoking, BadgeRevoked |
| Authentication | `pkg/authentication/application/service.go` | SSOGenerated |
| User | `pkg/user/application/profile.go` | ProfileUpdating, SettingsUpdating |
| Messenger | `pkg/messenger/application/friends.go` | FriendAdded, FriendRemoved |
| Management | `pkg/management/adapter/httpapi/session_routes.go` | SessionKicked |
| Management | `pkg/management/adapter/httpapi/hotel_routes.go` | HotelClosing, HotelReopening |

## New Endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/inventory/users/:user_id/credits` | Add/subtract credits |
| POST | `/inventory/users/:user_id/currencies/:type` | Add/subtract activity currency |
| DELETE | `/inventory/users/:user_id/badges/:code` | Revoke badge |

Routes split: `pkg/inventory/adapter/httpapi/routes.go` and `currency_routes.go`.
OpenAPI split: `openapi.go` (paths + schemas) and `openapi_ops.go` (operations).

## Wiring

`core/cli/serve.go` calls `SetEventFirer(services.fireSafe)` on every domain service at startup.
`core/cli/serve_services.go` exposes `fireSafe` nil-check wrapper and the `fire func(sdk.Event)` field.
`core/cli/serve_routes.go` passes `services.fireSafe` to management route registrars.

## AGENTS.md

Section 11 "Event System Rules" added with eight rules governing the mutation event pattern, file placement, priority registration, and firer wiring.

## Tests

### Unit Tests (24 tests)

| Package | Tests |
|---|---|
| `pkg/catalog/application/tests/event_test.go` | 6 (PageCreating, OfferCreating, VoucherRedeeming — cancel + allow) |
| `pkg/inventory/application/tests/event_test.go` | 8 (BadgeAwarding, BadgeRevoking, CreditsUpdating, CurrencyUpdating — cancel + allow) |
| `pkg/furniture/application/tests/event_test.go` | 4 (DefinitionCreating, DefinitionDeleting — cancel + allow) |
| `pkg/user/application/tests/event_test.go` | 4 (ProfileUpdating, SettingsUpdating — cancel + allow) |
| `pkg/messenger/application/tests/event_test.go` | 3 (FriendAdded fires, FriendRemoved cancel + allow) |
| `pkg/authentication/application/service_test.go` | 1 (SSOGenerated fires) |

### End-to-End Tests (4 tests)

| File | Tests |
|---|---|
| `e2e/11_events/11_events_test.go` | PluginCancelsPageCreation, PluginAllowsPageCreation, PluginCancelsBadgeAward, PluginReceivesFriendAddedEvent |

All tests use the real `coreplugin.Dispatcher` to verify plugin integration end-to-end.
