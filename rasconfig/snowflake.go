// Package rasconfig provides configuration helpers for database connections
// and environment variable management.
package rasconfig

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"time"

	sf "github.com/snowflakedb/gosnowflake"
)

// SnowflakeDBConfig holds configuration parameters for connecting to Snowflake.
// PrivateKey should be a base64-encoded PKCS8 private key (PEM or DER format).
type SnowflakeDBConfig struct {
	Host                  string
	Port                  int
	Database              string
	Schema                string
	Warehouse             string
	User                  string
	PrivateKey            string
	MaxConnections        int
	MaxIdleConnections    int
	MaxConnectionLifetime time.Duration
}

// NewSnowflakePool creates a connection pool to Snowflake with sensible defaults.
// Uses JWT authentication with the provided private key. Returns error on connection failure.
// Suggested Pool settings: 10 max open, 5 max idle, 30min lifetime.
func NewSnowflakePool(ctx context.Context, config *SnowflakeDBConfig) (*sql.DB, error) {
	dsn, err := buildSnowflakeDSN(config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Set Connection Defaults
	if config.MaxConnections == 0 {
		config.MaxConnections = 10
	}
	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = 5
	}
	if config.MaxConnectionLifetime == 0 {
		config.MaxConnectionLifetime = 30 * time.Minute
	}

	// Tune the pool – adjust values per your workload
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.MaxConnectionLifetime)

	// Validate connection
	if err := db.PingContext(ctx); err != nil {
		// Log detailed connection info for debugging
		slog.Error("Snowflake connection failed", "host", config.Host, "user", config.User, "port", config.Port, "err", err)
		return nil, fmt.Errorf("failed to ping Snowflake (Account: %s, User: %s): %w",
			config.Host, config.User, err)
	}

	return db, nil
}

func buildSnowflakeDSN(sfConfig *SnowflakeDBConfig) (string, error) {
	envPrivateKey := sfConfig.PrivateKey

	//Get Base64 Encoded Private Key
	privateKey, err := base64.StdEncoding.DecodeString(envPrivateKey)
	if err != nil {
		return "", err
	}

	//Parse it as a Rsa Private Key
	rsaPrivateKey, err := adaptToRsaPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	//Create Connection String
	dsn, err := sf.DSN(&sf.Config{
		Account:       sfConfig.Host,
		User:          sfConfig.User,
		Port:          sfConfig.Port,
		Authenticator: sf.AuthTypeJwt,
		PrivateKey:    rsaPrivateKey,
		Database:      sfConfig.Database,  // optional
		Schema:        sfConfig.Schema,    // optional
		Warehouse:     sfConfig.Warehouse, // optional
	})
	if err != nil {
		return "", err
	}
	return dsn, nil
}

// adaptToRsaPrivateKey parses a PKCS8 private key (PEM or DER format) into an RSA private key.
func adaptToRsaPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	// Try to decode as PEM first
	block, _ := pem.Decode(privateKey)

	var keyBytes []byte
	if block != nil && block.Type == "PRIVATE KEY" {
		// PEM format - use the decoded bytes
		keyBytes = block.Bytes
	} else {
		// Assume DER format (raw binary after base64 decode)
		keyBytes = privateKey
	}

	//Parse the PKCS8 PrivateKey Format
	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the PKCS8 private key: %w", err)
	}

	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("RSA key is invalid")
	}

	return rsaPrivateKey, nil
}
