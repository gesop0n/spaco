// Package factoryは、catalog module内部のuse caseとadapterを組み立てる。
package factory

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/adapter/atcoderproblems"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/adapter/postgres"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/adapter/rpc"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

// Configは、catalog moduleの外部接続設定を保持する。
type Config struct {
	// HTTPClientは、AtCoder Problemsへのアクセスに使う。未指定なら2分でtimeoutするclientを使う。
	HTTPClient *http.Client
	// AtCoderProblemsBaseURLは、未指定ならAtCoder Problemsの公開URLを使う。
	AtCoderProblemsBaseURL string
}

// Newは、PostgreSQL repository、AtCoder Problems client、use case、ConnectRPC handlerを接続する。
func New(pool *pgxpool.Pool, config Config) (*catalog.Module, error) {
	repository, err := postgres.NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}
	baseURL := config.AtCoderProblemsBaseURL
	if baseURL == "" {
		baseURL = atcoderproblems.DefaultBaseURL
	}
	client, err := atcoderproblems.NewClient(httpClient, baseURL)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}

	searchContests, err := usecase.NewSearchContests(repository)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	listContestProblems, err := usecase.NewListContestProblems(repository)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	findProblems, err := usecase.NewFindProblems(repository)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	ensureProblem, err := usecase.NewEnsureProblem(repository)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	syncCatalog, err := usecase.NewSyncCatalog(client, repository)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	handler, err := rpc.NewHandler(searchContests, listContestProblems)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}

	module, err := catalog.NewModule(handler, findProblems.Execute, ensureProblem.Execute, syncCatalog.Execute)
	if err != nil {
		return nil, fmt.Errorf("create catalog module: %w", err)
	}
	return module, nil
}
