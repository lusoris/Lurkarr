# Backend Best Practices — Deep Research Reference

## 1. Go Best Practices

### 1.1 Error Handling
- **Wrap errors** with context: `fmt.Errorf("failed to fetch queue: %w", err)`
- **Use sentinel errors** for expected conditions: `var ErrNotFound = errors.New("not found")`
- **Check with `errors.Is`/`errors.As`** instead of string comparison
- **Never ignore errors** — `_ = foo()` is a code smell; log or handle explicitly
- **Return early** on error — avoid deep nesting
- **Don't log and return** — do one or the other, not both (except at boundaries)

### 1.2 Context Usage
- **Always accept `context.Context`** as first parameter: `func Fetch(ctx context.Context, ...) error`
- **Set timeouts** for external calls: `ctx, cancel := context.WithTimeout(ctx, 30*time.Second)`
- **Always `defer cancel()`** after creating a derived context
- **Pass context through** — don't store in structs
- **Use `context.WithoutCancel`** when background work must outlive the caller

### 1.3 Concurrency
- **Use sync primitives correctly**: `sync.Mutex` for shared state, `sync.Once` for init
- **Prefer channels for communication**, mutexes for state protection
- **Always `defer mu.Unlock()`** after `mu.Lock()`
- **Use `errgroup.Group`** for concurrent operations with error propagation
- **Avoid goroutine leaks** — ensure all goroutines have an exit path via context or done channels
- **Use `select` with `ctx.Done()`** in long-running goroutines

### 1.4 Struct Design
- **Unexport struct fields** unless they need external access
- **Use functional options pattern** for complex constructors:
  ```go
  func NewClient(opts ...Option) *Client
  ```
- **Prefer composition** over inheritance (embedding)
- **Use interfaces for abstraction** — define them where they're consumed, not where implemented

### 1.5 Testing
- **Table-driven tests**: Use `[]struct{name string; ...}` pattern
- **Use `t.Run`** for subtests
- **Use `t.Parallel()`** where safe
- **Test behavior, not implementation** — assert on outputs, not internal calls
- **Integration tests**: Use `testcontainers` for real database testing
- **Test file naming**: `*_test.go` in same package (white-box) or `_test` package (black-box)
- **Generate mocks**: `mockgen` for interfaces at boundaries

---

## 2. Database Best Practices (PostgreSQL + pgx)

### 2.1 Connection Management
- **Always use `pgxpool.Pool`** — never share a single `pgx.Conn`
- **Configure pool size**: `pool_max_conns` ≈ 2-3× CPU cores, `pool_min_conns` ≈ 2-5
- **Set connection lifetime**: `pool_max_conn_lifetime=1h` to rotate connections
- **Health check period**: `pool_health_check_period=30s`

### 2.2 Query Patterns
- **Use parameterized queries** — NEVER concatenate SQL strings (injection prevention)
- **Use `pgx.CollectRows`** with `pgx.RowToStructByName` for scanning
- **Use `COALESCE`** for nullable columns to avoid nil pointer issues
- **Batch queries** with `pgx.Batch` for multiple operations
- **Use transactions** (`pool.BeginTx`) for multi-statement operations
- **Prefer `RETURNING`** clause over separate SELECT after INSERT/UPDATE

### 2.3 Migration Best Practices
- **Timestamp-based naming**: `YYYYMMDDHHMMSS_description.sql`
- **Never modify** deployed migrations — create new corrective migration
- **Always include Down** migration for reversibility
- **Use `IF NOT EXISTS`** for CREATE operations
- **Avoid `DROP COLUMN`** in production — use soft deprecation
- **Add columns as nullable first**, then backfill, then add NOT NULL constraint

### 2.4 Performance
- **Create indexes** for frequently queried columns: `CREATE INDEX CONCURRENTLY`
- **Use `EXPLAIN ANALYZE`** to verify query plans
- **Avoid `SELECT *`** — select only needed columns
- **Use pagination** with `LIMIT`/`OFFSET` or keyset pagination for large datasets

---

## 3. HTTP/API Best Practices

### 3.1 API Design
- **RESTful resource naming**: plural nouns (`/api/v1/apps`, `/api/v1/notifications`)
- **HTTP methods**: GET (read), POST (create/action), PUT (full update), PATCH (partial), DELETE
- **Status codes**: 200 OK, 201 Created, 204 No Content, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 422 Validation Error, 500 Internal Server Error
- **Error response format**: Consistent JSON `{"error": "message", "code": "ERROR_CODE"}`
- **Pagination**: Use `page`/`pageSize` or `cursor`-based with `Link` headers

