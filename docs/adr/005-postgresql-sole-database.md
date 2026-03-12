# ADR-005: PostgreSQL as Sole Database

**Status:** Accepted  
**Date:** 2026-01-15

## Context

Lurkarr needs persistent storage for users, sessions, app instances, settings, lurk history, stats, schedules, queue cleaner state (strikes, blocklist, scoring profiles), notification providers, and logs. The data model involves relational queries (joins across instances and stats), time-series aggregation (hourly caps, lurk_stats), and concurrent writes from multiple background services.

## Decision

Use PostgreSQL 17 as the sole database, accessed via pgx/v5 (native Go driver) with goose for schema migrations.

Key choices:
- **Single consolidated migration** (`001_initial.sql`) containing all tables, rather than incremental migrations per feature.
- **pgx/v5 native driver** instead of `database/sql` — direct access to PostgreSQL types (JSONB, arrays, intervals), connection pooling, and prepared statements.
- **goose migrations** with SQL files — no Go-code migrations, keeping the schema definition readable and diffable.

## Consequences

**Positive:**
- PostgreSQL handles concurrent writes from lurking engine, queue cleaner, auto-importer, and scheduler without locking contention.
- JSONB columns for flexible fields (notification provider config, scoring profiles).
- pgx/v5 connection pool with configurable `DB_MAX_CONNS` (auto-sized to CPU count × 4).
- goose embedded migrations run automatically at startup.

**Negative:**
- Requires a PostgreSQL instance — no SQLite fallback for simple deployments.
- Single consolidated migration makes it harder to reason about schema evolution over time (mitigated by subsequent incremental migrations for new features).
- pgx/v5 is PostgreSQL-specific — no portability to MySQL/SQLite without a rewrite.

## Alternatives Considered

- **SQLite:** Simpler deployment but doesn't handle concurrent writes well, lacks JSONB, and complicates Docker (needs volume for database file).
- **database/sql + lib/pq:** Standard library interface but misses pgx/v5 performance (binary protocol, prepared statements, batch queries) and PostgreSQL-specific features.
- **GORM/sqlx ORM:** Considered but rejected — raw pgx queries are simpler for this project's query patterns and avoid the ORM abstraction cost.
