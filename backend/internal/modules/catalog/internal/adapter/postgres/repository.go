// Package postgresは、catalog usecaseが要求するrepositoryをPostgreSQLで実装する。
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	catalogsqlc "github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/adapter/postgres/sqlc"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *catalogsqlc.Queries
}

var (
	_ usecase.ISearchContestsRepository      = (*Repository)(nil)
	_ usecase.IListContestProblemsRepository = (*Repository)(nil)
	_ usecase.IFindProblemsRepository        = (*Repository)(nil)
	_ usecase.IEnsureProblemRepository       = (*Repository)(nil)
	_ usecase.ISyncCatalogRepository         = (*Repository)(nil)
)

func NewRepository(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("create catalog repository: pool is required")
	}
	return &Repository{pool: pool, queries: catalogsqlc.New(pool)}, nil
}

func (r *Repository) SearchContests(ctx context.Context, query domain.ContestQuery) ([]domain.Contest, error) {
	rows, err := r.queries.SearchContests(ctx, query.LikePattern())
	if err != nil {
		return nil, fmt.Errorf("search contests: %w", err)
	}
	contests := make([]domain.Contest, 0, len(rows))
	for _, row := range rows {
		contest, err := rehydrateContest(row.ID, row.Title, row.StartAt, row.DurationSeconds, row.RateChange)
		if err != nil {
			return nil, err
		}
		contests = append(contests, contest)
	}
	return contests, nil
}

func (r *Repository) FindContestByID(ctx context.Context, id catalog.ContestID) (domain.Contest, error) {
	row, err := r.queries.FindContestByID(ctx, id.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Contest{}, usecase.ErrContestNotFound
	}
	if err != nil {
		return domain.Contest{}, fmt.Errorf("find contest by id: %w", err)
	}
	return rehydrateContest(row.ID, row.Title, row.StartAt, row.DurationSeconds, row.RateChange)
}

// ListContestProblemsは、コンテストでの問題番号とコンテストIDを持つ問題を返す。
// 他のコンテストと共有する問題も、指定したコンテストの問題として扱う。
func (r *Repository) ListContestProblems(ctx context.Context, contestID catalog.ContestID) ([]catalog.Problem, error) {
	rows, err := r.queries.ListContestProblems(ctx, contestID.String())
	if err != nil {
		return nil, fmt.Errorf("list contest problems: %w", err)
	}
	problems := make([]catalog.Problem, 0, len(rows))
	for _, row := range rows {
		problemID, err := catalog.ParseProblemID(row.ID)
		if err != nil {
			return nil, fmt.Errorf("parse stored problem id: %w", err)
		}
		problems = append(problems, catalog.NewProblem(problemID, contestID, row.ProblemIndex, valueOrEmpty(row.Name)))
	}
	return problems, nil
}

func (r *Repository) FindProblemsByIDs(ctx context.Context, ids []catalog.ProblemID) ([]catalog.Problem, error) {
	values := make([]string, len(ids))
	for index, id := range ids {
		values[index] = id.String()
	}
	rows, err := r.queries.FindProblemsByIDs(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("find problems by ids: %w", err)
	}
	problems := make([]catalog.Problem, 0, len(rows))
	for _, row := range rows {
		problem, err := rehydrateProblem(row.ID, row.ContestID, row.ProblemIndex, row.Name)
		if err != nil {
			return nil, err
		}
		problems = append(problems, problem)
	}
	return problems, nil
}

// EnsureProblemは、問題が未登録ならURLのコンテストで登録し、保存済みの情報を返す。
// 挿入と取得を別のstatementにして、同じ問題を並行して登録した場合も相手の行を読めるようにする。
func (r *Repository) EnsureProblem(ctx context.Context, ref catalog.ProblemRef) (catalog.Problem, error) {
	if err := r.queries.InsertProblemIfNotExists(ctx, catalogsqlc.InsertProblemIfNotExistsParams{
		ID:        ref.ProblemID().String(),
		ContestID: ref.ContestID().String(),
	}); err != nil {
		return catalog.Problem{}, fmt.Errorf("insert problem: %w", err)
	}
	row, err := r.queries.FindProblemByID(ctx, ref.ProblemID().String())
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.Problem{}, usecase.ErrProblemNotFound
	}
	if err != nil {
		return catalog.Problem{}, fmt.Errorf("find problem by id: %w", err)
	}
	return rehydrateProblem(row.ID, row.ContestID, row.ProblemIndex, row.Name)
}

