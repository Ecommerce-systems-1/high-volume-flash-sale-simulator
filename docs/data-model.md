# Data Model

## PostgreSQL Tables

```sql
-- Flash sale configuration
CREATE TABLE flash_sales (
    id          SERIAL PRIMARY KEY,
    product_id  VARCHAR(50) NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    initial_stock INTEGER NOT NULL CHECK (initial_stock > 0),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Orders created by successful reservations
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id         INTEGER NOT NULL REFERENCES flash_sales(id),
    user_id         VARCHAR(50) NOT NULL,
    reserved_at     TIMESTAMPTZ DEFAULT NOW(),
    status          VARCHAR(20) DEFAULT 'RESERVED' CHECK (status IN ('RESERVED','CANCELLED'))
);

-- Aggregated stats snapshots (for dashboard polling)
CREATE TABLE stats_snapshots (
    id              SERIAL PRIMARY KEY,
    sale_id         INTEGER NOT NULL REFERENCES flash_sales(id),
    snapshot_time   TIMESTAMPTZ DEFAULT NOW(),
    total_requests  INTEGER NOT NULL,
    successful      INTEGER NOT NULL,
    rejected_sold_out INTEGER NOT NULL,
    rejected_expired  INTEGER NOT NULL,
    rps             NUMERIC(10,2)
);
```

## Redis Keys

| Key Pattern | Type | Purpose |
|-------------|------|---------|
| `sale:{sale_id}:stock` | String (integer) | Atomic inventory counter |
| `sale:{sale_id}:requests` | String (integer) | Total request counter |
| `sale:{sale_id}:success` | String (integer) | Successful reservation counter |
| `sale:{sale_id}:active` | String | "1" if sale is active, expires at end_time |

## API Schemas

```json
// POST /api/reserve
// Request:
{ "user_id": "user_4821", "sale_id": 1 }

// Response 200:
{ "order_id": "a3f1b2c4-...", "product_name": "Air Max Limited", "stock_remaining": 47 }

// Response 409 (sold out):
{ "error": "sold_out", "message": "All units have been reserved" }

// Response 410 (expired):
{ "error": "sale_expired", "message": "Flash sale has ended" }

// GET /api/stats?sale_id=1
// Response 200:
{
  "sale_id": 1,
  "product_name": "Air Max Limited",
  "initial_stock": 100,
  "stock_remaining": 23,
  "total_requests": 1842,
  "successful": 77,
  "rejected_sold_out": 1203,
  "rejected_expired": 0,
  "rps": 312.4,
  "sale_active": true,
  "seconds_remaining": 284
}

// POST /api/admin/seed  (seeds a new flash sale, used at startup)
// Request: { "product_name": "Air Max Limited", "stock": 100, "duration_seconds": 300 }
// Response 200: { "sale_id": 1, "start_time": "...", "end_time": "..." }
```

## Validation Rules
- `user_id` must be non-empty string, max 50 chars
- `sale_id` must be positive integer
- `stock` in seed request: 1–10000
- `duration_seconds`: 60–3600