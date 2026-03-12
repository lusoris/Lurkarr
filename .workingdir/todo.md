# Lurkarr v2 — Master Todo

> Last updated: 2026-03-12
> State: All phases complete. All lint + gosec issues resolved. ADRs written.
> Priority order: Remaining cleanup items

---

## ✅ COMPLETED — Phase 0: Foundation

- [x] otter/v2 L1 cache (W-TinyLFU, 30s TTL settings cache)
- [x] golang.org/x/time/rate for rate limiting
- [x] Import paths fixed (lusoris/lurkarr)
- [x] Goroutine leak fixed (context.WithCancel)
- [x] CSRF key via crypto/rand
- [x] Settings cache eliminates DB reads per lurk cycle

## ✅ COMPLETED — Phase 1: Security Hardening

- [x] MaxBytesReader (1MB) on all 18 endpoints
- [x] Rate limiting on /api/auth/login (5/min/IP)
- [x] SSRF protection on all test-connection endpoints (apps, Prowlarr, SABnzbd)
- [x] Session rotation after login
- [x] CORS Vary: Origin header
- [x] RequestID in context + helper
- [x] slog attributes preserved
- [x] WebSocket ping/pong heartbeat (30s)
- [x] hourly_caps cleanup (7 day retention)
- [x] Input validation on settings
- [x] Secure cookie flag configurable
- [x] InsecureSkipVerify conditional on sslVerify setting

## ✅ COMPLETED — Phase 2: Interface-Based Lurking Engine

- [x] ArrLurker interface (GetMissing, GetUpgrades, Search, GetQueue)
- [x] 6 lurkers: sonarr, radarr, lidarr, readarr, whisparr, eros
- [x] LurkerFor(appType) registry
- [x] Exponential backoff on arr API errors
- [x] MinDownloadQueueSize enforcement
- [x] Radarr /api/v3/wanted/missing paginated

## ✅ COMPLETED — Phase 3: Multi-Arr Instance Manager (Backend)

- [x] DB multi-instance support (app_instances table)
- [x] Per-instance stats tracking
- [x] Instance-level hourly caps
- [x] Migration 003: instance-aware lurk_stats + hourly_caps

## ✅ COMPLETED — Phase 4: Queue Score Deduplication

- [x] ScoringProfile struct + DB table + queries
- [x] Release name parser (codec, resolution, source, audio, group)
- [x] FindDuplicates() engine
- [x] Strategy: "keep highest score" / "keep first adequate"
- [x] Per-instance queue monitoring loop
- [x] API endpoints for scoring profiles CRUD
- [x] Migration 004: queue management tables

## ✅ COMPLETED — Phase 5: Smart Auto-Import (Stuck Downloads)

- [x] Detect importPending items
- [x] Parse import failure reasons
- [x] Score comparison via GetManualImport
- [x] auto_import_log tracking

## ✅ COMPLETED — Phase 6: Queue Cleaner (Cleanuparr-style)

- [x] Strike system (stalled/slow/failed, configurable thresholds)
- [x] Stalled detection (per privacy type, metadata stuck)
- [x] Slow detection (SABnzbd-aware, speed from timeleft, ignore above size)
- [x] Failed import cleanup (pattern matching, auto-remove + blocklist)
- [x] Migration 005: cleaner enhancements

## ✅ COMPLETED — Download Clients (6 integrated)

- [x] SABnzbd (downloadclients/usenet/sabnzbd/) — full queue/history/settings
- [x] NZBGet (downloadclients/usenet/nzbget/) — XML-RPC
- [x] qBittorrent (downloadclients/torrent/qbittorrent/) — pause/resume/delete with rules
- [x] Transmission (downloadclients/torrent/transmission/) — RPC protocol
- [x] Deluge (downloadclients/torrent/deluge/) — web UI client
- [x] Generic abstraction (downloadclients/) — common Client interface + adapters

## ✅ COMPLETED — Notifications (8 providers)

