package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

// TEST_DATABASE_URLが設定された場合だけ、migration適用済みPostgreSQLで確認する。
func TestRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	// 本番と同じく、named prepared statementを使わないexec modeで確認する。
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	defer pool.Close()
	repository, err := NewRepository(pool)
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}

	// 他のデータと衝突しないよう、このtest専用のIDだけを作成・削除する。
	prefix := "itest" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupContext, `DELETE FROM contests WHERE id LIKE $1`, prefix+"%")
		_, _ = pool.Exec(cleanupContext, `DELETE FROM problems WHERE id LIKE $1`, prefix+"%")
	}()

	contestID := mustContestID(t, prefix+"abc")
	otherContestID := mustContestID(t, prefix+"arc")
	problemA := mustProblemID(t, prefix+"abc_a")
	problemB := mustProblemID(t, prefix+"abc_b")
	problemEx := mustProblemID(t, prefix+"abc_h")
	shared := mustProblemID(t, prefix+"arc_a")

	// AtCoder Problemsに無い問題も、URLのコンテストで仮登録できる。
	ref, err := catalog.ParseProblemURL("https://atcoder.jp/contests/" + contestID.String() + "/tasks/" + problemA.String())
	if err != nil {
		t.Fatalf("ParseProblemURL() error = %v", err)
	}
	stub, err := repository.EnsureProblem(ctx, ref)
	if err != nil {
		t.Fatalf("EnsureProblem() error = %v", err)
	}
	if _, ok := stub.Name(); ok || stub.ContestID() != contestID {
		t.Fatalf("stub problem = %+v", stub)
	}

	contest, err := domain.NewContest(contestID, "Integration Beginner Contest", time.Unix(1713614400, 0), 100*time.Minute, " ~ 1999")
	if err != nil {
		t.Fatalf("NewContest() error = %v", err)
	}
	permanent, err := domain.NewContest(otherContestID, "常設コンテスト", time.Unix(0, 0), 3153600000*time.Second, "-")
	if err != nil {
		t.Fatalf("NewContest() error = %v", err)
	}
	snapshot := domain.Snapshot{
		Contests: []domain.Contest{contest, permanent},
		Problems: []catalog.Problem{
			catalog.NewProblem(problemA, contestID, "A", "Past ABCs"),
			catalog.NewProblem(problemB, contestID, "B", "Dentist Aoki"),
			catalog.NewProblem(problemEx, contestID, "Ex", "Hard Problem"),
			catalog.NewProblem(shared, otherContestID, "A", "Shared Problem"),
		},
		ContestProblems: []domain.ContestProblem{
			mustContestProblem(t, contestID, problemEx, "Ex"),
			mustContestProblem(t, contestID, problemB, "B"),
			mustContestProblem(t, contestID, problemA, "A"),
			mustContestProblem(t, contestID, shared, "C"),
			mustContestProblem(t, otherContestID, shared, "A"),
		},
	}
	// 同じ内容で再実行しても、追加・更新だけで失敗しない。
	for range 2 {
		if err := repository.SaveSnapshot(ctx, snapshot); err != nil {
			t.Fatalf("SaveSnapshot() error = %v", err)
		}
	}

	filled, err := repository.EnsureProblem(ctx, ref)
	if err != nil {
		t.Fatalf("EnsureProblem() after sync error = %v", err)
	}
	if name, ok := filled.Name(); !ok || name != "Past ABCs" {
		t.Fatalf("filled problem name = %q, %v", name, ok)
	}

	contests, err := repository.SearchContests(ctx, mustContestQuery(t, " "+strings.ToUpper(prefix)+" ABC "))
	if err != nil {
		t.Fatalf("SearchContests() error = %v", err)
	}
	if len(contests) != 1 || contests[0].ID() != contestID || !contests[0].StartAt().Equal(contest.StartAt()) {
		t.Fatalf("SearchContests() = %+v", contests)
	}
	stored, err := repository.FindContestByID(ctx, otherContestID)
	if err != nil || stored.Duration() != permanent.Duration() {
		t.Fatalf("FindContestByID() = %+v, %v", stored, err)
	}
	if _, err := repository.FindContestByID(ctx, mustContestID(t, prefix+"missing")); !errors.Is(err, usecase.ErrContestNotFound) {
		t.Fatalf("missing contest error = %v, want ErrContestNotFound", err)
	}

	problems, err := repository.ListContestProblems(ctx, contestID)
	if err != nil {
		t.Fatalf("ListContestProblems() error = %v", err)
	}
	var indexes []string
	for _, problem := range problems {
		index, _ := problem.Index()
		indexes = append(indexes, index)
		if problem.ContestID() != contestID {
			t.Fatalf("problem %s contest = %s, want listing contest", problem.ID(), problem.ContestID())
		}
	}
	if strings.Join(indexes, ",") != "A,B,C,Ex" {
		t.Fatalf("problem indexes = %v, want A,B,C,Ex", indexes)
	}

	found, err := repository.FindProblemsByIDs(ctx, []catalog.ProblemID{shared, mustProblemID(t, prefix+"missing")})
	if err != nil || len(found) != 1 || found[0].ContestID() != otherContestID {
		t.Fatalf("FindProblemsByIDs() = %+v, %v", found, err)
	}
}

func mustContestQuery(t *testing.T, value string) domain.ContestQuery {
	t.Helper()
	query, err := domain.NewContestQuery(value)
	if err != nil {
		t.Fatalf("NewContestQuery() error = %v", err)
	}
	return query
}

func mustContestProblem(t *testing.T, contestID catalog.ContestID, problemID catalog.ProblemID, index string) domain.ContestProblem {
	t.Helper()
	pair, err := domain.NewContestProblem(contestID, problemID, index)
	if err != nil {
		t.Fatalf("NewContestProblem() error = %v", err)
	}
	return pair
}

func mustContestID(t *testing.T, value string) catalog.ContestID {
	t.Helper()
	id, err := catalog.ParseContestID(value)
	if err != nil {
		t.Fatalf("ParseContestID() error = %v", err)
	}
	return id
}

func mustProblemID(t *testing.T, value string) catalog.ProblemID {
	t.Helper()
	id, err := catalog.ParseProblemID(value)
	if err != nil {
		t.Fatalf("ParseProblemID() error = %v", err)
	}
	return id
}
