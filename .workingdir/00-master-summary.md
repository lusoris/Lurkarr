# Lurkarr Codebase Analysis — Master Summary

## Overall Assessment: 8/10 — Production-quality with targeted improvement opportunities

---

## Analysis Files

| File | Scope |
|------|-------|
| `01-go-package-optimization.md` | Dependency evaluation, stdlib vs external libraries |
| `02-code-deduplication.md` | 700+ lines of identified duplication patterns |
| `03-ui-ux-completeness.md` | Frontend ~85% complete, 4 critical coverage gaps |
| `04-codebase-quality.md` | Security, testing, performance, architecture |

---

## Prioritized Action Plan

### 🔴 Priority 1: Security Hardening

| # | Action | Effort | Source |
|---|--------|--------|--------|
| 1 | Add general API rate limiting (beyond login) | 3h | Quality §2 |
| 2 | Review all slog statements for secrets leakage | 2h | Quality §2 |
| 3 | Add transaction isolation to concurrent updates | 4h | Quality §8 |

### 🟠 Priority 2: UI Coverage Gaps

| # | Action | Effort | Source |
|---|--------|--------|--------|
| 4 | Seerr Requests page — request list, filters, reassignment | 8h | UI/UX §3 |
| 5 | SABnzbd integration in /downloads — queue, history, controls | 6h | UI/UX §3 |
| 6 | Prowlarr indexer list in Connections modal | 4h | UI/UX §4 |
| 7 | Dry-run mode visual toggle + warning banner | 2h | UI/UX §3 |
| 8 | State management UI (explain state vs stats) | 3h | UI/UX §3 |
| 9 | Whisparr v2/v3 labels (show full version names) | 1h | UI/UX §4 |

### 🟡 Priority 3: Code Deduplication (High-Impact)

| # | Action | Lines Saved | Effort | Source |
|---|--------|-------------|--------|--------|
| 10 | Extract `decodeJSON[T]()` generic helper | 105+ | 1h | Dedup §2.1 |
| 11 | Extract `parseUUIDFromPath()` helper | 50+ | 30m | Dedup §2.2 |
| 12 | Extract `getAndValidateAppType()` helper | 40+ | 30m | Dedup §2.4 |
| 13 | Standardize DB scans to `pgx.CollectRows` | 96+ | 3h | Dedup §5.1 |
| 14 | Extract download client `filterCompletedItems()` | 70 | 1h | Dedup §1.1 |

### 🟢 Priority 4: Developer Experience

| # | Action | Effort | Source |
|---|--------|--------|--------|
| 15 | Adopt testify/assert in tests | 8h | Packages §8 |
| 16 | Consolidate HTTP client factory | 4h | Packages §4 |
| 17 | Fix double-unmarshal in download clients | 4h | Packages §2 |
| 18 | Add errgroup for parallel operations | 2h | Packages §9 |
| 19 | Create scanning helpers for database | 3h | Packages §7 |
| 20 | Add gzip fallback compression | 2h | Packages §10 |

### 🔵 Priority 5: UX Polish

| # | Action | Effort | Source |
|---|--------|--------|--------|
| 21 | Sidebar navigation grouping | 4h | UI/UX §5 |
| 22 | Dashboard alerts (degraded health, failed schedules) | 4h | UI/UX §6 |
| 23 | Instance group mode tooltips | 1h | UI/UX §4 |
| 24 | Blocklist regex tester | 3h | UI/UX §4 |
| 25 | Schedule history inline display | 2h | UI/UX §4 |

---

## Key Numbers

| Metric | Value |
|--------|-------|
| Total backend endpoints | 80+ |
| Frontend routes | 15 |
| Frontend coverage | ~85% |
| Code duplication found | 700+ lines |
| Test files | 69 |
| Direct dependencies | 29 |
| Critical security issues | 0 |
| High security issues | 1 (rate limiting) |
| Major dep changes needed | 0 |

---

## What NOT To Do

- ❌ Don't switch to Chi/Echo router (stdlib ServeMux is sufficient)
- ❌ Don't add sonic/go-json (payloads too small to benefit)
- ❌ Don't switch to zap/logrus (slog is fine at this scale)
- ❌ Don't add Viper/envconfig (env vars are appropriate for cloud-native)
- ❌ Don't migrate to sqlc (too disruptive for 100+ query functions)
- ❌ Don't add go-playground/validator (lightweight helpers suffice)

---

## Architecture Strengths

1. **Clean separation**: internal/ packages well-organized, no circular deps
2. **Consistent patterns**: Error handling, API responses, status codes
3. **Security-first**: Bcrypt, CSRF, URL validation, body limits, parameterized SQL
4. **Container-native**: Multi-stage Docker, health probes, structured logging
5. **Minimal dependencies**: 29 direct deps, all actively maintained
6. **DI framework**: Uber fx provides clean lifecycle management
7. **Background services**: Proper context cancellation, graceful shutdown