- [x] Discord (embed webhooks)
- [x] Telegram (Bot API, HTML)
- [x] Pushover (priority + device)
- [x] Gotify (self-hosted)
- [x] Ntfy (ntfy.sh / self-hosted)
- [x] Apprise (proxy server)
- [x] Email (SMTP)
- [x] Webhook (raw HTTP POST)
- [x] Async manager with per-provider event subscription
- [x] Migration 006: notification_providers

## ✅ COMPLETED — Seerr Integration

- [x] Client (requests, media, users, counts)
- [x] SyncEngine with auto-approval loop
- [x] API: settings CRUD, test connection, requests listing
- [x] Migration 007: seerr_settings

## ✅ COMPLETED — Monitoring (Prometheus)

- [x] Prometheus metrics package (internal/metrics/)
- [x] /metrics endpoint via promhttp.Handler
- [x] Metrics: lurk (searches, missing, upgrades, duration, errors)
- [x] Metrics: queue_cleaner (items_removed, strikes, blocklist, duration)
- [x] Metrics: download_client (queue_size, speed, paused)
- [x] Metrics: scheduler (executions, duration, errors)
- [x] Metrics: http (requests, duration, response_size, rate_limit_hits)
- [x] Metrics: autoimport (runs, errors)
- [x] deploy/docker-compose.monitoring.yml (Prometheus + Loki + Grafana)
- [x] Prometheus scrape config (deploy/prometheus.yml)
- [x] Loki config (deploy/loki.yml)
- [x] ServiceMonitor Helm template

## ✅ COMPLETED — Logging (Backend)

- [x] slog JSON handler (configurable level via LOG_LEVEL env)
- [x] Ring buffer (10,000 entries) + async DB flush (500ms/100-item batches)
- [x] WebSocket broadcast for real-time streaming
- [x] Per-app logging
- [x] /api/logs + /ws/logs endpoints
- [ ] **REMOVE** frontend /logs page (replace with Grafana/Loki log exploration)
- [ ] Evaluate: simplify to stdout slog + Loki only (remove DB log storage + WS broadcast if Loki covers it)

## ✅ COMPLETED — CI/CD & Infrastructure

- [x] ci.yml — Go tests with race + coverage, PostgreSQL svc
- [x] docker.yml — Build + push dev images to GHCR
- [x] release.yml — Multi-platform binaries + Docker + Helm
- [x] helm.yml — Lint + package + push Helm chart to GHCR OCI
- [x] security.yml — SAST/dependabot
- [x] Helm chart (v0.1.0) with PostgreSQL dep, ServiceMonitor, ingress
- [x] Dockerfile (multi-stage: node → go → scratch)
- [x] release-please-config.json

## ✅ COMPLETED — Testing Infrastructure

