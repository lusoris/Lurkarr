# Backend Analysis — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Architecture Overview

- **Go 1.26.1** with **Uber FX** dependency injection
- **PostgreSQL 17** via pgx/v5 with goose migrations (7 migrations applied)
- **Multi-module structure**: 20+ internal packages
- **Entry point**: `cmd/lurkarr/main.go` with 8 FX modules

## Package Inventory

| Package | Purpose | Status |
|---------|---------|--------|
| `api` | HTTP handlers (50+ Store methods, 41 protected endpoints) | ✅ Complete |
| `arrclient` | *Arr app HTTP clients (Sonarr, Radarr, Lidarr, Readarr, Whisparr, Eros, Prowlarr) | ✅ Complete |
| `auth` | Authentication (local bcrypt + TOTP, OIDC w/ PKCE, proxy) | ✅ Complete |
| `autoimport` | Stuck import detection & resolution | ✅ Complete |
| `blocklist` | Community blocklist sync + rule matching (4 pattern types) | ✅ Complete |
| `cache` | Settings cache with TTL | ✅ Complete |
| `config` | Environment-based config | ✅ Complete |
| `database` | Models (25+), queries (5 files, 850+ LOC), migrations | ✅ Complete |
| `downloadclients` | Unified interface for 5 client types | ✅ Complete |
| `logging` | Structured JSON logging (slog wrapper) | ✅ Complete |
| `lurking` | Missing/upgrade search engine with exponential backoff | ✅ Complete |
| `metrics` | Prometheus instrumentation (counters, gauges, histograms) | ✅ Complete |
| `middleware` | Recovery, request ID, logging, CORS, rate limiting | ✅ Complete |
| `notifications` | 8 providers (Discord, Telegram, Pushover, Gotify, ntfy, Apprise, Email, Webhook) | ✅ Complete |
| `queuecleaner` | 6-category queue management (blocklist, dedup, strikes, imports, seeding, orphans) | ✅ Complete |
| `scheduler` | gocron/v2 cron scheduler with actions (enable/disable/cap) | ✅ Complete |
| `seerr` | Jellyseerr integration (sync + auto-approve) | ✅ Complete |
| `server` | Route registration + middleware stack | ✅ Complete |

## API Routes (41 Protected + Public)

### Public (no auth)
- `POST /api/auth/login` (rate-limited 5/min per IP)
- `GET/POST /api/auth/setup` (initial admin)
- `GET/POST /api/auth/oidc/*` (OIDC flow)
- `GET /healthz`, `GET /readyz` (K8s probes)
- `GET /metrics` (Prometheus)
- `GET /api/spec`, `GET /api/docs` (OpenAPI + Scalar UI)

### Protected (session + CSRF)
- **Auth**: logout, 2FA enable/disable/verify, recovery codes
- **User**: profile, update username/password
- **Sessions**: list, revoke single, revoke all
- **Admin**: list/create/delete users, reset password, toggle admin
- **Instances**: CRUD, health, test connection
- **Settings**: general GET/PUT, per-app GET/PUT
- **History**: list (paginated + FTS), delete by app
- **Stats**: all, reset, hourly caps
- **State**: reset times, clear state
- **Schedules**: CRUD, execution history
- **Queue Cleaner**: settings, scoring, blocklist/import logs
- **Blocklist**: sources CRUD, rules CRUD
- **Notifications**: providers CRUD, test
- **Download Clients**: instances CRUD, health, test
- **Prowlarr**: indexers, stats, test
- **SABnzbd**: queue, history, pause/resume, test
- **Seerr**: settings, test, requests, count

## Services Started on Boot
1. **Lurking Engine** — polls arr instances for missing/upgrade items
2. **Queue Cleaner** — monitors downloads for stalls, duplicates, orphans
3. **Auto-Importer** — resolves stuck imports
4. **Seerr Sync** — syncs requests with auto-approve
5. **Health Poller** — monitors instance health
6. **Maintenance** — cleans expired sessions, resets states, prunes logs hourly

## Known Issues

### Critical
- **Session revocation bug**: `HandleRevokeAllSessions` deletes ALL sessions including current one. Comment says "preserve current session" but code doesn't actually do this. User gets logged out.

### Medium
- **WebAuthn**: Model defined, migration table created, but NO handlers/endpoints/frontend exist. Dead code.
- **Legacy download client settings**: Old per-app table coexists with new multi-instance table. Unclear migration status.
- **Health poller results**: Not exposed in API responses — dashboard can't show cached health status.

### Low
- **Auto-importer string matching**: Uses hardcoded English strings for error detection. May miss localized messages.
- **No DB transaction wrapping**: Multi-step operations (e.g., 2FA enable + save recovery codes) aren't in a transaction.
- **queuecleaner/hardlink.go**: Uses `syscall.Stat_t` (Linux-only) — fails to compile on Windows.

## Test Status (as of 2026-03-13)

| Package | Tests | Status |
|---------|-------|--------|
| api | 50+ tests | ✅ Pass |
| arrclient | ~20 tests | ✅ Pass |
| auth | ~15 tests | ✅ Pass |
| autoimport | ~5 tests | ✅ Pass |
| blocklist | ~5 tests | ✅ Pass |
| config | ~5 tests | ✅ Pass |
| database | ~3 tests | ✅ Pass |
| downloadclients (5 pkgs) | ~15 tests | ✅ Pass |
| logging | ~3 tests | ✅ Pass |
| lurking | ~10 tests | ✅ Pass |
| middleware | ~10 tests | ✅ Pass |
| notifications | ~5 tests | ✅ Pass |
| queuecleaner | ~15 tests | ❌ Build fail (Windows) |
| scheduler | ~5 tests | ✅ Pass |
| seerr | ~5 tests | ✅ Pass |
| server | ~3 tests | ✅ Pass |

**19/20 packages pass** — only queuecleaner fails due to Linux-only syscall (works in Docker).
