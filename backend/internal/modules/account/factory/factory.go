// Package factory は、account module内部のuse caseとadapterを組み立てる。
package factory

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/account"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/adapter/postgres"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/adapter/rpc"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/usecase"
)

// Newは、PostgreSQL repository、use case、ConnectRPC handlerを接続する。
func New(pool *pgxpool.Pool) (*account.Module, error) {
	repository, err := postgres.NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	resolveUser, err := usecase.NewResolveUser(repository)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	getCurrentAccount, err := usecase.NewGetCurrentAccount(repository)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	updateProfile, err := usecase.NewUpdateProfile(repository)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	handler, err := rpc.NewHandler(getCurrentAccount, updateProfile)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}

	module, err := account.NewModule(handler, resolveUser.Execute)
	if err != nil {
		return nil, fmt.Errorf("create account module: %w", err)
	}
	return module, nil
}
