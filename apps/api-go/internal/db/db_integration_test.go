//go:build integration

package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnectIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gormDB, err := Connect(ctx, Config{DatabaseURL: databaseURL})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if err := Close(gormDB); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := HealthCheck(ctx, gormDB); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
}
