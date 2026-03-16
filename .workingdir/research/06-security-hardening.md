# Security Hardening — Deep Research Reference

## 1. OWASP Top 10 Compliance

### A01:2021 — Broken Access Control
- **Lurkarr Measures**:
  - Session-based authentication with secure cookies
  - Admin-only endpoints protected by middleware
  - CSRF protection on all state-changing endpoints
  - Rate limiting on authentication endpoints
- **Best Practices**:
  - Deny by default — explicitly allowlist public routes
  - Validate authorization on every endpoint, not just in middleware
  - Don't expose internal IDs that allow enumeration
  - Log all access control failures

### A02:2021 — Cryptographic Failures
- **Lurkarr Measures**:
  - bcrypt for password hashing (cost factor ≥ 10)
  - TOTP secrets stored securely
  - CSRF key as 32-byte random value
- **Best Practices**:
  - Use TLS in production (terminate at reverse proxy)
  - Never log API keys, passwords, or tokens
  - Use constant-time comparison for secrets (`subtle.ConstantTimeCompare`)
  - Rotate CSRF keys periodically

### A03:2021 — Injection
- **Lurkarr Measures**:
  - pgx uses parameterized queries exclusively
  - No SQL string concatenation
  - SvelteKit auto-escapes HTML output
- **Best Practices**:
  - Never use `fmt.Sprintf` for SQL queries
  - Validate and sanitize all user input at API boundary
  - Use allowlists for enum-like inputs
  - Never use `{@html}` with user-supplied content

### A04:2021 — Insecure Design
- **Lurkarr Measures**:
  - ADR documentation for architecture decisions
  - Multi-auth strategy (local, TOTP, OIDC, proxy, WebAuthn)
  - Strike system for queue cleaning (progressive penalty)
- **Best Practices**:
  - Threat modeling for new features
  - Principle of least privilege
  - Secure defaults (strict CSRF, secure cookies)

### A05:2021 — Security Misconfiguration
- **Lurkarr Measures**:
  - Security headers: `X-Content-Type-Options`, `X-Frame-Options`, `CSP`
  - No default credentials
  - Environment-based configuration
- **Best Practices**:
  - Disable debug mode in production
  - Remove stack traces from error responses
  - Set `Secure`, `HttpOnly`, `SameSite=Strict` on cookies
  - Review Docker image for unnecessary packages

### A06:2021 — Vulnerable Components
- **Best Practices**:
  - Regular `go get -u` and `npm update` for security patches
  - Use `govulncheck` for Go vulnerability scanning
  - Use `npm audit` for frontend vulnerability scanning
  - Pin dependency versions in go.sum and package-lock.json
  - Monitor GitHub Dependabot alerts

### A07:2021 — Identification & Authentication Failures
- **Lurkarr Measures**:
  - bcrypt password hashing
  - TOTP 2FA
  - WebAuthn/Passkeys
  - OIDC/SSO with PKCE
  - Proxy authentication (Authelia/Authentik)
  - Rate limiting on login
- **Best Practices**:
  - Enforce password complexity
  - Account lockout after N failed attempts
  - Session timeout and rotation
  - Use `SameSite=Strict` to prevent session theft

### A08:2021 — Software & Data Integrity
- **Best Practices**:
  - Verify checksums of downloaded dependencies
  - Use `go.sum` and `package-lock.json` for integrity verification
  - Sign container images
  - CI/CD pipeline security (least privilege tokens)

### A09:2021 — Security Logging & Monitoring
- **Lurkarr Measures**:
  - Structured logging with slog
  - Prometheus metrics for monitoring
  - Grafana dashboards with Loki log aggregation
- **Best Practices**:
  - Log authentication events (success and failure)
  - Log authorization failures
  - Log rate limit triggers
  - Alert on anomalous patterns
  - Never log sensitive data (credentials, tokens)

### A10:2021 — SSRF
- **Lurkarr Context**: Connects to user-configured URLs (arr apps, download clients)
- **Best Practices**:
  - Validate URLs at configuration time
  - Consider allowlisting internal network ranges
  - Set timeouts on all outbound HTTP requests
  - Don't follow redirects blindly

---

## 2. Authentication Architecture

### Multi-Auth Strategy (ADR-007)
```
┌──────────────────────────────────────┐
│           Authentication             │
├───────┬───────┬──────┬──────┬───────┤
│ Local │ TOTP  │ OIDC │Proxy │WebAuth│
│bcrypt │ 2FA   │ SSO  │ HDR  │Passkey│
└───────┴───────┴──────┴──────┴───────┘
```

### Session Security
- **Cookie Flags**: `Secure=true`, `HttpOnly=true`, `SameSite=Strict`, `Path=/`
- **Session Rotation**: New session ID after authentication
- **Session Timeout**: Configurable inactivity timeout
- **Concurrent Sessions**: Track and limit per user

---

## 3. HTTP Security Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Content-Type-Options` | `nosniff` | Prevent MIME-type sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `0` | Disable legacy XSS auditor (CSP preferred) |
| `Content-Security-Policy` | Restrictive policy | Prevent XSS, inline scripts |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referer leaking |
| `Permissions-Policy` | `camera=(), microphone=()` | Disable unnecessary APIs |

---

## 4. Rate Limiting Strategy

| Endpoint | Limit | Window | Key |
|----------|-------|--------|-----|
| `/api/v1/auth/login` | 5 attempts | 15 min | IP |
| `/api/v1/auth/totp/verify` | 5 attempts | 15 min | IP + User |
| `/api/v1/**` (authenticated) | 100 requests | 1 min | User |
| `/api/v1/**` (unauthenticated) | 20 requests | 1 min | IP |

---

## 5. External API Call Security

### HTTP Client Configuration
```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### API Key Storage
- Store encrypted in database
- Never log in plaintext
- Mask in API responses (show only last 4 characters)
- Validate format before storage
