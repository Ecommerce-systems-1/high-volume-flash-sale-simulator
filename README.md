# ⚡ High-Volume Flash Sale Simulator

A production-grade flash sale simulation system demonstrating atomic inventory reservation using Redis DECR, Go HTTP server for high-concurrency handling, and PostgreSQL for order persistence.

## Architecture

- **Go HTTP Server** — Handles `/api/reserve`, `/api/stats`, `/admin/seed` endpoints
- **Redis** — Atomic DECR for inventory reservation (single-threaded guarantee)
- **PostgreSQL** — Order persistence with UUID primary keys
- **Next.js Frontend** — Dark-themed dashboard with countdown, stock bar, live stats
- **Nginx** — Serves static frontend on port 7860, proxies `/api/*` to Go

## Quick Start

```bash
docker build -t flash-sale-simulator .
docker run -p 7860:7860 flash-sale-simulator
```

Open http://localhost:7860 to view the dashboard.

## Run Tests

```bash
cd backend && go test ./tests/ -v
```

## Load Testing

Requires [k6](https://k6.io/docs/getting-started/installation/):

```bash
k6 run --env BASE_URL=http://localhost:7860 load/flash_sale.js
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Go server port |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `POSTGRES_DSN` | `postgres://postgres:postgres@localhost:5432/flashsale?sslmode=disable` | PostgreSQL DSN |
| `DEFAULT_STOCK` | `100` | Initial stock for seeded sale |
| `SALE_DURATION` | `300` | Sale duration in seconds |