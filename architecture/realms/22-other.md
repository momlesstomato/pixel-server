# Realm: Other

Terminology note: references to services and NATS subjects in this file map to internal modules and internal contract topics in the single `pixelsv` binary unless explicitly marked as external adapter behavior.


> **Position:** 220 | **Phase:** 13 (Remaining) | **Packets:** 10 (7 c2s, 3 s2c)
> **Services:** game | **Status:** Not yet implemented

---

## Overview

The "Other" realm contains miscellaneous packets that don't fit neatly into other realms: mystery box opening, fireworks charging, rentable space extensions, phone verification, and utility operations. This is the smallest realm at 10 packets and the lowest priority for implementation.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 7 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `misc.mystery_box_open` | `itemId:int32` | Open mystery box (random reward) |
| -- | `misc.firework_charge` | `itemId:int32` | Charge firework (add fuel) |
| -- | `misc.rentable_extend` | `itemId`, `offerId` | Extend rental period |
| -- | `misc.phone_verify` | `phoneNumber:string` | Start phone verification |
| -- | `misc.phone_code` | `code:string` | Submit verification code |
| -- | `misc.report_room` | `roomId`, `reason` | Quick room report |
| -- | `misc.tutorial_skip` | _(none)_ | Skip tutorial |

### S2C -- 3 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `misc.mystery_box_result` | Mystery box opened, show reward |
| -- | `misc.phone_verify_result` | Phone verification status |
| -- | `misc.rentable_info` | Rental extension info |

---

## Implementation Analysis

### Mystery Box

```
1. User opens mystery box item
2. Validate item is a mystery box type
3. Consume the mystery box item
4. Random reward from mystery box pool (similar to recycler)
5. Create reward item in inventory
6. Send mystery_box_result with reward visualization
```

### Fireworks

Fireworks are room items that can be "charged" with fuel:
- Each charge costs credits (configurable)
- Each charge adds X uses
- When activated, firework plays animation and decrements use count
- At 0 uses, firework is inactive until recharged

### Rentable Spaces

Some room furniture can be rented for a period:
- Rental items have an expiry timestamp
- Users can extend before expiry
- Extension costs credits/points
- On expiry: item is returned to store/removed

### Phone Verification

Phone verification is an optional security feature:
- User submits phone number
- Server sends SMS code (via external service like Twilio)
- User submits code for verification
- Verified status unlocks features (e.g., marketplace access)

**pixel-server approach:** Phone verification is optional infrastructure. For MVP, stub the packets and return "verification not required". Implement when SMS provider is configured.

---

## Caveats

### 1. Mystery Box Fairness
Same requirements as dice and recycler: server-side randomness with cryptographic quality.

### 2. Firework Credit Drain
Without rate limiting, users could burn unlimited credits on firework charges. Implement a maximum charge level per firework.

### 3. Phone Number Privacy
Phone numbers are PII. If collected, they must be hashed/encrypted at rest and subject to GDPR data deletion requests.

---

## Dependencies

- **Phase 6 (Furniture)** -- mystery box and firework as item interactions
- **Phase 7 (Inventory)** -- reward item creation
- **External SMS service** -- for phone verification (optional)

---

## Testing Strategy

### Unit Tests
- Mystery box reward selection
- Firework charge/use counter
- Rental expiry calculation

### Integration Tests
- Mystery box full flow against real DB
- Rental extension and expiry
