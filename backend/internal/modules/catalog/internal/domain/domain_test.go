package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
)

func TestContestQueryIgnoresSpacesAndCaseAndEscapesWildcards(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		raw  string
		want string
	}{
		{raw: " ABC 350 ", want: "%abc350%"},
		{raw: "", want: "%%"},
		{raw: "100%_\\", want: `%100\%\_\\%`},
	} {
		query, err := NewContestQuery(test.raw)
		if err != nil {
			t.Fatalf("NewContestQuery(%q) error = %v", test.raw, err)
		}
		if got := query.LikePattern(); got != test.want {
			t.Errorf("LikePattern(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
	if _, err := NewContestQuery(strings.Repeat("a", 101)); !errors.Is(err, ErrInvalidContestQuery) {
		t.Fatalf("long query error = %v, want ErrInvalidContestQuery", err)
	}
}

func TestSnapshotNormalizedRemovesDuplicatesAndDanglingPairs(t *testing.T) {
	t.Parallel()

	abc042 := mustContestID(t, "abc042")
	arc058 := mustContestID(t, "arc058")
	shared := mustProblemID(t, "arc058_a")
	contest := mustContest(t, abc042)
	snapshot := Snapshot{
		Contests: []Contest{contest, contest},
		Problems: []catalog.Problem{
			catalog.NewProblem(shared, arc058, "A", "Iroha's Obsession"),
			catalog.NewProblem(shared, arc058, "A", "duplicated"),
		},
		ContestProblems: []ContestProblem{
			mustContestProblem(t, abc042, shared, "C"),
			mustContestProblem(t, abc042, shared, "C"),
			// arc058はこのSnapshotのコンテストに含まれない。
			mustContestProblem(t, arc058, shared, "A"),
		},
		Skipped: 1,
	}

	normalized := snapshot.Normalized()
	if len(normalized.Contests) != 1 || len(normalized.Problems) != 1 || len(normalized.ContestProblems) != 1 {
		t.Fatalf("Normalized() = %+v", normalized)
	}
	if name, _ := normalized.Problems[0].Name(); name != "Iroha's Obsession" {
		t.Fatalf("kept problem name = %q, want the first row", name)
	}
	if normalized.Skipped != 5 {
		t.Fatalf("Skipped = %d, want 5", normalized.Skipped)
	}
}

func mustContest(t *testing.T, id catalog.ContestID) Contest {
	t.Helper()
	contest, err := NewContest(id, "AtCoder Beginner Contest 042", time.Unix(1469275200, 0), 100*time.Minute, "All")
	if err != nil {
		t.Fatalf("NewContest() error = %v", err)
	}
	return contest
}

func mustContestProblem(t *testing.T, contestID catalog.ContestID, problemID catalog.ProblemID, index string) ContestProblem {
	t.Helper()
	pair, err := NewContestProblem(contestID, problemID, index)
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
