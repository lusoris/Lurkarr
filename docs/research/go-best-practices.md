# Go Best Practices & Coding Standards

## Project Structure (Lurkarr follows this)

```
cmd/lurkarr/          # Application entry point
internal/             # Private packages (not importable externally)
  api/                # HTTP handlers
  arrclient/          # *arr client implementations
  auth/               # Authentication & authorization
  cache/              # In-memory caching (otter)
  config/             # Configuration loading
  database/           # DB layer (pgx, goose migrations)
  lurking/            # Business logic: lurking engine
  logging/            # Structured logging
  metrics/            # Prometheus metrics
  middleware/         # HTTP middleware
  notifications/      # Notification providers
  queuecleaner/       # Queue management
  sabnzbd/            # SABnzbd client
  scheduler/          # Task scheduling
  server/             # HTTP server setup
deploy/               # Deployment configs (Docker, Helm, Grafana)
frontend/             # SvelteKit SPA
docs/                 # Documentation & research
```

## Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("fetch missing items for %s: %w", appType, err)
}

// Use errors.Is / errors.As for error checks
if errors.Is(err, context.Canceled) { return nil }

// Custom error types for API responses
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

## Context Usage

- Pass `context.Context` as first parameter to all functions that do I/O
- Use `context.WithTimeout` for external API calls
- Use `context.WithCancel` for goroutine lifecycle
- Store request-scoped values (RequestID) via `context.WithValue`

## Interface Design

```go
// Small, focused interfaces
type ArrLurker interface {
    GetMissing(ctx context.Context, url, apiKey string) ([]Item, error)
    GetUpgrades(ctx context.Context, url, apiKey string) ([]Item, error)
    Search(ctx context.Context, url, apiKey string, ids []int) error
    GetQueue(ctx context.Context, url, apiKey string) ([]QueueItem, error)
}

// Accept interfaces, return structs
func NewEngine(store Store, logger Logger) *Engine { ... }
```

## Concurrency Patterns

```go
// Worker pool with errgroup
g, ctx := errgroup.WithContext(ctx)
for _, item := range items {
    item := item
    g.Go(func() error {
        return process(ctx, item)
    })
}
if err := g.Wait(); err != nil { ... }

// Graceful shutdown pattern
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go func() {
    sig := <-signalCh
    cancel()
}()
```

## Database Patterns (pgx)

```go
// Always use parameterized queries
row := db.pool.QueryRow(ctx,
    "SELECT id, name FROM users WHERE email = $1", email)

// Batch operations
batch := &pgx.Batch{}
for _, id := range ids {
    batch.Queue("UPDATE items SET processed = true WHERE id = $1", id)
}
results := db.pool.SendBatch(ctx, batch)
defer results.Close()
```

## HTTP Handler Patterns

```go
func (s *State) handleGetSettings(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    settings, err := s.store.GetSettings(ctx)
    if err != nil {
        s.jsonError(w, "failed to get settings", http.StatusInternalServerError)
        return
    }
    s.jsonOK(w, settings)
}
```

## Configuration

- Environment variables for all config (12-factor)
- Sensible defaults for everything
- Validate at startup, fail fast
- No config files — env vars only

## Logging (slog)

```go
slog.Info("lurk completed",
    "app_type", appType,
    "instance", instanceID,
    "missing_found", count,
    "duration", elapsed,
)

// Use structured fields, not fmt.Sprintf
// BAD:  slog.Info(fmt.Sprintf("found %d items for %s", count, appType))
// GOOD: slog.Info("found items", "count", count, "app_type", appType)
```

## Performance

- Use `otter/v2` for hot-path caching (settings, frequently accessed data)
- Connection pool sizing: `max(4, numCPU)` for DB
- Batch DB operations where possible
- Use `context.WithTimeout` to prevent hung operations
- Profile with `pprof` when needed
