package config

import (
	"fmt"
	"os"
	"smartdelivery/apps/api-go/internal/apperr"
)

const (
	defaultAppEnv = "development"
	defaultPort   = "8080"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	Shopify     ShopifyConfig
}

type ShopifyConfig struct {
	APIKey    string
	APISecret string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", defaultAppEnv),
		Port:        getEnv("PORT", defaultPort),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Shopify: ShopifyConfig{
			APIKey:    os.Getenv("SHOPIFY_API_KEY"),
			APISecret: os.Getenv("SHOPIFY_API_SECRET"),
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
