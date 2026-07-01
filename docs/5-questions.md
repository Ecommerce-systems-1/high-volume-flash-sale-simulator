# 5-Questions

## What problem does this solve?
Engineers need to validate that their inventory reservation system can handle thousands of concurrent purchase attempts without overselling. Production flash sales regularly expose race conditions that development environments never surface.

## Who has this problem?
Platform engineers at mid-to-large ecommerce companies preparing for high-traffic events (Black Friday, product drops, limited-edition launches) who need empirical latency data and overselling guarantees before going live.

## What are we building?
A self-contained flash sale simulator: a Go HTTP server backed by Redis atomic DECR for inventory reservation and PostgreSQL for order persistence, plus a k6 load test that fires 10,000 virtual user requests, and a Next.js dashboard showing a countdown timer, simulated purchase button, real-time throughput/p99 chart, and stock-depleted indicator.

## Why this approach?
Redis SETNX/DECR atomics eliminate the need for distributed locks by leveraging single-threaded command execution. Go's goroutine model handles high concurrency cheaply. k6 is the industry standard for load testing HTTP endpoints with scriptable scenarios.

## How do we know it works?
1. k6 script shows zero oversells (orders created <= initial stock)
2. p99 latency < 50ms at 1,000 RPS under load test
3. Unit tests assert atomic reservation rejects when stock = 0
4. Integration tests confirm PostgreSQL order count matches Redis reservation count post-load