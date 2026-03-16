# UI/UX Completeness Analysis

## Executive Summary

The Lurkarr frontend is **~85% feature-complete**, with well-structured pages covering most major backend capabilities. The implementation is professional with proper loading states, error handling, and confirmation dialogs. However, there are notable **coverage gaps**, particularly around SABnzbd-specific functionality, Seerr request browsing, Prowlarr indexer visibility, and some visual feedback mechanisms.

---

## 1. Frontend Route Map (15 Routes)

| Route | Page | Purpose | Completeness |
|-------|------|---------|--------------|
| `/` | Dashboard | Overview, stats, health | ✅ 90% |
| `/login` | Login | Auth entry point | ✅ 100% |
| `/apps` | Connections | Arr instances, DL clients, services | ✅ 95% |
| `/lurk` | Lurk Settings | Per-app lurking config | ✅ 100% |
| `/scheduling` | Scheduling | CRUD for scheduled actions | ✅ 100% |
| `/history` | History | Lurk/blocklist/import/strike logs | ✅ 100% |
| `/activity` | Activity | Unified event feed | ✅ 100% |
| `/downloads` | Downloads | DL client live status | ✅ 100% |
| `/queue` | Queue | Queue cleaner settings, scoring, seeding | ✅ 90% |
| `/dedup` | Dedup | Cross-instance overlaps, groups, Seerr | ✅ 95% |
| `/notifications` | Notifications | Provider CRUD & events | ✅ 100% |
| `/monitoring` | Monitoring | Health probes, Prometheus, Grafana | ✅ 80% |
| `/settings` | Settings | General + OIDC SSO | ✅ 100% |
| `/admin/users` | Users | Admin user management | ✅ 100% |
| `/user` | Profile | Sessions, 2FA, passkeys | ✅ 100% |

---

## 2. Backend Endpoint Coverage (80+ Endpoints)

### Fully Covered (60+ endpoints)
- Authentication & User Management (13 endpoints) → 100%
- User Profile & Sessions (7 endpoints) → 100%
- Admin Management (7 endpoints) → 100%
- Settings (5 endpoints) → 100%
- Instances/Arr Apps (6 endpoints) → 100%
- History & Activity (3 endpoints) → 100%
- Statistics (3 endpoints) → 100%
- Scheduling (5 endpoints) → 100%
- Blocklist Management (8 endpoints) → 100%
- Notifications (7 endpoints) → 100%
- Download Clients (8 endpoints) → 100%

### Partially Covered
- Queue Cleaner (12 endpoints) → 90% — dry-run mode lacks visual toggle
- Instance Groups (10 endpoints) → 95% — mode not explained
- Monitoring (4 endpoints) → 80% — no embedded metrics

### Under-Covered
- **Seerr Integration** (7 endpoints) → **40%** — settings only, no request browser
- **SABnzbd Integration** (8 endpoints) → **25%** — settings only, no queue/history/controls
- **Prowlarr Integration** (5 endpoints) → **40%** — settings only, no indexer list/stats
- **State Management** (2 endpoints) → **0%** — no UI at all

---

## 3. Critical Coverage Gaps

### 🔴 Gap 1: Seerr Requests Not Browsable
- Backend has `GET /api/seerr/requests`, request count, reassignment
- Frontend only shows settings + duplicate scan
- **Impact**: Users must check Seerr directly for pending requests
- **Fix**: Create `/seerr` page with request list, filters, reassignment UI

### 🔴 Gap 2: SABnzbd Queue & History Not Visible
- Backend has 8 endpoints: queue, history, stats, pause/resume, test
- Frontend only shows settings configuration
- **Impact**: Users can't troubleshoot SABnzbd from Lurkarr
- **Fix**: Integrate into `/downloads` page or create `/sabnzbd` section

### 🔴 Gap 3: Dry-Run Mode Lacks Visual Feedback
- `dry_run` boolean exists in queue cleaner settings
- No UI toggle or warning banner when enabled
- **Impact**: Users may not realize they're in dry-run mode
- **Fix**: Add prominent toggle with red/yellow warning banner

### 🔴 Gap 4: State vs Stats Confusion
- Two separate concepts: stats = search counts; state = lurk progress
- No clear distinction in the UI
- **Impact**: Users don't know when to reset state vs stats
- **Fix**: Add help text, consider renaming "state" to "lurk progress"

---

## 4. Medium Priority Issues

### ⚠️ Whisparr v2/v3 Selection Unclear
- Selector shows "whisparr" and "eros" (internal codename)
- Non-technical users confused by "eros"
- **Fix**: Show "Whisparr v2 (Sonarr-based)" and "Whisparr v3 (Radarr-based)"

### ⚠️ Instance Groups Mode Not Explained
- "mode" field with no explanation of dependent vs independent
- Users create groups incorrectly
- **Fix**: Add tooltips with example use cases

### ⚠️ Prowlarr Indexer Sync Unchecked
- `sync_indexers` toggle exists but provides no feedback
- **Fix**: Show list of synced indexers; add "Sync now" button

### ⚠️ Blocklist Patterns No Regex Preview
- Accepts regex patterns with no tester or example
- **Fix**: Add regex tester input showing sample matches

