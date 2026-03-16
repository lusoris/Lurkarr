# Go Packages — Deep Research Reference

## Direct Dependencies (go.mod)

### 1. Database & Migrations

#### `github.com/jackc/pgx/v5` v5.8.0
- **Purpose**: PostgreSQL driver and toolkit (pure Go)
- **Key Features**: Connection pooling (`pgxpool`), `pgtype` for rich PG types, COPY protocol, prepared statements, batch queries, row scanning
- **Best Practices**:
  - Always use `pgxpool.Pool` for concurrent access (not `pgx.Conn`)
  - Use context timeouts on every query
  - Prefer `pgx.CollectRows` / `pgx.RowTo*` helpers for scanning
  - Use prepared statements for frequently-executed queries
  - Handle `pgx.ErrNoRows` explicitly
  - Connection string format: `postgres://user:pass@host:5432/db?sslmode=disable`
  - Pool config: `pool_max_conns=25`, `pool_min_conns=5`

#### `github.com/pressly/goose/v3` v3.27.0
- **Purpose**: Database migration tool
- **Migration Naming**: `YYYYMMDDHHMMSS_description.sql` (timestamp-based)
- **Best Practices**:
  - Never modify existing migrations — always create new ones
  - Use `-- +goose Up` and `-- +goose Down` annotations
  - Use `-- +goose NO TRANSACTION` for DDL that can't run in transactions
  - Keep migrations atomic and reversible when possible
  - Use `goose.SetDialect("postgres")` before running

### 2. Dependency Injection

#### `go.uber.org/fx` v1.24.0
- **Purpose**: Dependency injection framework
- **Pattern Used**: Module-based organization with `fx.Module`, `fx.Provide`, `fx.Invoke`
- **Best Practices**:
  - One module per domain area (config, database, server, etc.)
  - Use `fx.Lifecycle` for startup/shutdown hooks
  - Use `fx.In`/`fx.Out` structs for multi-dependency constructors
  - Prefer `fx.Annotate` with `fx.As` for interface bindings
  - Use `fx.Supply` for value types, `fx.Provide` for constructors
  - Avoid `fx.Replace` unless testing
  - Lifecycle hooks: `OnStart` runs in order, `OnStop` runs in reverse

### 3. Web Framework & Middleware

#### `github.com/gorilla/csrf` v1.7.3
- **Purpose**: CSRF protection middleware
- **Pattern**: Double-submit cookie with SPA token header
- **Best Practices**:
  - Use `csrf.Protect(key, csrf.Secure(true))` in production
  - Expose CSRF token via dedicated GET endpoint
  - Frontend sends token in `X-CSRF-Token` header
  - Set `csrf.SameSite(csrf.SameSiteStrictMode)` for additional protection
  - Use `csrf.TrustedOrigins` for cross-origin if needed

#### `golang.org/x/time` (rate)
- **Purpose**: Rate limiting via `rate.Limiter`
- **Usage**: Token bucket algorithm
- **Best Practices**:
  - Per-IP rate limiting for login endpoints
  - Per-user rate limiting for API endpoints
  - Use `rate.NewLimiter(rate.Every(time.Second), burst)`
  - Check with `limiter.Allow()` or `limiter.Wait(ctx)`

### 4. Authentication

#### `github.com/pquerna/otp` v1.5.0
- **Purpose**: TOTP/HOTP one-time password generation and validation
- **Best Practices**:
  - Use `totp.GenerateOpts` with Issuer and AccountName
  - Store secret encrypted at rest
  - Allow 1-step time skew (`totp.ValidateOpts{Skew: 1}`)

#### `github.com/go-webauthn/webauthn` v0.16.1
- **Purpose**: WebAuthn/FIDO2 passkey authentication
- **Best Practices**:
  - Configure `RPDisplayName`, `RPID`, `RPOrigins` correctly
  - Store credential data (ID, public key, sign count) in database
  - Validate sign count to detect cloned authenticators
  - Support multiple credentials per user

#### `github.com/coreos/go-oidc/v3` v3.14.1
- **Purpose**: OpenID Connect client (OIDC)
- **Best Practices**:
  - Use `oidc.NewProvider` for auto-discovery via `.well-known/openid-configuration`
  - Verify ID tokens with `provider.Verifier`
  - Use PKCE flow (`S256` code challenge) for SPAs
  - Store nonce in session and validate

