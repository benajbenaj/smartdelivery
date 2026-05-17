package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"smartdelivery/apps/api-go/internal/config"
	"smartdelivery/apps/api-go/internal/db"
	"smartdelivery/apps/api-go/internal/httpserver"
	"smartdelivery/apps/api-go/internal/repository"
	"smartdelivery/apps/api-go/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := db.Connect(ctx, db.Config{DatabaseURL: cfg.DatabaseURL})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(database); err != nil {
			slog.Warn("failed to close database", "error", err)
		}
	}()

	ruleRepository := repository.NewDeliveryRuleRepository(database)
	ruleService := service.NewDeliveryRuleService(ruleRepository)

	if err := httpserver.Run(ctx, httpserver.Config{
		Addr:          cfg.HTTPAddr(),
		DeliveryRules: ruleService,
	}, slog.Default()); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
