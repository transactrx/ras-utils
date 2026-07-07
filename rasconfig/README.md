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

- `GetEnvironmentVariableOrDefault(key, defaultValue string) string`
- `GetEnvironmentVariableOrDefaultInt(key string, defaultValue int) int`
- `GetEnvironmentVariableOrDefaultDuration(key, defaultValue string) time.Duration`
- `GetEnvironmentVariableOrPanic(key, panicMessage string) string`
- `InitDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error)`
- `InitReadOnlyDbPool(ctx context.Context, cfg *DBConfig) (*pgxpool.Pool, error)`
- `NewSnowflakePool(ctx context.Context, cfg *SnowflakeDBConfig) (*sql.DB, error)`