#### `golang.org/x/oauth2` v0.30.0
- **Purpose**: OAuth2 client library (used with go-oidc)
- **Best Practices**:
  - Use `oauth2.Config` with `AuthCodeURL` and `Exchange`
  - Always set `state` parameter and validate on callback
  - Use `httpClient` with timeout via context

#### `golang.org/x/crypto` v0.38.0
- **Purpose**: bcrypt password hashing
- **Best Practices**:
  - Use `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)` (cost >= 10)
  - Compare with `bcrypt.CompareHashAndPassword`
  - Never store plain passwords; always hash before storage

### 5. Scheduling

#### `github.com/go-co-op/gocron/v2` v2.19.1
- **Purpose**: Job scheduling with cron expressions
- **Best Practices**:
  - Use `gocron.CronJob(expression, false)` for cron-style schedules
  - Use `gocron.DurationJob(interval)` for recurring intervals
  - Handle panics in jobs (use `gocron.WithEventListeners`)
  - Use WithSingletonMode for non-overlapping jobs
  - Graceful shutdown: `scheduler.StopJobs()` then `scheduler.Shutdown()`
  - Provide descriptive names via `gocron.WithName`

### 6. Observability

#### `github.com/prometheus/client_golang` v1.23.2
- **Purpose**: Prometheus metrics instrumentation
- **Metric Types**:
  - `Counter`: monotonically increasing (requests, errors)
  - `Gauge`: current value (queue size, active connections)
  - `Histogram`: distributions (request duration)
- **Best Practices**:
  - Follow naming: `lurkarr_<subsystem>_<name>_<unit>`
  - Use `prometheus.MustRegister` or custom registries
  - Avoid high-cardinality labels
  - Use `promhttp.Handler()` for `/metrics` endpoint

### 7. Development & Testing

#### `go.uber.org/mock` v0.5.2
- **Purpose**: Mock generation for interfaces
- **Usage**: `mockgen` CLI generates mock implementations
- **Best Practices**:
  - Use `//go:generate mockgen` directives
  - Mock at interface boundaries, not concrete types
  - Use `gomock.InOrder` for ordered expectations
  - Use `gomock.Any()` for flexible matching
  - Clean up: `ctrl.Finish()` via `defer`

#### `github.com/testcontainers/testcontainers-go` v0.37.0
- **Purpose**: Docker-based integration testing
- **Best Practices**:
  - Use for database integration tests (real PostgreSQL)
  - `testcontainers.GenericContainer` with `postgres:17` image
  - Cleanup with `defer container.Terminate(ctx)`
  - Use `container.MappedPort` for dynamic port allocation
  - Parallel tests with separate containers

### 8. Utilities

#### `github.com/skip2/go-qrcode` v0.0.0
- **Purpose**: QR code generation (for TOTP setup)
- **Usage**: Generate QR code PNG from TOTP URI

#### `github.com/autobrr/go-rtorrent` v1.12.0
- **Purpose**: rTorrent XMLRPC client
- **Usage**: Connect to rTorrent for download client integration

### 9. Indirect but Critical Dependencies

#### `github.com/fxamacker/cbor/v2` v2.8.0
- **Purpose**: CBOR encoding for WebAuthn attestation/assertion

#### `github.com/go-jose/go-jose/v4` v4.0.4
- **Purpose**: JOSE/JWT handling for OIDC tokens

#### `github.com/jackc/pgpassfile`, `github.com/jackc/pgservicefile`
- **Purpose**: PostgreSQL password and service file support (pgx internals)

---

## Dependency Version Matrix

| Package | Current | Purpose |
|---------|---------|---------|
| pgx/v5 | 5.8.0 | PostgreSQL driver |
| goose/v3 | 3.27.0 | DB migrations |
| uber/fx | 1.24.0 | Dependency injection |
| gorilla/csrf | 1.7.3 | CSRF protection |
| gocron/v2 | 2.19.1 | Job scheduling |
| prometheus | 1.23.2 | Metrics |
| go-webauthn | 0.16.1 | WebAuthn/Passkeys |
| go-oidc/v3 | 3.14.1 | OIDC/SSO |
| pquerna/otp | 1.5.0 | TOTP 2FA |
| uber/mock | 0.5.2 | Mocking |
| testcontainers | 0.37.0 | Integration tests |
| x/crypto | 0.38.0 | bcrypt |
| x/oauth2 | 0.30.0 | OAuth2 |
| x/time | 0.11.0 | Rate limiting |
