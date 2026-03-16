# Lurkarr — Active TODO

## Phase 1: Dev Stack Overhaul ✅

### 1.1 Multi-Instance Arr Apps ✅
- [x] Add second Sonarr instance (`sonarr-anime` — for anime profile testing)
- [x] Add second Radarr instance (`radarr-4k` — for quality hierarchy dedup testing)
- [x] Add Whisparr v2 instance (hotio `whisparr:v2`, port 6968)
- [x] Verify both Whisparr v2 + v3 are selectable in Lurkarr UI

### 1.2 Missing Services ✅
- [x] Add Jellyfin (hotio, port 8097 — 8096 conflicts with host)
- [x] Add rTorrent + ruTorrent (crazymax image, XMLRPC on 8000, web on 8484)
- [x] ~~Configure Seerr → Jellyfin connection~~ → moved to Phase 7.2

### 1.3 Dev Stack Cleanup ✅
- [x] Organize volumes with clear naming
- [x] Add shared media library structure (/data/media/tv, /data/media/movies, etc.)
- [x] Comment/document every service in docker-compose.dev.yml
- [x] Ensure all services use hotio images where available (per convention)
- [x] YAML anchors (x-common) for shared env

### 1.4 Seed Data & Test Fixtures ✅
- [x] Create `dev/seed.sh` — API-driven setup script (all instances, DL clients, Prowlarr, media)
- [x] Fix sample media add — wrong TVDB IDs, qualityProfileId, JSON whitespace
- [x] Fix pagination mock bug in lurking tests (QueueResponse without records → infinite loop)
- [x] ~~Create test data for queue cleaner~~ → moved to Phase 7.2
- [x] ~~Create test data for dedup~~ → moved to Phase 7.2

---

## Phase 2: New App Integrations

### 2.1 Bazarr (Subtitle Management) → Phase 7.3
- [x] Add bazarr service to dev stack ✅ (already in docker-compose.dev.yml)
- [x] ~~Remaining Bazarr work~~ → moved to Phase 7.3

### 2.2 Kapowarr / Stash → Phase 7.4
- [x] ~~Research tasks~~ → moved to Phase 7.4

---

## Phase 3: Frontend Unification ✅

### 3.1 Completed Fixes ✅
- [x] Bulk GET /api/instances endpoint (1 call instead of 6)
- [x] Instance store uses bulk endpoint with caching
- [x] Dashboard + Apps page use shared instance cache
- [x] 429 retry with jitter in api.ts
- [x] Rate limiter removed from authenticated routes
- [x] Dedup page double /api/api/ prefix fixed
- [x] Downloads page polling 5s → 15s
- [x] InstanceSwitcher redesigned with app-colored tabs
- [x] Toggle switch color now works
- [x] Queue sub-tabs unified with same visual style
- [x] Services grid → 3-column
- [x] Min → Max Download Queue Size rename (backend + frontend + migration 047)

### 3.2 Section Header Fixes ✅
Gold standard: `<h3 class="text-sm font-semibold text-foreground mb-3">`
- [x] **settings/+page.svelte** — 5 h2 headers → h3/text-sm/mb-3
- [x] **apps/+page.svelte** — 3 h2 headers → h3/text-sm/mb-3
- [x] **user/+page.svelte** — 5 h2 headers → h3/text-sm/mb-3
- [x] **monitoring/+page.svelte** — 3 h2 tags → h3 tags + mb-4→mb-3
- [x] **+page.svelte (dashboard)** — 4 h2 headers → h3/text-foreground
- [x] **dedup/+page.svelte** — selectedGroup.name h2 → h3

### 3.3 Queue Page Tab Component Fix ✅
- [x] Replaced custom inline-flex tab buttons with shared `<Tabs>` component
- [x] Fixed all section headers: text-xs/muted-foreground/uppercase → text-sm/text-foreground
- [x] "Global Blocklist Management" h2/text-base → h3/text-sm

