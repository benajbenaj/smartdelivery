package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Addr                string
	ReadHeaderTimeout   time.Duration
	ShutdownTimeout     time.Duration
	RequestTimeout      time.Duration
	DeliveryRules       DeliveryRuleService
	ShopifyInstallation ShopifyInstallationService
}

func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	cfg = cfg.withDefaults()

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(Dependencies{DeliveryRules: cfg.DeliveryRules, ShopifyInstallation: cfg.ShopifyInstallation, RequestTimeout: cfg.RequestTimeout}),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	err := <-errCh
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (cfg Config) withDefaults() Config {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 3 * time.Second
	}
	return cfg
}
