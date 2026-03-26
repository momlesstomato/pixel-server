# Administration: HTTP API and CLI

## HTTP API

The catalog HTTP API is mounted on the `/api/v1/catalog` prefix. All endpoints
return JSON and use HTTP-standard status codes. See the OpenAPI specification
at `/swagger` for interactive documentation.

### Page endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/catalog/pages` | List all catalog pages |
| `GET` | `/api/v1/catalog/pages/:id` | Get one page by ID |
| `POST` | `/api/v1/catalog/pages` | Create a new catalog page |

#### GET /api/v1/catalog/pages

Returns all pages as a JSON array. Pages are returned in database order;
the client is responsible for its own tree assembly if needed.

**Response 200**
```json
[
  {
    "ID": 1,
    "ParentID": null,
    "Caption": "Front Page",
    "IconImage": 213,
    "PageLayout": "frontpage4",
    "Visible": true,
    "Enabled": true,
    "MinPermission": "",
    "ClubOnly": false,
    "OrderNum": 1,
    "Images": ["catalog_frontpage_headline_shop_GENERAL", ""],
    "Texts": ["Welcome to the shop!", "Redeem a voucher here:"],
    "CreatedAt": "2024-01-01T00:00:00Z",
    "UpdatedAt": "2024-01-01T00:00:00Z"
  }
]
```

#### GET /api/v1/catalog/pages/:id

Returns a single page record. Returns `404` when the ID does not exist.

#### POST /api/v1/catalog/pages

Creates a new catalog page. The request body must be a JSON object matching
the `CatalogPage` domain entity.

**Request body** (required fields):
- `Caption` — display title
- `PageLayout` — layout template name
- `Images` — image asset key array
- `Texts` — text content array

**Response 201** — created page with assigned ID.

### Offer endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/catalog/pages/:id/offers` | List all offers for a page |

#### GET /api/v1/catalog/pages/:id/offers

Returns all offers belonging to the specified page as a JSON array. Returns
`404` when the page ID does not exist.

**Response 200**
```json
[
  {
    "ID": 1,
    "PageID": 3,
    "ItemDefinitionID": 55,
    "CatalogName": "Iced Table",
    "CostPrimary": 4,
    "CostPrimaryType": 0,
    "CostSecondary": 0,
    "CostSecondaryType": 0,
    "Amount": 1,
    "LimitedTotal": 0,
    "LimitedSells": 0,
    "OfferActive": true,
    "ExtraData": "",
    "BadgeID": "",
    "ClubOnly": false,
    "OrderNum": 10
  }
]
```

### Voucher endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/catalog/vouchers` | List all vouchers |
| `POST` | `/api/v1/catalog/vouchers/redeem` | Redeem a voucher for a user |

#### GET /api/v1/catalog/vouchers

Returns all voucher records. Includes both enabled and disabled vouchers.

#### POST /api/v1/catalog/vouchers/redeem

Redeems a voucher on behalf of a user. Intended for administrative use
(server-side reward dispatch) as well as internal bridging from the realtime
redeem flow.

**Request body:**
```json
{
  "code": "WELCOME2024",
  "user_id": 42
}
```

**Response 200** — the redeemed voucher record.

### Error responses

| HTTP status | Domain error |
|------------|--------------|
| `400 Bad Request` | Malformed request body or invalid ID parameter |
| `403 Forbidden` | Voucher disabled, offer inactive |
| `404 Not Found` | Page, offer, or voucher not found |
| `409 Conflict` | Voucher exhausted or already redeemed by this user |
| `500 Internal Server Error` | Unexpected service failure |

All error responses carry a JSON body:
```json
{ "error": "catalog page not found" }
```

Every response includes a `X-Ray-ID` header with a unique trace identifier.
Errors are logged with the associated `ray_id` via the structured logger for
cross-service correlation.

---

## CLI

The catalog CLI is registered under the `catalog` subcommand of the root
`pixel-server` binary. All subcommands accept the following persistent flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--env-file` | `.env` | Path to the environment configuration file |
| `--env-prefix` | (empty) | Optional prefix for environment variable keys |

### pixel-server catalog pages-list

Writes all catalog pages as a JSON array to stdout.

```
pixel-server catalog pages-list [--env-file .env]
```

**Output** — same schema as `GET /api/v1/catalog/pages`.

### pixel-server catalog pages-get [id]

Writes one catalog page by ID as a JSON object to stdout.

```
pixel-server catalog pages-get 3 [--env-file .env]
```

Returns a non-zero exit code and error message when the page is not found.

### pixel-server catalog offers-list [id]

Writes all offers for the specified page ID as a JSON array to stdout.

```
pixel-server catalog offers-list 3 [--env-file .env]
```

Returns a non-zero exit code when the page is not found or the ID is invalid.

### Example workflow

Create a new category page from the command line and seed its offers via
the HTTP API:

```bash
# 1. Inspect existing pages to determine tree structure
pixel-server catalog pages-list | jq '.[] | {ID, Caption, ParentID}'

# 2. Create a new page (using the HTTP API for full create support)
curl -s -X POST http://localhost:8080/api/v1/catalog/pages \
  -H "Content-Type: application/json" \
  -d '{
    "Caption": "Winter Furniture",
    "PageLayout": "default_3x3",
    "IconImage": 2,
    "Visible": true,
    "Enabled": true,
    "Images": ["winter_header", "winter_teaser"],
    "Texts": ["Furnish your winter room!", ""],
    "OrderNum": 50
  }'

# 3. Verify the page was created
pixel-server catalog pages-get 5

# 4. List offers for the new page (expect empty array)
pixel-server catalog offers-list 5
```

---

## Configuration

The catalog service itself has no dedicated configuration section. It depends
on the global PostgreSQL and Redis configuration sections.

Redis caching is activated at startup in `core/cli/serve_economy.go` with:

```go
catalog.SetCache(runtime.Redis, catalogapplication.CacheConfig{
    Prefix: "catalog",
    TTL:    5 * time.Minute,
})
```

When `runtime.Redis` is nil (Redis not configured), `SetCache` is a no-op and
all reads go directly to PostgreSQL.

---

## Persistence

### Schema

The catalog persists three entity types in PostgreSQL:

| Table | Entity |
|-------|--------|
| `catalog_pages` | `CatalogPage` |
| `catalog_offers` | `CatalogOffer` |
| `catalog_vouchers` | `Voucher` |
| `catalog_voucher_redemptions` | `VoucherRedemption` |

Migrations and seeds live under:
```
pkg/catalog/infrastructure/migration/
pkg/catalog/infrastructure/seed/
```

Each schema change is delivered as a new migration step with a unique ID.
Seed data provides bootstrap pages and offers for development and testing.

### Seeding

The seed step creates a minimal `frontpage4` root page plus one default
category with an example offer. This satisfies the client's expectation of
at least one root catalog node present on first launch.

To apply migrations and seeds:
```bash
pixel-server db migrate up
pixel-server db seed up
```

To roll back:
```bash
pixel-server db seed down
pixel-server db migrate down
```
