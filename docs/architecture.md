# Architecture

## System Diagram

```
┌───────────────────────────────────────────────────────┐
│                    Docker Container                      │
│                                                         │
│  ┌──────────┐    ┌────────────────┐    ┌────────────┐  │
│  │ Next.js  │───▶│  Go HTTP       │───▶│    Redis    │  │
│  │  :3000   │    │  Server :8080  │    │  (atomic ops)│  │
│  │(static)  │    │                │    └────────────┘  │
│  └──────────┘    │  /reserve      │                       │
│       │          │  /stats        │    ┌────────────┐  │
│  nginx:7860      │  /admin/seed   │───▶│  PostgreSQL  │  │
│  proxies to      └────────────────┘    │  (orders,    │  │
│  Next.js static                      │   snapshots) │  │
│  + /api/* to Go                      └────────────┘  │
│                                                         │
│  k6 (load test — run manually outside container)        │
└───────────────────────────────────────────────────────┘
```

## Startup Flow

1. PostgreSQL container starts; `init.sql` creates tables
2. Redis starts with default config
3. Go server starts; connects to both, retries with backoff (max 10s)
4. Go server calls `seedDefaultSale()` — inserts one flash sale row, sets Redis stock key to 100, sets `sale:1:active` with TTL = `duration_seconds`
5. Next.js build runs (`next build && next export`) → static files in `/out`
6. Nginx serves `/out` on port 7860, proxies `/api/*` to Go on 8080
7. Frontend auto-loads `GET /api/stats?sale_id=1` every 500ms

## CAP Theorem Classification

**CP (Consistency + Partition Tolerance)** for reservation; eventual consistency for stats.

Redis operates as a CP system for inventory: a single Redis instance guarantees atomic DECR, so two concurrent requests cannot both see stock=1 and both succeed. Stats snapshots are eventually consistent (written async, polled by frontend) — stale by up to 500ms is acceptable. PostgreSQL uses optimistic locking for order inserts.

## Failure Modes

| Failure | Behavior | Recovery |
|---------|----------|----------|
| Redis unavailable | `/reserve` returns 503, requests logged | Auto-reconnect with exponential backoff |
| PostgreSQL unavailable | Reservation succeeds in Redis, order write queued in memory channel (best-effort) | Drain channel on reconnect; alert if channel full |
| Stock key missing (Redis eviction) | DECR on missing key returns -1; server re-checks PostgreSQL order count to set Redis key | Re-seed key from DB on negative result |
| Go server OOM | Container restarts; Redis retains stock count; sale resumes | Docker restart policy: `always` |
| Sale expired mid-request | `sale:active` TTL expires; server returns 410 | No action needed; TTL is authoritative |

## Performance Budget

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| p50 latency (reserve) | < 5ms | k6 `http_req_duration{p(50)}` |
| p99 latency (reserve) | < 50ms | k6 `http_req_duration{p(99)}` |
| Throughput | > 1,000 RPS | k6 `http_reqs` counter |
| Oversell rate | 0% | `orders.count` == `initial_stock - stock_remaining` |
| Stats poll latency | < 50ms | Browser DevTools |