# Uber FX — Dependency Injection for Go

> Version: v1.24.0 | License: MIT | Imported by: 9,366+ packages
> Docs: https://uber-go.github.io/fx/
> API: https://pkg.go.dev/go.uber.org/fx

## Overview

Fx is Uber's production DI framework for Go. Eliminates globals, reduces boilerplate, auto-wires dependency graphs.

## Core Concepts

### Application Lifecycle

```go
fx.New(opts...).Run()
// or manual: app.Start(ctx) → app.Stop(ctx)
```

- `fx.Provide(constructors...)` — register constructors (lazy, cached as singletons)
- `fx.Invoke(funcs...)` — eagerly run at startup, trigger constructor instantiation
- `fx.Supply(values...)` — provide pre-built values directly
- `fx.Module(name, opts...)` — group options into named logical unit
- `fx.Lifecycle` — register OnStart/OnStop hooks for servers, background goroutines

### Constructor Pattern

```go
// Constructor: depends on *Config, produces *Server
func NewServer(cfg *Config) (*Server, error) { ... }
fx.Provide(NewServer)
```

### Lifecycle Hooks (crucial for Lurkarr)

```go
fx.Provide(
    fx.Annotate(
        NewServer,
        fx.OnStart(func(ctx context.Context, s *Server) error {
            return s.Start()
        }),
        fx.OnStop(func(ctx context.Context, s *Server) error {
            return s.Shutdown(ctx)
        }),
    ),
)
```

### Parameter Structs (many deps)

```go
type Params struct {
    fx.In
    DB     *database.DB
    Logger *slog.Logger
    Config *config.Config
}
func NewEngine(p Params) *Engine { ... }
```

### Result Structs (many outputs)

```go
type Result struct {
    fx.Out
    Engine  *Engine
    Cleaner *Cleaner
}
```

### Named Values

```go
fx.Provide(fx.Annotate(NewReadDB, fx.ResultTags(`name:"ro"`)))
fx.Provide(fx.Annotate(NewWriteDB, fx.ResultTags(`name:"rw"`)))
```

### Value Groups (for handlers, hunters, etc.)

```go
// Provider
type HunterResult struct {
    fx.Out
    Hunter ArrHunter `group:"hunters"`
}

// Consumer
type EngineParams struct {
    fx.In
    Hunters []ArrHunter `group:"hunters"`
}
```

### Testing with fxtest

```go
import "go.uber.org/fx/fxtest"

func TestApp(t *testing.T) {
    app := fxtest.New(t,
        fx.Provide(NewDB),
        fx.Invoke(func(db *DB) { /* test */ }),
    )
    app.RequireStart()
    defer app.RequireStop()
}
```

### fx.ValidateApp — validate dependency graph without running

```go
err := fx.ValidateApp(opts...)
```

## Lurkarr Migration Plan

Current main.go manually wires:
1. config.Load()
2. database.New()
3. logging.NewHub() + logging.New()
4. notifications.NewManager()
5. hunting.New() + Start()
6. scheduler.New() + Start()
7. queuecleaner.New() + Start()
8. autoimport.New() + Start()
9. seerr.NewSyncEngine() + Start()
10. server.New() + Start()
11. Hourly cleanup goroutine
12. Signal handling + graceful shutdown

Target: Each becomes an fx.Module with Provide + Lifecycle hooks.

```go
// Example module
var DatabaseModule = fx.Module("database",
    fx.Provide(database.New),
)

var HuntingModule = fx.Module("hunting",
    fx.Provide(
        fx.Annotate(
            hunting.New,
            fx.OnStart(func(ctx context.Context, e *hunting.Engine) error {
                e.Start(ctx)
                return nil
            }),
            fx.OnStop(func(ctx context.Context, e *hunting.Engine) error {
                e.Stop()
                return nil
            }),
        ),
    ),
)
```

## Key Decisions

- Use `fx.Module` per package for logical grouping
- Use `fx.Lifecycle` hooks instead of manual defer chains
- Use value groups for arr hunters (ArrHunter interface)
- WithLogger to integrate slog with fx event logging
- RecoverFromPanics() for production safety
