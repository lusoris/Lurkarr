# Code Deduplication Analysis

## Summary

**Total estimated duplication: 700+ lines of nearly identical code** across download clients, API handlers, background services, database queries, and tests.

---

## 1. Download Client Adapters — HIGH PRIORITY

### 1.1 GetHistory Filter Pattern (70 lines)

**Files**: All 5 torrent adapters (qbittorrent, transmission, deluge, rtorrent, utorrent)

All torrent client adapters implement identical GetHistory filtering:
```
Get all items → filter where Progress >= 1.0 → return completed items
```

~14 lines per adapter × 5 adapters = 70 lines duplicated

**Fix**: Extract to shared helper:
```go
func filterCompletedItems(items []DownloadItem) []DownloadItem {
    var completed []DownloadItem
    for _, item := range items {
        if item.Progress >= 1.0 {
            completed = append(completed, item)
        }
    }
    return completed
}
```

### 1.2 Pause/Resume Iteration Pattern (112 lines)

**Files**: transmission, deluge, rtorrent, utorrent adapters

All adapters repeat nearly identical PauseAll/ResumeAll patterns:
- Get all torrents
- Extract IDs into slice
- Call client pause/resume

~28 lines per adapter × 4 adapters = 112 lines

**Fix**: Create a mixin/wrapper with generic pause/resume via callback:
```go
type GenericPauseResumeAdapter struct {
    GetAllIDs func(ctx context.Context) ([]string, error)
    DoPause   func(ctx context.Context, ids []string) error
    DoResume  func(ctx context.Context, ids []string) error
}
```

### 1.3 Status Mapping Functions (25 lines)

**Files**: transmission, rtorrent, utorrent adapters

Each maps client-specific status codes to normalized strings independently via similar switch statements.

**Fix**: Create a status mapper registry pattern

---

## 2. API Handlers — HIGH PRIORITY

### 2.1 Request Parsing & Validation Boilerplate (105+ lines)

**Files**: apps.go, download_clients.go, notifications.go, blocklist.go, queue.go, instance_groups.go, + 10 more

Identical pattern repeats 15+ times:
```go
limitBody(w, r)
var req struct { ... }
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
    return
}
```

~7 lines × 15+ handlers = 105+ lines

**Fix**: Generic decoder:
```go
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
    limitBody(w, r)
    var req T
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
        return req, false
    }
    return req, true
}
```

### 2.2 UUID Parsing Error Pattern (50+ lines)

**Files**: apps.go, download_clients.go, blocklist.go, notifications.go, instance_groups.go, + 5 more

Identical UUID parsing with identical error response appears 10+ times:
```go
id, err := uuid.Parse(r.PathValue("id"))
if err != nil {
    writeJSON(w, http.StatusBadRequest, errorResponse("invalid [type] ID"))
    return
}
```

**Fix**:
```go
func parseUUIDFromPath(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
    id, err := uuid.Parse(r.PathValue(param))
    if err != nil {
        writeJSON(w, http.StatusBadRequest, errorResponse("invalid "+param))
        return uuid.UUID{}, false
    }
    return id, true
}
```

### 2.3 App Type Validation Pattern (40+ lines)

**Files**: queue.go, apps.go, history.go, instance_groups.go, + 4 more

Same validation appears 8+ times:
```go
appType := r.PathValue("app")
if !database.ValidAppType(appType) {
    writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
    return
}
```

**Fix**:
```go
func getAndValidateAppType(w http.ResponseWriter, r *http.Request) (database.AppType, bool)
```

### 2.4 Mask Secrets Pattern

**Files**: apps.go, download_clients.go, notifications.go

Handlers mask API keys/passwords identically after retrieval. Should move masking to response mapper methods.

---

## 3. ArrClient Methods — MEDIUM PRIORITY

### 3.1 GetMissing/GetCutoffUnmet Pattern (70 lines)

**Files**: radarr.go, sonarr.go, lidarr.go, readarr.go, whisparr.go

Every app type repeats nearly identical methods:
```go
func (c *Client) [AppName]GetMissing(ctx context.Context) ([]Record, error) {
    records, err := getAllPages[Record](ctx, c, apiVersion+"/wanted/missing?...")
    if err != nil {
        return nil, fmt.Errorf("[app] get missing: %w", err)
    }
    return records, nil
}
```

~7 lines × 5 apps × 2 methods = 70 lines

