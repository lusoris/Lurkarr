# Codebase Audit Results

**Date**: Audited after pull to latest (bcf6f9c)
**Backend**: `go vet ./...` — PASS (0 issues)
**Backend tests**: `go test ./... -short` — 21 packages, ALL PASS
**Frontend tests**: `npx vitest run` — 8 files, 124 tests, ALL PASS

---

## Issues Found & Fixed

### Fixed: Frontend Test Failure — `@lucide/svelte` Not Installed
- **File**: `frontend/package.json`
- **Issue**: `@lucide/svelte` declared in `devDependencies` but not installed in `node_modules`. 27 shadcn-svelte components import from `@lucide/svelte/icons/*`. This blocked `more-components.test.ts` entirely (0 of 37 tests ran).
- **Fix**: `npm install` to install the declared dependency
- **Status**: ✅ Fixed

### Fixed: Frontend Test Failure — Select Custom Class Assertion
- **File**: `frontend/src/lib/components/ui/more-components.test.ts` line 73
- **Issue**: Test asserted `container.querySelector('label.my-select')` but `Select.svelte` applies `class` to a `<div>` wrapper, not a `<label>`.
- **Fix**: Changed selector to `'div.my-select'`
- **Status**: ✅ Fixed

---

## Audit Findings (Not Fixed — Actionable Items)

### Warning: Missing HTTP Security Headers
- **File**: `internal/server/server.go`
- **Issue**: No middleware setting standard security headers:
  - `X-Content-Type-Options: nosniff` — missing
  - `X-Frame-Options: DENY` — missing
  - `Content-Security-Policy` — missing
  - `Strict-Transport-Security` — missing (for TLS deployments)
- **Recommendation**: Add a single security headers middleware to the HTTP chain
- **Severity**: Warning

### Warning: uTorrent JSON Unmarshal Suppression
- **File**: `internal/downloadclients/torrent/utorrent/client.go` ~lines 173-178
- **Issue**: `_ = json.Unmarshal(...)` on 6+ fields (Status, Size, Progress, Downloaded, Uploaded, Ratio). Silent data corruption if API response format changes.
- **Recommendation**: Log unmarshal failures; don't silently use zero values
- **Severity**: Warning

### Info: Dual Lucide Package Dependency
- **File**: `frontend/package.json`
- **Issue**: Both `lucide-svelte` (v0.577.0 in dependencies) and `@lucide/svelte` (v0.561.0 in devDependencies) are declared. shadcn-svelte components use `@lucide/svelte`, custom components use `lucide-svelte`.
- **Recommendation**: Consolidate to `@lucide/svelte` (the official scoped package going forward) and migrate the 25 `lucide-svelte` imports
- **Severity**: Info

### Info: Readarr is Officially Retired
- **Discovery**: Servarr team has officially retired Readarr (metadata unusable)
- **Impact**: Lurkarr still supports Readarr connections
- **Recommendation**: Consider adding a UI notice when Readarr instances are configured; gracefully handle degraded functionality
- **Severity**: Info

---

## Positive Findings (Good Practices Confirmed)

| Area | Status |
|------|--------|
| SQL injection prevention | ✅ All queries parameterized via pgx ($1, $2 placeholders) |
| Context threading | ✅ All external calls accept context.Context as first param |
| No stored contexts in structs | ✅ Clean context lifecycle |
| HTTP client timeouts | ✅ All clients have proper timeouts (10-30s) |
| Graceful shutdown | ✅ All components have fx.Lifecycle OnStop hooks |
| Rate limiting | ✅ Per-IP login (5/min) and API (300/min) rate limits |
| CSRF protection | ✅ gorilla/csrf double-submit cookie pattern |
| Password hashing | ✅ bcrypt with proper cost |
| No sensitive data in logs | ✅ No passwords, tokens, or keys logged |
| Error handling discipline | ✅ Consistent log-or-return, proper error wrapping |
| Structured logging | ✅ slog with key-value pairs throughout |
| Multi-auth strategy | ✅ Local + TOTP + OIDC + Proxy + WebAuthn |
| Test coverage | ✅ 21 Go packages + 8 frontend test files |