### 3.4 Minor Font/Style Fixes ✅
- [x] **notifications/+page.svelte** — 3 modal headers font-medium → font-semibold
- [x] **seerr/+page.svelte** — duplicates header font-medium → font-semibold + mb-3

### 3.5 Missing Frontend Features ✅
- [x] State management UI — exposed GET /api/state + POST /api/state/reset on lurk settings page with per-instance reset buttons
- [x] Stats dashboard — GET /api/stats + GET /api/stats/hourly-caps on monitoring page with reset all button

---

## Phase 4: Code Quality & Testing

### 4.1 Dedup Helpers ✅
- [x] `decodeJSON[T]()`, `parseUUID()`, `validAppTypeParam()`, `filterCompleted()` — all extracted

### 4.2 Security ✅
- [x] Silent error swallowing fixed in auth.go + passkey.go (slog.Error added)
- [x] Slog secret audit — no secrets logged anywhere
- [x] Transaction isolation for HandleSetup — `SetupFirstUser()` composite method wraps CreateUser + UpsertGeneralSettings in a single pgx transaction

### 4.3 Unit Test Coverage ✅ (119/119 handlers tested — 100%)
Previous session added 38 new tests covering 12 of the 13 remaining untested handlers.
HandleFinishLogin now tested via WebAuthnProvider interface extraction + 7 new tests.

**All handlers tested:**
- [x] apps.go: HandleListAllInstances, HandleHealthCheckInstance, HandleTestConnection
- [x] auth.go: HandleSetupCheck
- [x] notifications.go: HandleGetNotificationHistory
- [x] queue.go: HandleGetStrikeLog, HandleGetDownloadClientSettings, HandleUpdateDownloadClientSettings
- [x] queue.go: HandleListSeedingRuleGroups, HandleCreateSeedingRuleGroup, HandleUpdateSeedingRuleGroup, HandleDeleteSeedingRuleGroup
- [x] seerr.go: HandleScanDuplicates
- [x] passkey.go: HandleFinishLogin (WebAuthnProvider interface + parse injection, 7 tests: parse error, no session, validation fail, cookie error, success, session rotation, sign count error)

### 4.4 Integration & E2E Tests → Phase 7.5
- [x] ~~Integration/E2E tests~~ → moved to Phase 7.5

### 4.5 OpenAPI Spec Verification ✅
- [x] Verify all 119 handlers have corresponding spec entries — 50 missing routes added, 2 orphaned removed
- [x] Verify request/response schemas match actual implementation — 17 new schemas, QueueCleanerSettings +34 fields, GeneralSettings/SeerrSettings/Enable2FAResponse fixed
- [x] Spec rewritten: 2503 → 4115 lines, 123/123 routes covered, all tests passing

---

## Phase 5: Deep Audit — Bug Fixes ✅

Comprehensive audit found 8 backend bugs and 10 frontend issues. All critical/high bugs fixed.

### 5.1 Backend Bug Fixes ✅
- [x] **CRITICAL: Dashboard seerr count crash** — Frontend `+page.svelte` read `res.count` but backend returns `res.total` → `TypeError`. Fixed: `res.count` → `res.total`.
- [x] **HIGH: Seerr cleanup pagination skip** — `cleanupFulfilledRequests` deleted while paginating with increasing `skip`, causing items to shift and be missed. Fixed: two-phase collect-then-delete approach.
- [x] **HIGH: MetadataStuckMinutes ignored** — Value was only used as boolean (>0 = enabled), actual minutes never compared. Fixed: added `Added time.Time` to `QueueRecord`, now compares `time.Since(record.Added) >= threshold`.
- [x] **HIGH: FindDuplicates "adequate" strategy fallback** — When no item met threshold, kept first item by queue order instead of highest-scored. Fixed: falls back to highest-score when no item meets `AdequateThreshold`.
- [x] **HIGH: HandleSetup non-atomic** — CreateUser + UpsertSettings + Session without transaction. Fixed: `SetupFirstUser()` composite method wraps user creation + settings upsert in single pgx transaction. Session cookie stays separate (non-critical failure — user can re-login).
- [x] **HIGH: Handle2FAEnable non-atomic** — SetTOTPSecret before SetRecoveryCodes. If recovery codes failed, user had TOTP active with no recovery codes. Fixed: save recovery codes first, then activate TOTP.
- [x] **MEDIUM: 50% progress `continue` skipped all detection** — Downloads >50% complete with "healthy" status skipped slow speed detection entirely. Fixed: removed `continue` so `detectProblem` still evaluates.
- [x] **MEDIUM: Webhook URL validation (SSRF)** — Discord/webhook/Gotify/Ntfy/Apprise notification URLs had no validation. Fixed: added `validateProviderURLs()` using existing `validateAPIURL()` (http/https only, no embedded credentials).

