# Lurkarr Codebase Analysis & Improvement Plan

> Generated: 2026-03-14 | Branch: develop
> Reference projects: Huntarr (archived), Newtarr (v6.6.3 fork), Cleanuparr (v2.8.1)

---

## Table of Contents

1. [What Lurkarr Does Today](#1-what-lurkarr-does-today)
2. [Reference Project Comparison](#2-reference-project-comparison)
3. [Module-by-Module Logic Analysis](#3-module-by-module-logic-analysis)
4. [Bugs & Issues](#4-bugs--issues)
5. [Where Lurkarr Already Wins](#5-where-lurkarr-already-wins)
6. [Where Lurkarr Falls Short](#6-where-lurkarr-falls-short) + Deep Competitor Analysis (6.1-6.10)
7. [Improvement Plan](#7-improvement-plan) (original)
8. [Web Research: Additional Competitor Analysis](#6-web-research-additional-competitor-analysis) — Decluttarr, Swaparr, Recyclarr, Autobrr
9. [Web Research: Multi-Arr Management Design](#7-web-research-multi-arr-management-design)
10. [Web Research: UI/UX Design](#8-web-research-uiux-design-best-practices)
11. [Revised Improvement Plan](#9-revised-improvement-plan-post-research) — Phases 0-7 with status tracking

---

## 1. What Lurkarr Does Today

Lurkarr is a unified media automation service that combines the functionality of multiple separate tools into one Go binary with a PostgreSQL backend and SvelteKit frontend.

### Core Services (5 background goroutines)

| Service | Purpose | Cycle |
|---------|---------|-------|
| **Lurking Engine** | Searches *arr apps for missing media + quality upgrades | Configurable (default 15 min) |
| **Queue Cleaner** | Detects and removes problem downloads (stalled, slow, failed, dupes) | Configurable per app |
| **Auto-Importer** | Detects stuck "importPending" items, triggers rescans | Fixed 5 min |
| **Seerr Sync** | Polls Seerr for pending requests, optionally auto-approves | Configurable (default 30 min) |
| **Health Poller** | Monitors *arr instance health, exposes Prometheus metrics | 60 sec |

### Supported Apps

- Sonarr, Radarr, Lidarr, Readarr, Whisparr v2, Whisparr v3 (eros)
- Download clients: qBittorrent, Transmission, Deluge, SABnzbd, NZBGet
- Notifications: Discord, Telegram, Pushover, Gotify, Ntfy, Apprise, Email, Webhook
- Integrations: Prowlarr (indexer stats), Seerr (request management)

### Architecture

- **Language:** Go with Uber Fx dependency injection
- **Database:** PostgreSQL (pgx/v5, goose migrations)
- **Frontend:** SvelteKit 5 + Tailwind CSS (embedded in binary, Brotli pre-compressed)
- **Auth:** Local (bcrypt), TOTP, WebAuthn/Passkeys, OIDC, Proxy auth
- **Observability:** Prometheus metrics, structured JSON logging, Grafana dashboards

---

## 2. Reference Project Comparison

### Huntarr (Python, archived — significant security vulnerabilities found)

**What it did:**
- Lurking for missing content + quality upgrades across *arr apps
- Multi-instance support, hourly API caps, scheduler
- v7+ added Requestarr (media request system), Prowlarr integration
- v9+ added built-in NZB/torrent clients, Movie Hunt/TV Hunt (full *arr replacement)
- Flask-based web UI, file-based config (JSON on disk)

**Why it died:** Security vulnerabilities (exposed API keys, hardcoded secrets, no CSRF, weak password hashing), telemetry controversies, author rage-quit.

### Newtarr (Python, fork of Huntarr v6.6.3)

**What it is:**
- Minimal fork: just lurking (missing + upgrades) + multi-instance + hourly caps
- No queue cleaning, no blocklist, no notifications, no download client integration
- Python/Flask, file-based config, NO database
- Designed for SSO-proxied deployments (ElfHosted)

### Cleanuparr (C#, 2.1k stars, actively maintained)

**What it does — features Lurkarr must match or beat:**
- Strike system for stalled/slow/failed downloads
- Blocklist/whitelist management (community lists + custom rules)
- Known malware pattern detection (community-maintained)
- Trigger re-search after removal (downloads replacement content automatically)
- Seeding enforcement (ratio + time limits)
- Orphan/no-hardlink detection and cleanup
- Cross-seed awareness
- Ignore rules: torrent hashes, categories, tags, trackers
- Per-provider notifications on strike/removal
- Supports: qBittorrent, Transmission, Deluge, uTorrent, rTorrent
- OIDC authentication (just added)
- Detailed UI with per-item strike history, removal logs

---

### Feature Matrix: Lurkarr vs References

| Feature | Huntarr | Newtarr | Cleanuparr | **Lurkarr** | Notes |
|---------|---------|---------|------------|-------------|-------|
| **LURKING** | | | | | |
| Missing content search | ✅ | ✅ | ❌ | ✅ | Core feature |
| Quality upgrade search | ✅ | ✅ | ❌ | ✅ | Core feature |
| Multi-instance per app | ✅ | ✅ | ✅ | ✅ | |
| Hourly API cap | ✅ | ✅ | ❌ | ✅ | |
| Smart prioritization | ❌ | ❌ | ❌ | ✅ | ✅ Selection modes (P3.1) |
| Min queue size check | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Processed item tracking | ✅ (file) | ✅ (file) | ❌ | ✅ (DB) | Lurkarr's is better (persistent DB) |
| State reset window | ✅ | ✅ | ❌ | ✅ | |
| **QUEUE CLEANING** | | | | | |
| Strike system | ❌ | ❌ | ✅ | ✅ | |
| Stalled detection | ❌ | ❌ | ✅ | ✅ | |
| Slow download detection | ❌ | ❌ | ✅ | ✅ | |
| Metadata stuck detection | ❌ | ❌ | ✅ | ✅ | |
| Failed import cleanup | ❌ | ❌ | ✅ | ✅ | |
| Duplicate detection (scoring) | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Private/public tracker awareness | ❌ | ❌ | ✅ | ✅ | |
| SABnzbd status integration | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Re-search after removal | ❌ | ❌ | ✅ | ✅ | ✅ P2.1 + P3.3 |
| Known malware detection | ❌ | ❌ | ✅ | ❌ | **Gap — community-maintained patterns** |
| Ignore rules (hash/category/tag/tracker) | ❌ | ❌ | ✅ | ✅ | ✅ P2.2 |
| Intra-queue deduplication (quality) | ❌ | ❌ | ❌ | ❌ | **Gap — Deduparr has this** |
| Pack-aware operations | ❌ | ❌ | ❌ | ❌ | **Gap — Deduparr has this** |
| Per-item strike history in UI | ❌ | ❌ | ✅ | ❌ | **Gap** |
| **SEEDING** | | | | | |
| Ratio enforcement | ❌ | ❌ | ✅ | ✅ | |
| Time-based enforcement | ❌ | ❌ | ✅ | ✅ | |
| AND/OR mode | ❌ | ❌ | ✅ | ✅ | |
| Cross-seed awareness | ❌ | ❌ | ✅ | ✅ | |
| Hardlink protection | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| **ORPHAN CLEANUP** | | | | | |
| Orphan detection | ❌ | ❌ | ✅ | ✅ | |
| Grace period | ❌ | ❌ | ✅ | ✅ | |
| Cross-arr sync | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Dry-run mode | ❌ | ❌ | ✅ | ✅ | ✅ P0.1 |
| **BLOCKLIST** | | | | | |
| Community source sync | ❌ | ❌ | ✅ | ✅ | |
| ETag support | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Regex patterns | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Title contains | ❌ | ❌ | ✅ | ✅ | |
| Release group | ❌ | ❌ | ✅ | ✅ | |
| Indexer rules | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Whitelist (allow-list) | ❌ | ❌ | ✅ | ❌ | **Gap** |
| Health endpoints | ❌ | ❌ | ❌ | ❌ | **Gap — Deduparr+Swaparr have** |
| RecycleBin mode | ❌ | ❌ | ❌ | ❌ | **Gap — qbit_manage has** |
| Unregistered torrent detection | ❌ | ❌ | ❌ | ❌ | **Gap — qbit_manage has** |
| **SEERR / REQUESTS** | | | | | |
| Request monitoring | ✅ (Requestarr) | ❌ | ❌ | ✅ (basic) | |
| Auto-approve | ✅ | ❌ | ❌ | ✅ (blind) | |
| Multi-arr instance routing | ✅ | ❌ | ❌ | ❌ | **Gap** |
| Duplicate request cleanup | ❌ | ❌ | ❌ | ❌ | **Gap — user requested** |
| Instance/quality reassignment | ❌ | ❌ | ❌ | ❌ | **Gap — user requested** |
| **INFRASTRUCTURE** | | | | | |
| Language | Python | Python | C# | **Go** | Lurkarr: fastest, smallest binary |
| Database | SQLite/JSON files | JSON files | SQLite | **PostgreSQL** | Lurkarr: most robust |
| Auth: Local | ✅ (weak) | ❌ | ✅ | ✅ (bcrypt) | |
| Auth: Passkeys/WebAuthn | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Auth: TOTP (2FA) | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Auth: OIDC | ❌ (Plex SSO only) | ❌ | ✅ (new) | ✅ | |
| CSRF protection | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Prometheus metrics | ❌ | ❌ | ❌ | ✅ | Deduparr also has Prometheus |
| API docs (OpenAPI/Scalar) | ❌ | ❌ | ❌ | ✅ | **Lurkarr unique** |
| Scheduling (cron) | ✅ | ✅ | ❌ | ✅ | |
| Download client: rTorrent | ❌ | ❌ | ✅ | ❌ | **Gap** |
| Download client: uTorrent | ❌ | ❌ | ✅ | ❌ | **Gap** |
| Notifications | ✅ | ❌ | ✅ | ✅ | Feature parity |

---

## 3. Module-by-Module Logic Analysis

### 3.1 Lurking Engine

**Current algorithm:**
```
Loop forever per app type:
  For each instance:
    1. Check hourly cap → skip if reached
    2. Check state reset window → clear processed items if expired
    3. Check min queue size → skip if queue full
    4. Get missing items from *arr API (page size 1000)
    5. Filter out already-processed items
    6. Random/sequential select N items
    7. Trigger search command per item
    8. Mark as processed in DB
    9. Increment hourly hit counter
  Sleep configurable seconds
  On error: exponential backoff (2^N sec, max 5 min)
```

**Problems:**
- **P1: No prioritization** — Pure random selection. No consideration of: how long content has been missing, series/movie rating, number of missing episodes in series, release date proximity, previous search failure count. A newly-added movie gets the same chance as something missing for 2 years.
- **P2: Blunt state reset** — All processed items cleared at once after StatefulResetHours. Should be per-item TTL instead.
- **P3: Per-app-type backoff** — One failing Sonarr instance causes ALL Sonarr instances to back off.
- **P4: Page size 1000 hardcoded** — Libraries with >1000 missing items only search from first page.
- **P5: No search failure tracking** — If a search for item X consistently finds nothing, it's re-searched with same priority as items that might have results.
- **P6: No re-search after queue cleaner removal** — When queue cleaner removes a bad download, the lurking engine doesn't know to re-trigger a search for that media.

### 3.2 Queue Cleaner

**Current detection chain:**
```
1. SABnzbd override (if usenet: check SAB status, skip Queued/Grabbing)
2. Metadata stuck (Size==0 && Status in downloading/delay)
3. Stalled (Status==warning, private/public aware)
4. Slow (speed < threshold, with large-download exemption)
→ Strike on detection, remove at max strikes
5. Failed import (keyword matching on status messages)
6. Seeding enforcement (ratio/time limits)
7. Orphan detection (not in any *arr queue)
```

**Problems:**
- **P7: No re-search after removal** — Cleanuparr's killer feature. When a stalled download is removed + blocklisted, Cleanuparr triggers a new search so the *arr finds a different release. Lurkarr just removes and walks away.
- **P8: No ignore rules** — No way to exclude specific torrent hashes, categories, tags, or trackers from being processed. Cleanuparr has this.
- **P9: No whitelist/allowlist** — Only blocklist rules exist. Can't say "always keep releases from this group."
- **P10: Hardcoded public tracker list** — 18 indexer names that will rot. Should be configurable or use API flags only.
- **P11: Cross-arr sync by title only** — Title string matching is fragile. Should use download hash or release name normalization.
- **P12: No dry-run mode** — Can't preview what would be removed without actually removing.
- **P13: Speed estimation uses arr's `timeleft` guess** — arr's timeleft calculation can be wildly inaccurate. Should use download client's actual speed data instead.
- **P14: Strike add + count not in transaction** — Race condition between AddStrike() and CountStrikes().
- **P15: No per-item strike history in UI** — Users can't see why something was struck/removed.

### 3.3 Auto-Importer

**Current algorithm:**
```
Every 5 minutes:
  For each instance:
    Get queue items with TrackedDownloadState == "importPending"
    Filter by status message keywords ("unable to import", "import failed", etc.)
    Try to get manual import options → LOG that it's available (NO ACTION)
    Fallback: trigger rescan (re-search the series/movie)
```

**Problems:**
- **P16: Logs manual import availability but doesn't act** — The code checks if manual import is possible, logs it, then does nothing. Should actually trigger the manual import.
- **P17: Fixed 5-min interval** — Not configurable. No backoff on errors.
- **P18: Rescan is a sledgehammer** — Rescanning the entire series/movie is overkill when only one episode/file is stuck.
- **P19: No distinction between "file exists but can't import" vs "no file at all"** — Different problems need different solutions.

### 3.4 Seerr Integration

**Current implementation: basic poll + blind auto-approve.**

**What it should do (per user requirements):**
- **Multi-arr instance management** — When same media exists across multiple *arr instances (e.g., 1080p Radarr + 4K Radarr), help manage/route requests to the right one.
- **Duplicate request cleanup** — Detect when the same media is requested to multiple instances unnecessarily.
- **Instance/quality reassignment** — Move a request from one instance/quality profile to another.
- **Capacity-aware approval** — Don't approve if the target instance is overloaded or has no disk space.

**Current implementation is fundamentally wrong for these goals.** The entire Seerr module needs a redesign.

### 3.5 Blocklist

**Current implementation:**
- Pattern types: release_group, title_contains, title_regex, indexer
- Community source sync with ETag support
- First-match-wins evaluation

**Problems:**
- **P20: No regex complexity/timeout limits** — User-supplied regex can cause catastrophic backtracking (DoS).
- **P21: Regex compiled at matcher creation, no cache invalidation** — If rules change, old compiled regex still used until next cleaner cycle.
- **P22: No whitelist/allowlist support** — Can't exempt known-good releases from being matched.
- **P23: Community source sync deletes all rules from source, then re-inserts** — Not atomic. If crash between delete and insert, rules are lost.
- **P24: No malware pattern detection** — Cleanuparr has community-maintained malware patterns (*.lnk, *.zipx detection). Lurkarr doesn't.

### 3.6 Notifications

**Current implementation is solid.** 8 providers, per-event filtering, concurrent delivery.

**Problems:**
- **P25: Minor race condition** — Provider list copied under lock, but if unregistered between copy and send, could hit stale pointer.
- **P26: No notification history/log in UI** — Users can't see past notifications.
- **P27: No notification templates** — Message format is hardcoded per provider.

### 3.7 Download Clients

**Current implementation:** Unified interface for 5 client types.

**Problems:**
- **P28: Missing rTorrent support** — Cleanuparr has it, common in the *arr community.
- **P29: Missing uTorrent support** — Same as above.
- **P30: No speed trend analysis** — Fetches point-in-time speed only. Rolling average would give much better stalled/slow detection.
- **P31: No category/tag filtering** — Can't tell "only manage downloads in category X."

### 3.8 Authentication & Security

**Lurkarr is already the strongest here.** Passkeys, TOTP, OIDC, CSRF — none of the reference projects have all of these.

**Remaining issues:**
- **P32: CSRF key not persisted** — Random key on restart = all sessions invalidated.
- **P33: No account lockout** — Mitigated by rate limiting, but proper lockout is better.
- **P34: Plaintext HTTP CSRF bypass** — `PlaintextHTTPRequest()` disables Referer/Origin validation.

---

## 4. Bugs & Issues

### Critical (must fix)

| # | Bug | Location | Impact |
|---|-----|----------|--------|
| B1 | CSRF key not persisted — all sessions lost on restart | `cmd/lurkarr/providers.go` | Users forced to re-login after every restart |
| B2 | Strike add + count not in DB transaction | `internal/queuecleaner/cleaner.go` | Race condition on strike count, items removed at wrong count |
| B3 | Regex patterns have no complexity limit or timeout | `internal/blocklist/matcher.go` | DoS vector — malicious regex blocks queue cleaner |

### High (should fix)

| # | Bug | Location | Impact |
|---|-----|----------|--------|
| B4 | Public tracker list hardcoded (18 names, will rot) | `internal/queuecleaner/cleaner.go` | New public trackers treated as private → wrong strike rules |
| B5 | Cross-arr sync matches by title string only | `internal/queuecleaner/cleaner.go` | Different formatting = missed sync |
| B6 | Auto-importer logs manual import option but doesn't trigger it | `internal/autoimport/importer.go` | Stuck imports stay stuck forever |
| B7 | Per-app-type backoff (one bad instance backs off all instances of that type) | `internal/lurking/engine.go` | One flaky Sonarr instance stops all Sonarr lurking |
| B8 | Seerr does blind auto-approve with no capacity check | `internal/seerr/sync.go` | Could overwhelm instances with approved requests |
| B9 | Blocklist sync deletes then re-inserts (not atomic) | `internal/blocklist/syncer.go` | Crash = rules lost until next sync |

### Medium

| # | Bug | Location | Impact |
|---|-----|----------|--------|
| B10 | Page size 1000 hardcoded for *arr API calls | `internal/arrclient/` | Libraries with >1000 missing items don't get full coverage |
| B11 | Speed estimation uses arr's timeleft instead of client data | `internal/queuecleaner/cleaner.go` | Inaccurate speed detection |
| B12 | Auto-importer fixed 5-min interval, not configurable | `internal/autoimport/importer.go` | Can't tune for different setups |
| B13 | Notification race on provider unregister | `internal/notifications/notifications.go` | Unlikely panic on precise timing |
| B14 | No DB query timeouts | All DB calls | Stuck queries hang indefinitely |
| B15 | Orphan with AddedAt=0 and no CompletedAt skips grace period | `internal/queuecleaner/cleaner.go` | Premature deletion of new downloads |

---

## 5. Where Lurkarr Already Wins

These are areas where Lurkarr is already **objectively better** than all reference projects:

1. **Security** — Passkeys + TOTP + OIDC + CSRF + bcrypt. Huntarr had hardcoded secrets and exposed API keys. Cleanuparr only recently added OIDC. Neither has 2FA or passkeys.

2. **Infrastructure** — Go binary + PostgreSQL is faster, uses less memory, and is more reliable than Python+JSON files or C#+SQLite. Single Docker container with embeddedrontend.

3. **Observability** — Prometheus metrics, structured JSON logging, Grafana dashboards. None of the others have this.

4. **API design** — OpenAPI spec, Scalar docs, consistent REST patterns. Full API-first design.

5. **Queue deduplication scoring** — Neither Huntarr nor Cleanuparr can detect and resolve duplicate downloads by quality scoring. This is Lurkarr-unique.

6. **SABnzbd status integration** — Lurkarr checks SABnzbd's actual queue status before striking, preventing false positives. Cleanuparr doesn't do this.

7. **Cross-arr sync** — Propagating removals across sibling instances is unique to Lurkarr.

8. **Hardlink protection** — Detecting hardlinks before deleting files prevents breaking other tools' references.

9. **ETag-based blocklist sync** — Efficient, avoids re-downloading unchanged community lists.

10. **Indexer-based blocklist rules** — Can block all releases from a specific indexer.

11. **Scheduling system** — Built-in cron scheduler for enable/disable/API-cap actions. Cleanuparr doesn't have this.

12. **Lurking + cleaning in one service** — Huntarr said "use Cleanuparr alongside Huntarr." Lurkarr does both, plus more, in one container.

---

## 6. Where Lurkarr Falls Short

These are the gaps that make Lurkarr look "too much like" the references instead of being clearly superior:

### 6.1 Missing "Killer Features" from Cleanuparr

1. **Re-search after removal** — When queue cleaner removes a bad download and blocklists it, Cleanuparr automatically triggers a new search in the *arr app so it finds a replacement. Lurkarr just removes and leaves the media unmoved. This is Cleanuparr's #1 feature and Lurkarr doesn't have it.

2. **Ignore rules** — Ability to exclude specific torrent hashes, categories, tags, or tracker URLs from being processed by the queue cleaner. Critical for cross-seed setups and manual downloads.

3. **Whitelist/allowlist** — Opposite of blocklist. "Always keep releases from group X" or "Never touch downloads from tracker Y."

4. **Dry-run mode** — Preview what would be cleaned/removed without actually doing it. Essential for building user trust.

5. **Per-item strike history in UI** — See exactly why each item was struck, when, how many strikes, and what happened.

6. **Malware pattern detection** — Community-maintained patterns for known-malicious files (*.lnk, *.zipx, etc.) that get into *arr queues.

### 6.2 Missing Intelligence (not in any reference — opportunity to be BETTER)

7. **Smart search prioritization** — Instead of random selection, weight by: days missing, series/movie popularity (TMDB rating), number of missing episodes in series (prefer completing a season), release date proximity (new releases more likely to have results), previous search failure count (deprioritize items that never find results).

8. **Search failure tracking** — Track how many times a search for item X returned nothing. Deprioritize chronic failures, surface them in UI as "hard to find."

9. **Adaptive strike thresholds** — Instead of fixed max_strikes for all items, adjust based on: private vs public tracker, download age, number of available alternatives, whether manual import is possible.

10. **Speed trend analysis** — Instead of point-in-time speed checks, track rolling average over multiple cycles. Detect "slowly dying" downloads that are technically above threshold but trending to zero.

11. **Download queue capacity awareness** — Don't trigger searches when download queue is at capacity or disk is running low. The min-queue-size check exists but doesn't check disk space.

12. **Related media grouping** — When lurking, prefer to search for remaining episodes of a season that's already 80% complete vs a random episode from a series that's 10% complete.

### 6.3 Seerr Integration is Fundamentally Wrong

The current implementation (poll + auto-approve) doesn't match the user's intended use case:

13. **Multi-arr instance management** — When the same media exists across multiple *arr instances (e.g., 1080p Radarr + 4K Radarr), Lurkarr should understand these relationships and help manage requests across them.

14. **Duplicate request cleanup** — If the same movie is requested to both the 1080p and 4K Radarr instances but only needs to be in one, detect and clean up duplicates.

15. **Instance/quality reassignment** — Move a request from one *arr instance to another, or change the quality profile, from the Lurkarr UI.

16. **Capacity-aware approval** — Check instance health, disk space, queue depth before approving requests.

### 6.4 Download Client Gaps

17. **rTorrent support** — Common in the selfhosted community, Cleanuparr has it.
18. **uTorrent support** — Same.
19. **Category/tag filtering** — "Only manage downloads in category `tv-sonarr`."

### 6.5 UI/UX Gaps

20. **No notification history** — Can't see what notifications were sent.
21. **No strike/removal log browser in UI** — Data is in DB but no page to view it.
22. **No "why was this removed" explanation per item** — Users need to understand and trust the automation.

---

## 7. Improvement Plan

### Phase 1: Fix Critical Bugs (stability + correctness)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 1.1 | Persist CSRF key to DB (generate once, store, reuse on restart) | Critical | S |
| 1.2 | Wrap strike add+count+delete in DB transaction | Critical | S |
| 1.3 | Add regex complexity limit + match timeout for blocklist rules | Critical | M |
| 1.4 | Per-instance backoff instead of per-app-type | High | M |
| 1.5 | Make public tracker list configurable (DB-stored, editable via UI) | High | M |
| 1.6 | Fix blocklist sync atomicity (transaction: delete + re-insert) | High | S |
| 1.7 | Add DB query context timeouts | Medium | S |
| 1.8 | Pagination for *arr API calls (handle >1000 items) | Medium | M |

### Phase 2: Match Cleanuparr's Best Features (parity + surpass)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 2.1 | **Re-search after removal** — trigger *arr search when queue cleaner removes+blocklists a download | Critical | M |
| 2.2 | **Ignore rules** — configurable exclusion by torrent hash, category, tag, tracker | High | M |
| 2.3 | **Whitelist/allowlist** — opposite of blocklist, "never touch these" | High | M |
| 2.4 | **Dry-run mode** — preview cleaner actions without executing | High | M |
| 2.5 | **Per-item strike history UI** — show strike log per queue item in frontend | High | M |
| 2.6 | **Malware pattern detection** — community-sourced file extension + name patterns | Medium | M |

### Phase 3: Intelligence Features (be genuinely better)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 3.1 | **Smart search prioritization** — weight by days-missing, popularity, season completion %, release date | High | L |
| 3.2 | **Search failure tracking** — track failed searches per item, deprioritize chronically unavailable content | High | M |
| 3.3 | **Adaptive strike thresholds** — adjust max strikes based on tracker type, download age, alternative availability | Medium | M |
| 3.4 | **Speed trend analysis** — rolling average speed over multiple cycles instead of point-in-time | Medium | M |
| 3.5 | **Season completion priority** — prefer searching for remaining episodes of nearly-complete seasons | Medium | M |
| 3.6 | **Auto-importer actually imports** — trigger manual import instead of just logging it | High | M |

### Phase 4: Seerr Redesign (multi-arr management)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 4.1 | **Design new Seerr integration** — understand instance relationships, instance groups | High | L |
| 4.2 | **Duplicate request detection** — same media across multiple instances | High | M |
| 4.3 | **Instance/quality reassignment** — move requests between instances | Medium | L |
| 4.4 | **Capacity-aware approval** — check health/disk/queue before approving | Medium | M |

### Phase 5: Polish & Completeness

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 5.1 | Add rTorrent download client support | Medium | M |
| 5.2 | Add category/tag filtering for download client management | Medium | M |
| 5.3 | Notification history page in UI | Low | M |
| 5.4 | Strike/removal log browser page in UI | Medium | M |
| 5.5 | Configurable auto-import interval | Low | S |
| 5.6 | Account lockout after failed login attempts | Low | S |
| 5.7 | Notification templates (customizable message format) | Low | L |

### Size Legend
- **S** = Small (< 100 lines changed)
- **M** = Medium (100-500 lines)
- **L** = Large (500+ lines)

---

## 6. Web Research: Additional Competitor Analysis

### 6.1 Decluttarr (798 stars, Python)
Active project with several features Lurkarr doesn't have yet:

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **TEST_RUN (dry-run mode)** | Logs all actions but doesn't execute deletions; essential for user confidence | **Missing** — no dry-run |
| **PROTECTED_TAG** | Torrents/items tagged with this tag are never removed by any cleaner logic | **Missing** — no whitelist/ignore rules |
| **OBSOLETE_TAG** | Instead of deleting, tag item so user can manually handle later (allows seed targets to be met) | **Missing** — all removals are immediate |
| **Adaptive slowness detection** | Auto-disables slow detection when current bandwidth exceeds 80% of configured limit (assumes everything is slow because pipe is full) | **Missing** — Lurkarr uses static speed thresholds |
| **DETECT_DELETIONS** | Monitors media folders for externally-deleted files (e.g., Plex/Tautulli cleanup), refreshes arr when detected | **Missing** — no filesystem awareness |
| **REMOVE_BAD_FILES** | Malware/sample detection with `keep_archives` option (keep .rar while removing bad executables) | **Partial** — Lurkarr has no file-level inspection |
| **REMOVE_UNMONITORED** | Clean up downloads for media that was unmonitored after download started | **Missing** |
| **MIN_DAYS_BETWEEN_SEARCHES** | Configurable cooldown between re-searches per media item to avoid hammering indexers | **Missing** — no re-search cooldown |
| **MAX_CONCURRENT_SEARCHES** | Limits simultaneous searches across all media | **Missing** — no concurrency control |
| **Per-job max_strikes** | Override the global strike count for specific stall reasons | **Missing** — global-only strikes |
| **message_patterns** | Wildcard matching on failed import messages (e.g., `*codec*unsupported*`) | **Missing** — no message-based filtering |
| **Tag-instead-of-remove pattern** | OBSOLETE_TAG allows seeding obligations to be completed before manual cleanup | **Missing** |

**Key takeaway:** Decluttarr's adaptive bandwidth detection, protected tags, and the tag-instead-of-delete pattern are the most impactful features Lurkarr should adopt. The dry-run mode is essential for user trust on first deployment.

### 6.2 Swaparr (382 stars, Rust)
Simpler tool, but has a few patterns worth noting:

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **MAX_DOWNLOAD_TIME** | Single timer-based removal (no strike system, just "too old → remove") | Lurkarr's strike system is already more sophisticated |
| **IGNORE_ABOVE_SIZE** | Skip items above a certain file size (large downloads take longer) | **Missing** — no size-based ignore |
| **STRIKE_QUEUED** | Strike items stuck in "queued" state (not downloading, not stalled — just queued) | **Partial** — Lurkarr handles "stalled" but queued items may slip through |
| **REMOVE_FROM_CLIENT** | Option to also remove from download client (not just arr) | Lurkarr already has `removeFromClient` option |
| **DRY_RUN** | Same as Decluttarr's TEST_RUN | **Missing** |

### 6.3 Recyclarr (1.9k stars, C#)
Syncs TRaSH Guides quality profiles and custom formats to Radarr/Sonarr. Not a competitor to Lurkarr's functionality, but relevant because:
- Shows that quality profile management across instances is a real user need
- Multi-instance support is standard
- Custom format scoring matters for quality decisions

### 6.4 Autobrr (2.6k stars, Go+React)
Relevant as a UI/architecture reference (same Go backend + web UI pattern):
- Go backend with PostgreSQL + SQLite support
- Multi-instance arr client support
- OIDC auth support
- Prometheus metrics
- React frontend with dark/light mode toggle
- Clean filter rule builder UI — **Lurkarr's blocklist/filter UI should be at least this good**

### 6.5 Consolidated: Missing Features Lurkarr Must Implement

**Critical (differentiators for launch):**
1. **Dry-run mode** — Every cleaner tool has this. Users won't trust a tool that deletes things without a safety net on first run.
2. **Protected tags/whitelist** — Users need to exempt specific items from all cleanup logic.
3. **Re-search after removal** — When queue cleaner removes a stalled download, automatically trigger a new search in the arr. Cleanuparr's #1 feature.
4. **Ignore rules** — By hash, category, tag, tracker, size. All competitors offer this.
5. **Adaptive speed detection** — Don't flag downloads as slow when the entire pipe is saturated.

**High Priority:**
6. **Tag-instead-of-delete pattern** — Let seeding obligations complete before removal.
7. **Per-job strike overrides** — Different stall types deserve different patience levels.
8. **Message pattern matching** — Failed imports have specific messages that should trigger specific actions.
9. **Size-based ignore** — Large files (50GB+ 4K remux) naturally take longer; don't penalize them.
10. **Search cooldown** — MIN_DAYS_BETWEEN_SEARCHES to avoid hammering indexers.
11. **Concurrent search limits** — MAX_CONCURRENT_SEARCHES to be a good indexer citizen.

**Medium Priority:**
12. **Filesystem deletion detection** — Monitor for externally-deleted media files.
13. **Unmonitored download cleanup** — Remove downloads for media that was unmonitored.
14. **Strike queued items** — Items stuck in "queued" (not stalled) state should also get strikes.

### 6.6 Deduparr (Rust, ~92 commits, MIT — git.enoent.fr/kernald/deduparr)

**Purpose:** Single-purpose tool that deduplicates *arr download queues. When the same movie/episode is queued at multiple quality levels (e.g., slow 720p grab + newer 1080p grab), it removes the lower quality duplicate, keeping the best.

**Supported apps:** Radarr + Sonarr only (no Lidarr/Readarr/Whisparr/Eros).

**Core Algorithm — Quality-Aware Queue Deduplication:**
1. Fetch queue from *arr instance
2. Group items by media ID (movie_id for Radarr, series_id+episode_id for Sonarr)
3. Detect "packs" — downloads containing multiple episodes (same download_id appears >1 time)
4. For each group with duplicates, compare quality scores: **custom format score** (authoritative if present) → **resolution** (2160>1080>720>480) → **source rank** (Remux=4, Bluray=3, WEB=2, HDTV=1)
5. Keep the best quality, remove the rest via arr API
6. **Smart pack handling:** Packs are atomic — a pack is removed ONLY if ALL items in it have better individual sources. If even one episode in the pack has no better source, the pack is kept entirely. "Partial" duplicates are tracked separately.
7. Uses a "seen tracker" so the same duplicate is only counted once in metrics, preventing metric inflation on subsequent runs

**Key Features:**

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **Intra-instance queue dedup** | Removes lower-quality duplicates within same instance's queue | **Missing** — Lurkarr's dupe detection is for cross-arr sync, not intra-queue |
| **Quality scoring** (CF → resolution → source) | Custom format score is authoritative; falls back to resolution then source type | **Partial** — Lurkarr has dupe scoring but not for intra-queue dedup |
| **Pack-aware dedup** | Multi-episode packs treated atomically (keep if needed for ANY episode) | **Missing** — no pack awareness |
| **Partial duplicate tracking** | Packs that are superseded for some (not all) episodes tracked as "partial" | **Missing** |
| **Seen tracker** (in-memory/DB) | Prevents double-counting duplicates across runs; clears on removal for re-detection | **Missing** |
| **DryRun / Interactive / Remove modes** | Three action modes including per-item interactive confirmation | **Partial** — Lurkarr has dry-run but no interactive per-item |
| **Prometheus metrics** | `duplicates_found_total{instance,status}`, `items_removed_total`, `api_errors_total` | Lurkarr already has Prometheus |
| **Metrics persistence** (SQLite/PostgreSQL/MySQL) | Duplicate counts persisted to DB for Grafana dashboards | **Missing** — Lurkarr metrics are in-memory only |
| **Health endpoints** (`/health/live`, `/health/ready`) | Ready check verifies connectivity to all configured instances | **Missing** — Lurkarr has no health endpoints |
| **Grafana dashboard JSON included** | Ships a ready-made Grafana dashboard | Lurkarr has dashboards too |
| **K8s CronJob or long-running service** | Supports both run-once and daemon modes | Lurkarr is daemon-only |

**Key Takeaways for Lurkarr:**
1. **Intra-queue deduplication is a completely different problem** from Lurkarr's cross-instance dedup (Phase 4). Deduparr solves same-instance, same-queue duplicate downloads. This should be a **new Phase 2 feature** for the queue cleaner.
2. **Pack-aware logic is essential** — removing a season pack when only some episodes are superseded breaks the download. Lurkarr needs this if adding intra-queue dedup.
3. **Custom format score should be authoritative** when comparing quality — this matches how the *arr ecosystem works (users configure CF scores to express quality preferences).
4. **Health endpoints** (`/health/live`, `/health/ready`) are table stakes for K8s/Docker deployments. Lurkarr should add these.
5. **Metrics persistence** — storing Prometheus-style counters in the database for long-term retention is a good pattern.

### 6.7 ArrQueueCleaner (27 stars, TypeScript — github.com/thelegendtubaguy/ArrQueueCleaner)

**Purpose:** Focused Sonarr queue cleaner with rule-based removal. Newer project (6 months old), but has a few interesting design decisions.

**Key Differentiators:**

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **Per-instance rule overrides** | Each Sonarr instance can override global rules (e.g., `removeNotAnUpgrade: true` for 4K instance only) | **Partial** — Lurkarr has per-job strike overrides but not per-instance rule toggles |
| **Episode count mismatch detection** | Detects when on-disk file spans more episodes than the release | **Missing** — Lurkarr doesn't check episode count consistency |
| **Series ID mismatch detection** | Catches releases where the series ID doesn't match the expected series | **Missing** |
| **Undetermined sample detection** | Removes items where Sonarr can't determine if file is a sample | **Missing** |
| **Per-rule blocklist toggle** | Each rule independently controls whether removed items are blocklisted | **Missing** — Lurkarr's blocklist decision is more coarse |
| **Opt-in safety** | No rules enabled by default — user must explicitly enable each one | Lurkarr uses defaults |
| **Cron-based scheduling** | Standard cron expression for timing (e.g., `*/5 * * * *`) | Lurkarr already has cron |
| **Multi-instance config** | JSON array of instances with name/host/apiKey/rules per entry | Lurkarr already has this |

**Key Takeaways:**
1. **Per-rule blocklist toggle** — useful granularity. "Quality blocked" items should be blocklisted (prevent re-grab), but "not an upgrade" items shouldn't (might be valid on retry).
2. **Episode count mismatch / Series ID mismatch** — niche but real problems users encounter. Could be added as additional message pattern matches.
3. **Opt-in safety** (no rules enabled by default) is a good trust-building pattern, similar to Lurkarr's dry-run mode.

### 6.8 Maintainerr (1.7k stars, TypeScript — github.com/Maintainerr/Maintainerr)

**Purpose:** Media lifecycle management. Configures rules based on Plex/Jellyfin/Seerr/Radarr/Sonarr/Tautulli data to build collections for "leaving soon" and eventually remove unwatched/stale content. **This is NOT a download queue manager** — it manages the media library itself.

**Relevant for Lurkarr:**

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **Rule engine with cross-service parameters** | Rules can combine data from Plex (play count, last watched) + Seerr (request date, requester) + Radarr/Sonarr (quality, size, added date) | **Missing** — Lurkarr doesn't consider Plex/Tautulli data |
| **"Leaving soon" collection-based lifecycle** | Shows content in a "Leaving Soon" Plex collection for X days before deletion | **Missing** — interesting UX pattern |
| **Manual include/exclude support** | Override rules for specific items | **Partial** — Lurkarr has protected tags |
| **Jellyfin + Plex support** | Dual media server support | Lurkarr doesn't integrate with media servers |
| **Seerr request cleanup** | Clears Seerr requests when media is removed | **Missing** — Lurkarr's Seerr integration doesn't clean up |
| **Unmonitor + remove from arr** | Can unmonitor in Radarr/Sonarr AND delete files | Lurkarr can unmonitor but lifecycle is simpler |

**Key Takeaways:**
1. **Maintainerr is complementary, not competitive.** It handles media library lifecycle; Lurkarr handles download queue management + lurking. They work together.
2. **Seerr request cleanup on media removal** — worth adding to Lurkarr's Phase 4 Seerr redesign.
3. **Cross-service rules** (combining Plex watch data + Seerr request data + arr metadata) is a powerful concept. Lurkarr could optionally integrate with Plex/Tautulli for smarter prioritization (e.g., don't search for content nobody watches).
4. **The "Leaving Soon" pattern** (warn before delete) maps to Lurkarr's existing tag-instead-of-delete feature.

### 6.9 qbit_manage (1.5k stars, Python — github.com/StuffAnThings/qbit_manage)

**Purpose:** Comprehensive qBittorrent management tool. Tags, categorizes, removes orphaned data, handles unregistered torrents, enforces share limits, and detects hardlink status.

**Key Features:**

| Feature | How It Works | Lurkarr Status |
|---------|-------------|---------------|
| **Tag by tracker URL** | Auto-tags torrents based on which tracker they're using | **Missing** — Lurkarr doesn't auto-tag |
| **Category by save path** | Assigns categories based on where files are saved | **Missing** |
| **Remove unregistered torrents** | Detects torrents removed from tracker (404/unregistered) and removes them | **Missing** — Lurkarr doesn't check tracker registration status |
| **Recheck paused torrents** | Rechecks paused torrents and resumes if complete (sorted by size — smallest first) | **Missing** |
| **No-hardlink detection (tag)** | Tags torrents that have no hardlinks outside root folder | **Partial** — Lurkarr has hardlink protection but differs in approach |
| **Share limits by group** | Filtered by tags/categories; min seed time + max ratio + max seeding time with optional cleanup | **Partial** — Lurkarr has ratio/time enforcement but not group-based |
| **RecycleBin** | Moves files to recycle bin instead of permanent deletion | **Missing** — Lurkarr does permanent deletion |
| **Category change rules** (`cat_change`) | Auto-change categories based on current category (useful for autobrr workflows) | **Missing** |
| **Cross-seed awareness** | Won't delete data if it's being cross-seeded | ✅ Lurkarr has this |
| **Orphaned file detection** | Finds files on disk not tracked by qBittorrent | **Partial** — Lurkarr's orphan detection is arr-queue-based, not filesystem-based |
| **Web UI** (Tauri desktop + web) | Built-in web interface for management | Lurkarr already has web UI |
| **Webhook notifications** (Notifiarr + Apprise) | Notification on actions | Lurkarr already has 8 notification providers |

**Key Takeaways:**
1. **RecycleBin pattern** — moving to a recycle folder instead of permanent deletion is a significant safety net. Pairs well with Lurkarr's dry-run and tag-instead-of-delete approach.
2. **Unregistered torrent detection** — checking if a torrent is still registered with its tracker is a detection method Lurkarr doesn't have. Unregistered torrents will never complete.
3. **Share limits by group** (tags/categories) — more granular than Lurkarr's global seeding enforcement. Different trackers have different seeding requirements.
4. **Recheck paused torrents** — a useful self-healing feature for torrents that got paused due to errors.
5. **Category-based workflow** — qbit_manage handles autobrr → category → management workflow. Lurkarr could benefit from category/tag-based pipeline awareness.

### 6.10 Expanded Consolidated Findings (Post-Deep-Research)

After analyzing **9 competing/complementary tools** (Cleanuparr, Decluttarr, Deduparr, Swaparr, ArrQueueCleaner, Maintainerr, qbit_manage, Recyclarr, Autobrr), here are **additional features Lurkarr should implement** beyond the existing roadmap:

**New Critical Features:**

| # | Feature | Source | Impact |
|---|---------|--------|--------|
| N1 | **Intra-queue deduplication** — remove lower-quality duplicate grabs within same instance | Deduparr | Prevents wasted bandwidth downloading same content at multiple quality levels |
| N2 | **Pack-aware operations** — treat multi-episode packs atomically in all decisions | Deduparr | Prevents breaking season packs by removing individual items |
| N3 | **Health endpoints** (`/health/live`, `/health/ready` with instance connectivity check) | Deduparr | Table stakes for K8s/Docker deployments, enables proper orchestration |
| N4 | **RecycleBin** — move files to recycle folder instead of permanent deletion | qbit_manage | Safety net; reduces risk of data loss |
| N5 | **Unregistered torrent detection** — detect torrents removed from tracker | qbit_manage | These torrents will never complete; instant removal candidate |

**New High-Priority Features:**

| # | Feature | Source | Impact |
|---|---------|--------|--------|
| N6 | **Share limits by group** (tags/categories/trackers) — different seeding rules per group | qbit_manage | Private trackers have different rules than public; essential for multi-tracker users |
| N7 | **Per-rule blocklist toggle** — independently control whether each removal reason triggers blocklisting | ArrQueueCleaner | More granular; "quality blocked" should blocklist but "not an upgrade" shouldn't |
| N8 | **Seerr request cleanup** — clear/close Seerr requests when media is removed/unmonitored | Maintainerr | Prevents stale requests cluttering Seerr |
| N9 | **Cutoff unmet search** — search for items where quality cutoff hasn't been met (not just missing) | Decluttarr | Proactively upgrades existing content to desired quality |
| N10 | **Ignore download clients** — skip items from specific download client names | Decluttarr | Some clients are managed by other tools (e.g., autobrr) |
| N11 | **Recheck paused torrents** — recheck and auto-resume paused torrents if complete | qbit_manage | Self-healing for erroneously paused downloads |
| N12 | **Episode/series ID mismatch detection** — catch releases with wrong metadata | ArrQueueCleaner | Prevents importing wrong content |

**New Medium-Priority Features:**

| # | Feature | Source | Impact |
|---|---------|--------|--------|
| N13 | **Metrics persistence** — store counters in DB for long-term Grafana dashboards | Deduparr | Current in-memory metrics reset on restart |
| N14 | **Run-once mode** (K8s CronJob) — execute once and exit, in addition to daemon mode | Deduparr | Useful for K8s CronJob deployments |
| N15 | **Keep archives option** for bad file removal (.rar, .zip kept for unpackerr) | Decluttarr | Prevents breaking unpackerr workflows |
| N16 | **Tag-by-tracker auto-tagging** — auto-tag downloads based on tracker URL | qbit_manage | Enables tracker-based rule targeting |
| N17 | **Plex/Tautulli integration** (optional) — use watch data for prioritization | Maintainerr | Search for content people actually watch; deprioritize unwatched |

---

## 7. Web Research: Multi-Arr Management Design

### 7.1 The Problem Space

Users run multiple arr instances for legitimate reasons:
- **Quality separation:** 1080p instance vs 4K instance (most common)
- **Content type separation:** Standard TV vs Anime (different indexers, naming, quality profiles)
- **Language separation:** English vs other languages
- **User separation:** Family-safe vs adult content (Whisparr v2/v3)

Current solutions (TRaSH Guides approach):
- Use arr's built-in **Import Lists** to sync between instances (profiles/tags/full sync)
- Separate download client categories per instance
- Separate root folders per instance
- **NO automation for cross-instance deduplication** — this is entirely manual today

**This is Lurkarr's opportunity.** No existing tool automates cross-instance quality arbitration.

### 7.2 User's Multi-Arr Example (Canonical Use Case)

Setup:
```
sonarr-standard  → quality: up to 1080p
sonarr-4k        → quality: 4K only
sonarr-anime     → quality: up to 1080p (anime profiles)
```

Quality hierarchy: `4K > 1080p > 720p > SD`

Rules:
1. If content exists in 4K AND 1080p → remove 1080p content + unmonitor in standard
2. If new 1080p request comes in for something already in 4K → reject/remove the 1080p request
3. 4K always wins when content is the same
4. Anime instance is independent (different content type, no overlap)

### 7.3 Multi-Arr Management Modes (Configurable)

**Mode 1: Quality Hierarchy (user's example)**
- Define quality rank: `4K > 1080p > 720p`
- Map instances to quality tiers: `{sonarr-4k: "4K", sonarr-standard: "1080p"}`
- Higher quality always wins → lower quality duplicate is unmonitored + content removed
- Works per-media-item (match by TVDB/TMDB ID)

**Mode 2: Primary/Secondary**
- One instance is "primary" (source of truth)
- Secondary instances sync from primary via Lurkarr (not arr's import lists)
- Lurkarr manages the sync with more intelligence than arr's built-in list sync
- De-duplication: primary wins on conflicts

**Mode 3: Independent with Overlap Detection**
- Instances are independent (different purposes)
- Lurkarr only detects and reports overlaps — doesn't auto-remove
- Dashboard shows: "Movie X exists in Instance A (1080p) and Instance B (4K)"
- User decides what to do — notification/report only

**Mode 4: Split Seasons (advanced)**
- For the user's case: older seasons in 1080p, newer in 4K
- Define split rules: `seasons 1-3: sonarr-standard, seasons 4+: sonarr-4k`
- Lurkarr monitors and enforces the split

### 7.4 Multi-Arr Architecture Design

```
┌──────────────────────────────────────────────┐
│              Instance Groups                  │
│                                               │
│  Group "TV Shows"                             │
│  ├── sonarr-standard (tier: 1080p, rank: 2)  │
│  ├── sonarr-4k       (tier: 4K,    rank: 1)  │
│  └── sonarr-anime    (tier: 1080p, independent) │
│                                               │
│  Group "Movies"                               │
│  ├── radarr-hd       (tier: 1080p, rank: 2)  │
│  └── radarr-4k       (tier: 4K,    rank: 1)  │
│                                               │
│  Mode: quality_hierarchy                      │
│  Match by: TVDB/TMDB ID                       │
│  Action: unmonitor + remove_content on loser  │
│  Ignore: instances marked "independent"       │
└──────────────────────────────────────────────┘
```

**Database model:**
```
instance_groups
  id, name, mode (quality_hierarchy|primary_secondary|overlap_detect|split_season)
  
instance_group_members
  id, group_id, app_id, tier_name, tier_rank, is_independent
  
cross_instance_media
  id, group_id, tmdb_id, tvdb_id, title
  instance_id, quality_tier, has_content (bool), monitored (bool)
  
cross_instance_actions (audit log)  
  id, media_id, action (unmonitor|remove_content|reject_request)
  source_instance_id, target_instance_id, reason, executed_at
```

**Sync process (runs on schedule):**
1. For each instance group, fetch all media from each member instance
2. Match by TVDB/TMDB ID to find overlaps
3. Apply mode-specific rules:
   - `quality_hierarchy`: higher rank wins, lower rank gets unmonitored + content removed
   - `primary_secondary`: primary always wins
   - `overlap_detect`: just log + notify
   - `split_season`: check season assignments vs rules
4. Execute actions (or log in dry-run mode)
5. Write to audit log for dashboard display

### 7.5 Seerr Integration Redesign

Current Lurkarr Seerr integration: just auto-approves requests. **This is wrong.**

Redesigned Seerr integration for multi-arr management:

1. **Request routing** — When a request comes through Seerr:
   - Check which instances already have this media
   - Check which quality tier the request is for
   - Route to the correct instance based on group rules
   - If higher quality already exists, auto-reject or notify user

2. **Duplicate detection** — Scan Seerr requests against all arr instances:
   - Flag requests for content that already exists in a higher-quality instance
   - Offer "upgrade" instead of "duplicate" (request 4K of something already in 1080p)

3. **Request cleanup** — Remove/close Seerr requests for content that was:
   - Moved to a higher quality instance
   - Unmonitored as part of quality hierarchy enforcement

4. **Quality profile reassignment** — Move a request from one instance to another:
   - User requests in standard → actually should be 4K → Lurkarr can reassign

---

## 8. Web Research: UI/UX Design Best Practices

### 8.1 Tech Stack Context
- **SvelteKit 5** with Svelte 5 runes (`$state`, `$derived`, `$effect`)
- **Tailwind CSS 4.2** (utility-first, oklch colors)
- **bits-ui** (headless components: Dialog, Switch, Select, etc.)
- **tailwind-variants** (variant-based styling for Button, Badge)
- **lucide-svelte** (icon library)
- **shadcn-svelte** patterns available (Sidebar, Card, Chart, DataTable, etc.)

### 8.2 Dashboard Layout Recommendations

**Use shadcn-svelte Sidebar component:**
The shadcn-svelte library provides a mature, composable sidebar with:
- Collapsible to icons mode (perfect for Lurkarr's many sections)
- Built-in mobile support (slides from side)
- Keyboard shortcut (Cmd+B / Ctrl+B)
- `inset` variant for modern look
- Grouped menu items with labels, badges, and sub-menus
- Footer section for user/account
- **Lurkarr should adopt this pattern** — it's production-ready and matches our stack

**Recommended layout:**
```
┌─────────────────────────────────────────────┐
│ [=] Lurkarr           [search] [user] [🔔]  │
├──────┬──────────────────────────────────────┤
│      │                                      │
│ Nav  │  Content Area                        │
│      │                                      │
│ 📊   │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐   │
│ Dash │  │ KPI │ │ KPI │ │ KPI │ │ KPI │   │
│      │  └─────┘ └─────┘ └─────┘ └─────┘   │
│ 🔍   │                                      │
│ Lurk │  ┌──────────────────────────────┐    │
│      │  │   Chart / Activity Feed      │    │
│ 📦   │  └──────────────────────────────┘    │
│ Queue│                                      │
│      │  ┌──────────────────────────────┐    │
│ 📥   │  │   Data Table                 │    │
│ Down │  └──────────────────────────────┘    │
│      │                                      │
│ 📱   │                                      │
│ Apps │                                      │
│      │                                      │
│ ⚙️   │                                      │
│ Set  │                                      │
├──────┴──────────────────────────────────────┤
│ Footer: version, health status              │
└─────────────────────────────────────────────┘
```

### 8.3 Key UI/UX Patterns to Implement

**1. Status Cards with Sparklines**
- Each connected arr instance gets a status card
- Show: name, health (green/yellow/red), queue count, disk space
- Small sparkline chart showing activity over last 24h
- shadcn Card + Chart components

**2. Activity Feed / Timeline**
- Chronological feed of all actions taken by Lurkarr
- Icons per action type: search triggered, download removed, import completed, blocklist added
- Filterable by instance, action type, time range
- Use shadcn-svelte's data table with virtual scrolling

**3. Instance Selector / Switcher**
- Top-level switcher to filter all views by instance
- "All Instances" as default
- per-instance filtering
- Use shadcn DropdownMenu in header or sidebar

**4. Action Confirmation Dialogs**
- For destructive actions: removing downloads, clearing queue items, blocklist entries
- Use bits-ui Dialog with clear warning state
- Show what will happen before confirmation
- Dry-run preview: "This would remove 5 items" before actual execution

**5. Toast Notifications**
- Non-blocking feedback for background operations
- Success (green), warning (amber), error (red)
- Auto-dismiss with undo option for reversible actions

**6. Data Tables with Column Controls**
- Queue, History, Downloads pages need sortable, filterable tables
- Column visibility toggle
- Persistent filter/sort preferences (localStorage)
- Row actions via dropdown menu (shadcn DropdownMenu pattern)

**7. Settings Pages**
- Use tabs/sections for different setting groups
- Inline validation with immediate feedback
- Connection test buttons for arr instances, download clients, notification providers
- Toggle switches for feature enable/disable (bits-ui Switch)

**8. Dark Mode First**
- Media server users overwhelmingly prefer dark mode
- Use oklch colors for consistent perception
- Separate sidebar colors from main content (shadcn sidebar theming pattern)
- Ensure sufficient contrast ratios (WCAG AA minimum)

### 8.4 Specific Component Recommendations

| Page | Current Issues | Recommended Components |
|------|---------------|----------------------|
| Dashboard | No KPIs, no overview | shadcn Card + Chart + stat counters |
| Queue | Basic table | shadcn DataTable with sorting, filtering, column toggle, row actions |
| Lurk | Just shows config | Add activity feed, last search timestamps per media, search stats |
| Apps | Instance cards lack status | Status cards with health indicator, connection status, disk space bar |
| Downloads | Basic listing | DataTable with progress bars, speed indicators, ETA, client type icon |
| History | Chronological list | Timeline component with icons, filterable, expandable details |
| Settings | Long form | Tabbed sections, inline validation, test buttons, feature toggles |
| Blocklist | Just a list | DataTable with regex tester, import/export, category columns |
| Notifications | Config only | Add notification history feed, test notification button, delivery status |

### 8.5 Accessibility & Usability Priorities

1. **Keyboard navigation** — All interactive elements reachable via Tab, sidebar via keyboard shortcut
2. **Screen reader support** — Use bits-ui's built-in aria attributes, add sr-only labels
3. **Loading states** — Skeleton loaders (shadcn Sidebar.MenuSkeleton pattern) while fetching data
4. **Error states** — Clear error messages with retry actions, not just blank pages
5. **Empty states** — Helpful messages when no data exists ("No queue items. Your downloads are clean!")
6. **Responsive design** — Sidebar collapses on mobile, data tables stack or scroll horizontally
7. **Consistent action placement** — Primary actions top-right, destructive actions require confirmation
8. **Breadcrumbs** — For nested views (Settings > Notifications > Discord)

### 8.6 Lurkarr-Specific UX Patterns

**Whisparr v2/v3 Unified View:**
As per conventions, show a single "Whisparr" section with a v2/v3 toggle selector.
```svelte
<Tabs.Root value="v3">
  <Tabs.List>
    <Tabs.Trigger value="v2">v2 (Sonarr-based)</Tabs.Trigger>
    <Tabs.Trigger value="v3">v3 (Radarr-based)</Tabs.Trigger>
  </Tabs.List>
</Tabs.Root>
```

**Multi-Instance Group Configuration:**
- Visual drag-and-drop grouping of instances
- Quality tier assignment via select/dropdown per instance
- Real-time preview of how rules would apply
- "Simulate" button (dry-run a dedup pass and show results)

**Cross-Instance Dedup Dashboard:**
- Matrix view: rows = media, columns = instances
- Cell shows: quality, file size, monitored status
- Color coding: green (winner), red (will be removed), gray (not present)
- Bulk actions: "Apply recommendations" with confirmation

---

## 9. Revised Improvement Plan (Post-Research)

### Phase 0: Safety & Trust (NEW — must be first)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 0.1 | **Dry-run mode** — global toggle, log all actions without executing | Critical | M | ✅ `2ccfe2b` |
| 0.2 | **Protected tags** — items with configured tag(s) are exempt from all cleanup | Critical | S | ✅ `20288a7` |
| 0.3 | **Action audit log** — persistent log of every action Lurkarr takes, viewable in UI | Critical | M | ✅ `1282c9b` |

### Phase 1: Critical Bug Fixes (unchanged)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 1.1 | **Persist CSRF key** to database/config (B1) | Critical | S | ✅ `73764a2` |
| 1.2 | **Transaction wrapping** for strike add+count (B2) | Critical | S | ✅ `05971da` |
| 1.3 | **Regex complexity limits** — cap pattern length + use timeout context (B3) | Critical | S | ✅ `79a662a` |
| 1.4 | **Validate arr API URLs** — prevent SSRF via instance URLs (B4) | High | S | ✅ `e6416f9` |
| 1.5 | **Fix lurking random bias** — use crypto/rand or Fisher-Yates shuffle (B5) | High | S | ✅ Already correct |
| 1.6 | **Fix cross-arr title-only matching** — use TVDB/TMDB IDs (B8) | High | M | ✅ `023cf18` |

### Phase 2: Queue Cleaner Upgrade (expanded with research)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 2.1 | **Re-search after removal** — trigger arr search when item is removed | Critical | M | ✅ `d001b4a` |
| 2.2 | **Ignore rules** — by hash, category, tag, tracker, size | Critical | M | ✅ `7876419` |
| 2.3 | **Adaptive speed detection** — disable slow flagging when bandwidth >80% of limit | High | M | ✅ `cf7e501` |
| 2.4 | **Per-job strike overrides** — different max strikes per stall reason | High | S | ✅ `c26090f` |
| 2.5 | **Size-based ignore** — configurable threshold for large files | High | S | ✅ `babdc90` |
| 2.6 | **Tag-instead-of-delete** — tag items as "obsolete" instead of immediate removal | High | M | ✅ `dedb372` |
| 2.7 | **Message pattern matching** — failed import message wildcards | Medium | M | ✅ `3a670e9` |
| 2.8 | **Strike queued items** — items stuck in "queued" state (not just stalled) | Medium | S | ✅ `decaa14` |
| 2.9 | **Search cooldown** — MIN_DAYS_BETWEEN_SEARCHES per media item | Medium | S | ✅ `cde560f` |
| 2.10 | **Concurrent search limits** — MAX_CONCURRENT_SEARCHES | Medium | S | ✅ `8464840` |

### Phase 3: Lurking Engine Improvements (unchanged + additions)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 3.1 | **Prioritization** — newest/oldest/random/least-recently-searched modes | High | M | ✅ `1031f2d` |
| 3.2 | **Per-instance backoff** — separate cooldowns per arr instance | High | M | ✅ `b30c3b1` |
| 3.3 | **Re-search on removal integration** — coordinate with queue cleaner | High | S | ✅ `73c33c1` |
| 3.4 | **Search failure tracking** — track and deprioritize consistently-failing searches | Medium | M | ✅ `b62ddba` |
| 3.5 | **Filesystem deletion detection** — monitor for externally-deleted media | Medium | L |
| 3.6 | **Unmonitored download cleanup** — clean up downloads for unmonitored media | Medium | M |

### Phase 4: Multi-Arr Management (NEW — major feature)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 4.1 | **Instance groups** — DB model + API for grouping instances with quality tiers | Critical | L |
| 4.2 | **Quality hierarchy mode** — automatic cross-instance deduplication by quality rank | Critical | L |
| 4.3 | **Cross-instance media matching** — by TVDB/TMDB ID, not title | Critical | M |
| 4.4 | **Seerr request routing** — route requests to correct instance based on group rules | High | L |
| 4.5 | **Duplicate request detection** — flag/reject duplicate requests across instances | High | M |
| 4.6 | **Overlap detection mode** — notify-only mode for conservative users | High | M |
| 4.7 | **Cross-instance dedup dashboard** — matrix view showing media across instances | High | L |
| 4.8 | **Split season rules** — assign seasons to different instances | Medium | L |
| 4.9 | **Request quality reassignment** — move requests between instances | Medium | M |

### Phase 5: UI/UX Overhaul (NEW)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 5.1 | **Adopt shadcn Sidebar** — replace current nav with composable sidebar component | High | M |
| 5.2 | **Dashboard KPI cards** — status cards with instance health, queue counts, disk space | High | M |
| 5.3 | **Activity feed/timeline** — chronological log of all Lurkarr actions | High | M |
| 5.4 | **Instance switcher** — top-level filter by instance across all views | High | S |
| 5.5 | **DataTable upgrade** — sorting, filtering, column toggle, row actions for Queue/History/Blocklist | High | L |
| 5.6 | **Toast notifications** — non-blocking feedback for background operations | Medium | S |
| 5.7 | **Settings tabs** — tabbed sections with inline validation and test buttons | Medium | M |
| 5.8 | **Loading/empty/error states** — skeleton loaders, helpful empty states, retry actions | Medium | M |
| 5.9 | **Whisparr v2/v3 unified view** — single section with version toggle | Medium | S |
| 5.10 | **Cross-instance dedup matrix view** | Medium | L |

### Phase 6: Polish & Completeness

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 6.1 | Add rTorrent download client support | Medium | M |
| 6.2 | Notification history + delivery status UI | Medium | M |
| 6.3 | Notification templates (customizable message format) | Low | L |
| 6.4 | Account lockout after failed login attempts | Low | S |
| 6.5 | Strike/removal log browser page in UI | Medium | M |
| 6.6 | Configurable auto-import interval | Low | S |
| 6.7 | **Health endpoints** (`/health/live`, `/health/ready`) — liveness + instance connectivity | High | S |
| 6.8 | **RecycleBin mode** — move files to recycle folder instead of permanent deletion | Medium | M |
| 6.9 | **Unregistered torrent detection** — detect torrents removed from tracker (404/unregistered) | Medium | M |
| 6.10 | **Recheck paused torrents** — recheck and auto-resume paused torrents if complete | Low | S |
| 6.11 | **Metrics persistence** — store key counters in DB for long-term retention across restarts | Low | M |
| 6.12 | **Seerr request cleanup** — close/clear Seerr requests when media is removed/unmonitored | Medium | M |
| 6.13 | **Run-once mode** — execute once and exit (for K8s CronJob deployments) | Low | S |

### Phase 7: Queue Intelligence (NEW — from Deduparr/Decluttarr analysis)

| # | Task | Priority | Effort |
|---|------|----------|--------|
| 7.1 | **Intra-queue deduplication** — detect and remove lower-quality duplicate grabs within same instance queue | High | L |
| 7.2 | **Pack-aware operations** — treat multi-episode packs atomically in dedup + all cleanup decisions | High | M |
| 7.3 | **Quality scoring engine** — custom format score (authoritative) → resolution → source rank | High | M |
| 7.4 | **Cutoff unmet search** — search for items where quality cutoff hasn't been met (proactive upgrades) | Medium | M |
| 7.5 | **Share limits by group** — different seeding rules per tag/category/tracker group | Medium | M |
| 7.6 | **Per-rule blocklist toggle** — independently control whether each removal reason triggers blocklisting | Medium | S |
| 7.7 | **Ignore download clients** — skip items from specific download client names | Medium | S |
| 7.8 | **Episode/series ID mismatch detection** — catch releases with wrong metadata | Low | S |
| 7.9 | **Keep archives option** — preserve .rar/.zip for unpackerr when removing bad files | Low | S |

### Revised Size Legend
- **S** = Small (< 100 lines changed)
- **M** = Medium (100-500 lines)
- **L** = Large (500+ lines)

### Reference Projects Analyzed

| Project | Stars | Language | Focus | URL |
|---------|-------|----------|-------|-----|
| Cleanuparr | 2.1k | C# | Queue cleaning, strike system, blocklist | github.com/Cleanuparr/Cleanuparr |
| Recyclarr | 1.9k | C# | TRaSH Guides quality profile sync | github.com/recyclarr/recyclarr |
| Maintainerr | 1.7k | TypeScript | Media library lifecycle management | github.com/Maintainerr/Maintainerr |
| qbit_manage | 1.5k | Python | qBittorrent management, tagging, share limits | github.com/StuffAnThings/qbit_manage |
| Decluttarr | 798 | Python | Queue cleaning, multi-job, adaptive speed | github.com/ManiMatter/decluttarr |
| Swaparr | 382 | Rust | Simple stalled download removal | github.com/ThijmenGThN/swaparr |
| Autobrr | 2.6k | Go | Filter-based auto-downloading | github.com/autobrr/autobrr |
| Deduparr | ~10 | Rust | Quality-based queue deduplication | git.enoent.fr/kernald/deduparr |
| ArrQueueCleaner | 27 | TypeScript | Rule-based Sonarr queue cleaning | github.com/thelegendtubaguy/ArrQueueCleaner |
