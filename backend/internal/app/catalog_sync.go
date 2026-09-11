package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	catalogfactory "github.com/gesop0n/spaco/backend/internal/modules/catalog/factory"
)

// RunCatalogSyncは、AtCoder Problemsのコンテスト・問題情報をDBへ取り込む。
// API serverとは別のプロセスとして、手動または定期実行で使う。
func RunCatalogSync(ctx context.Context, databaseURL string) (catalog.SyncResult, error) {
	databaseConfig, err := databasePoolConfig(databaseURL)
	if err != nil {
		return catalog.SyncResult{}, err
	}
	database, err := pgxpool.NewWithConfig(ctx, databaseConfig)
	if err != nil {
		return catalog.SyncResult{}, fmt.Errorf("create database pool: %w", err)
	}
	defer database.Close()
	if err := database.Ping(ctx); err != nil {
		return catalog.SyncResult{}, fmt.Errorf("ping database: %w", err)
	}

	catalogModule, err := catalogfactory.New(database, catalogfactory.Config{})
	if err != nil {
		return catalog.SyncResult{}, err
	}
	return catalogModule.SyncFromAtCoderProblems(ctx)
}
