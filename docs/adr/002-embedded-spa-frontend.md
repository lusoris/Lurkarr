# ADR-002: Embedded SPA Frontend in Go Binary

**Status:** Accepted  
**Date:** 2026-03-12

## Context

Lurkarr's frontend is a SvelteKit 5 static SPA. Originally the frontend was built during Docker image creation and copied into the image, but the Go server never served it — requiring a separate web server or reverse proxy to serve the frontend.

A single-binary deployment model (one binary serves both API and UI) is simpler to deploy, especially for non-Docker users.

## Decision

Use Go's `//go:embed` directive to embed the entire `frontend/build/` directory into the Go binary. Serve it via `http.FileServerFS` with SPA fallback (any non-file path returns `index.html`).

The embed is in `frontend/embed.go` with `//go:embed all:build`. A `BuildFS()` function returns `nil` when only `.gitkeep` exists (fresh clone without a frontend build), so the server gracefully skips SPA serving.

## Consequences

**Positive:**
- Single binary serves everything — no nginx sidecar or separate file server needed.
- Works in scratch/distroless containers with zero external dependencies.
- Immutable assets (`_app/immutable/`) get far-future cache headers automatically.
- `BuildFS()` returning nil on fresh clones prevents embed failures during pure-backend development.

**Negative:**
- Binary size increases by ~1-2MB (compressed SvelteKit build output).
- Frontend rebuilds require recompiling the Go binary.
- `.gitkeep` in `frontend/build/` needs a `.gitignore` exception and must be re-added after `npm run build` (SvelteKit replaces the directory).

## Alternatives Considered

- **Separate nginx container:** Standard approach but adds deployment complexity.
- **Embed via build tags:** More flexible but adds build matrix complexity.
- **Runtime file serving from disk:** Simpler but requires the build directory to exist at runtime, breaking single-binary deployment.
