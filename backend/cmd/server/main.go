package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gesop0n/spaco/backend/internal/app"
	"github.com/gesop0n/spaco/backend/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	// application全体のslogをJSONへ統一し、RPCログを収集基盤で扱いやすくする。
	// tokenやrequest bodyなどの機密情報は、各ログ出力側で属性に含めない。
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// .envが存在しなければ無視。開発環境用
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := app.NewServer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	return server.Run(ctx)
}
