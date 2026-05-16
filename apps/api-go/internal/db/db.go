package db

import (
	"context"
	"fmt"
	"time"

	"smartdelivery/apps/api-go/internal/apperr"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = time.Hour
)

type Config struct {
	DatabaseURL     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Connect(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("%w: DATABASE_URL", apperr.ErrMissingConfig)
	}

	cfg = cfg.withDefaults()

	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := HealthCheck(ctx, gormDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return gormDB, nil
}

func HealthCheck(ctx context.Context, gormDB *gorm.DB) error {
	var result int
	if err := gormDB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("postgres health check: %w", err)
	}
	if result != 1 {
		return fmt.Errorf("postgres health check: expected 1, got %d", result)
	}
	return nil
}

func Close(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get postgres pool: %w", err)
	}
	return sqlDB.Close()
}

func (cfg Config) withDefaults() Config {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = defaultMaxOpenConns
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = defaultConnMaxLifetime
	}
	return cfg
}
