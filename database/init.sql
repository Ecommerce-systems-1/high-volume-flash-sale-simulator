CREATE TABLE IF NOT EXISTS flash_sales (
    id          SERIAL PRIMARY KEY,
    product_id  VARCHAR(50) NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    initial_stock INTEGER NOT NULL CHECK (initial_stock > 0),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id         INTEGER NOT NULL REFERENCES flash_sales(id),
    user_id         VARCHAR(50) NOT NULL,
    reserved_at     TIMESTAMPTZ DEFAULT NOW(),
    status          VARCHAR(20) DEFAULT 'RESERVED' CHECK (status IN ('RESERVED','CANCELLED'))
);

CREATE TABLE IF NOT EXISTS stats_snapshots (
    id              SERIAL PRIMARY KEY,
    sale_id         INTEGER NOT NULL REFERENCES flash_sales(id),
    snapshot_time   TIMESTAMPTZ DEFAULT NOW(),
    total_requests  INTEGER NOT NULL,
    successful      INTEGER NOT NULL,
    rejected_sold_out INTEGER NOT NULL,
    rejected_expired  INTEGER NOT NULL,
    rps             NUMERIC(10,2)
);