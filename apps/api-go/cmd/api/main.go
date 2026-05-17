package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"smartdelivery/apps/api-go/internal/audit"
	"smartdelivery/apps/api-go/internal/config"
	"smartdelivery/apps/api-go/internal/db"
	"smartdelivery/apps/api-go/internal/httpserver"
	"smartdelivery/apps/api-go/internal/repository"
	"smartdelivery/apps/api-go/internal/service"
	"smartdelivery/apps/api-go/internal/shopify"
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
	shopRepository := repository.NewShopRepository(database)
	auditWorker := audit.NewWorker(audit.NewStore(database), audit.Config{
		BufferSize:       cfg.Audit.BufferSize,
		OverflowBehavior: audit.OverflowBehavior(cfg.Audit.OverflowBehavior),
		Logger:           slog.Default(),
	})
	auditCtx, stopAudit := context.WithCancel(context.Background())
	defer stopAudit()

	auditErrCh := make(chan error, 1)
	go func() {
		auditErrCh <- auditWorker.Run(auditCtx)
	}()

	ruleService := service.NewDeliveryRuleServiceWithAudit(ruleRepository, auditWorker)
	shopifyOAuthService := shopify.NewOAuthService(shopRepository, shopify.OAuthConfig{
		APIKey: cfg.Shopify.APIKey,
	})

	err = httpserver.Run(ctx, httpserver.Config{
		Addr:                cfg.HTTPAddr(),
		DeliveryRules:       ruleService,
		ShopifyInstallation: shopifyOAuthService,
	}, slog.Default())

	stopAudit()
	if auditErr := <-auditErrCh; auditErr != nil && !errors.Is(auditErr, context.Canceled) {
		slog.Error("audit worker stopped", "error", auditErr)
	}

	if err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
