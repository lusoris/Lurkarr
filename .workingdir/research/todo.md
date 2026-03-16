# Research & Audit Findings — Master TODO

> Generated from deep codebase audit + web research comparison.  
> Priority: Critical > Warning > Info. Within each level, items sorted by impact.

---

## Critical

- [x] **Add HTTP security headers middleware** — `internal/server/server.go` (line ~565)  
  Missing `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Content-Security-Policy`, `Referrer-Policy`, `Permissions-Policy`.  
  Add a single `SecurityHeaders` middleware to the `middleware.Chain(...)` call.

- [x] **Add `Strict-Transport-Security` header** — `internal/server/server.go`  
  Should be conditional (only when behind TLS or a trusted reverse proxy). `max-age=31536000; includeSubDomains`.

---

## Warning — Security & Config

- [x] **Validate `CSRF_KEY` env var** — `internal/config/config.go` (~line 62)  
  Empty `CSRF_KEY` means gorilla/csrf generates a random key each restart → all user sessions invalidate on every deploy. Should log a warning or fail if empty.

- [x] **Prevent rate limit of 0** — `internal/config/config.go` (~lines 84-85)  
  `LOGIN_RATE_LIMIT=0` or `API_RATE_LIMIT=0` silently disables rate limiting. Add a minimum floor (e.g., 1).

- [x] **Validate `LOG_LEVEL` against known values** — `internal/config/config.go`  
  Accept only `debug`, `info`, `warn`, `error`. Log a warning and default to `info` for unknown values.

---

## Warning — Database

- [x] **Add index on `sessions.user_id`** — new migration  
  `DELETE FROM sessions WHERE user_id = $1` and session lookup queries do sequential scans on sessions table without this index.

- [x] **Add index on `schedule_executions.executed_at`** — new migration  
  Queries sorting by `executed_at DESC` benefit from this index.

- [x] **Add index on `notification_history.provider_type`** — new migration  
  Filterable by provider type; currently no index.

---

## Warning — Backend Test Coverage

- [ ] **Add tests for `internal/downloadclients/` adapter layer** — 8 source files, 0 tests  
  The individual clients (qbittorrent, transmission, etc.) are tested but the adapter/registry layer is not.

- [ ] **Add tests for `internal/healthpoller/`** — 1 source file, 0 tests

- [ ] **Add tests for `internal/bazarrclient/`** — 1 source file, 0 tests

- [ ] **Add tests for `internal/kapowarrclient/`** — 1 source file, 0 tests

- [ ] **Add tests for `internal/shokoclient/`** — 1 source file, 0 tests

- [ ] **Add tests for `internal/downloadclients/torrent/rtorrent/`** — 1 source file, 0 tests

- [ ] **Add tests for `internal/downloadclients/torrent/utorrent/`** — 1 source file, 0 tests

- [ ] **Increase test coverage for `internal/database/`** — 15 source files, only 2 test files  
  Query logic (especially complex queries in `queries.go`, `queries_apps.go`, etc.) is mostly untested.

- [ ] **Increase test coverage for `internal/notifications/`** — 11 source files, only 2 test files  
  8 notification provider implementations have zero unit tests.

---

## Warning — Frontend Test Coverage

- [ ] **Add tests for route pages** — 0 of 17 routes tested  
  Untested: dashboard (`+page.svelte`), activity, admin/users, apps, dedup, downloads, help, history, login, lurk, monitoring, notifications, queue, scheduling, seerr, settings, user.

- [ ] **Add tests for custom lib components** — 0 of 7 tested  
  Untested: `Breadcrumbs.svelte`, `CommandPalette.svelte`, `HelpDrawer.svelte`, `InstanceSwitcher.svelte`, `ScrollToTop.svelte`, `Sidebar.svelte`, `Toast.svelte`.

---

## Warning — Frontend Code Quality

- [x] **Fix empty `catch {}` blocks in dashboard** — `frontend/src/routes/+page.svelte`  
  ~12 empty catch blocks silently swallow network errors. Users see stale/empty data with no failure indication. At minimum, log to console or show a toast.

- [x] **Fix empty `catch {}` blocks in other pages** — `settings/+page.svelte`, `dedup/+page.svelte`, `monitoring/+page.svelte`  
  Same pattern — silent failure on API errors.

- [x] **Replace `catch (e: any)` with `catch (e: unknown)`** — `frontend/src/routes/admin/users/+page.svelte` (lines 62, 81, 91, 113)  
  Use `e instanceof Error` checks instead of `any` cast.