### ⚠️ Scoring Profiles Buried
- Located on same page as queue cleaner settings
- **Fix**: Dedicated "Scoring" tab or clearer section

### ⚠️ Schedule History Not Easily Accessible
- Requires extra click on history button
- **Fix**: Show recent executions inline under each schedule

---

## 5. Navigation Analysis

### Current Sidebar Structure
```
Dashboard
Connections (/apps)
Lurk Settings (/lurk)
Scheduling (/scheduling)
History (/history)
Activity (/activity)
Downloads (/downloads)
Queue (/queue)
Dedup (/dedup)
Notifications (/notifications)
Monitoring (/monitoring)
Settings (/settings)
Users (/admin/users) [admin only]
Profile (/user)
```

### Issues
- **"Connections" vs "Apps"**: Label is generic
- **No grouping**: All 14 items flat
- **No Seerr, SABnzbd, Prowlarr** pages in menu

### Recommended Reorganization
```
Dashboard

[Configuration]
  Connections (/apps)
  Lurk Settings (/lurk)
  Queue Settings (/queue)
  Scheduling (/scheduling)

[Operations]
  Downloads (/downloads)
  Dedup (/dedup)
  Seerr Requests (/seerr) **NEW**

[Monitoring]
  History (/history)
  Activity (/activity)
  Monitoring (/monitoring)

[Administration]
  Notifications (/notifications)
  Settings (/settings)
  Users (/admin/users)

[Account]
  Profile (/user)
```

---

## 6. Dashboard Analysis

### What's Good
- Stats per app (searches, upgrades)
- Hourly API cap tracking
- Instance health status
- Download client overview
- Service connection status (Prowlarr, Seerr)

### What's Missing
- No SABnzbd widget
- Seerr count only — no latest requests
- No Prowlarr indexer preview
- No alerts/degradation indicators
- No recently failed schedules
- No download queue warnings (min_download_queue_size)

---

## 7. Component Library Assessment

| Component | Status | Notes |
|-----------|--------|-------|
| Button | ✅ | Multiple variants |
| Input | ✅ | Text, password, number |
| Select | ✅ | Dropdown with optgroups |
| Toggle | ✅ | Checkbox-style |
| Tabs | ✅ | Tab navigation |
| Modal | ✅ | Confirmation & content |
| Card | ✅ | Bordered container |
| Badge | ✅ | Status variants |
| PageHeader | ✅ | Title + description |
| DataTable | ✅ | Sortable + pagination |
| Skeleton | ✅ | Loading placeholders |
| EmptyState | ✅ | Empty state messages |

### Missing Components
- ❌ Form validation component (validation is inline)
- ❌ Tooltip component (hints as small text only)
- ❌ Progress bar (polls use spinners only)
- ❌ Breadcrumb navigation
- ❌ Collapse/accordion
- ❌ Code display / syntax highlighting

---

## 8. API Client Quality

### Strengths
- Centralized API client with proper error handling
- CSRF token management with auto retry on 401
- Auto-redirect to login on auth failure
- Generic types for type-safe responses

### Limitations
- No request caching (every fetch is fresh)
- No retry logic for network errors (only CSRF retry)
- No request cancellation for long operations
- No rate-limit handling (no 429 backoff)

---

## 9. Component Maturity

| Category | Maturity | Notes |
|----------|----------|-------|
| Forms | 🟢 High | Solid text, select, toggle; inline validation |
| Tables | 🟢 High | Sorting, pagination, good UX |
| Dialogs | 🟢 High | Confirmation, creation, deletion |
| Navigation | 🟡 Medium | Works but needs grouping; no breadcrumbs |
| Status Display | 🟡 Medium | Badge/color usage inconsistent |
| Loading States | 🟡 Medium | Skeleton screens; could add progress bars |
| Error Handling | 🟢 High | Toast notifications; error boundaries |
| Mobile UX | 🟡 Medium | Responsive layout; some pages cramped |

---

## 10. Priority Actions

### High Priority (New pages/features needed)
| Feature | Backend | Frontend | Action |
|---------|---------|----------|--------|
| Seerr Requests | ✅ 7 endpoints | ❌ | Create `/seerr` page |
| SABnzbd Dashboard | ✅ 8 endpoints | ❌ | Add to `/downloads` or new page |
| Prowlarr Indexers | ✅ 5 endpoints | ❌ | Add indexer list to Prowlarr modal |
| Dry-Run Mode | ✅ Setting | ❌ toggle | Add toggle + warning banner |
| State Management | ✅ 2 endpoints | ❌ | Add to settings or new section |

### Medium Priority (UX improvements)
| Feature | Action |
|---------|--------|
| Dashboard Alerts | Show degraded health, failed schedules |
| Download Queue Warnings | Alert on low queue size |
| API Rate Limit Visual | Show hourly cap usage with threshold |
| Recent Searches Widget | Last 5 searches on dashboard |
| Sidebar Grouping | Organize into collapsible sections |

### Low Priority (Polish)
| Feature | Action |
|---------|--------|
| Breadcrumb Navigation | Context path display |
| Keyboard Shortcuts | Cmd+K for quick nav |
| Export/Import Settings | YAML/JSON backup |
| Health Status Standardization | Unified icon + badge combo |
