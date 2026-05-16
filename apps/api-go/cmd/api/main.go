package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"smartdelivery/apps/api-go/internal/config"
	"smartdelivery/apps/api-go/internal/httpserver"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := httpserver.Run(ctx, httpserver.Config{Addr: cfg.HTTPAddr()}, slog.Default()); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
