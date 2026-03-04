# Realm: Camera & Photos

> **Position:** 190 | **Phase:** 13 (Remaining) | **Packets:** 15 (8 c2s, 7 s2c)
> **Services:** game | **Status:** Not yet implemented

---

## Overview

Camera & Photos manages room snapshot capture, server-side photo rendering, photo publishing to profiles, purchasing printed photos, and thumbnail generation. This realm requires an external rendering pipeline (or a clever workaround) since the server must produce a room image from ECS state.

**Roadmap reference:** [009-packet-roadmap.md](../009-packet-roadmap.md) Phase 13

---

## Packet Inventory

### C2S -- 8 packets

| ID | Name | Fields | Summary |
|----|------|--------|---------|
| -- | `camera.capture` | `roomId:int32` | Request room snapshot |
| -- | `camera.render` | `renderData:string` | Submit client-rendered photo data |
| -- | `camera.publish` | `photoId:int32` | Publish photo to profile |
| -- | `camera.purchase` | `photoId:int32` | Purchase printed photo (furniture item) |
| -- | `camera.get_price` | _(none)_ | Get photo purchase price |
| -- | `camera.delete` | `photoId:int32` | Delete photo |
| -- | `camera.thumbnail` | `roomId:int32` | Request room thumbnail |
| -- | `camera.get_photos` | _(none)_ | Get user's photo gallery |

### S2C -- 7 packets

| ID | Name | Summary |
|----|------|---------|
| -- | `camera.capture_result` | Snapshot captured, render ID assigned |
| -- | `camera.render_result` | Rendered photo URL/data |
| -- | `camera.publish_result` | Photo published to profile |
| -- | `camera.purchase_result` | Photo item created in inventory |
| -- | `camera.price` | Photo purchase price |
| -- | `camera.photos` | User's photo gallery |
| -- | `camera.thumbnail` | Room thumbnail data |

---

## Implementation Analysis

### Photo Pipeline

```
1. Client captures viewport (local screenshot)
2. Client sends render data (JSON with room state, entity positions, furniture)
3. Server either:
   a. Stores client-rendered image (base64 or binary upload) -- SIMPLER
   b. Server-side renders from ECS state snapshot -- COMPLEX
4. Assign photo ID, store in photos table
5. User can publish to profile or purchase as furniture item
```

**Recommended approach for pixel-server:** Accept client-rendered images. Server-side rendering is a massive engineering effort (essentially reimplementing the client's renderer). Client-submitted images should be:
- Size-limited (max 500KB)
- Dimension-limited (max 800x600)
- Format-validated (PNG/JPEG only)
- Content-scanned for abuse (optional, out of scope for MVP)

### Storage

Photos stored as files on disk or object storage (S3-compatible):
- `photos/<userId>/<photoId>.png`
- Metadata in PostgreSQL: `photos` table (id, user_id, room_id, created_at, published)

### Room Thumbnails

Room thumbnails are auto-generated snapshots used in navigator cards:
- Generated on room creation and on-demand refresh
- Lower resolution (200x150)
- Cached with long TTL (1 hour)

---

## Caveats

### 1. Image Upload Abuse
Without content moderation, users can upload inappropriate images. Implement:
- File size limit
- Rate limit (1 photo per 30 seconds)
- Admin review queue for published photos

### 2. Storage Growth
Photos accumulate indefinitely. Implement retention: unpublished photos deleted after 30 days. Published photos retained indefinitely but count toward a per-user limit (50 photos).

---

## Dependencies

- **Phase 3 (Room)** -- room state for snapshot context
- **File storage** -- local disk or S3-compatible object storage

---

## Testing Strategy

### Unit Tests
- Image validation (size, format, dimensions)
- Photo CRUD operations

### Integration Tests
- Photo upload, storage, and retrieval
- Photo purchase creates furniture item
