# Navigator API

## REST Endpoints

All navigator REST endpoints require API key authentication via the
configured API key header.

### Categories

#### `GET /api/navigator/categories`

Returns all navigator categories ordered by `order_num`.

**Response** `200 OK`
```json
[
  {
    "id": 1,
    "caption": "Public Rooms",
    "visible": true,
    "order_num": 0,
    "icon_image": 0,
    "category_type": "public"
  }
]
```

#### `POST /api/navigator/categories`

Creates a new navigator category.

**Request Body**
```json
{
  "caption": "Events",
  "visible": true,
  "category_type": "public"
}
```

**Response** `201 Created`

#### `DELETE /api/navigator/categories/:id`

Deletes a navigator category by ID.

**Response** `204 No Content`

### Rooms

#### `GET /api/navigator/rooms`

Returns paginated rooms with optional filter parameters.

| Parameter | Type | Description |
|-----------|------|-------------|
| `category_id` | int | Filter by category |
| `search` | string | ILIKE search on name and tags |
| `owner_id` | int | Filter by owner |
| `offset` | int | Pagination offset |
| `limit` | int | Page size (default 50) |

**Response** `200 OK`
```json
{
  "rooms": [...],
  "total": 42
}
```

#### `GET /api/navigator/rooms/:id`

Returns one room by ID.

**Response** `200 OK`

#### `DELETE /api/navigator/rooms/:id`

Deletes a room by ID.

**Response** `204 No Content`

## CLI Commands

### `pixel-server navigator categories list`

Lists all navigator categories in JSON format.

### `pixel-server navigator rooms get <id>`

Retrieves one room by ID in JSON format.

## OpenAPI

All navigator endpoints are documented in OpenAPI format and exposed via
the Swagger UI route. The OpenAPI paths are merged during serve startup in
`registerServeHTTPRoutes`.
