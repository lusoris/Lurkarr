# Bug Tracker — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Critical Bugs

### BUG-001: Session Revoke All Logs Out Current User
**Location:** `internal/api/sessions.go` — `HandleRevokeAllSessions`
**Symptom:** "Revoke all other sessions" actually revokes ALL sessions including current
**Root cause:** Code deletes all user sessions, but doesn't preserve/recreate the current one
**Impact:** User gets logged out when clicking "Revoke All Others"
**Fix:** Filter current session from deletion OR recreate it after deletion

### BUG-002: queuecleaner fails to compile on Windows
**Location:** `internal/queuecleaner/hardlink.go:61`
**Symptom:** `undefined: syscall.Stat_t` on Windows
**Root cause:** Uses Linux-specific syscall for hardlink detection
**Impact:** Tests can't run on Windows (works in Docker/Linux)
**Fix:** Add build tags (`//go:build linux`) and a no-op stub for other platforms

## Medium Priority

### BUG-003: WebAuthn Dead Code
**Location:** `internal/database/models.go:50-58`, migration 007
**Symptom:** `WebAuthnCredential` model defined, `webauthn_credentials` table created, but no handlers, no endpoints, no frontend
**Impact:** Dead code confusion. Table exists with no way to populate it.
**Fix:** Either implement or remove

### BUG-004: Legacy Download Client Settings Still in Schema
**Location:** `internal/database/queries_queue.go` — `GetDownloadClientSettings`/`UpdateDownloadClientSettings`
**Symptom:** Two download client patterns co-exist: legacy per-app `download_client_settings` and new multi-instance `download_client_instances`
**Impact:** Potential confusion / data inconsistency  
**Fix:** Migrate all consumers to multi-instance, drop legacy table

### BUG-005: No DB Transactions for Multi-Step Operations
**Location:** `internal/api/auth.go:196-210` (2FA enable: save secret + save recovery codes)
**Symptom:** If recovery code save fails after TOTP secret is already saved, user has 2FA enabled but no recovery codes
**Impact:** Partial state corruption possible
**Fix:** Wrap multi-step DB operations in transactions

## Low Priority

### BUG-006: Auto-Importer Error Detection Uses Hardcoded English Strings
**Location:** `internal/autoimport/importer.go:130-137`
**Symptom:** Uses `strings.Contains(lower, "unable to import")` etc.
**Impact:** May miss non-English error messages from arr apps
**Fix:** Use arr API status codes/enums instead of string matching

### BUG-007: Health Poller Results Not Exposed
**Location:** Health poller runs every 5min but results are NOT available via API
**Symptom:** Dashboard has to make individual health checks per instance (slow)
**Impact:** Performance — N+1 API calls from frontend when cached data is available
**Fix:** Add a `GET /instances/health` endpoint returning cached health status

### BUG-008: Inconsistent Error Responses
**Location:** Various API handlers
**Symptom:** Many handlers return generic "failed to create instance" (500) instead of specific messages (409 for conflict, 422 for validation)
**Impact:** Poor user experience — frontend can't show specific errors
**Fix:** Use appropriate HTTP status codes and pass through relevant error info
