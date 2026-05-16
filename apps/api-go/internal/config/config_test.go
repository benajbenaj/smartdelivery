package config

import (
	"errors"
	"smartdelivery/apps/api-go/internal/apperr"
	"testing"
)

func TestLoadReturnsConfigFromEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SHOPIFY_API_KEY", "test-key")
	t.Setenv("SHOPIFY_API_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("AppEnv = %q, want test", cfg.AppEnv)
	}
	if cfg.HTTPAddr() != ":9090" {
		t.Fatalf("HTTPAddr() = %q, want :9090", cfg.HTTPAddr())
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("DatabaseURL = %q, want postgres://example", cfg.DatabaseURL)
	}
	if cfg.Shopify.APIKey != "test-key" {
		t.Fatalf("Shopify.APIKey = %q, want test-key", cfg.Shopify.APIKey)
	}
	if cfg.Shopify.APISecret != "test-secret" {
		t.Fatalf("Shopify.APISecret = %q, want test-secret", cfg.Shopify.APISecret)
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SHOPIFY_API_KEY", "test-key")
	t.Setenv("SHOPIFY_API_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != defaultAppEnv {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, defaultAppEnv)
	}
	if cfg.Port != defaultPort {
		t.Fatalf("Port = %q, want %q", cfg.Port, defaultPort)
	}
}

func TestLoadReturnsMissingConfigError(t *testing.T) {
	_, err := Load()
	if !errors.Is(err, apperr.ErrMissingConfig) {
		t.Fatalf("Load() error = %v, want ErrMissingConfig", err)
	}
}
