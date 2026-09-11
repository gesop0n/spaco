// Package catalogは、AtCoderのコンテスト・問題の基本情報を扱うmoduleである。
// 他moduleが使う問題IDなどの値と、組み立て済みmoduleの公開窓口をこのパッケージに置く。
package catalog

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidContestID  = errors.New("invalid contest id")
	ErrInvalidProblemID  = errors.New("invalid problem id")
	ErrInvalidProblemURL = errors.New("invalid problem url")
)

// identifierPatternは、AtCoderのコンテストIDと問題IDに使われる文字である。
// APG4bのように大文字を含むIDもあるため、大文字と小文字を区別して保持する。
var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// ContestIDは、AtCoderのコンテストIDである。例: abc350
type ContestID struct{ value string }

func ParseContestID(value string) (ContestID, error) {
	if !identifierPattern.MatchString(value) {
		return ContestID{}, fmt.Errorf("%w: %q", ErrInvalidContestID, value)
	}
	return ContestID{value: value}, nil
}

func (id ContestID) String() string { return id.value }
func (id ContestID) IsZero() bool   { return id.value == "" }

// ProblemIDは、AtCoderの問題IDである。例: abc350_d
type ProblemID struct{ value string }

func ParseProblemID(value string) (ProblemID, error) {
	if !identifierPattern.MatchString(value) {
		return ProblemID{}, fmt.Errorf("%w: %q", ErrInvalidProblemID, value)
	}
	return ProblemID{value: value}, nil
}

func (id ProblemID) String() string { return id.value }
func (id ProblemID) IsZero() bool   { return id.value == "" }

// ProblemRefは、問題URLから読み取ったコンテストIDと問題IDの組である。
type ProblemRef struct {
	contestID ContestID
	problemID ProblemID
}

// ParseProblemURLは、https://atcoder.jp/contests/{contest}/tasks/{problem} の形のURLを読み取る。
// 形を確認するだけで、AtCoderへアクセスして問題の実在は確認しない。
func ParseProblemURL(raw string) (ProblemRef, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "atcoder.jp") ||
		parsed.User != nil || parsed.Opaque != "" {
		return ProblemRef{}, fmt.Errorf("%w: must be an https://atcoder.jp URL", ErrInvalidProblemURL)
	}
	segments := strings.Split(strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), "/"), "/")
	if len(segments) != 4 || segments[0] != "contests" || segments[2] != "tasks" {
		return ProblemRef{}, fmt.Errorf(
			"%w: path must be /contests/{contest}/tasks/{problem}",
			ErrInvalidProblemURL,
		)
	}
	contestID, err := ParseContestID(segments[1])
	if err != nil {
		return ProblemRef{}, fmt.Errorf("%w: %w", ErrInvalidProblemURL, err)
	}
	problemID, err := ParseProblemID(segments[3])
	if err != nil {
		return ProblemRef{}, fmt.Errorf("%w: %w", ErrInvalidProblemURL, err)
	}
	return ProblemRef{contestID: contestID, problemID: problemID}, nil
}

func (ref ProblemRef) ContestID() ContestID { return ref.contestID }
func (ref ProblemRef) ProblemID() ProblemID { return ref.problemID }

// Problemは、他moduleへ公開する問題の基本情報である。
type Problem struct {
	id        ProblemID
	contestID ContestID
	index     string
	name      string
}

// NewProblemは、問題の基本情報を返す。まだ補完していないindexとnameには空文字を渡す。
func NewProblem(id ProblemID, contestID ContestID, index, name string) Problem {
	return Problem{id: id, contestID: contestID, index: index, name: name}
}

func (p Problem) ID() ProblemID        { return p.id }
func (p Problem) ContestID() ContestID { return p.contestID }

// Indexは、コンテスト内の問題番号を返す。URLから登録し、まだ補完していなければfalseを返す。
func (p Problem) Index() (string, bool) { return p.index, p.index != "" }

// Nameは、問題名を返す。URLから登録し、まだ補完していなければfalseを返す。
func (p Problem) Name() (string, bool) { return p.name, p.name != "" }

// URLは、AtCoderの問題ページのURLを返す。IDはURLで安全な文字だけで構成される。
func (p Problem) URL() string {
	return "https://atcoder.jp/contests/" + p.contestID.value + "/tasks/" + p.id.value
}