- [ ] **Add client-side form validation** — across all settings/admin/apps forms  
  No `required` attributes, no URL format validation, no min/max bounds, no field-level error messages. Add HTML5 validation attributes + inline error feedback.

- [x] **Add `aria-label` to icon-only buttons** (~9 instances)  
  Screen readers can't identify purpose. Key files: `ScrollToTop.svelte`, `queue/+page.svelte` (edit/delete icons), `+page.svelte` (reset icon), `apps/+page.svelte` (remove icon).

---

## Warning — Backend Robustness

- [x] **Fix silent `json.Unmarshal` suppression in uTorrent client** — `internal/downloadclients/torrent/utorrent/client.go` (~lines 173-178)  
  6x `_ = json.Unmarshal(...)` for Status, Size, Progress, Downloaded, Uploaded, Ratio. Log a warning on parse failure rather than silently using zero values.

- [x] **Add notification retry logic** — `internal/notifications/notifications.go`  
  Currently fire-once with no retry. Failed sends are logged but never retried. Add at least 1 retry with backoff for transient errors (5xx, timeout).

- [x] **Add external API call duration metrics** — arr clients, bazarr, seerr, shoko, kapowarr  
  No histogram tracking outbound HTTP latency. Key for diagnosing slow external APIs. Add `lurkarr_external_api_duration_seconds` histogram with labels `{client, method}`.

---

## Info — Frontend Cleanup

- [x] **Consolidate lucide icon packages** — `frontend/package.json`  
  Both `lucide-svelte` (dependencies) and `@lucide/svelte` (devDependencies) declared. 25 custom files use `lucide-svelte`, 27 shadcn-svelte files use `@lucide/svelte`. Migrate all to `@lucide/svelte` (the official scoped package going forward) and remove `lucide-svelte`.

- [x] **Remove unused `@sveltejs/adapter-auto`** — `frontend/package.json` (devDependencies)  
  Declared but unused; `@sveltejs/adapter-static` is the actual adapter.

- [x] **Add descriptive `alt` text to app logo images** — dashboard, apps page, InstanceSwitcher  
  Currently `alt=""` (decorative), but these logos appear next to meaningful app names. Use `alt="{appName} logo"`.

- [ ] **Add `aria-live="polite"` regions for dynamic content** — search results, data refresh indicators, filter result counts  
  Screen readers don't announce when data finishes loading or search results change.

- [ ] **Type `any` usages to clean up** — `apps/+page.svelte` (line 70, Prowlarr indexer), `seerr/+page.svelte` (line 75, duplicates response), `DataTable.svelte` (line 6)  
  Replace with proper typed interfaces.

---

## Info — Backend Cleanup & Hardening

- [x] **Expose database connection pool metrics** — pgx pool stats → Prometheus  
  Use `pool.Stat()` to export idle connections, acquired connections, max connections, wait time. Helps diagnose connection exhaustion.

- [ ] **Add notification circuit breaker** — `internal/notifications/notifications.go`  
  A persistently failing provider (e.g., Discord returning 5xx for hours) is called on every event. Consider disabling after N consecutive failures with periodic retry.

- [x] **Add `govulncheck` / `gosec` Makefile target** — `Makefile`  
  No security scanning target exists. Add `security: govulncheck ./... && gosec ./...`.

- [x] **Add coverage Makefile target** — `Makefile`  
  A `coverage/` directory exists but no target to generate reports. Add `cover: go test -coverprofile=coverage/coverage.out ./... && go tool cover -html=...`.

- [x] **Clean up `openapi.yaml.bak`** — `internal/openapi/`  
  Backup file suggests manual spec editing. Consider removing or adding to `.gitignore`.

- [ ] **Display Readarr retirement notice in UI** — `frontend/src/routes/apps/+page.svelte`  
  Readarr is officially retired by the Servarr team (metadata unusable). When a user configures a Readarr instance, show a warning banner.

---

## Info — Observability Gaps

- [ ] **Add blocklist sync metrics** — `internal/blocklist/sync.go`  
  No instrumentation for sync duration, error count, or items synced.

- [ ] **Add active sessions gauge** — `internal/metrics/`  
  Track current active session count to detect session leak issues.

---

---

## Warning — Multi-Instance Cross-Arr Gaps

> Based on real-world multi-instance patterns from TRaSH Guides, Servarr wiki, and community usage.  
> People run multi-instance arrs for: 4K/1080p split, anime vs regular, language-specific libraries,
> private tracker separation, streaming-optimized vs archival quality profiles.

