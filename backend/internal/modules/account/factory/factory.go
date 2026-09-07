// Package factory は、account module内部のuse caseとadapterを組み立てる。
package factory

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/account"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/adapter/postgres"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/adapter/rpc"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/application"
)

// Newは、PostgreSQL repository、application service、ConnectRPC handlerを接続する。
func New(pool *pgxpool.Pool) (*account.Module, error) {
	repository, err := postgres.NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	service, err := application.NewService(repository)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	handler, err := rpc.NewHandler(service)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}

	module, err := account.NewModule(handler, service.ResolveUser)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	return module, nil
}