### 5.2 Frontend Type Mismatches ✅
- [x] `SeerrSettings` — added `cleanup_enabled`, `cleanup_after_days`
- [x] `GeneralSettings` — added `auto_import_interval_minutes`
- [x] `QueueCleanerSettings` — added 9 missing fields: `deletion_detection_enabled`, `unmonitored_cleanup_enabled`, `unregistered_enabled`, `max_strikes_unregistered`, `recheck_paused_enabled`, `recycle_bin_enabled`, `recycle_bin_path`, `ignored_release_groups`, `public_tracker_list`
- [x] `SABnzbdSettings` — added `id`, `timeout`, `category`

### 5.3 Test Updates ✅
- [x] Updated `TestHandle2FAEnable_Success` and `TestHandle2FAEnable_SaveError` for new recovery-codes-first ordering
- [x] Full test suite passing (all packages)
- [x] Frontend build passing

---

## Phase 6: Quality Hardening ✅

### Round 1
- [x] **Fix silently discarded DB errors** — 7 sites across prowlarr.go, apps.go, state.go, passkey.go, server.go now log or return errors instead of using `_ :=`
- [x] **Fix uTorrent unchecked errors** — `parseTorrent()` now returns errors for critical fields (Hash, Name); remaining fields use explicit `_ =` discard; retry `io.ReadAll` checked
- [x] **Fix notification nilerr** — improved comment in `validateProviderURLs` clarifying intentional `return nil` design
- [x] **Remove dead code** — removed `filterCompleted[T]` (helpers.go), `testUser` (testenv_test.go), `rtorrentSeedingTime` (rtorrent.go) + unused imports
- [x] **Fix import shadowing** — renamed `url` → `pageURL` in arrclient/client.go, `blocklist` → `addToBlocklist` in queuecleaner/cleaner.go
- [x] **Add missing nolint explanations** — `server.go` bare `//nolint:errcheck` now has `// best-effort copy`
- [x] **Fix escaped quotes** — corrected `\"` artifacts in server.go OIDC slog call

### Round 2
- [x] **Fix uTorrent fetchToken io.ReadAll** — now properly returns error instead of falling through to "token not found"
- [x] **Fix OIDC settings seed error** — `UpdateOIDCSettings` failure now logged with `slog.Error` instead of silently discarded
- [x] **Activity feed error logging** — all 7 data sources (lurk history, cross-instance, blocklist, auto-import, schedule, strike, notification) now log warnings on query failure instead of silently skipping

### Round 3
- [x] **Fix json.Marshal error discards** — CreateTag, TagMedia (arrclient/client.go), ReadarrSearchBook, WhisparrSearchEpisode, ErosSearchMovie now return marshal errors
- [x] **Fix arrclient/seerr io.ReadAll in error paths** — both doRequest (arrclient) and doAction (seerr) now check io.ReadAll errors and include them in error messages
- [x] **Fix sabnzbd ParseFloat discard** — `strconv.ParseFloat` failure now logged with slog.Warn including nzo_id and raw value
- [x] **Fix rtorrent IsActive discard** — `IsActive` error now logged with slog.Warn including torrent hash

