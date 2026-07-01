# BRD (Business Requirements Document)

## Problem Statement
Flash sale events create sudden demand spikes where thousands of users simultaneously attempt to purchase limited inventory. Without atomic reservation semantics, systems oversell stock, leading to order cancellations, refund costs, and customer trust damage.

## High-Level Requirements

| ID | Requirement | Measure |
|----|-------------|---------|
| R01 | Inventory reservation must be atomic | Zero oversells across 10,000 concurrent requests |
| R02 | Reservation endpoint p99 latency | < 50ms at 1,000 RPS |
| R03 | System rejects reservation when stock = 0 | HTTP 409 with `{"error":"sold_out"}` |
| R04 | Every successful reservation creates a PostgreSQL order | Order count = reservations granted |
| R05 | Flash sale has configurable duration | Countdown visible on frontend; requests rejected after expiry |
| R06 | k6 load test script ships with the repo | `k6 run load/flash_sale.js` runs without modification |
| R07 | Frontend shows real-time throughput | Dashboard polls `/api/stats` every 500ms |

## User Personas

1. **Platform Engineer (primary)** — Runs k6 load test, inspects latency histogram and oversell count. Wants concrete numbers to share with stakeholders.
2. **Engineering Manager** — Views HuggingFace demo without running code. Clicks the simulated purchase button and watches stock decrement live.
3. **Tech Recruiter / Interviewer** — Reviews codebase to assess concurrency patterns, Redis usage, and test coverage. Evaluates Go idioms.

## Out of Scope
- Real payment processing
- Multi-region Redis replication
- Actual email/SMS notifications
- User authentication
- Multiple concurrent flash sale events