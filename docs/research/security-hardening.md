# Go Security Hardening Best Practices

## OWASP Top 10 — Go-Specific Mitigations

### 1. Broken Access Control
- Session-based auth with CSRF tokens (gorilla/csrf) ✅
- Role-based checks per endpoint
- Session rotation after login ✅
- httpOnly + Secure + SameSite cookies

### 2. Cryptographic Failures
- Use `crypto/rand` for secrets, tokens, keys ✅
- bcrypt/argon2 for password hashing (golang.org/x/crypto) ✅
- TLS 1.2+ minimum in HTTP clients
- Never log API keys or passwords

### 3. Injection
- **SQL:** Use parameterized queries with pgx (`$1, $2...`) ✅
- **XSS:** Escape output in templates, CSP headers
- **Command Injection:** Never use `os/exec` with user input
- **SSRF:** Validate URLs, block private IPs ✅

### 4. Insecure Design
- Input validation at API boundaries ✅
- Rate limiting on auth endpoints ✅
- MaxBytesReader on all request bodies ✅
- Timeouts on all HTTP clients and DB connections

### 5. Security Misconfiguration
- Minimal Docker image (scratch) ✅
- No debug endpoints in production
- Secure defaults for all settings
- CORS restricted to allowed origins ✅

### 6. Vulnerable Components
- Dependabot for dependency updates ✅
- `govulncheck` for Go vulnerability scanning
- Pin dependency versions in go.mod ✅

### 7. Authentication Failures
- TOTP-based 2FA ✅
- Rate limiting on login (5/min/IP) ✅
- Account lockout after failed attempts
- Constant-time comparison for secrets

### 8. Data Integrity
- CSRF protection on all state-changing endpoints ✅
- Verify webhook signatures
- Goose for migration integrity

### 9. Logging & Monitoring
- Structured logging with slog ✅
- Prometheus metrics for security events ✅
- Rate limit hit tracking ✅
- Don't log sensitive data

### 10. SSRF
- Block private IPs in test-connection endpoints ✅
- URL validation + allowlists for external connections
- DNS rebinding protection

## Go-Specific Security Patterns

### HTTP Client Hardening

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### Request Body Limiting

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
```

### Context Timeouts

```go
ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
defer cancel()
```

### Password Hashing

```go
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword(hash, []byte(input))
```

### Constant-Time Comparison

```go
import "crypto/subtle"
if subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 1 {
    // not equal
}
```

## Security Scanning Tools

| Tool | Purpose |
|------|---------|
| `govulncheck` | Go vulnerability database check |
| `gosec` | Go security linter |
| `staticcheck` | Static analysis (includes security rules) |
| `trivy` | Container image vulnerability scanning |
| `golangci-lint` | Meta-linter (includes gosec, staticcheck) |

## Docker Security

- Use `scratch` or `distroless` base ✅
- Run as non-root (add USER instruction)
- No shell in production image ✅
- Copy only the binary + certs ✅
- Set `GOFLAGS=-trimpath` for reproducible builds
- Use `CGO_ENABLED=0` ✅

## Database Security

- Connection via IAM or strong passwords
- SSL/TLS for DB connections (sslmode=require in production)
- Connection pool limits to prevent exhaustion ✅
- Prepared statements (pgx uses them by default)
- Migration checksums (goose tracks them)
