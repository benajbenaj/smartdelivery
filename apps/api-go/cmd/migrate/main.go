package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"smartdelivery/apps/api-go/internal/db"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: migrate up|down|status")
	}

	database, err := db.Connect(ctx, db.Config{DatabaseURL: os.Getenv("DATABASE_URL")})
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(database); err != nil {
			slog.Warn("failed to close database", "error", err)
		}
	}()

	switch args[1] {
	case "up":
		return db.MigrateUp(ctx, database)
	case "down":
		return db.MigrateDown(ctx, database)
	case "status":
		status, err := db.CheckMigrationStatus(ctx, database)
		if err != nil {
			return err
		}
		for _, table := range status.Tables {
			fmt.Printf("%s: %t\n", table.Name, table.Exists)
		}
		return nil
	default:
		return fmt.Errorf("unknown migration command %q", args[1])
	}
}
