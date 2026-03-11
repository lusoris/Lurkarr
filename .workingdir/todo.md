# Lurkarr v2 — Master Todo

> Last updated: 2026-03-11
> State: Phases 0–6 complete, Phase 9a/9b complete, partial Phase 7. Phase 15 (docs/research) ~90% done. Phase 16 (code audit) ~95% done. Refactors: hunting→lurking rename, download client restructure, migration consolidation, Whisparr/Eros API rewrite, Seerr naming unification.
> Priority order: Remaining Phase 16 cleanup → Phase 7 (Download Cleaner) → feature work

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
- [ ] **`internal/downloadclients`** — unified Client interface + 5 adapters built for Phase 7 (Download Cleaner Advanced). Not dead — pending integration.
- [ ] `deluge.AddTorrentByURL` / `transmission.AddTorrentByURL` — exported but never called (Phase 7 planned use)
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

### Seeding Rules (Torrent Clients)
- [ ] Per-category seeding rules (max ratio, min/max seed time)
- [ ] Delete source files option
- [ ] Wire integrated torrent clients (qBit, Transmission, Deluge) into cleaner

### Unlinked Download Handling
- [ ] Detect orphan downloads not linked to any arr item
- [ ] Hardlink detection (don't remove if hardlinked)
- [ ] Cross-seed awareness

## Phase 8: Malware & Blocklist

- [ ] Community blocklist sync (configurable list URLs)
- [ ] Block known bad release groups / patterns
- [ ] Remove matching downloads from queue
- [ ] Cross-Arr blocklist sync (propagate across instances of same type)

## Phase 9a: Authentication & Reverse Proxy Support

> **STATUS: CORE COMPLETE** — OIDC login (authz code + PKCE), proxy auth hardening (trusted proxy IP validation), reverse proxy support (BASE_PATH, XFF, XFP) all implemented. Remaining: group/role mapping, multi-provider, multi-header, WebSocket, docs.

### OIDC / SSO Support
- [x] OIDC provider configuration (issuer URL, client ID, client secret, scopes)
- [x] OIDC login flow (authorization code + PKCE)
- [x] Token validation + refresh (ID token → local session mapping)
- [x] Auto-create local user on first OIDC login (optional, configurable)
- [ ] Group/role claim mapping (e.g., admin group → Lurkarr admin)
- [ ] Support multiple providers (Authentik, Keycloak, Authelia, Dex, Google, etc.)
- [x] `/api/auth/oidc/callback` endpoint
- [x] Frontend login page: "Sign in with SSO" button alongside local login
- [x] DB migration for OIDC fields (auth_provider, external_id on users table)

### Proxy Authentication Hardening
- [x] Trusted proxy IP allowlist (`TRUSTED_PROXIES` env — CIDR ranges, default: private ranges only)
- [x] Reject proxy auth headers from untrusted source IPs
- [ ] Support multiple proxy header formats (Remote-User, X-Forwarded-User, X-authentik-username, etc.)
- [x] Auto-create user on first proxy auth if not exists (configurable)
- [ ] Proxy auth + CSRF interaction audit (bypass CSRF when proxy auth active?)
- [x] Log warning when proxy auth enabled without trusted proxy config

### Reverse Proxy Support
- [x] Base path / sub-path support (`BASE_PATH` env, e.g. `/lurkarr/`) — prefix all routes + static assets
- [x] Trusted proxy config for `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Real-IP` (rate limiter validates source IP)
- [x] Respect `X-Forwarded-Proto` for secure cookie decisions
- [ ] WebSocket upgrade behind reverse proxy (wss:// handling, connection upgrade headers)
- [x] Health check endpoint (`/api/health`) bypasses auth — for load balancer probes
- [ ] Document reverse proxy configs (Traefik, Caddy, nginx, HAProxy) in README or docs/

## Phase 9b: Uber FX Dependency Injection ⚡ PRIORITY — DO BEFORE FEATURE WORK

> **STATUS: COMPLETE** — fx.New() app lifecycle with modules, providers, and lifecycle hooks

- [x] Add go.uber.org/fx dependency
- [x] Define fx.Module per package (config, database, logging, notifications, scheduler, server, services, maintenance)
- [x] Refactor main.go from manual wiring to fx.New() app lifecycle
- [x] Use fx.Provide / fx.Invoke for service startup
- [x] Add fx.Lifecycle hooks for graceful shutdown (replaces manual defer chains + signal handling)
- [x] Health check integration (existing /api/health endpoint, fx handles signal-based shutdown)

## Phase 10: Grafana Dashboards (Professional)

> **STATUS: PARTIAL** — 1 overview dashboard exists (460 lines), metrics endpoint + full monitoring stack deployed

- [ ] Lurking Dashboard — per-app/instance search rates, missing/upgrade trends, error rates, duration histograms
- [ ] Queue Cleaner Dashboard — strikes issued, items removed, blocklist additions, stalled/slow/failed breakdown
- [ ] Download Clients Dashboard — queue sizes, speeds (up/down), paused states, per-client comparison
- [ ] Auto-Import Dashboard — import runs, successes vs errors, score distributions
- [ ] Scheduler Dashboard — task executions, durations, errors, rule activity timeline
- [ ] HTTP/API Dashboard — request rates, latencies (p50/p95/p99), error rates, rate limit hits
- [ ] Notifications Dashboard — send counts, per-provider success/failure rates
- [ ] System Overview Dashboard — Go runtime metrics (goroutines, memory, GC), DB connection pool stats
- [ ] Loki Log Dashboard — structured log exploration, error aggregation (REPLACES frontend /logs page)
- [ ] Arr Stack Overview Dashboard — Sonarr/Radarr/Lidarr/etc health, queue sizes, activity status
- [ ] Dashboard variables — instance picker, time range, app_type filter across all dashboards

## Phase 11: OpenAPI Spec + ogen Codegen

- [ ] Write openapi.yaml spec for full Lurkarr API
- [ ] Generate server stubs with ogen
- [ ] Rewrite API handlers as ogen interface implementations
- [ ] Built-in request validation (replaces manual input validation)
- [ ] Type-safe request/response structs

## Phase 12: Scalar API Documentation

- [ ] Serve OpenAPI spec at /api/spec
- [ ] Add Scalar HTML page at /api/docs (interactive API reference)
- [ ] Zero-config: reads same openapi.yaml used by ogen

## Phase 13: Coder Development Environment

> **STATUS: NOT STARTED** — No template, devcontainer, or .tf files exist
> **Coder instance:** https://code.dev.cauda.dev (K8s-based)

- [ ] Create Coder template (Terraform-based, K8s provisioner) for Lurkarr development
- [ ] Include: Go toolchain, Node.js, PostgreSQL sidecar, Docker-in-Docker
- [ ] Pre-install VS Code extensions (Go, Svelte, Tailwind, GitLens)
- [ ] Auto-clone repo from GitHub (already linked)
- [ ] Include monitoring stack (Prometheus + Grafana) in dev environment
- [ ] Environment variables + secrets management
- [ ] Onboard project on Coder instance (https://code.dev.cauda.dev)
- [ ] Document template usage in README

## Phase 14: Frontend Gaps

- [ ] Instance management UI (add/remove/name per app type) — Phase 3 backend done
- [ ] Notifications settings page (provider CRUD, event config, test send) — backend done
- [ ] Seerr settings page (URL, API key, sync config, request viewer) — backend done
- [ ] Download clients settings page (all 6 clients) — backend done
- [ ] Monitoring/Grafana embed or link page
- [ ] Auto-import config UI (enable/disable per instance, score threshold)
- [ ] Cross-instance dedup detection settings

## Phase 15: Documentation & Research

> **STATUS: IN PROGRESS** — 11 research docs created in docs/research/

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
- [ ] Architecture decision records (ADRs)
- [ ] User-facing documentation (setup guide, configuration reference)

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
