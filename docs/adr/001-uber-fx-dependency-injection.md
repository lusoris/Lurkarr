# ADR-001: Uber FX for Dependency Injection

**Status:** Accepted  
**Date:** 2026-02-15

## Context

Lurkarr has 15+ packages with complex interdependencies (database, logging, notifications, scheduler, server, auth, queue cleaner, download clients, etc.). The original `main.go` manually wired 30+ constructors with explicit `defer` chains for shutdown. Adding new services required editing a growing chain of provider calls and remembering cleanup order.

## Decision

Use [Uber FX](https://github.com/uber-go/fx) as the dependency injection and application lifecycle framework.

## Consequences

**Positive:**
- Providers declare their dependencies via constructor signatures; FX resolves the dependency graph automatically.
- `fx.Lifecycle` hooks replace manual `defer` chains — shutdown runs in correct reverse order.
- Adding a new service means writing a provider function and registering it in the module; no changes to `main.go` wiring.
- Signal handling (SIGINT/SIGTERM) and graceful shutdown are built in.

**Negative:**
- Runtime dependency resolution means missing providers surface as startup errors, not compile errors.
- Stack traces from FX can be verbose and harder to read.
- Adds a learning curve for contributors unfamiliar with DI containers.

## Alternatives Considered

- **Manual wiring:** Simple but doesn't scale — 30+ constructors with ordering concerns.
- **Wire (compile-time DI):** Generates code at build time. Considered but rejected: FX's lifecycle hooks are more valuable than compile-time safety for this project's size.
- **dig (FX's underlying container):** Lower-level, lacks lifecycle management.