### Round 4
- [x] **Fix JSON Content-Type on error responses** — 23 sites across auth (oidc.go, middleware.go), middleware (middleware.go, ratelimit.go), and server.go now send `application/json` Content-Type with JSON error bodies. Added `jsonError()` helper in auth package.
- [x] **Fix test error discards** — 5 `io.ReadAll` + `json.Unmarshal` discards in notification test mock handlers now properly checked with `t.Fatal`
- [x] All tests passing (api, auth, notifications), dev image rebuilt and healthy

---

## Phase 7: Feature Completion & Gaps

### 7.1 Dedup / Multi-Arr — Frontend Group Management UI ✅
Backend is 100% complete (internal/crossarr, instance_groups API, database methods, scanner).
Frontend dedup page exists but was **unusable** — no UI to create or manage instance groups. Now fixed.
- [x] Add "Instance Groups" section to `/apps` (Connections) page
  - List groups per app type (Sonarr, Radarr)
  - Create group form: name, mode (quality_hierarchy / overlap_detect / split_season)
  - Edit/Delete group buttons
- [x] Create member management modal/section
  - Select instances to add to group
  - Set quality_rank per member
  - Toggle is_independent flag
  - Save → PUT /api/instance-groups/by-id/{id}/members
- [x] Fix dedup page "Go to Connections" CTA (now includes help text explaining how to create groups)
- [x] Add "Scan for Overlaps" button to dedup page (calls scanner) — already existed
- [x] Seed test instance groups in `dev/seed.sh` (Radarr + Radarr 4K group, Sonarr + Sonarr Anime group)

### 7.2 Seed Data Gaps ✅
Dev seed script enhanced with comprehensive test data for all features.
- [x] Add Sonarr overlap data: Attack on Titan added to both Sonarr instances for dedup testing
- [x] Seed queue cleaner settings for sonarr/radarr (dry-run mode, stalled/slow/seeding/orphan detection)
- [x] Seed lurk app settings for sonarr/radarr/lidarr/readarr (API limits, sync intervals)
- [x] Add Dev Webhook notification channel → httpbin.org with 4 event types
- [x] Document Seerr manual setup steps in seed.sh summary section
- [x] Updated seed summary with all new categories

### 7.3 Bazarr Integration ✅
Bazarr integrated as a Service (like Prowlarr/Seerr), not an AppType — it manages subtitles, not media.

**Research & Design:**
- [x] Created docs/research/bazarr.md — API docs, auth (config.ini apikey), integration type decision

**Backend:**
- [x] Database migration 048_bazarr_settings.sql (singleton settings table: url, api_key, enabled, timeout)
- [x] BazarrSettings model + MaskedAPIKey() in models.go
- [x] queries_bazarr.go — GetBazarrSettings() / UpdateBazarrSettings()
- [x] bazarrclient package — Client struct with TestConnection, GetHealth, GetWantedEpisodes/Movies, GetEpisodeHistory/MovieHistory
- [x] BazarrHandler (bazarr.go) — HandleTestConnection, HandleGetWanted, HandleGetHealth, HandleGetHistory
- [x] BazarrSettings handlers (bazarr_settings.go) — HandleGetSettings (masked key), HandleUpdateSettings
- [x] Store interface + mock updated with Bazarr methods
- [x] 6 routes registered in server.go (settings GET/PUT, test POST, wanted/health/history GET)
- [x] OpenAPI spec updated with BazarrSettings schema + 6 endpoints

**Frontend:**
- [x] BazarrSettings type added to types.ts
- [x] Bazarr logo downloaded (128x128 PNG from official repo)
- [x] index.ts mappings: logo, accentBorder (green-400), bgColor (green-500), hoverBg, website, Tailwind safelist
- [x] Apps page: state vars, loadServices() fetch, saveBazarr/testBazarr functions
- [x] Apps page: Bazarr items in both dropdown menus (header + empty state)
- [x] Apps page: Bazarr service card with health badge + "Subtitle manager" label
- [x] Apps page: Bazarr settings modal (toggle, URL, API key, timeout, save/test)