// SaveSnapshotは、AtCoder Problemsの情報を1つのtransactionで追加・更新する。
func (r *Repository) SaveSnapshot(ctx context.Context, snapshot domain.Snapshot) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin save snapshot transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	queries := r.queries.WithTx(tx)

	contests := catalogsqlc.UpsertContestsParams{
		Ids:              make([]string, len(snapshot.Contests)),
		Titles:           make([]string, len(snapshot.Contests)),
		StartAts:         make([]time.Time, len(snapshot.Contests)),
		DurationsSeconds: make([]int64, len(snapshot.Contests)),
		RateChanges:      make([]string, len(snapshot.Contests)),
	}
	for index, contest := range snapshot.Contests {
		contests.Ids[index] = contest.ID().String()
		contests.Titles[index] = contest.Title()
		contests.StartAts[index] = contest.StartAt()
		contests.DurationsSeconds[index] = int64(contest.Duration() / time.Second)
		contests.RateChanges[index] = contest.RateChange()
	}
	if err := queries.UpsertContests(ctx, contests); err != nil {
		return fmt.Errorf("upsert contests: %w", err)
	}

	problems := catalogsqlc.UpsertProblemsParams{
		Ids:            make([]string, len(snapshot.Problems)),
		ContestIds:     make([]string, len(snapshot.Problems)),
		ProblemIndexes: make([]string, len(snapshot.Problems)),
		Names:          make([]string, len(snapshot.Problems)),
	}
	for index, problem := range snapshot.Problems {
		problems.Ids[index] = problem.ID().String()
		problems.ContestIds[index] = problem.ContestID().String()
		problems.ProblemIndexes[index], _ = problem.Index()
		problems.Names[index], _ = problem.Name()
	}
	if err := queries.UpsertProblems(ctx, problems); err != nil {
		return fmt.Errorf("upsert problems: %w", err)
	}

	pairs := catalogsqlc.UpsertContestProblemsParams{
		ContestIds:     make([]string, len(snapshot.ContestProblems)),
		ProblemIds:     make([]string, len(snapshot.ContestProblems)),
		ProblemIndexes: make([]string, len(snapshot.ContestProblems)),
	}
	for index, pair := range snapshot.ContestProblems {
		pairs.ContestIds[index] = pair.ContestID().String()
		pairs.ProblemIds[index] = pair.ProblemID().String()
		pairs.ProblemIndexes[index] = pair.Index()
	}
	if err := queries.UpsertContestProblems(ctx, pairs); err != nil {
		return fmt.Errorf("upsert contest problems: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit save snapshot: %w", err)
	}
	return nil
}

func rehydrateContest(
	id string,
	title string,
	startAt time.Time,
	durationSeconds int64,
	rateChange string,
) (domain.Contest, error) {
	contestID, err := catalog.ParseContestID(id)
	if err != nil {
		return domain.Contest{}, fmt.Errorf("parse stored contest id: %w", err)
	}
	contest, err := domain.NewContest(contestID, title, startAt, time.Duration(durationSeconds)*time.Second, rateChange)
	if err != nil {
		return domain.Contest{}, fmt.Errorf("rehydrate stored contest: %w", err)
	}
	return contest, nil
}

func rehydrateProblem(id, contestID string, index, name *string) (catalog.Problem, error) {
	problemID, err := catalog.ParseProblemID(id)
	if err != nil {
		return catalog.Problem{}, fmt.Errorf("parse stored problem id: %w", err)
	}
	parsedContestID, err := catalog.ParseContestID(contestID)
	if err != nil {
		return catalog.Problem{}, fmt.Errorf("parse stored contest id: %w", err)
	}
	return catalog.NewProblem(problemID, parsedContestID, valueOrEmpty(index), valueOrEmpty(name)), nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
