# PostgreSQL + pgx + goose

## pgx/v5 — PostgreSQL Driver

> Version: v5.8.0
> Docs: https://pkg.go.dev/github.com/jackc/pgx/v5

### Connection Pool

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
if err != nil { return err }
defer pool.Close()
```

### Pool Config

```go
config, _ := pgxpool.ParseConfig(databaseURL)
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 30 * time.Minute
config.MaxConnIdleTime = 5 * time.Minute
pool, _ := pgxpool.NewWithConfig(ctx, config)
```

### Query Patterns

```go
// Single row
var name string
err := pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", id).Scan(&name)

// Multiple rows
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
defer rows.Close()
for rows.Next() {
    var id int; var name string
    rows.Scan(&id, &name)
}

// Exec (insert/update/delete)
tag, err := pool.Exec(ctx, "DELETE FROM items WHERE expired_at < $1", time.Now())
fmt.Println(tag.RowsAffected())
```

### Batch Operations

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO items (name) VALUES ($1)", "item1")
batch.Queue("INSERT INTO items (name) VALUES ($1)", "item2")
results := pool.SendBatch(ctx, batch)
defer results.Close()
```

### Transactions

```go
tx, err := pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx) // no-op if committed

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
if err != nil { return err }
_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
if err != nil { return err }

return tx.Commit(ctx)
```

### JSON/JSONB

```go
var settings map[string]any
err := pool.QueryRow(ctx, "SELECT settings FROM users WHERE id = $1", id).Scan(&settings)
```

### Row Scanning to Structs

```go
import "github.com/jackc/pgx/v5/pgxutil" // or use pgx.RowToStructByName

rows, _ := pool.Query(ctx, "SELECT id, name, email FROM users")
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

## goose/v3 — Database Migrations

> Version: v3.27.0
> Docs: https://pressly.github.io/goose/

### Migration File Format

```
internal/database/migrations/
├── 001_initial.sql
├── 002_prowlarr_sabnzbd.sql
├── 003_instance_aware_stats.sql
├── 004_queue_management.sql
├── 005_cleaner_enhancements.sql
├── 006_notifications.sql
└── 007_seerr_settings.sql
```

### SQL Migration Template

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS notifications;
```

### Programmatic Usage (Lurkarr)

```go
import (
    "github.com/pressly/goose/v3"
    "embed"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB) error {
    goose.SetBaseFS(embedMigrations)
    return goose.Up(db, "migrations")
}
```

### CLI Commands

```bash
goose -dir migrations postgres "postgres://user:pass@localhost/db" status
goose -dir migrations postgres "postgres://user:pass@localhost/db" up
goose -dir migrations postgres "postgres://user:pass@localhost/db" down
goose -dir migrations postgres "postgres://user:pass@localhost/db" create add_feature sql
```

## PostgreSQL 17 Notes

### Key Features

- Incremental backup
- JSON_TABLE() function
- MERGE improvements
- Improved COPY performance

### Indexes for Lurkarr

```sql
-- Composite index for common query patterns
CREATE INDEX idx_lurk_history_app_instance
ON lurk_history (app_type, instance_id, created_at DESC);

-- Partial index for active items
CREATE INDEX idx_queue_items_active
ON queue_items (status) WHERE status != 'completed';

-- GIN index for JSONB
CREATE INDEX idx_notifications_config
ON notifications USING GIN (config);
```

### Connection String

```
postgres://lurkarr:password@localhost:5432/lurkarr?sslmode=disable
# Production: sslmode=require or sslmode=verify-full
```
