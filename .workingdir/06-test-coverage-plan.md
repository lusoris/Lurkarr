# Test Coverage Plan — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Current State

### Backend (Go)
- **48 test files** across 20 packages
- **19/20 packages pass** (queuecleaner fails on Windows only — syscall.Stat_t)
- Good coverage on: api handlers, arrclient, auth, middleware, lurking engine
- Missing coverage: new auth features (sessions, admin, recovery), query files, server integration

### Frontend (SvelteKit)
- **0 test files** — no testing infrastructure set up
- No Vitest, no @testing-library/svelte

## Backend Test Gaps

### Priority 1: New Auth Features (No Tests)

#### `internal/api/sessions.go`
```
Tests needed:
- TestHandleListSessions_Success
- TestHandleListSessions_NoUser (401)
- TestHandleRevokeSession_Success
- TestHandleRevokeSession_NotOwned (404)
- TestHandleRevokeSession_NoUser (401)
- TestHandleRevokeAllSessions_Success
- TestHandleRevokeAllSessions_NoUser (401)
```

#### `internal/api/admin.go`
```
Tests needed:
- TestHandleListUsers_Success
- TestHandleListUsers_NonAdmin (403)
- TestHandleCreateUser_Success
- TestHandleCreateUser_WeakPassword
- TestHandleCreateUser_DuplicateUsername
- TestHandleCreateUser_NonAdmin
- TestHandleDeleteUser_Success
- TestHandleDeleteUser_Self (400)
- TestHandleDeleteUser_NonAdmin
- TestHandleResetUserPassword_Success
- TestHandleResetUserPassword_NonAdmin
- TestHandleToggleAdmin_Success
- TestHandleToggleAdmin_Self (400)
- TestHandleToggleAdmin_NonAdmin
```

#### `internal/auth/recovery.go`
```
Tests needed:
- TestGenerateRecoveryCodes_Returns10Codes
- TestGenerateRecoveryCodes_CodesAreUnique
- TestGenerateRecoveryCodes_HashedDifferFromPlain
- TestValidateRecoveryCode_ValidCode
- TestValidateRecoveryCode_InvalidCode
- TestValidateRecoveryCode_EmptyList
- TestValidateRecoveryCode_ReturnsCorrectIndex
```

### Priority 2: Database Query Tests

These are integration tests that would need a test database:

#### `internal/database/queries.go`
```
- TestCreateUser / TestGetUserByUsername / TestGetUserByID
- TestCreateSession / TestGetSession / TestDeleteSession
- TestListUserSessions / TestDeleteUserSessions
- TestCreateAppInstance / TestListInstances / TestDeleteInstance
- TestUpsertAppSettings / TestGetAppSettings
- TestUpsertGeneralSettings
- TestCheckProcessedItem / TestMarkProcessedItem / TestResetProcessedItems
- TestAddLurkHistory / TestListLurkHistory (with FTS)
- TestIncrementStats / TestResetStats
- TestCheckHourlyCap / TestIncrementHourlyCap
- TestSchedule CRUD
```

#### `internal/database/queries_blocklist.go`
```
- TestBlocklistSource CRUD
- TestBlocklistRule CRUD
- TestListEnabledRules
- TestDeleteRulesBySource
```

#### `internal/database/queries_notifications.go`
```
- TestNotificationProvider CRUD
- TestListEnabledProviders
```

### Priority 3: Missing Handler Tests

#### `internal/api/apps.go`
```
Additional tests:
- TestHandleTestConnection_SSRF (private IP blocked)
- TestHandleCreateInstance_DuplicateName
- TestHandleUpdateInstance_NotFound
```

#### `internal/api/settings.go`
```
- TestHandleGetGeneralSettings_MaskedSecretKey
- TestHandleSaveGeneralSettings_ValidationErrors
```

### Priority 4: Service Tests

#### `internal/notifications/` — provider-specific tests
```
- TestDiscordSend / TestDiscordTest
- TestTelegramSend / TestTelegramTest
- TestEmailSend / TestEmailTest
- etc. for all 8 providers (HTTP mock server)
```

## Frontend Test Plan

### Setup Required
```bash
cd frontend
npm install -D vitest @testing-library/svelte jsdom
```

Add to `vite.config.ts`:
```typescript
test: {
  environment: 'jsdom',
  include: ['src/**/*.test.ts'],
}
```

### Priority 1: API Client & Stores
```
src/lib/api.test.ts
- test CSRF token management
- test auto-redirect on 401
- test error handling (APIError)
- test GET/POST/PUT/DELETE methods

src/lib/stores/auth.svelte.test.ts
- test check() loads user
- test login() sends credentials
- test logout() clears state
- test 401 handling

src/lib/stores/toast.svelte.test.ts
- test success/error/info/warning create toasts
- test auto-dismiss timing
- test remove() clears specific toast
```

### Priority 2: UI Components
```
src/lib/components/ui/Button.test.ts
- test variants render correct classes
- test loading spinner
- test disabled state
- test click handler

src/lib/components/ui/Input.test.ts
- test value binding
- test label/hint/error rendering
- test disabled state

src/lib/components/ui/Modal.test.ts
- test open/close binding
- test escape key closes
- test backdrop click closes

src/lib/components/ui/Toggle.test.ts
- test checked binding
- test label rendering
```

### Priority 3: Page-Level Tests
```
Minimal smoke tests per page:
- Renders without error
- Shows loading state
- Displays data after fetch
- Handles API errors gracefully
```

## Estimated Effort

| Area | New Tests | Effort |
|------|-----------|--------|
| Backend: new auth handlers | ~25 tests | Medium |
| Backend: recovery codes | ~7 tests | Small |
| Backend: DB integration | ~40 tests | Large (needs test DB) |
| Backend: missing handler tests | ~15 tests | Medium |
| Frontend: setup | Infrastructure | Small |
| Frontend: API + stores | ~15 tests | Small |
| Frontend: components | ~15 tests | Small |
| Frontend: page smoke tests | ~13 tests | Medium |
| **Total** | **~130 tests** | |