**Seed:**
- [x] Bazarr API key extraction from config.ini in seed.sh
- [x] Bazarr auto-registration via PUT /api/bazarr/settings

**Tests:** All 119 Go tests pass. Frontend build clean.

### 7.4 Research — Future Integrations
- [ ] Kapowarr API research (comics/manga — evaluate arr-compatibility)
- [ ] Stash GraphQL API research (adult content manager)

### 7.5 Integration & E2E Tests
- [ ] Go integration tests with real dev stack (API round-trips using httptest + live DB)
- [ ] Frontend E2E tests (Playwright against dev stack)

---

## Phase 8: Frontend Design & UX Polish

### 8.1 Design Unification ✅
All raw HTML interactive elements replaced with shadcn wrapper components across every route.
- [x] Audit all pages for raw `<button>`, `<input>`, `<select>` — zero remaining in routes
- [x] Fix `ConfirmAction.svelte` — now uses AlertDialog primitives (modal confirmation with destructive button)
- [x] Fix HealthBadge / ConnectionCard — uses `Badge.svelte` + semantic colors
- [x] All forms use `Input`, `Select`, `Toggle` wrappers consistently
- [x] All pages follow consistent spacing, card usage, section header patterns

### 8.2 shadcn Component Cleanup ✅
Full shadcn primitive adoption across the frontend.
- [x] Audit `frontend/src/lib/components/ui/` — all components use shadcn primitives
- [x] All interactive elements use shadcn-based wrappers
- [x] All modals use `Modal.svelte` (shadcn Dialog wrapper)
- [x] All hardcoded colors replaced with CSS variables (`text-foreground`, `bg-muted`, `text-destructive`, etc.)

### 8.5 shadcn Deep Integration ✅
Installed and integrated all useful shadcn-svelte primitives. Two rounds of installation — initial 7, then 10 more after comprehensive audit.

**Round 1 (7 components):**
- [x] **Alert** — `variant="destructive"` for offline/error banners, custom `variant="warning"` for yellow callouts (dry-run, recovery codes). Applied to dashboard, login, queue, user pages.
- [x] **Alert Dialog** — Rewrote `ConfirmAction.svelte` to use AlertDialog primitives. All ~15 confirm actions (dashboard, queue, user, history) now get proper accessible modal confirmations.
- [x] **Tooltip** — Sidebar nav items, sign out, footer links show tooltips when sidebar collapsed.
- [x] **Progress** — Both arr-client and SABnzbd download progress bars use `<Progress>` with success color on completion.
- [x] **Avatar** — Initials avatar in admin users table + sidebar footer with username link to profile.
- [x] **Scroll Area** — Scheduling history modal, Prowlarr indexers list.
- [x] **Dropdown Menu** — Apps page "Add Connection" dropdown (header + empty state) uses shadcn DropdownMenu.

**Round 2 (10 components — comprehensive audit):**
- [x] **Textarea** — notifications page body template field → shadcn `<Textarea>`
- [x] **Label** — Input.svelte, Select.svelte, notifications page → shadcn `<Label>`
- [x] **Collapsible** — queue page (per-reason blocklist overrides) + user page (manual TOTP entry) → `Collapsible.Root/Trigger/Content`
- [x] **Sonner** — Toast.svelte now uses shadcn sonner wrapper (auto theme via mode-watcher)
- [x] **Pagination** — DataTable.svelte → full numbered pagination with ellipsis, prev/next
- [x] **Toggle Group** — scheduling page day-of-week picker → `ToggleGroup.Root type="multiple"`
- [x] **Popover** — HealthBadge.svelte → click badge to see Status + Version info
- [x] **Sidebar** — complete sidebar rewrite using shadcn Sidebar primitives (S.Root, S.Header, S.Content, S.Group, S.Menu, S.MenuButton, S.Footer, S.Rail). Mobile handled internally by built-in Sheet.
- [x] **Command** — command palette (Cmd+K / Ctrl+K) for quick navigation to all pages
- [x] **Toggle** — installed as dependency for toggle-group

