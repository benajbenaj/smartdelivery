package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"smartdelivery/apps/api-go/internal/httpserver"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := httpserver.Run(ctx, httpserver.Config{Addr: ":8080"}, slog.Default()); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