- [ ] **Enforce split-season rules in lurk engine & queue cleaner** — `internal/lurking/engine.go`, `internal/queuecleaner/cleaner.go`  
  Split-season rules exist in DB (`split_season_rules` table, migration 029) and have full API (`GET/POST/DELETE`), but **no module reads or enforces them**. The lurk engine should respect season assignments (only search for missing episodes on the assigned instance), and the queue cleaner should not cross-instance clean items covered by a split rule. Currently a dead feature.

- [ ] **Add split-season rule management UI** — `frontend/src/routes/dedup/+page.svelte` or new sub-route  
  Split-season rules have API endpoints but **zero frontend UI**. Users must use raw API calls to create/manage rules. Add a collapsible section per media item in the dedup overlap matrix showing season ranges and allowing assignment to specific instances.

- [ ] **Add instance health/status dashboard** — new component or section in existing admin/monitoring page  
  No way to see at-a-glance which instances are healthy, responding, or degraded. Users with 4-6+ instances (common in multi-4K setups) need: connectivity status, last successful API call, response time, disk space, queue size per instance. TRaSH Guides document that instance health monitoring is a major pain point.

- [ ] **Add "action buttons" to overlap matrix** — `frontend/src/routes/dedup/+page.svelte`  
  The overlap matrix correctly detects duplicates but provides **no actions**. Users must manually open each arr instance to unmonitor/delete. Add: "Remove from lower-ranked instance", "Unmonitor on instance X", "Move to instance Y" action buttons per media row.

- [ ] **Add cross-instance search fallback** — `internal/lurking/engine.go`  
  If a search fails on instance A (indexer error, rate limit), the same media is never tried on instance B. In multi-instance setups where instances may share indexers via Prowlarr, a cross-instance search retry would improve completion rates.

- [ ] **Add instance group-level configurable settings** — `internal/database/models.go`, `instance_groups` table  
  Group mode (`quality_hierarchy`, `overlap_detect`, `split_season`) is stored but all hierarchy logic is hardcoded (rank 1 = always best). Real users need: configurable behavior for what happens to lower-ranked copies (keep, unmonitor, delete), minimum quality threshold before routing, Seerr auto-decline toggle per group.

- [ ] **Add quality profile sync awareness** — `internal/arrclient/`  
  Multi-instance users following TRaSH Guides use Import Lists to sync libraries between instances (Option 1: full sync, Option 2: profile-based cherry-pick, Option 3: tag-based). Lurkarr should detect when instances are already synced via Import Lists and factor this into dedup decisions to avoid conflicting with arr-native sync.

- [ ] **Add cross-instance blocklist sync dashboard** — `internal/queuecleaner/cleaner.go`, frontend  
  `CrossArrSync` setting propagates blocklist removals across instances but has **no visibility**. Add a log/history view showing: which items were synced, from which instance to which, and why. The `cross_instance_actions` table partially covers this but only for Seerr routing, not for cleaner blocklist syncs.

- [ ] **Add Seerr auto-decline for flagged duplicates** — `internal/seerr/router.go`  
  `ScanForDuplicates` flags requests but takes **no action**. Users must manually decline each flagged request in Seerr. Add an optional "auto-decline" mode that uses Seerr API to decline requests that match duplicate criteria, with a dry-run preview first.

---

## Warning — Frontend UX Improvements

> Based on dashboard/admin UX best practices, data-heavy panel patterns, and Material Design interaction states.

- [ ] **Add skeleton loading states to all data pages** — dashboard, history, queue, downloads, activity, dedup  
  Currently pages show nothing or a brief empty state while data loads. Replace with skeleton placeholders (animated gray blocks matching the layout) so users perceive faster loading. Use shadcn-svelte `Skeleton` component (already in deps).

- [ ] **Add error boundary with retry for API failures** — all route pages  
  Empty catch blocks mean users see blank sections with no way to recover. Wrap data-fetching sections in an error boundary pattern: show error message with "Retry" button, log to console, optionally show toast. This directly addresses the 15+ empty catch blocks.

- [ ] **Add optimistic UI updates for toggle/switch operations** — settings, apps, notification providers  
  Currently every toggle (enable/disable instance, enable/disable notification) requires a full save + wait for server response. Implement optimistic updates: flip the UI immediately, send request in background, revert on error with toast.