**Fix**: Parameterized method factory using generics

### 3.2 Search Command Pattern

**Files**: sonarr.go, radarr.go

Similar search command building pattern (~8 lines × 3 search methods)

**Fix**: Extract to a command executor method

### 3.3 GetQueue/GetQueueEnriched Pattern (16 lines)

**Files**: radarr.go, sonarr.go

Nearly identical except for query parameters.

---

## 4. Background Service Loop Patterns — MEDIUM PRIORITY

### 4.1 Instance Iteration Pattern (80+ lines)

**Files**: queuecleaner/cleaner.go, autoimport/importer.go, lurking/engine.go, healthpoller/poller.go

All 4 services repeat identical instance iteration+setup:
1. For each app type → get enabled instances
2. Handle error logging
3. For each instance → create client → call operation

~20 lines per service × 4 services = 80+ lines

**Fix**: Create iteration utility:
```go
func IterateEnabledInstances(ctx context.Context, db Store, appType AppType,
    callback func(inst AppInstance) error) error
```

### 4.2 Client Creation Pattern (20 lines)

**Files**: autoimport/importer.go, healthpoller/poller.go, queuecleaner/cleaner.go, lurking/engine.go

All repeat:
```go
client := arrclient.NewClient(
    inst.APIURL, inst.APIKey,
    time.Duration(genSettings.APITimeout)*time.Second,
    genSettings.SSLVerify,
)
```

~5 lines × 4 places

**Fix**: Factory method in config/settings module

---

## 5. Database Query Scan Pattern — MEDIUM PRIORITY

### 5.1 List + Scan Loop Pattern (96+ lines)

**Files**: queries_blocklist.go, queries_notifications.go, queries_download_clients.go, + 5 more

Same scan loop pattern repeats across 8+ query functions:
```go
rows, err := db.Pool.Query(ctx, `SELECT ...`)
if err != nil { return nil, fmt.Errorf(...) }
defer rows.Close()
var items []ItemType
for rows.Next() {
    var item ItemType
    if err := rows.Scan(...); err != nil { ... }
    items = append(items, item)
}
return items, rows.Err()
```

~12 lines × 8+ functions = 96+ lines

**Fix**: Standardize on `pgx.CollectRows` (already used in some queries):
```go
return pgx.CollectRows(rows, pgx.RowToStructByPos[Item])
```

---

## 6. Test Helper Duplication — MEDIUM PRIORITY

### 6.1 Handler Test Setup (120+ lines)

**Files**: api_test.go, activity_test.go, download_clients_test.go, + 5 more

Each test file duplicates mock setup:
```go
ctrl := gomock.NewController(t)
store := NewMockStore(ctrl)
// ... multiple EXPECT().Return(...) setup lines
h := &SomeHandler{DB: store}
w := httptest.NewRecorder()
h.SomeMethod(w, httptest.NewRequest(...))
```

~15 lines per test file × 8 files = 120+ lines

**Fix**: Create test helper factory per handler type

---

## Priority Summary

| Category | Pattern | Occurrences | Lines | Priority |
|----------|---------|-------------|-------|----------|
| Download Client: GetHistory | Identical filter | 5 adapters | 70 | HIGH |
| Download Client: Pause/Resume | Iteration & ID extraction | 4 adapters | 112 | HIGH |
| API Handler: Decode JSON | Request parsing | 15+ handlers | 105+ | HIGH |
| API Handler: UUID parsing | Error handling | 10+ places | 50+ | HIGH |
| API Handler: App type validation | Validation boilerplate | 8+ handlers | 40+ | HIGH |
| ArrClient: GetMissing/GetCutoff | Generic methods | 10 methods | 70 | MEDIUM |
| Background Services: Instance loops | Iteration pattern | 4 services | 80+ | MEDIUM |
| Database: Scan loops | Row iteration | 8+ queries | 96+ | MEDIUM |
| Test Helpers | Mock setup | 8+ files | 120+ | MEDIUM |

## Quick Wins (Easiest to Implement)

1. **Extract API handler UUID parsing helper** — 50+ lines saved, trivial change
2. **Extract app type validation** — 40+ lines saved, trivial change
3. **Standardize database scans to use `pgx.CollectRows`** — 96 lines consolidated
4. **Create request decoder helper** — 105+ lines saved, uses generics
5. **Extract GetHistory filter** — 70 lines saved, shared helper function
