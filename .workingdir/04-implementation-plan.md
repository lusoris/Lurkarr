# Implementation Plan — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Phase 1: Bug Fixes (Critical)

### 1.1 Fix Session Revoke All (BUG-001)
- Update `HandleRevokeAllSessions` to exclude current session from deletion
- Add `DeleteUserSessionsExcept(ctx, userID, currentSessionID)` query
- Update tests

### 1.2 Fix hardlink.go Build Tags (BUG-002)
- Split `hardlink.go` into `hardlink_linux.go` + `hardlink_other.go`
- Add `//go:build linux` and `//go:build !linux` tags
- Stub returns `false` on non-Linux (safe default)

## Phase 2: Connections Page Redesign (User Request)

### Current Problem
Page shows ALL 5 arr app sections with per-type "Add" buttons, even when no instances exist.
User wants: single "Add" button → type dropdown → only show categories that have instances.

### 2.1 New Layout: Dynamic Connections
```
┌─────────────────────────────────────────────┐
│ Connections                    [+ Add New ▼] │
├─────────────────────────────────────────────┤
│                                              │
│ (empty state: illustration + "Add your      │
│  first connection to get started")           │
│                                              │
│ After adding:                                │
│ ┌─ Sonarr ──────────────────────────────┐   │
│ │ [instance card] [instance card]       │   │
│ └───────────────────────────────────────┘   │
│ ┌─ SABnzbd ─────────────────────────────┐   │
│ │ [instance card]                       │   │
│ └───────────────────────────────────────┘   │
│ ┌─ Services ────────────────────────────┐   │
│ │ [Prowlarr card] [Seerr card]         │   │
│ └───────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### 2.2 Add Dropdown Menu
- Single "+ Add" button with dropdown:
  - Section header: "Arr Apps"
    - Sonarr, Radarr, Lidarr, Readarr, Whisparr
  - Section header: "Download Clients"
    - qBittorrent, Transmission, Deluge, SABnzbd, NZBGet
- Selecting opens the existing modal with type pre-filled
- Services (Prowlarr/Seerr) always visible at bottom

## Phase 3: Settings Layout Improvements

### 3.1 Lurk Settings Page — Add Sections
Current: flat list of inputs
Target: grouped with headers (like queue page)
```
Search Behavior
├── Lurk Missing Count / Lurk Upgrade Count (2-col)
├── Missing Mode / Upgrade Mode (2-col)
Rate Limiting
├── Sleep Duration (seconds)
├── Hourly API Cap
Filters
├── Monitored Only (toggle)
├── Skip Future Releases (toggle)
Advanced
├── Random Selection (toggle)
├── Debug Mode (toggle)
```

### 3.2 Notifications Modal — Group Config Fields
Current: flat list of all config fields
Target: grouped by purpose for complex providers (email)
```
Provider Type / Name / Enabled
─── Connection ───
Host / Port (2-col)
TLS toggle
─── Authentication ───
Username / Password (2-col)
─── Message ───
From / To (2-col)
Subject
─── Events ───
Event checkboxes (2-col grid)
```

## Phase 4: Wire Up Missing Features

### 4.1 Blocklist Source Management
- Backend has full CRUD: `GET/POST /blocklist/sources`, `PUT/DELETE /blocklist/sources/{id}`
- Frontend only shows blocklist LOG — no UI for managing sources
- Add: Source list with add/edit/delete, sync status, enable/disable

### 4.2 Seerr Requests Display
- Backend has `GET /seerr/requests`, `GET /seerr/count`
- No frontend for this — add a section to dashboard or seerr service view

### 4.3 State Reset Per Instance
- Backend has `POST /state/{app_type}/instances/{id}/reset`
- Frontend only has global reset on dashboard
- Add per-instance reset button on dashboard cards

### 4.4 SABnzbd Dashboard Widgets
- Backend has `GET /sabnzbd/queue`, `GET /sabnzbd/history`, `GET /sabnzbd/stats`
- Could add enhanced SABnzbd view on downloads page

## Phase 5: Full Test Suite

### 5.1 Backend Test Coverage Goals
Current: ~150+ tests across 20 packages, all passing (except queuecleaner on Windows)
Gaps to fill:
- [ ] `api/sessions.go` — no tests for new session handler
- [ ] `api/admin.go` — no tests for new admin handler 
- [ ] `auth/recovery.go` — no tests for recovery code generation/validation
- [ ] `database/queries_download_clients.go` — no unit tests (integration only)
- [ ] `database/queries_blocklist.go` — no unit tests
- [ ] `database/queries_notifications.go` — no unit tests
- [ ] `database/queries_seerr.go` — no unit tests
- [ ] `server/server_test.go` — route registration tests only, no integration

### 5.2 Frontend Test Setup
Currently: **zero frontend tests**
Plan:
- Add Vitest (already compatible with Vite)
- Add @testing-library/svelte for component tests
- Test priority:
  1. API client (`$lib/api.ts`)
  2. Auth store (`$lib/stores/auth.svelte.ts`)
  3. Toast store (`$lib/stores/toast.svelte.ts`)
  4. UI components (Button, Input, Toggle, Modal)
  5. Page-level integration tests

## Phase 6: UI Consistency Pass

### 6.1 Audit All Visual Elements
- Verify all pages use same spacing scale
- Verify all cards have consistent padding
- Verify all section headers use same typography
- Verify all tables have consistent styling
- Verify all empty states look the same

### 6.2 Add Missing Polish
- Loading skeletons on all pages (some have, some don't)
- Error states on all pages (currently silent failures on some)
- Confirm dialogs for all destructive actions
- Keyboard shortcuts (Ctrl+S for save, Escape for modals)

## Priority Order

1. **Phase 1** — Bug fixes (critical session bug)
2. **Phase 2** — Connections page (user-requested, high impact)
3. **Phase 3** — Settings layouts (user-requested)
4. **Phase 4** — Wire up missing features
5. **Phase 5** — Test suite
6. **Phase 6** — UI consistency
