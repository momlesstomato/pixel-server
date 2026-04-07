# Wire Protocol

All catalog interactions use the Pixel binary frame protocol. Each packet
carries a two-byte header (the packet ID) followed by typed fields encoded
with `codec.NewWriter`. Lengths are big-endian integers.

## Client-to-Server Packets

### catalog.get_index — c2s 1195

Sent when the user opens the catalog. Requests the full page tree.

| Field | Type | Description |
|-------|------|-------------|
| `catalogType` | string | Catalog variant: `"NORMAL"` or `"BUILDERS_CLUB"` |

The server responds with `catalog.index` (s2c 1032).

---

### catalog.get_page — c2s 412

Sent when the user navigates to a catalog page.

| Field | Type | Description |
|-------|------|-------------|
| `pageId` | int32 | Catalog page identifier |
| `offerId` | int32 | Specific offer to focus (`-1` for all offers) |
| `catalogType` | string | Catalog variant string |

The server responds with `catalog.page` (s2c 804).

---

### catalog.purchase — c2s 3492

Sent when the user confirms a purchase.

| Field | Type | Description |
|-------|------|-------------|
| `pageId` | int32 | Catalog page of the offer |
| `offerId` | int32 | Offer identifier |
| `extraData` | string | Item-specific personalisation data (pet name, text, etc.) |
| `amount` | int32 | Quantity to purchase |

On success: `catalog.purchase_ok` (s2c 869).
On failure: `catalog.purchase_error` (s2c 1404) or `catalog.purchase_not_allowed` (s2c 3770).

---

### catalog.purchase_gift — c2s 1411

Sent when the user sends a gift to another user.

| Field | Type | Description |
|-------|------|-------------|
| `pageId` | int32 | Catalog page of the offer |
| `itemId` | int32 | Offer identifier |
| `extraData` | string | Item-specific personalisation data |
| `receiverName` | string | Username of the gift recipient |
| `giftMessage` | string | Message to include with the gift |
| `spriteId` | int32 | Wrapping paper sprite identifier |
| `boxId` | int32 | Gift box style identifier |
| `ribbonId` | int32 | Ribbon style identifier |
| `showMyFace` | bool | Include sender's avatar face on the gift |

On success: `catalog.purchase_ok` (s2c 869). On failure: error packets above.

---

### catalog.redeem_voucher — c2s 339

Sent when the user submits a voucher code.

| Field | Type | Description |
|-------|------|-------------|
| `voucherCode` | string | Voucher code to redeem |

On success: `catalog.voucher_ok` (s2c 3336).
On failure: `catalog.voucher_error` (s2c 714).

---

### catalog.get_gift_wrapping_config — c2s 418

Sent before opening the gift dialog. No fields. Server responds with
`catalog.gift_wrapping_config` (s2c 2234).

---

## Server-to-Client Packets

### catalog.index — s2c 1032

Delivers the full page navigation tree.

The packet begins with the root tree node, then:

```
bool   newAdditionsAvailable
string catalogType              (echoed from request)
```

Each tree node (root and all descendants) is encoded recursively:

```
bool    visible
int32   icon
int32   pageId          (-1 when the page is disabled / non-navigable)
string  pageName        (link key, used internally for navigation events)
string  localization    (caption displayed in the UI tree)
int32   offerIdCount
int32[] offerIds        (offer IDs directly reachable from this node)
int32   childCount
node[]  children        (recursive, same structure)
```

The root node uses `pageId = -1` and `pageName = ""`.

---

### catalog.page — s2c 804

Delivers the content of one catalog page.

```
int32  pageId
string catalogType
string layoutCode
int32  imageCount
string[] images
int32  textCount
string[] texts
int32  offerCount
offer[] offers
int32  -1               (sentinel, always -1)
bool   false            (acceptSeasonCurrencyAsCredits, always false unless configured)
```

For `frontpage4` pages, after the sentinel and boolean, a front-page
promotions list is appended:

```
int32  promotionCount
entry[] promotions as:
  int32  id
  string caption
  string image
  int32  unknown
  string pageLink
  string pageId
```

Each **offer** inside the offers array is encoded as:

```
int32  offerId
string localizationId   (CatalogName)
bool   isRentable       (always false in current implementation)
int32  priceCredits
int32  priceActivityPoints
int32  activityPointType
bool   giftable
int32  productCount
product[] products      (see Product Encoding below)
int32  clubLevel        (0 = no club required, 1 = basic, 2 = VIP)
bool   bundlePurchaseAllowed
bool   isPet
string previewImage     (e.g. "catalogue/pet_lion.png", empty if none)
```

**Product encoding** by type:

*Floor / wall item (`"s"` / `"i"`) and avatar effect (`"e"`):*
```
string "s" | "i" | "e"
int32  spriteId
string extraParam       (colour preset, variant string, or empty)
int32  amount
bool   isLimited
if isLimited:
  int32 seriesSize
  int32 remaining
```

*Badge (`"b"`):*
```
string "b"
string badgeCode
```

---

### catalog.purchase_ok — s2c 869

Confirms a successful purchase. Contains the purchased offer record using the
same encoding as an offer inside `catalog.page`.

---

### catalog.purchase_error — s2c 1404

Reports a failed purchase.

| Field | Type | Description |
|-------|------|-------------|
| `errorCode` | int32 | `0` generic, `1` insufficient credits, `2` insufficient activity points, `3` not available |

---

### catalog.purchase_not_allowed — s2c 3770

Reports that this user may not purchase at all (parental controls, guest
account restrictions, etc.).

| Field | Type | Description |
|-------|------|-------------|
| `errorCode` | int32 | Specific permission denial reason code |

