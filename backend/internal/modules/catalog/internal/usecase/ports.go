// Package usecaseは、catalog moduleのユースケースと必要なportを定義する。
package usecase

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
)

// PostgreSQLやHTTP clientの型は、いずれのinterfaceにも持ち込まない。

type ISearchContestsRepository interface {
	SearchContests(context.Context, domain.ContestQuery) ([]domain.Contest, error)
}

type IListContestProblemsRepository interface {
	FindContestByID(context.Context, catalog.ContestID) (domain.Contest, error)
	ListContestProblems(context.Context, catalog.ContestID) ([]catalog.Problem, error)
}

type IFindProblemsRepository interface {
	FindProblemsByIDs(context.Context, []catalog.ProblemID) ([]catalog.Problem, error)
}

type IEnsureProblemRepository interface {
	EnsureProblem(context.Context, catalog.ProblemRef) (catalog.Problem, error)
}

type ISyncCatalogRepository interface {
	SaveSnapshot(context.Context, domain.Snapshot) error
}

type IAtCoderProblemsClient interface {
	FetchSnapshot(context.Context) (domain.Snapshot, error)
}
