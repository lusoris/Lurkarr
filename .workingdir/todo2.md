# Lurkarr — Implementation Progress Tracker

**Started:** 2026-03-13 | **Branch:** develop

---

## Phase 1: Critical Bug Fixes
- [x] **1.1** Fix "Revoke All Sessions" logging out current user (BUG-001)
- [x] **1.2** Fix hardlink.go build tags for cross-platform (BUG-002)

## Phase 2: Connections Page Redesign
- [x] **2.1** Single "+ Add" button with categorized dropdown
- [x] **2.2** Hide empty categories — only show sections that have instances
- [x] **2.3** Always show Services (Prowlarr/Seerr) at bottom
- [x] **2.4** Empty state illustration when no connections exist

## Phase 3: Settings Layout Improvements
- [x] **3.1** Lurk settings — add section headers (Lurking Behaviour, API & Command, Security)
- [x] **3.2** Notifications modal — group events by category, email fields side-by-side

## Phase 4: Wire Up Missing Features
- [x] **4.1** Blocklist source management UI (CRUD)
- [x] **4.2** Blocklist rule management UI (CRUD)
- [x] **4.3** State reset per instance (button on dashboard cards)
- [x] **4.4** History delete by app type
- [x] **4.5** Seerr request count on dashboard/seerr card

## Phase 5: Full Test Suite
- [x] **5.1** Backend: sessions handler tests (10 tests)
- [x] **5.2** Backend: admin handler tests (14 tests)
- [x] **5.3** Backend: recovery code tests (6 tests)
- [x] **5.4** Backend: hardlink cross-platform test (5 tests)
- [x] **5.5** Frontend: setup Vitest + testing-library
- [x] **5.6** Frontend: API client tests (9 tests)
- [x] **5.7** Frontend: auth store tests (6 tests)
- [x] **5.8** Frontend: toast store tests (8 tests)
- [x] **5.9** Frontend: UI component tests (23 tests)

## Phase 6: UI Consistency Pass
- [x] **6.1** Loading skeletons on all pages
- [x] **6.2** Error states on all pages
- [x] **6.3** Confirm dialogs for all destructive actions
- [x] **6.4** Consistent spacing/padding/typography audit

---

## Change Log

| When | Item | What Changed |
|------|------|-------------|
| Phase 1 | 1.1 | Added `DeleteUserSessionsExcept` to Store/DB, fixed `HandleRevokeAllSessions` to keep current session |
| Phase 1 | 1.2 | Added `//go:build linux` to hardlink.go, created hardlink_other.go stub, tagged hardlink_test.go |
| Phase 2 | 2.1-2.4 | Rewrote apps page: single Add Connection dropdown, hide empty sections, empty state, Services always visible |
| Phase 3 | 3.1-3.2 | Settings: 3 sectioned Cards; Notifications: grouped events, email grid |
| Phase 4 | 4.3 | Dashboard: per-instance reset button with refresh icon, show instance breakdown for single-instance apps too |
| Phase 4 | 4.4 | History: app type filter dropdown, delete-by-app buttons with inline confirm, proper {items,total} response parsing, total count display |
| Phase 4 | 4.5 | Dashboard: Seerr request count card in new "Services" section (fetches GET /seerr/requests/count) |
| Phase 4 | 4.1-4.2 | Queue page: Global Blocklist Management section with Sources CRUD (add/edit/delete/toggle) and Rules CRUD (add/delete with pattern type selector, source attribution) |
| Phase 5 | 5.1-5.3 | Backend: 30 new Go tests — sessions handlers (10), admin handlers (14), recovery codes (6) |
| Phase 5 | 5.4 | Backend: 5 cross-platform hardlink tests in hardlink_common_test.go (no build tag, runs everywhere) |
| Phase 5 | 5.5-5.9 | Frontend: Vitest 4.1 setup, 60 tests across 5 files — index.test.ts (14), api.test.ts (9), auth.test.ts (6), toast.test.ts (8), components.test.ts (23) |
| Phase 6 | 6.1 | Loading skeletons: lurk settings (6 pulse rows), queue sources/rules (3 pulse rows each), queue tabs (4 pulse rows), user profile (4 cards), admin users (3 rows) |
| Phase 6 | 6.3 | Confirm dialogs: dashboard reset stats (inline Yes/No), dashboard per-instance reset (inline Yes/No), queue delete source (inline), queue delete rule (inline), user disable 2FA (inline), user revoke session/all (inline) |