---

### catalog.published — s2c 1866

Broadcast to all connected clients when the catalog is updated. No fields.
The client invalidates its cached catalog data and may prompt the user
to refresh.

---

### catalog.voucher_ok — s2c 3336

Sent when a voucher is redeemed successfully. No additional fields beyond
the redemption confirmation.

---

### catalog.voucher_error — s2c 714

Sent when voucher redemption fails. The error code indicates the specific
failure reason.

---

### catalog.gift_wrapping_config — s2c 2234

Lists available wrapping paper, box, and ribbon options for gift purchases.

---

## Complete Packet Reference

| Packet name | ID | Direction | Summary |
|-------------|-----|-----------|---------|
| `catalog.get_index` | 1195 | C2S | Request the page tree |
| `catalog.get_page` | 412 | C2S | Request page content |
| `catalog.purchase` | 3492 | C2S | Purchase an offer |
| `catalog.purchase_gift` | 1411 | C2S | Purchase as a gift |
| `catalog.redeem_voucher` | 339 | C2S | Redeem a voucher code |
| `catalog.get_gift_wrapping_config` | 418 | C2S | Request wrapping options |
| `catalog.get_gift` | 2436 | C2S | Request gift delivery details |
| `catalog.check_giftable` | 1347 | C2S | Check if offer is giftable |
| `catalog.get_club_offers` | 3285 | C2S | Request club subscription offers |
| `catalog.get_club_gift_info` | 487 | C2S | Request HC monthly gift selector data |
| `catalog.get_direct_club_buy` | 801 | C2S | Request direct/SMS club-buy availability |
| `catalog.select_club_gift` | 2276 | C2S | Claim one HC monthly gift by name |
| `catalog.get_product_offer` | 2594 | C2S | Request one offer by ID |
| `catalog.get_pet_breeds` | 1756 | C2S | Request pet breed palettes |
| `catalog.index` | 1032 | S2C | Page tree response |
| `catalog.page` | 804 | S2C | Page content response |
| `catalog.club_gift_info` | 619 | S2C | HC monthly gift selector payload |
| `catalog.club_offers` | 2405 | S2C | HC subscription offer list |
| `catalog.direct_sms_club_buy` | 195 | S2C | Direct/SMS club-buy availability payload |
| `catalog.club_gift_selected` | 659 | S2C | Confirm one claimed HC gift |
| `catalog.purchase_ok` | 869 | S2C | Purchase success |
| `catalog.purchase_error` | 1404 | S2C | Purchase failure |
| `catalog.purchase_not_allowed` | 3770 | S2C | Purchase permission denied |
| `catalog.published` | 1866 | S2C | Catalog updated broadcast |
| `catalog.gift_wrapping_config` | 2234 | S2C | Wrapping options |
| `catalog.pet_breeds` | 3331 | S2C | Pet breed palette |

## Subscription Packets

Subscription packets belong to the `subscription-offers` realm but are closely
related to the catalog flow.

| Packet name | ID | Direction | Summary |
|-------------|-----|-----------|---------|
| `user.get_subscription` | 3166 | C2S | Request subscription status |
| `user.subscription` | 954 | S2C | Deliver subscription state |
| `user.get_kickback_info` | 869 | C2S | Request HC payday snapshot |
| `user.kickback_info` | 3277 | S2C | Deliver HC payday and streak state |

### user.subscription — s2c 954

Sent after login and after any subscription change.

| Field | Type | Description |
|-------|------|-------------|
| `productName` | string | Club product identifier (e.g. `"club_habbo"`) |
| `daysToPeriodEnd` | int32 | Days remaining in current period |
| `memberPeriods` | int32 | Completed subscription periods |
| `periodsSubscribedAhead` | int32 | Pre-paid future periods |
| `responseType` | int32 | `1` = login refresh, `2` = new purchase, `3` = discount available |
| `hasEverBeenMember` | bool | Has user ever had a club subscription |
| `isVip` | bool | Current VIP status |
| `pastClubDays` | int32 | Total historical club days |
| `pastVipDays` | int32 | Total historical VIP days |
| `minutesUntilExpiration` | int32 | Minutes until current period expires |

When a player has no active subscription, the server sends a zero-state
packet (`daysToPeriodEnd=0`, `memberPeriods=0`, etc.) with `responseType=1`
rather than an error. This matches the client's expectation that a subscription
response is always delivered on login.

### user.kickback_info — s2c 3277

Sent when the HC Center asks for payday status.

| Field | Type | Description |
|-------|------|-------------|
| `currentHCStreak` | int32 | Current continuous HC age in days |
| `firstSubscriptionDate` | string | First recorded HC join date (`YYYY-MM-DD`) |
| `kickbackPercentage` | float64 | Percentage configured for spend kickback |
| `totalCreditsMissed` | int32 | Lifetime cycle spend that yielded no reward |
| `totalCreditsRewarded` | int32 | Lifetime payday credits granted |
| `totalCreditsSpent` | int32 | Credits spent in the current payday cycle |
| `creditRewardForStreakBonus` | int32 | Reward currently accrued from streak bonus |
| `creditRewardForMonthlySpent` | int32 | Reward currently accrued from spend kickback |
| `timeUntilPayday` | int32 | Seconds remaining until next payday |

### catalog.club_gift_info — s2c 619

Nitro expects the club-gift selector payload as:

1. `daysUntilNextGift`
2. `giftsAvailable`
3. `CatalogPageMessageOfferData[]`
4. trailing gift metadata map entries: `offerId`, `vipOnly`, `daysRequired`, `isSelectable`

Sending only zero counters without the offer and metadata sections is not
enough for the HC gift selector to render.
