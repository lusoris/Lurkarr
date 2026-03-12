# ADR-007: Multi-Auth Strategy (Local + OIDC + Proxy)

**Status:** Accepted  
**Date:** 2026-02-20

## Context

Lurkarr needs to support diverse deployment scenarios: standalone (local auth), behind an SSO provider (OIDC), and behind a reverse proxy with pre-authentication (e.g. Authelia, Authentik). These auth methods must coexist — a single instance may use local auth for some users and OIDC for others.

## Decision

Support three authentication methods simultaneously, evaluated in order by the `RequireAuth` middleware:

1. **Proxy auth** — If `PROXY_AUTH=true`, check for configured headers (e.g. `Remote-User`) from trusted proxy IPs only. Auto-create local user on first login if configured.
2. **Session cookie** — Standard cookie-based sessions (`lurkarr_session`). Used by both local login and OIDC callback.
3. **OIDC** — Authorization code + PKCE flow via `go-oidc/v3`. Standard OIDC discovery makes it provider-agnostic (Authentik, Keycloak, Authelia, Dex, Google, etc.). Group claims map to admin role.

Key security decisions:
- Proxy auth headers are only trusted from IPs matching `TRUSTED_PROXIES` CIDR ranges (defaults to RFC1918 private ranges).
- CSRF protection is enforced for all auth methods, including proxy auth (defense-in-depth).
- Session rotation occurs after login to prevent fixation.
- `X-Forwarded-Proto` is respected for secure cookie decisions.

## Consequences

**Positive:**
- Covers all common self-hosted deployment patterns.
- Provider-agnostic OIDC — no provider-specific code needed.
- Proxy auth + CSRF prevents cross-origin attacks through the proxy.
- Auto-create simplifies onboarding for both OIDC and proxy auth users.

**Negative:**
- Simultaneous multi-OIDC-provider support (e.g. Google + Authentik) is not yet implemented — deferred as a separate feature.
- Proxy auth trusts IP ranges, which can be spoofed in misconfigured networks.
- Three auth paths increase the surface area for security bugs.
