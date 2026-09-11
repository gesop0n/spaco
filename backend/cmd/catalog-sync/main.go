package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gesop0n/spaco/backend/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("catalog sync failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// .envが存在しなければ無視。開発環境用。設定済みの環境変数は上書きしない。
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startedAt := time.Now()
	result, err := app.RunCatalogSync(ctx, databaseURL)
	if err != nil {
		return err
	}
	slog.Info(
		"catalog synced",
		"contests", result.Contests,
		"problems", result.Problems,
		"contest_problems", result.ContestProblems,
		"skipped", result.Skipped,
		"duration", time.Since(startedAt).String(),
	)
	return nil
}