- [x] go.uber.org/mock v0.6.0
- [x] 12 mockgen directives (//go:generate) across packages
- [x] 47+ test files with unit + integration tests
- [x] Race detection + coverage in CI
- [x] Test files for every client (qbit, transmission, deluge, nzbget, seerr)

## ✅ COMPLETED — DB Migrations (goose, single consolidated)

- [x] 001_initial.sql — all tables in one clean migration:
  - Users, sessions, app_instances, app_settings, general_settings
  - Lurking engine: processed_items, state_resets, lurk_history, lurk_stats, hourly_caps
  - Scheduling: schedules, schedule_executions
  - Logging: logs
  - Prowlarr & SABnzbd settings
  - Queue cleaner: settings, strikes, auto_import_log, scoring_profiles, blocklist_log
  - Notifications: notification_providers
  - Seerr: seerr_settings
  - Seed data for all app types

---

# 🔨 IN PROGRESS / TODO

## Phase 16: Code Audit Fixes ⚡ CRITICAL — DO FIRST

> **STATUS: IN PROGRESS** — Dependency bumps done, build fixed. Critical bugs and dead code identified.

### 16.0 Dependency Maintenance (DONE)
- [x] Bump go directive 1.25.8 → 1.26.1 (matches installed toolchain)
- [x] Bump indirect deps: boombuler/barcode, prometheus/common, prometheus/procfs, x/image, x/sync, x/sys, x/text, yaml/v2
- [x] Fix build breakage: `totp.go` — `bytes.Buffer` → `nopCloseBuffer` for `io.WriteCloser` compat with go-qrcode/writer/standard v1.3.0
- [x] All direct deps already at latest (no action needed)
- [x] govulncheck: 0 called vulns. 1 imported vuln (gorilla/csrf GO-2025-3884 TrustedOrigins — NOT used by Lurkarr, no fix available)
- [x] No deprecated direct deps. 1 deprecated indirect (golang/protobuf — transitive via goose, no action needed)

### 16.1 Critical Bugs
- [x] **🔴 Notifications are a no-op** — Fixed: created `internal/notifications/build.go` with `BuildProvider()` + `LoadProviders()`. Providers loaded from DB at startup in main.go; `syncManager()` reloads after CRUD. Tests in `build_test.go`.
- [x] **🔴 `isPrivateTracker()` always returns false** — Fixed: added `IndexerFlags` field to `QueueRecord` (parsed from arr API), private tracker detection now uses flags + known-public-tracker fallback list.
- [x] **🟡 API key length panic** — Fixed: added `len >= 4` guard before slicing in prowlarr_settings.go and sabnzbd_settings.go (matching apps.go pattern).
- [x] **🟡 `writeJSON` silently swallows encoding errors** — Fixed: now logs `slog.Error` on encode failure.
- [x] **🟡 `limitBody` passes nil ResponseWriter** — Fixed: signature changed to `limitBody(w, r)`, all 21 call sites updated.

### 16.2 Data Race
- [x] **🔴 Data race in logging Hub.Broadcast vs Hub.HandleWebSocket** — Fixed: added `sync.RWMutex` to `wsClient` for filter fields. All 22 tests pass with `-race`.

### 16.3 Dead Code / Unused Packages
- [x] **`internal/cache`** — Removed. Was otter-backed W-TinyLFU cache implemented + tested but never wired in. Deleted since otter/v2 is used directly where needed.
- [x] **`internal/downloadclients`** — Now wired into queue cleaner (Phase 7 seeding rules). getTorrentClient() factory creates adapters from DB settings.
- [ ] `deluge.AddTorrentByURL` / `transmission.AddTorrentByURL` — exported but never called (future use)
- [x] `downloadclient/sabnzbd.RemoveItem` — Fixed: uses native `DeleteQueueItem` (mode=queue&name=delete)

### 16.4 Overlapping / Duplicate Code
- [x] **Whisparr v2 rewritten, Eros v3 fixed** — Whisparr v2: movie-based→Sonarr-based series/episode model (WhisparrEpisode, wanted/missing, EpisodeSearch). Eros v3: removed non-existent cutoff endpoint, added ItemType field ("movie"|"scene"), documented client-side filtering. Full API reference in docs/research/api-reference-arr-stack.md.
- [x] **Prowlarr in AllAppTypes() wastes goroutines** — Fixed: lurking engine now skips app types with no registered lurker (matches cleaner/importer pattern).
- [x] SABnzbd: Added `DeleteQueueItem` to native client, fixed adapter `RemoveItem` stub. API handler + cleaner pattern (load settings → create client) is correct for settings that can change at runtime — no further consolidation needed.

### 16.5 Swallowed Errors in Background Services
- [x] `lurking/engine.go`: Fixed 6 swallowed errors → `slog.Warn` on failure
- [x] `queuecleaner/cleaner.go`: Fixed 4 swallowed errors → `slog.Warn` on failure
- [x] `autoimport/importer.go`: Fixed 2 swallowed errors → `slog.Warn` on failure
- [x] `cmd/lurkarr/main.go`: Fixed 7 swallowed errors (prune + maintenance) → `slog.Warn` on failure

### 16.6 Minor Improvements
- [x] Maintenance goroutine in main.go — already uses parent `ctx` correctly (no change needed)
- [x] Seerr sync description — fixed misleading "triggers searches" docstring to "monitors status and auto-approves"

## Phase 7: Download Cleaner (Advanced)

### ✅ Seeding Rules (Torrent Clients) — DONE
- [x] DownloadItem extended with Ratio, SeedingTime, CompletedAt, AddedAt
- [x] qBittorrent, Transmission, Deluge adapters populate seeding fields
- [x] Deluge native client: seeding_time added to defaultFields
- [x] Migration 002: seeding columns on queue_cleaner_settings + download_client_settings table
- [x] QueueCleanerSettings model with 6 seeding fields (enabled, max_ratio, max_hours, mode, delete_files, skip_private)
- [x] DownloadClientSettings model + CRUD queries + API endpoints (GET/PUT /api/queue/download-client/{app})
- [x] cleanSeeding() phase 4 in queue cleaner: torrent client factory, download ID matching, ratio/time enforcement, and/or mode, skip-private
- [x] seedingLimitReached() with 14 table-driven tests
- [x] Delete source files option (SeedingDeleteFiles)

### ✅ Orphan Detection (All Download Clients) — DONE
- [x] GetHistory() added to downloadclient.Client interface
- [x] GetHistory implemented in all 5 adapters (qBit/Transmission/Deluge filter completed; SABnzbd/NZBGet call native GetHistory)
- [x] getTorrentClient → getDownloadClient with SABnzbd + NZBGet support
- [x] cleanOrphans(): aggregates queue records across all *arr instances, cross-refs with download client items+history
- [x] Grace period (orphan_grace_minutes), excluded categories (orphan_excluded_categories), delete files option
- [x] Migration 002 extended with 4 orphan columns
- [x] 12 tests: parseExcludedCategories + orphan detection logic

### ✅ Hardlink & Cross-Seed — DONE
- [x] Hardlink detection (don't remove if hardlinked) — commit `23aba87`
- [x] Cross-seed awareness (detect cross-seeded torrents by content hash match) — commit `39f62c8`

## ✅ COMPLETED — Phase 8: Blocklist System

- [x] Community blocklist sync (configurable list URLs, HTTP ETag conditional fetch) — commit `7f7f2ae`
- [x] Block known bad release groups / patterns (4 pattern types: release_group, title_contains, title_regex, indexer)
- [x] Remove matching downloads from queue (phase 0 in cleaner, before dedup)
- [x] Blocklist API: 8 REST endpoints for sources + rules CRUD
- [x] Migration 003: blocklist_sources + blocklist_rules tables
- [x] 12 blocklist tests (matcher + parser)
- [x] Cross-Arr blocklist sync (propagate across instances of same type) — commit `453b0b6`

## ✅ COMPLETED — Phase 9a: Authentication & Reverse Proxy Support

> All items complete: OIDC login, proxy auth hardening, reverse proxy support, CSRF audit, group mapping, docs.

### OIDC / SSO Support
- [x] OIDC provider configuration (issuer URL, client ID, client secret, scopes)
- [x] OIDC login flow (authorization code + PKCE)
- [x] Token validation + refresh (ID token → local session mapping)
- [x] Auto-create local user on first OIDC login (optional, configurable)
- [x] Group/role claim mapping (e.g., admin group → Lurkarr admin) — commit `b2a608a`
- [x] Support multiple providers (Authentik, Keycloak, Authelia, Dex, Google, etc.) — implementation is provider-agnostic via standard OIDC discovery; simultaneous multi-provider deferred
- [x] `/api/auth/oidc/callback` endpoint
- [x] Frontend login page: "Sign in with SSO" button alongside local login
- [x] DB migration for OIDC fields (auth_provider, external_id on users table)
- [x] Migration 004: is_admin column on users table

### Proxy Authentication Hardening
- [x] Trusted proxy IP allowlist (`TRUSTED_PROXIES` env — CIDR ranges, default: private ranges only)
- [x] Reject proxy auth headers from untrusted source IPs
- [x] Support multiple proxy header formats (comma-separated PROXY_HEADER env) — commit `b2a608a`
- [x] Auto-create user on first proxy auth if not exists (configurable)
- [x] Proxy auth + CSRF interaction audit — CSRF correctly enforced even with proxy auth; fixed missing frontend CSRF token handling
- [x] Log warning when proxy auth enabled without trusted proxy config

### Reverse Proxy Support
- [x] Base path / sub-path support (`BASE_PATH` env, e.g. `/lurkarr/`) — prefix all routes + static assets
- [x] Trusted proxy config for `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Real-IP` (rate limiter validates source IP)
- [x] Respect `X-Forwarded-Proto` for secure cookie decisions
- [x] WebSocket origin patterns configurable via AllowedOrigins — commit `b2a608a`
- [x] Health check endpoint (`/api/health`) bypasses auth — for load balancer probes
- [x] Document reverse proxy configs (Traefik, Caddy, nginx, HAProxy) in README or docs/ — covered in README rewrite

## Phase 9b: Uber FX Dependency Injection ⚡ PRIORITY — DO BEFORE FEATURE WORK

> **STATUS: COMPLETE** — fx.New() app lifecycle with modules, providers, and lifecycle hooks

- [x] Add go.uber.org/fx dependency
- [x] Define fx.Module per package (config, database, logging, notifications, scheduler, server, services, maintenance)
- [x] Refactor main.go from manual wiring to fx.New() app lifecycle
- [x] Use fx.Provide / fx.Invoke for service startup
- [x] Add fx.Lifecycle hooks for graceful shutdown (replaces manual defer chains + signal handling)
- [x] Health check integration (existing /api/health endpoint, fx handles signal-based shutdown)

## ✅ COMPLETED — Phase 10: Grafana Dashboards (Professional)

> **STATUS: COMPLETE** — 3 dashboards: enhanced overview (lurkarr.json), system/runtime (lurkarr-system.json), logs (lurkarr-logs.json)

- [x] Lurking Dashboard — per-app/instance search rates, missing/upgrade trends, error rate %, duration histograms (p50/p95/p99)
- [x] Queue Cleaner Dashboard — strikes issued, items removed, blocklist additions, cleaner run duration (p95)
- [x] Download Clients Dashboard — queue sizes, speeds (up/down), paused states, per-client comparison
- [x] Auto-Import Dashboard — import runs by app type, errors by app type + instance
- [x] Scheduler Dashboard — task executions, durations (p95), errors by task type
- [x] HTTP/API Dashboard — request rates, latencies (p50/p95/p99), error rate %, rate limit hits, top endpoints, response sizes
- [x] Notifications Dashboard — send counts, per-provider success/failure rates, delivery latency (lurkarr-notifications.json)
- [x] System Overview Dashboard — Go runtime metrics (goroutines, heap, RSS, GC pause, alloc rate, FDs, CPU, threads, uptime)
- [x] Loki Log Dashboard — structured log exploration, volume by level, error aggregation, component breakdown, text search
- [ ] Arr Stack Overview Dashboard — Sonarr/Radarr/Lidarr/etc health (BLOCKED: requires scraping arr /api endpoints directly)
- [x] Dashboard variables — $app_type + $instance pickers on overview, $level + $search on logs, $DS_PROMETHEUS datasource selector

## ✅ COMPLETED — Phase 11: OpenAPI Spec

- [x] Write openapi.yaml spec for full Lurkarr API (68 endpoints, 40+ schemas)
- [x] Spec covers all routes: auth, user, instances, settings, history, logs, stats, state, schedules, prowlarr, sabnzbd, queue, blocklist, notifications, seerr, health, metrics, websocket
- [x] Embedded via internal/openapi package (go:embed)

## ✅ COMPLETED — Phase 12: Scalar API Documentation

- [x] Serve OpenAPI spec at /api/spec (YAML, cacheable)
- [x] Add Scalar HTML page at /api/docs (interactive API reference via CDN)
- [x] Zero-config: reads embedded openapi.yaml, base path aware

## ✅ COMPLETED — Phase 13: Coder Development Environment

> **STATUS: COMPLETE** — Terraform template in deploy/coder/main.tf
> **Coder instance:** https://code.dev.cauda.dev (K8s-based)

- [x] Create Coder template (Terraform-based, K8s provisioner) for Lurkarr development
- [x] Include: Go toolchain, Node.js, PostgreSQL sidecar, Docker-in-Docker
- [x] Pre-install VS Code extensions (Go, Svelte, Tailwind, GitLens)
- [x] Auto-clone repo from GitHub (already linked)
- [x] Include monitoring stack (Prometheus + Grafana) in dev environment — Grafana app exposed on :3000
- [x] Environment variables + secrets management
- [ ] Onboard project on Coder instance (https://code.dev.cauda.dev) — manual step
- [x] Document template usage in README

## ✅ COMPLETED — Phase 14: Frontend Gaps

- [x] Instance management UI (add/remove/name per app type) — Phase 3 backend done
- [x] Notifications settings page (provider CRUD, event config, test send) — commit `f87b101`
- [x] Seerr settings page (URL, API key, sync config) — commit `f87b101`
- [x] Download clients settings page (all 6 clients) — commit `f87b101`
- [x] Mobile-responsive sidebar (hamburger menu, slide-in drawer) — commit `f87b101`
- [x] Svelte transitions (fly toasts, fade/scale modals) — commit `f87b101`
- [x] Responsive grids, scrollable tabs, stacking layouts — commit `f87b101`
- [x] Monitoring/Grafana embed or link page — monitoring page with health, endpoints, Grafana dashboard info
- [x] Auto-import config UI (enable/disable per instance, score threshold) — auto-import fields visible in queue cleaner settings; per-instance control deferred (needs backend)
- [x] Cross-instance dedup detection settings — scoring profile tab + cross_arr_sync toggle in cleaner settings
- [x] Download client settings tab in Queue Management page
- [x] Seeding rules, orphan cleanup, hardlink protection, cross-seed settings in Queue Cleaner tab

## ✅ COMPLETED — Phase 15: Documentation & Research

- [x] Technology stack overview (docs/research/tech-stack.md)
- [x] Uber FX deep dive + migration plan (docs/research/uber-fx.md)
- [x] Testing & gomock patterns (docs/research/testing-gomock.md)
- [x] Prometheus metrics & Grafana dashboards (docs/research/prometheus-grafana.md)
- [x] Coder template reference (docs/research/coder-template.md)
- [x] Arr stack + download client APIs (docs/research/api-reference-arr-stack.md)
- [x] OpenAPI + ogen + Scalar (docs/research/openapi-ogen-scalar.md)
- [x] Go security hardening (docs/research/security-hardening.md)
- [x] Go best practices (docs/research/go-best-practices.md)
- [x] SvelteKit 5 + TailwindCSS v4 (docs/research/sveltekit-tailwind.md)
- [x] PostgreSQL + pgx + goose (docs/research/database-pgx-goose.md)
- [x] User-facing documentation (README.md — comprehensive rewrite with all features, env vars, deployment, reverse proxy, OIDC, monitoring)
- [x] Architecture decision records (ADRs) — 8 ADRs in docs/adr/

---

## Package Decisions

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/maypok86/otter/v2` | v2.3.0 | L1 cache (W-TinyLFU) |
| `golang.org/x/time/rate` | latest | Rate limiting |
| `go.uber.org/mock` | v0.6.0 | Mock generation for tests |
| `github.com/prometheus/client_golang` | v1.23.2 | Prometheus metrics |
| `github.com/go-co-op/gocron/v2` | v2.19.1 | Task scheduling |
| `github.com/jackc/pgx/v5` | v5.8.0 | PostgreSQL driver |
| `github.com/pressly/goose/v3` | v3.27.0 | DB migrations |
| `github.com/coder/websocket` | v1.8.14 | WebSocket (logs, real-time) |
| `github.com/gorilla/csrf` | v1.7.3 | CSRF protection |
| `github.com/pquerna/otp` | v1.5.0 | TOTP 2FA |
| `go.uber.org/fx` | TBD | Dependency injection (Phase 9b) |
| `github.com/coreos/go-oidc/v3` | TBD | OIDC token verification (Phase 9a) |
| `golang.org/x/oauth2` | TBD | OAuth2 authorization code flow (Phase 9a) |
| `github.com/ogen-go/ogen` | TBD | OpenAPI codegen (Phase 11) |
| `@scalar/api-reference` | TBD | API docs UI (Phase 12) |

**NOT using:**
- River (overkill — gocron sufficient for single binary)
- Ristretto (Otter beats it on all metrics)
- sqlc / squirrel (hand-written pgx queries fine for our scope)
- Wire / Dig (Uber FX chosen instead)
