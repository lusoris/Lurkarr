# ADR-008: SvelteKit 5 Static SPA with TailwindCSS v4

**Status:** Accepted  
**Date:** 2026-01-20

## Context

Lurkarr needs a web frontend for configuration, monitoring, and management. The frontend communicates with the Go backend via REST API and WebSocket. It must work when embedded in the Go binary (static files served by `http.FileServerFS`).

## Decision

Use SvelteKit 5 with `adapter-static` to produce a pre-rendered SPA, styled with TailwindCSS v4.

Key choices:
- **Svelte 5 runes** (`$state`, `$derived`, `$effect`, `$props`) for reactive state — no stores for component-local state.
- **adapter-static** with `fallback: 'index.html'` — produces a static SPA that the Go server can serve with index.html fallback for client-side routing.
- **TailwindCSS v4** with `@theme` for custom design tokens (lurk-* green brand palette, surface-* neutral grays).
- **Dark mode default** — matches media preference, with manual toggle.
- **Vite 7** for build tooling with Brotli pre-compression.

## Consequences

**Positive:**
- Static output embeds cleanly into Go binary via `//go:embed`.
- Svelte 5 runes eliminate boilerplate (no `$:` reactive statements, no writable stores for local state).
- TailwindCSS v4's `@theme` provides a consistent design system without custom CSS.
- Pre-compressed `.br` files reduce transfer size.
- No SSR complexity — pure client-side SPA with API calls.

**Negative:**
- Static SPA means no server-side rendering — initial load shows a blank page until JS loads.
- `adapter-static` requires explicit route pre-rendering configuration for any non-SPA pages.
- TailwindCSS v4 is relatively new — fewer community examples and plugins compared to v3.

## Alternatives Considered

- **React/Next.js:** More ecosystem support but heavier bundle, more boilerplate for this project's scope.
- **Vue/Nuxt:** Similar trade-offs to React; Svelte's smaller bundle and simpler reactivity model won.
- **HTMX + Go templates:** Simpler stack but lacks the interactivity needed for real-time log streaming, drag-and-drop scheduling, and complex form UIs.
