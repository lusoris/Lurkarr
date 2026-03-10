package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	// goose requires database/sql for migrations
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrations embed.FS

// DB wraps a pgx connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool and runs migrations.
func New(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	config.MaxConns = 25

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{Pool: pool}
	if err := db.migrate(databaseURL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("database connected and migrated", "max_conns", config.MaxConns)
	return db, nil
}

func (db *DB) migrate(databaseURL string) error {
	goose.SetBaseFS(migrations)

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql connection for migrations: %w", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// HealthCheck pings the database.
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