### 3.2 Security (OWASP)
- **CSRF Protection**: Double-submit cookie pattern (gorilla/csrf) for all state-changing endpoints
- **CORS**: Restrictive `Access-Control-Allow-Origin` — no wildcards in production
- **Rate Limiting**: Per-IP and per-user rate limits on auth endpoints
- **Input Validation**: Validate all input at API boundary; reject unknown fields
- **Authentication**: bcrypt for passwords, TOTP for 2FA, OIDC for SSO, WebAuthn for passkeys
- **Session Management**: Secure, HttpOnly, SameSite=Strict cookies
- **Headers**: Set `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Strict-Transport-Security`
- **SQL Injection**: Always use parameterized queries (pgx handles this)
- **XSS**: SvelteKit auto-escapes; never use `{@html}` with user content

### 3.3 Middleware Chain
Recommended order:
1. Recovery/Panic handler
2. Request ID generation
3. Logging
4. Metrics (Prometheus)
5. CORS
6. Rate limiting
7. Authentication
8. CSRF
9. Authorization
10. Request handler

---

## 4. Architecture Patterns

### 4.1 Dependency Injection (uber/fx)
- **Module per domain**: config, database, logging, server, scheduler, notifications, services
- **Lifecycle management**: `fx.Lifecycle` hooks for graceful start/stop
- **Constructor pattern**: Functions that accept dependencies, return interface implementations
- **Interface boundaries**: Define in consumer package, implement in provider
- **Testing**: Use `fx.Replace` or `fx.Decorate` to swap implementations

### 4.2 Clean Architecture Layers
1. **API Layer** (`internal/api/`): HTTP handlers, request/response mapping, validation
2. **Service Layer** (`internal/lurking/`, `internal/queuecleaner/`, etc.): Business logic
3. **Store Layer** (`internal/database/`): Data access, SQL queries
4. **Client Layer** (`internal/arrclient/`, `internal/downloadclients/`): External API communication
5. **Config Layer** (`internal/config/`): Environment configuration

### 4.3 Graceful Shutdown
- Listen for `SIGINT`/`SIGTERM`
- Stop accepting new requests
- Drain in-flight requests (timeout)
- Close database connections
- Stop scheduler jobs
- Log shutdown completion
- uber/fx handles this via lifecycle hooks

### 4.4 Configuration
- **Environment variables** for all config (12-factor app)
- **Sensible defaults** for optional config
- **Validation at startup** — fail fast on invalid config
- **Never log secrets** — mask API keys, passwords in logs

---

## 5. Logging Best Practices

### 5.1 Structured Logging
- **Use `slog`** (Go 1.21+ standard library) or `zerolog`/`zap`
- **Key-value pairs**: `slog.Info("fetched queue", "app", appName, "count", len(items))`
- **Log levels**: Debug (verbose), Info (normal operations), Warn (recoverable issues), Error (failures)
- **Consistent field names**: `app`, `client`, `duration`, `error`, `count`, `status`

### 5.2 What to Log
- **DO log**: Service startup/shutdown, config validation (no secrets), external API calls (method + URL + status + duration), errors with context, scheduled job execution, authentication events
- **DON'T log**: Passwords, API keys, tokens, full request/response bodies, PII

### 5.3 Observability Correlation
- **Request IDs**: Generate UUID per request, include in all logs and error responses
- **Trace through services**: Pass request ID to external API calls
- **Prometheus labels**: Match log field names for Grafana correlation with Loki

---

## 6. Scheduling Best Practices

### 6.1 Job Design
- **Idempotent jobs**: Running twice should produce same result
- **Singleton mode**: Prevent overlapping executions
- **Timeout**: Every job should have a context timeout
- **Error handling**: Log failures, don't crash the scheduler
- **Backoff**: Exponential backoff for retryable failures

### 6.2 Queue Cleaner — Strike System
- **Pattern**: Progressive penalty system instead of immediate removal
- **Strikes**: Track failures per queue item, remove after threshold
- **Benefits**: Tolerant of temporary issues, only removes persistent problems
- **Documentation**: ADR-004

---

## 7. Docker Best Practices

### 7.1 Multi-Stage Build
- **Stage 1**: Build frontend (`node:lts-alpine`)
- **Stage 2**: Build Go binary (`golang:1.26-alpine`)
- **Stage 3**: Runtime (`alpine:3.x` — minimal attack surface)
- **Best Practices**:
  - Pin base image versions
  - Run as non-root user
  - Use `.dockerignore` to exclude dev files
  - Single binary with embedded frontend
  - Health check via `HEALTHCHECK CMD`

### 7.2 Compose Configuration
- **Separate compose files**: `docker-compose.yml` (production), `docker-compose.dev.yml` (development)
- **Environment variables**: Use `.env` file, never hardcode secrets
- **Volumes**: Named volumes for persistent data (PostgreSQL)
- **Networks**: Isolated network for service communication
