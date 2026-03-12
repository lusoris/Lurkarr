# ADR-006: Logging Architecture — DB Ring Buffer + WebSocket Broadcast

**Status:** Accepted (under evaluation for simplification)  
**Date:** 2026-02-01

## Context

Lurkarr generates structured logs from multiple background services (lurking engine, queue cleaner, auto-importer, scheduler). Users need to view logs in real-time via the web UI and query historical logs. The system also supports a full Loki-based observability stack for production deployments.

## Decision

Implement a hybrid logging architecture:

1. **slog JSON handler** — All logs go through Go's `slog` with a JSON handler. Log level is configurable via `LOG_LEVEL` env var.
2. **Ring buffer** — A 10,000-entry buffered channel captures log entries asynchronously from slog.
3. **Async DB flush** — A background flusher writes batches to PostgreSQL every 500ms or every 100 entries.
4. **WebSocket broadcast** — Each log entry is simultaneously broadcast to connected WebSocket clients with per-client app/level filtering.
5. **Frontend /logs page** — Real-time log viewer with text search, level filtering, and auto-scroll.

## Consequences

**Positive:**
- Real-time log streaming to the browser without polling.
- Historical log retention in the database for API queries.
- Per-client level/app filtering reduces WebSocket traffic.
- Non-blocking: ring buffer drops entries silently if writes are too slow, preventing backpressure on logging callers.

**Negative:**
- Dual storage (DB + Loki) is redundant for production deployments with a full observability stack.
- DB log storage grows unbounded without pruning (currently relies on periodic cleanup in maintenance goroutine).
- WebSocket hub adds complexity (mutex-protected client set, ping/pong heartbeat).

## Open Evaluation

For deployments with Loki, the DB log storage and WebSocket broadcast may be unnecessary overhead. A future simplification could:
- Remove DB log storage (keep stdout slog + Loki scraping only)
- Replace the /logs page with a Grafana/Loki embed or link
- Keep WebSocket broadcast as optional for users without Loki

This is tracked in the project todo as an evaluation item.
