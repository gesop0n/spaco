package catalog

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/generated/spaco/catalog/v1/catalogv1connect"
)

// FindProblemsFuncは、登録済みの問題をIDから探す関数である。
type FindProblemsFunc func(context.Context, []ProblemID) ([]Problem, error)

// EnsureProblemFuncは、問題URLから読み取った問題を返し、未登録なら登録する関数である。
type EnsureProblemFunc func(context.Context, ProblemRef) (Problem, error)

// SyncFuncは、AtCoder Problemsからコンテスト・問題情報を取り込む関数である。
type SyncFunc func(context.Context) (SyncResult, error)

// SyncResultは、AtCoder Problemsから取り込んだ件数である。
type SyncResult struct {
	Contests        int
	Problems        int
	ContestProblems int
	// Skippedは、IDの形式が不正などの理由で取り込まなかった行数である。
	Skipped int
}

// Moduleは、組み立て済みcatalog moduleの公開窓口である。
// use caseやrepositoryの具象型はmodule外へ公開しない。
type Module struct {
	handler       catalogv1connect.CatalogServiceHandler
	findProblems  FindProblemsFunc
	ensureProblem EnsureProblemFunc
	sync          SyncFunc
}

func NewModule(
	handler catalogv1connect.CatalogServiceHandler,
	findProblems FindProblemsFunc,
	ensureProblem EnsureProblemFunc,
	sync SyncFunc,
) (*Module, error) {
	if handler == nil {
		return nil, errors.New("create catalog module: handler is required")
	}
	if findProblems == nil || ensureProblem == nil || sync == nil {
		return nil, errors.New("create catalog module: problem operations are required")
	}
	return &Module{
		handler:       handler,
		findProblems:  findProblems,
		ensureProblem: ensureProblem,
		sync:          sync,
	}, nil
}

// FindProblemsは、登録済みの問題を返す。存在しないIDは結果に含めない。
func (m *Module) FindProblems(ctx context.Context, ids []ProblemID) ([]Problem, error) {
	return m.findProblems(ctx, ids)
}

// EnsureProblemは、問題URLから読み取った問題を返す。未登録なら最小限の情報で登録する。
func (m *Module) EnsureProblem(ctx context.Context, ref ProblemRef) (Problem, error) {
	return m.ensureProblem(ctx, ref)
}

// SyncFromAtCoderProblemsは、AtCoder Problemsのコンテスト・問題情報を取り込む。
func (m *Module) SyncFromAtCoderProblems(ctx context.Context) (SyncResult, error) {
	return m.sync(ctx)
}

// ConnectHandlerは、CatalogServiceをmountするpathとHTTP handlerを返す。
func (m *Module) ConnectHandler(options ...connect.HandlerOption) (string, http.Handler) {
	return catalogv1connect.NewCatalogServiceHandler(m.handler, options...)
}
