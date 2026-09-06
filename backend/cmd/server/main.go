package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/gesop0n/spaco/backend/internal/config"
	"github.com/joho/godotenv"
)

func main() {
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

	fmt.Println(cfg)

	return nil
}