- [ ] **Improve keyboard navigation & shortcuts** — global  
  The app has a `CommandPalette.svelte` component — extend it with common actions: "Go to Settings" (`Ctrl+,`), "Go to Queue" (`Ctrl+Q`), "Toggle sidebar" (`Ctrl+B`), "Search" (`Ctrl+K`). Ensure all interactive elements are keyboard-focusable with visible focus rings (the `:focus-visible` ring audit).

- [ ] **Add responsive table patterns for small screens** — queue, history, downloads, activity  
  Tables with 5+ columns break on tablet/mobile. Implement: responsive card layout below breakpoint, or horizontal scroll with sticky first column, or collapsible row details. Check all `DataTable.svelte` usages.

- [ ] **Add data refresh indicator & auto-refresh** — dashboard, queue, monitoring  
  No visual indication of when data was last fetched. Add "Last updated X seconds ago" timestamp + auto-refresh toggle (30s/60s/off). Queue and monitoring pages especially benefit from near-real-time updates.

- [ ] **Add empty state illustrations** — all list/table pages  
  When no instances configured, no queue items, no history, etc., show helpful empty states with illustration + call-to-action ("Add your first Sonarr instance" with link to apps page). Currently most pages just show blank space.

- [ ] **Add bulk actions for queue management** — `frontend/src/routes/queue/+page.svelte`  
  Queue page operations are per-item only. Multi-instance users with large queues need: select all, select by instance, bulk remove, bulk blocklist, bulk retry. Use checkbox column + floating action bar pattern.

- [ ] **Add toast notifications for background operations** — global  
  `svelte-sonner` is installed but underutilized. All API mutations (save settings, add instance, run lurk, clean queue) should show success/error toasts. Currently most operations have no feedback beyond the catch block.

- [ ] **Add confirmation dialogs for destructive actions** — delete instance, remove media, blocklist, clear queue  
  Several destructive operations happen on single click with no confirmation. Add a confirmation dialog for: delete app instance, bulk queue operations, blocklist sync, and any delete/remove action.

- [ ] **Add breadcrumb navigation consistency** — `Breadcrumbs.svelte` exists but usage is inconsistent  
  Some pages use breadcrumbs, others don't. Standardize across all routes for consistent wayfinding, especially in nested routes like `admin/users`.

- [ ] **Add progressive disclosure for settings pages** — `frontend/src/routes/settings/+page.svelte`  
  Settings page likely has many options on one long page. Group into collapsible sections with clear headers (General, Security, Notifications, Advanced). Only show advanced options when explicitly expanded.

- [ ] **Improve form validation UX** — all forms with inputs  
  Beyond adding HTML5 validation attributes (covered in existing todo), implement: real-time inline validation on blur, URL reachability test for API endpoints ("Test Connection" buttons already exist — add visual feedback), password strength indicator for local auth.

---

## Info — Multi-Instance Ecosystem Awareness

- [ ] **Document Recyclarr integration path** — `.workingdir/research/` or docs  
  Recyclarr syncs TRaSH Guide quality profiles/custom formats to arr instances automatically. Many multi-instance users rely on it. Lurkarr should document compatibility (operates on same API endpoints) and potentially detect Recyclarr-managed instances to avoid conflicting changes.

- [ ] **Consider Notifiarr/Unpackerr/Kometa awareness** — informational  
  Multi-instance users commonly also run Notifiarr (centralized notifications), Unpackerr (extract archives), and Kometa (Plex metadata management). Lurkarr's notification system could overlap with Notifiarr — document how they coexist. Kometa tags could theoretically integrate with Lurkarr's instance group tags.

- [ ] **Add per-instance label/tag display in UI** — across all instance-aware pages  
  Multi-instance users use naming conventions: "Radarr-4K", "Radarr-Anime", "Sonarr-DE" (German). The UI should prominently show instance names/tags (with optional color coding) in queue, history, downloads, and dedup pages so users can quickly identify which instance is involved.

- [ ] **Add download client category validation** — `internal/downloadclients/`  
  TRaSH Guides and Servarr wiki emphasize that each instance MUST use a separate download client category. Lurkarr could validate this: warn if two instances of the same app type share the same download client category, since this causes cross-contamination.

---

## Done (Fixed During Audit)

- [x] **Frontend test failure — `@lucide/svelte` not installed** — `npm install` resolved
- [x] **Frontend test failure — Select custom class assertion** — changed `label.my-select` → `div.my-select` in `more-components.test.ts`
