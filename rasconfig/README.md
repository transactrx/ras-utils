# rasconfig

Database configuration and environment variable helpers for Go applications.

## Features

- Environment variable helpers with type-safe defaults
- PostgreSQL connection pool setup (pgxpool)
- Read/write and read-only pool separation
- Snowflake connection pool with JWT authentication

## Installation

```go
import "github.com/transactrx/ras-utils/rasconfig"
```

## Usage

### Environment Variables

```go
// With defaults
host := rasconfig.GetEnvironmentVariableOrDefault("DB_HOST", "localhost")
port := rasconfig.GetEnvironmentVariableOrDefaultInt("DB_PORT", 5432)
timeout := rasconfig.GetEnvironmentVariableOrDefaultDuration("DB_TIMEOUT", "30s")

// Required (panics if missing)
apiKey := rasconfig.GetEnvironmentVariableOrPanic("API_KEY", "API_KEY is required")
```

### Type-Safe Generic Environment Variables

For type-safe parsing of environment variables to any primitive type:

```go
// With default - returns default if not set or empty, error if parse fails
enabled, err := rasconfig.TryGetEnvironmentVariableOrDefault("FEATURE_ENABLED", false)
maxRetries, err := rasconfig.TryGetEnvironmentVariableOrDefault[int16]("MAX_RETRIES", 3)
threshold, err := rasconfig.TryGetEnvironmentVariableOrDefault("THRESHOLD", 0.95)

// Without default - returns ErrEnvNotFound if not set or empty
port, err := rasconfig.TryGetEnvironmentVariable[uint16]("PORT")
if errors.Is(err, rasconfig.ErrEnvNotFound) {
    // handle missing env var
}
```

Supported types: `bool`, `string`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`.

### PostgreSQL Connection Pool

```go
cfg := &rasconfig.DBConfig{
    Host:                  "localhost",
    ReadOnlyHost:          "readonly.localhost",
    Port:                  "5432",
    DatabaseName:          "mydb",
    User:                  "user",
    Password:              "pass",
    MaxConnections:        10,
    MinConnections:        2,
    MaxConnectionLifetime: time.Hour,
    MaxConnectionIdleTime: 30 * time.Minute,
    ConnectionTimeout:     5 * time.Second,
}

pool, err := rasconfig.InitDbPool(ctx, cfg)
readOnlyPool, err := rasconfig.InitReadOnlyDbPool(ctx, cfg)
```

### Snowflake Connection Pool

```go
sfCfg := &rasconfig.SnowflakeDBConfig{
    Host:                  "account.snowflakecomputing.com",
    Port:                  443,
    Database:              "MY_DB",
    Schema:                "PUBLIC",
    Warehouse:             "MY_WAREHOUSE",
    User:                  "my_user",
    PrivateKey:            "base64-encoded-pkcs8-private-key",
    MaxConnections:        10,
    MaxIdleConnections:    5,
    MaxConnectionLifetime: 30 * time.Minute,
}

db, err := rasconfig.NewSnowflakePool(ctx, sfCfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// PrivateKey must be base64-encoded PKCS8 format (PEM or DER).
// Generate with: openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt | base64 -w0
```

## API Reference

### Types

- `DBConfig` - PostgreSQL connection configuration
- `SnowflakeDBConfig` - Snowflake connection configuration

### Functions

**Environment Variables:**
- `GetEnvironmentVariableOrDefault(key, defaultValue string) string`
- `GetEnvironmentVariableOrDefaultInt(key string, defaultValue int) int`
- `GetEnvironmentVariableOrDefaultDuration(key, defaultValue string) time.Duration`
- `GetEnvironmentVariableOrPanic(key, panicMessage string) string`
- `TryGetEnvironmentVariable[T primitive](key string) (T, error)` - returns `ErrEnvNotFound` if missing
- `TryGetEnvironmentVariableOrDefault[T primitive](key string, defaultValue T) (T, error)` - returns default if missing, error if parse fails

**Database Pools:**
- `InitDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error)`
- `InitReadOnlyDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error)`
- `NewSnowflakePool(ctx context.Context, cfg *SnowflakeDBConfig) (*sql.DB, error)`

### Errors

- `ErrEnvNotFound` - returned by `TryGetEnvironmentVariable` when the environment variable is not set or empty
