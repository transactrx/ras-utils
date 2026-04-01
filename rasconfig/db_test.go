package rasconfig

import (
	"context"
	"testing"
	"time"
)

func TestDBConfig(t *testing.T) {
	t.Run("struct fields are set correctly", func(t *testing.T) {
		cfg := &DBConfig{
			Host:                  "localhost",
			ReadOnlyHost:          "readonly.localhost",
			Port:                  "5432",
			DatabaseName:          "testdb",
			User:                  "testuser",
			Password:              "testpass",
			MaxConnections:        10,
			MinConnections:        2,
			MaxConnectionLifetime: time.Hour,
			MaxConnectionIdleTime: 30 * time.Minute,
			ConnectionTimeout:     5 * time.Second,
		}

		if cfg.Host != "localhost" {
			t.Errorf("expected localhost, got %s", cfg.Host)
		}
		if cfg.ReadOnlyHost != "readonly.localhost" {
			t.Errorf("expected readonly.localhost, got %s", cfg.ReadOnlyHost)
		}
		if cfg.Port != "5432" {
			t.Errorf("expected 5432, got %s", cfg.Port)
		}
		if cfg.DatabaseName != "testdb" {
			t.Errorf("expected testdb, got %s", cfg.DatabaseName)
		}
		if cfg.MaxConnections != 10 {
			t.Errorf("expected 10, got %d", cfg.MaxConnections)
		}
		if cfg.MinConnections != 2 {
			t.Errorf("expected 2, got %d", cfg.MinConnections)
		}
	})
}

func TestInitDbPool_InvalidHost(t *testing.T) {
	cfg := &DBConfig{
		Host:                  "invalid-host-that-does-not-exist",
		Port:                  "5432",
		DatabaseName:          "testdb",
		User:                  "testuser",
		Password:              "testpass",
		MaxConnections:        5,
		MinConnections:        1,
		MaxConnectionLifetime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
		ConnectionTimeout:     1 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := InitDbPool(ctx, cfg)
	if err == nil {
		pool.Close()
		t.Error("expected error for invalid host")
	}
}

func TestInitReadOnlyDbPool_EmptyHost(t *testing.T) {
	cfg := &DBConfig{
		Host:                  "localhost",
		ReadOnlyHost:          "",
		Port:                  "5432",
		DatabaseName:          "testdb",
		User:                  "testuser",
		Password:              "testpass",
		MaxConnections:        5,
		MinConnections:        1,
		MaxConnectionLifetime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
		ConnectionTimeout:     1 * time.Second,
	}

	ctx := context.Background()

	_, err := InitReadOnlyDbPool(ctx, cfg)
	if err == nil {
		t.Error("expected error for empty ReadOnlyHost")
	}
	if err.Error() != "ReadOnlyHost is not configured" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestInitReadOnlyDbPool_InvalidHost(t *testing.T) {
	cfg := &DBConfig{
		Host:                  "localhost",
		ReadOnlyHost:          "invalid-readonly-host-that-does-not-exist",
		Port:                  "5432",
		DatabaseName:          "testdb",
		User:                  "testuser",
		Password:              "testpass",
		MaxConnections:        5,
		MinConnections:        1,
		MaxConnectionLifetime: time.Hour,
		MaxConnectionIdleTime: 30 * time.Minute,
		ConnectionTimeout:     1 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := InitReadOnlyDbPool(ctx, cfg)
	if err == nil {
		pool.Close()
		t.Error("expected error for invalid read-only host")
	}
}
