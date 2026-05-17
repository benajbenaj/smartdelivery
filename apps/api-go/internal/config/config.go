package config

import (
	"fmt"
	"os"
	"strconv"

	"smartdelivery/apps/api-go/internal/apperr"
)

const (
	defaultAppEnv                = "development"
	defaultPort                  = "8080"
	defaultAuditBufferSize       = 100
	defaultAuditOverflowBehavior = "block"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	Shopify     ShopifyConfig
	Audit       AuditConfig
}

type ShopifyConfig struct {
	APIKey    string
	APISecret string
}

type AuditConfig struct {
	BufferSize       int
	OverflowBehavior string
}

func Load() (Config, error) {
	auditBufferSize, err := getEnvInt("AUDIT_BUFFER_SIZE", defaultAuditBufferSize)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:      getEnv("APP_ENV", defaultAppEnv),
		Port:        getEnv("PORT", defaultPort),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Shopify: ShopifyConfig{
			APIKey:    os.Getenv("SHOPIFY_API_KEY"),
			APISecret: os.Getenv("SHOPIFY_API_SECRET"),
		},
		Audit: AuditConfig{
			BufferSize:       auditBufferSize,
			OverflowBehavior: getEnv("AUDIT_OVERFLOW_BEHAVIOR", defaultAuditOverflowBehavior),
		},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("%w: DATABASE_URL", apperr.ErrMissingConfig)
	}
	if cfg.Shopify.APIKey == "" {
		return Config{}, fmt.Errorf("%w: SHOPIFY_API_KEY", apperr.ErrMissingConfig)
	}
	if cfg.Shopify.APISecret == "" {
		return Config{}, fmt.Errorf("%w: SHOPIFY_API_SECRET", apperr.ErrMissingConfig)
	}
	if cfg.Audit.OverflowBehavior != "block" && cfg.Audit.OverflowBehavior != "drop" {
		return Config{}, fmt.Errorf("%w: AUDIT_OVERFLOW_BEHAVIOR", apperr.ErrValidation)
	}

	return cfg, nil
}

func (cfg Config) HTTPAddr() string {
	return ":" + cfg.Port
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%w: %s", apperr.ErrValidation, key)
	}
	return parsed, nil
}
