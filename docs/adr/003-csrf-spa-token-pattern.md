# ADR-003: CSRF Protection with SPA Token Header Pattern

**Status:** Accepted  
**Date:** 2026-03-12

## Context

Lurkarr uses `gorilla/csrf` for CSRF protection on all state-changing endpoints. The standard approach (hidden form field or cookie) doesn't work cleanly with a SPA that communicates via `fetch` API calls.

Additionally, CSRF must remain active even when proxy authentication is enabled (e.g. Authelia). An attacker on the same network could craft a cross-origin request, and the proxy would auto-inject auth headers, making CSRF the only defense.

## Decision

Use a response header pattern:
1. A `csrfInjectToken` middleware sets `X-CSRF-Token` on every authenticated response (via `csrf.Token(r)`).
2. The frontend captures this token from response headers and sends it back with POST/PUT/DELETE requests.
3. `gorilla/csrf` validates the token from the `X-CSRF-Token` request header.

The middleware chain is: `CSRFProtect → csrfInjectToken → RequireAuth → handler`.

## Consequences

**Positive:**
- Works seamlessly with `fetch` and SPA routing — no DOM parsing or meta tags needed.
- Token is always fresh (updated on every response).
- CSRF is enforced regardless of auth method (local, OIDC, or proxy auth) — defense-in-depth.

**Negative:**
- Frontend must handle the token (capture from response, send with mutations).
- Token is exposed in response headers — acceptable since it's per-session and requires the session cookie to be useful.

## Alternatives Considered

- **Double-submit cookie:** Simpler but `gorilla/csrf` already handles this internally; the header pattern gives us explicit control.
- **SameSite=Strict cookies only:** Insufficient — doesn't protect against subdomain attacks, and some proxy auth setups may relay cookies cross-origin.
- **Disable CSRF for proxy auth:** Rejected — removes a critical defense layer against cross-origin attacks through the proxy.
