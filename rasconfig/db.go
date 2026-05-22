// Package rasconfig provides configuration utilities for database connections
// and environment variable handling.
package rasconfig

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBConfig holds configuration for PostgreSQL database connections.
type DBConfig struct {
	Host                  string
	ReadOnlyHost          string
	Port                  string
	DatabaseName          string
	User                  string
	Password              string
	MaxConnections        int
	MinConnections        int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
	ConnectionTimeout     time.Duration
}

// InitDbPool creates and returns a PostgreSQL connection pool using the provided configuration.
// It verifies connectivity by pinging the database before returning.
func InitDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error) {
	slog.Debug("Config stuff", "debug config", cfg)

	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.DatabaseName,
		cfg.User,
		cfg.Password,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnectionLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolConfig.HealthCheckPeriod = cfg.ConnectionTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("database connection pool established", "host", cfg.Host, "database", cfg.DatabaseName)

	return pool, nil
}

// InitReadOnlyDbPool creates a connection pool to the read-only replica specified in [DBConfig.ReadOnlyHost].
// It returns an error if ReadOnlyHost is not configured.
func InitReadOnlyDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error) {
	if cfg.ReadOnlyHost == "" {
		return nil, fmt.Errorf("ReadOnlyHost is not configured")
	}

	slog.Debug("Config stuff (read-only)", "debug config", cfg)

	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.ReadOnlyHost,
		cfg.Port,
		cfg.DatabaseName,
		cfg.User,
		cfg.Password,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnectionLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolConfig.HealthCheckPeriod = cfg.ConnectionTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create read-only connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping read-only database: %w", err)
	}

	slog.Info("read-only database connection pool established", "host", cfg.ReadOnlyHost, "database", cfg.DatabaseName)

	return pool, nil
}
