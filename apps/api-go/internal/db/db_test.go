package db

import (
	"context"
	"errors"
	"smartdelivery/apps/api-go/internal/apperr"
	"testing"
)

func TestConnectRequiresDatabaseURL(t *testing.T) {
	_, err := Connect(context.Background(), Config{})
	if !errors.Is(err, apperr.ErrMissingConfig) {
		t.Fatalf("Connect() error = %v, want ErrMissingConfig", err)
	}
}

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{}.withDefaults()

	if cfg.MaxOpenConns != defaultMaxOpenConns {
		t.Fatalf("MaxOpenConns = %d, want %d", cfg.MaxOpenConns, defaultMaxOpenConns)
	}
	if cfg.MaxIdleConns != defaultMaxIdleConns {
		t.Fatalf("MaxIdleConns = %d, want %d", cfg.MaxIdleConns, defaultMaxIdleConns)
	}
	if cfg.ConnMaxLifetime != defaultConnMaxLifetime {
		t.Fatalf("ConnMaxLifetime = %s, want %s", cfg.ConnMaxLifetime, defaultConnMaxLifetime)
	}
}
