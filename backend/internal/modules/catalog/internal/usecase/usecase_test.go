package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
)

type repositoryStub struct {
	searchContests      func(context.Context, domain.ContestQuery) ([]domain.Contest, error)
	findContestByID     func(context.Context, catalog.ContestID) (domain.Contest, error)
	listContestProblems func(context.Context, catalog.ContestID) ([]catalog.Problem, error)
	findProblemsByIDs   func(context.Context, []catalog.ProblemID) ([]catalog.Problem, error)
	saveSnapshot        func(context.Context, domain.Snapshot) error
}

func (s repositoryStub) SearchContests(ctx context.Context, query domain.ContestQuery) ([]domain.Contest, error) {
	return s.searchContests(ctx, query)
}

func (s repositoryStub) FindContestByID(ctx context.Context, id catalog.ContestID) (domain.Contest, error) {
	return s.findContestByID(ctx, id)
}

func (s repositoryStub) ListContestProblems(ctx context.Context, id catalog.ContestID) ([]catalog.Problem, error) {
	return s.listContestProblems(ctx, id)
}

func (s repositoryStub) FindProblemsByIDs(ctx context.Context, ids []catalog.ProblemID) ([]catalog.Problem, error) {
	return s.findProblemsByIDs(ctx, ids)
}

func (s repositoryStub) SaveSnapshot(ctx context.Context, snapshot domain.Snapshot) error {
	return s.saveSnapshot(ctx, snapshot)
}

type clientStub struct {
	snapshot domain.Snapshot
	err      error
}

func (s clientStub) FetchSnapshot(context.Context) (domain.Snapshot, error) {
	return s.snapshot, s.err
}

func TestSearchContestsRejectsLongQueryBeforeRepository(t *testing.T) {
	t.Parallel()

	searchContests, err := NewSearchContests(repositoryStub{
		searchContests: func(context.Context, domain.ContestQuery) ([]domain.Contest, error) {
			t.Fatal("repository must not be called")
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("NewSearchContests() error = %v", err)
	}
	if _, err := searchContests.Execute(context.Background(), strings.Repeat("a", 101)); !errors.Is(err, domain.ErrInvalidContestQuery) {
		t.Fatalf("Execute() error = %v, want ErrInvalidContestQuery", err)
	}
}

func TestListContestProblemsValidatesContestID(t *testing.T) {
	t.Parallel()

	listContestProblems, err := NewListContestProblems(repositoryStub{})
	if err != nil {
		t.Fatalf("NewListContestProblems() error = %v", err)
	}
	if _, err := listContestProblems.Execute(context.Background(), "abc/350"); !errors.Is(err, catalog.ErrInvalidContestID) {
		t.Fatalf("Execute() error = %v, want ErrInvalidContestID", err)
	}
}

func TestListContestProblemsReturnsNotFound(t *testing.T) {
	t.Parallel()

	listContestProblems, err := NewListContestProblems(repositoryStub{
		findContestByID: func(context.Context, catalog.ContestID) (domain.Contest, error) {
			return domain.Contest{}, ErrContestNotFound
		},
	})
	if err != nil {
		t.Fatalf("NewListContestProblems() error = %v", err)
	}
	if _, err := listContestProblems.Execute(context.Background(), " abc350 "); !errors.Is(err, ErrContestNotFound) {
		t.Fatalf("Execute() error = %v, want ErrContestNotFound", err)
	}
}

func TestFindProblemsDeduplicatesAndSkipsEmptyInput(t *testing.T) {
	t.Parallel()

	calls := 0
	findProblems, err := NewFindProblems(repositoryStub{
		findProblemsByIDs: func(_ context.Context, ids []catalog.ProblemID) ([]catalog.Problem, error) {
			calls++
			if len(ids) != 2 {
				t.Fatalf("ids = %v, want 2 unique ids", ids)
			}
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("NewFindProblems() error = %v", err)
	}
	if _, err := findProblems.Execute(context.Background(), nil); err != nil || calls != 0 {
		t.Fatalf("empty Execute() error = %v, calls = %d", err, calls)
	}
	a := mustProblemID(t, "abc350_a")
	b := mustProblemID(t, "abc350_b")
	if _, err := findProblems.Execute(context.Background(), []catalog.ProblemID{a, b, a}); err != nil || calls != 1 {
		t.Fatalf("Execute() error = %v, calls = %d", err, calls)
	}
	if _, err := findProblems.Execute(context.Background(), []catalog.ProblemID{{}}); !errors.Is(err, catalog.ErrInvalidProblemID) {
		t.Fatalf("zero id error = %v, want ErrInvalidProblemID", err)
	}
}

func TestSyncCatalogRejectsEmptySnapshot(t *testing.T) {
	t.Parallel()

	syncCatalog, err := NewSyncCatalog(clientStub{}, repositoryStub{
		saveSnapshot: func(context.Context, domain.Snapshot) error {
			t.Fatal("empty snapshot must not be saved")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewSyncCatalog() error = %v", err)
	}
	if _, err := syncCatalog.Execute(context.Background()); !errors.Is(err, ErrEmptySnapshot) {
		t.Fatalf("Execute() error = %v, want ErrEmptySnapshot", err)
	}
}

func TestSyncCatalogSavesNormalizedSnapshot(t *testing.T) {
	t.Parallel()

	contestID := mustContestID(t, "abc350")
	problemID := mustProblemID(t, "abc350_a")
	contest, err := domain.NewContest(contestID, "AtCoder Beginner Contest 350", time.Unix(1713614400, 0), 100*time.Minute, " ~ 1999")
	if err != nil {
		t.Fatalf("NewContest() error = %v", err)
	}
	problem := catalog.NewProblem(problemID, contestID, "A", "Past ABCs")
	pair, err := domain.NewContestProblem(contestID, problemID, "A")
	if err != nil {
		t.Fatalf("NewContestProblem() error = %v", err)
	}
	var saved domain.Snapshot
	syncCatalog, err := NewSyncCatalog(
		clientStub{snapshot: domain.Snapshot{
			Contests:        []domain.Contest{contest},
			Problems:        []catalog.Problem{problem, problem},
			ContestProblems: []domain.ContestProblem{pair},
		}},
		repositoryStub{saveSnapshot: func(_ context.Context, snapshot domain.Snapshot) error {
			saved = snapshot
			return nil
		}},
	)
	if err != nil {
		t.Fatalf("NewSyncCatalog() error = %v", err)
	}
	result, err := syncCatalog.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := catalog.SyncResult{Contests: 1, Problems: 1, ContestProblems: 1, Skipped: 1}
	if result != want || len(saved.Problems) != 1 {
		t.Fatalf("Execute() = %+v, saved %d problems; want %+v", result, len(saved.Problems), want)
	}
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