**Installed shadcn primitives (32 total):**
accordion, alert, alert-dialog, avatar, badge, breadcrumb, button, card, checkbox, collapsible, command, dialog, dropdown-menu, input, label, pagination, popover, progress, scroll-area, select, separator, sheet, sidebar, skeleton, sonner, switch, table, tabs, textarea, toggle, toggle-group, tooltip

### 8.3 Dashboard Improvements ✅
Dashboard enhanced with at-a-glance utility.
- [x] Quick-action navigation buttons below header (Lurk Settings, Queue Cleaner, Scheduling)
- [x] Hourly API cap progress bars with color coding (≥90% red, ≥70% yellow) using shadcn Progress
- [x] Recent activity feed — last 5 events with source icons, badges, timestamps, "View All" link
- [x] App settings loaded per app type for cap limit display
- [x] Polling refreshes activity alongside stats every 30s
- [x] Removed unused appTabLabel import

### 8.4 Large Config Page Structure ✅
Config-heavy pages restructured with collapsible sections and scroll-to-top for mobile usability.
- [x] Created `CollapsibleCard.svelte` — reusable Card wrapper with Collapsible header (ChevronRight rotates on open), defaults open=true
- [x] Created `ScrollToTop.svelte` — fixed bottom-right button, appears after 400px scroll, smooth scroll to top
- [x] Queue page: all cleaner sections (Stall Detection, Strike System, Actions, Failed Imports, Metadata Mismatch, Seeding Rules, Orphan Cleanup, Advanced) + scoring sections (Preferences, Weights) converted to CollapsibleCard
- [x] Settings page: General tab sections (Lurking Behaviour, API & Command Execution, Security) + SSO tab sections (OpenID Connect Provider, User Management) converted to CollapsibleCard
- [x] ScrollToTop added to 10 long pages: queue, settings, lurk, apps, notifications, downloads, scheduling, user, monitoring, history

### 8.6 Breadcrumb Navigation ✅
Route-aware breadcrumb navigation integrated into root layout.
- [x] Install shadcn Breadcrumb component
- [x] Add Breadcrumbs.svelte wrapper with route-aware auto-generation (capitalizes path segments, handles separators)
- [x] Integrate into root layout (inside S.Inset, above page content)
- [x] Works on all screen sizes — hidden on mobile header, visible in main content

### 8.7 In-App Help / FAQ System ✅
Comprehensive searchable Help/FAQ page covering all features.
- [x] Dedicated `/help` route with 12 accordion sections (Getting Started, Connections, Lurk Settings, Queue Cleaner, Scheduling, Downloads, Seerr, Dedup, Notifications, Security, Monitoring, General Settings)
- [x] 40+ FAQ items with detailed answers
- [x] Real-time search filter across questions and answers
- [x] Per-section icons and FAQ counts via Badge
- [x] Help link in sidebar nav
- [x] Uses shadcn Accordion, Input (search), Card, Badge components

### 8.8 Command Palette ✅
Cmd+K / Ctrl+K command palette for quick page navigation.
- [x] CommandPalette.svelte using shadcn Command Dialog component
- [x] All 16 sidebar nav items indexed with searchable keywords
- [x] Grouped by category (Pages, Configuration, Operations, Monitoring, Admin)
- [x] Admin-only items filtered based on user role
- [x] Integrated into root layout

---

## Current Sprint Focus
> Phase 1–6 complete. 119/119 unit tests. All backend features implemented.
> Phase 7.1 complete — dedup fully usable with instance group management on Connections page.
> Phase 7.2, 7.3 complete — seed data comprehensive, Bazarr fully integrated (backend + frontend + seed).
> Phase 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8 complete — full shadcn design system (32 primitives), breadcrumbs, dashboard improvements, collapsible config pages, help/FAQ, command palette.
> Priority: 7.4–7.5
