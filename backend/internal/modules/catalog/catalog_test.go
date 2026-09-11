package catalog

import (
	"errors"
	"testing"
)

func TestParseProblemURLAcceptsAtCoderProblemPages(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		raw       string
		contestID string
		problemID string
	}{
		{
			raw:       " https://atcoder.jp/contests/abc350/tasks/abc350_a/?lang=ja#task-statement ",
			contestID: "abc350",
			problemID: "abc350_a",
		},
		// 大文字を含むIDや、別コンテストの問題を共有するURLもそのまま扱う。
		{raw: "https://atcoder.jp/contests/APG4b/tasks/APG4b_a", contestID: "APG4b", problemID: "APG4b_a"},
		{raw: "https://atcoder.jp/contests/abc042/tasks/arc058_a", contestID: "abc042", problemID: "arc058_a"},
	} {
		ref, err := ParseProblemURL(test.raw)
		if err != nil {
			t.Fatalf("ParseProblemURL(%q) error = %v", test.raw, err)
		}
		if ref.ContestID().String() != test.contestID || ref.ProblemID().String() != test.problemID {
			t.Fatalf("ParseProblemURL(%q) = %s, %s", test.raw, ref.ContestID(), ref.ProblemID())
		}
	}
}

func TestParseProblemURLRejectsOtherURLs(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"",
		"javascript:alert(1)",
		"http://atcoder.jp/contests/abc350/tasks/abc350_a",
		"https://atcoder.jp.evil.test/contests/abc350/tasks/abc350_a",
		"https://atcoder.jp@evil.test/contests/abc350/tasks/abc350_a",
		"https://user:pass@atcoder.jp/contests/abc350/tasks/abc350_a",
		"https://atcoder.jp:8443/contests/abc350/tasks/abc350_a",
		"https://atcoder.jp/contests/abc350",
		"https://atcoder.jp/contests/abc350/tasks/a/extra",
		"https://atcoder.jp//contests/abc350/tasks/abc350_a",
		"https://atcoder.jp/contests/abc350/tasks/abc%2F350",
	} {
		if _, err := ParseProblemURL(raw); !errors.Is(err, ErrInvalidProblemURL) {
			t.Errorf("ParseProblemURL(%q) error = %v, want ErrInvalidProblemURL", raw, err)
		}
	}
}

func TestProblemURLAndOptionalFields(t *testing.T) {
	t.Parallel()

	problemID, err := ParseProblemID("arc058_a")
	if err != nil {
		t.Fatalf("ParseProblemID() error = %v", err)
	}
	contestID, err := ParseContestID("abc042")
	if err != nil {
		t.Fatalf("ParseContestID() error = %v", err)
	}
	problem := NewProblem(problemID, contestID, "", "")
	if problem.URL() != "https://atcoder.jp/contests/abc042/tasks/arc058_a" {
		t.Fatalf("URL() = %q", problem.URL())
	}
	if _, ok := problem.Index(); ok {
		t.Fatal("Index() of an unfilled problem must be unset")
	}
	if _, ok := problem.Name(); ok {
		t.Fatal("Name() of an unfilled problem must be unset")
	}
	if _, err := ParseProblemID("abc 350"); !errors.Is(err, ErrInvalidProblemID) {
		t.Fatalf("ParseProblemID() error = %v, want ErrInvalidProblemID", err)
	}
}
