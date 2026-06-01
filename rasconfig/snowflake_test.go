package rasconfig

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

func newTestSnowflakeConfig(key string) *SnowflakeDBConfig {
	return &SnowflakeDBConfig{
		Host:       "testaccount",
		Port:       443,
		Database:   "testdb",
		Schema:     "PUBLIC",
		Warehouse:  "testwh",
		User:       "testuser",
		PrivateKey: key,
	}
}

func TestNewSnowflakePool(t *testing.T) {
	t.Run("invalid base64 key returns error", func(t *testing.T) {
		cfg := newTestSnowflakeConfig("not-valid-base64!!!")

		db, err := NewSnowflakePool(context.Background(), cfg)
		if err == nil {
			if db != nil {
				_ = db.Close()
			}
			t.Fatal("expected error for invalid base64 key")
		}
	})

	t.Run("invalid RSA key returns error", func(t *testing.T) {
		invalidKey := base64.StdEncoding.EncodeToString([]byte("this is not a valid key"))
		cfg := newTestSnowflakeConfig(invalidKey)

		db, err := NewSnowflakePool(context.Background(), cfg)
		if err == nil {
			if db != nil {
				_ = db.Close()
			}
			t.Fatal("expected error for invalid RSA key")
		}
		if !strings.Contains(err.Error(), "PKCS8") {
			t.Errorf("expected PKCS8 parse error, got: %v", err)
		}
	})

	t.Run("invalid host returns connection error", func(t *testing.T) {
		cfg := newTestSnowflakeConfig(generateTestBase64Key(t))
		cfg.Host = "invalid-account-that-does-not-exist"

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := NewSnowflakePool(ctx, cfg)
		if err == nil {
			if db != nil {
				_ = db.Close()
			}
			t.Fatal("expected error for invalid host")
		}
	})

	t.Run("valid DER key creates pool", func(t *testing.T) {
		cfg := newTestSnowflakeConfig(generateTestBase64Key(t))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := NewSnowflakePool(ctx, cfg)
		if err == nil {
			_ = db.Close()
			t.Skip("unexpectedly connected to Snowflake")
		}
		if strings.Contains(err.Error(), "base64") || strings.Contains(err.Error(), "PKCS8") {
			t.Errorf("expected connection error, got key error: %v", err)
		}
	})

	t.Run("valid PEM key creates pool", func(t *testing.T) {
		cfg := newTestSnowflakeConfig(generateTestBase64PEMKey(t))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := NewSnowflakePool(ctx, cfg)
		if err == nil {
			_ = db.Close()
			t.Skip("unexpectedly connected to Snowflake")
		}
		if strings.Contains(err.Error(), "base64") || strings.Contains(err.Error(), "PKCS8") {
			t.Errorf("expected connection error, got key error: %v", err)
		}
	})

	t.Run("custom pool settings are accepted", func(t *testing.T) {
		cfg := newTestSnowflakeConfig(generateTestBase64Key(t))
		cfg.MaxConnections = 20
		cfg.MaxIdleConnections = 10
		cfg.MaxConnectionLifetime = 1 * time.Hour

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err := NewSnowflakePool(ctx, cfg)
		if err == nil {
			_ = db.Close()
			t.Skip("unexpectedly connected to Snowflake")
		}
		if strings.Contains(err.Error(), "base64") || strings.Contains(err.Error(), "PKCS8") {
			t.Errorf("expected connection error, got config error: %v", err)
		}
	})
}

func TestAdaptToRsaPrivateKey(t *testing.T) {
	t.Run("valid DER format key", func(t *testing.T) {
		privKey := generateTestPrivateKey(t)
		derBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			t.Fatalf("failed to marshal private key: %v", err)
		}

		result, err := adaptToRsaPrivateKey(derBytes)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil RSA key")
		}
	})

	t.Run("valid PEM format key", func(t *testing.T) {
		privKey := generateTestPrivateKey(t)
		derBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			t.Fatalf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: derBytes,
		}
		pemBytes := pem.EncodeToMemory(pemBlock)

		result, err := adaptToRsaPrivateKey(pemBytes)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil RSA key")
		}
	})

	t.Run("invalid key bytes", func(t *testing.T) {
		_, err := adaptToRsaPrivateKey([]byte("not a valid key"))
		if err == nil {
			t.Error("expected error for invalid key")
		}
	})
}

func generateTestPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return privKey
}

func generateTestBase64Key(t *testing.T) string {
	t.Helper()
	privKey := generateTestPrivateKey(t)
	derBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(derBytes)
}

func generateTestBase64PEMKey(t *testing.T) string {
	t.Helper()
	privKey := generateTestPrivateKey(t)
	derBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derBytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)
	return base64.StdEncoding.EncodeToString(pemBytes)
}
