// Package domainは、catalog moduleのコンテスト・問題に関するルールを定義する。
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
)

var (
	ErrInvalidContest        = errors.New("invalid contest")
	ErrInvalidContestProblem = errors.New("invalid contest problem")
	ErrInvalidContestQuery   = errors.New("invalid contest query")
)

// Contestは、AtCoder Problemsから取り込んだコンテストである。
type Contest struct {
	id         catalog.ContestID
	title      string
	startAt    time.Time
	duration   time.Duration
	rateChange string
}

func NewContest(
	id catalog.ContestID,
	title string,
	startAt time.Time,
	duration time.Duration,
	rateChange string,
) (Contest, error) {
	if strings.TrimSpace(title) == "" {
		return Contest{}, fmt.Errorf("%w: title is required", ErrInvalidContest)
	}
	if duration < 0 {
		return Contest{}, fmt.Errorf("%w: duration must not be negative", ErrInvalidContest)
	}
	return Contest{
		id:         id,
		title:      title,
		startAt:    startAt.UTC(),
		duration:   duration,
		rateChange: rateChange,
	}, nil
}

func (c Contest) ID() catalog.ContestID   { return c.id }
func (c Contest) Title() string           { return c.title }
func (c Contest) StartAt() time.Time      { return c.startAt }
func (c Contest) Duration() time.Duration { return c.duration }
func (c Contest) RateChange() string      { return c.rateChange }

// ContestProblemは、コンテストと、そのコンテストでの問題番号の対応である。
type ContestProblem struct {
	contestID catalog.ContestID
	problemID catalog.ProblemID
	index     string
}

func NewContestProblem(
	contestID catalog.ContestID,
	problemID catalog.ProblemID,
	index string,
) (ContestProblem, error) {
	if strings.TrimSpace(index) == "" {
		return ContestProblem{}, fmt.Errorf("%w: problem index is required", ErrInvalidContestProblem)
	}
	return ContestProblem{contestID: contestID, problemID: problemID, index: index}, nil
}

func (p ContestProblem) ContestID() catalog.ContestID { return p.contestID }
func (p ContestProblem) ProblemID() catalog.ProblemID { return p.problemID }
func (p ContestProblem) Index() string                { return p.index }

const maximumContestQueryLength = 100

// ContestQueryは、コンテスト検索の入力を、空白と大文字・小文字の違いを無視する形にしたものである。
// 「ABC 350」と入力しても、IDのabc350やタイトルに一致させる。
type ContestQuery struct{ normalized string }

func NewContestQuery(raw string) (ContestQuery, error) {
	if !utf8.ValidString(raw) || utf8.RuneCountInString(raw) > maximumContestQueryLength {
		return ContestQuery{}, fmt.Errorf(
			"%w: must be valid text up to %d characters",
			ErrInvalidContestQuery,
			maximumContestQueryLength,
		)
	}
	return ContestQuery{normalized: strings.Join(strings.Fields(strings.ToLower(raw)), "")}, nil
}

// LikePatternは、PostgreSQLのLIKEで部分一致させるパターンを返す。
// 入力中の%と_は、ワイルドカードではなく文字として扱う。
func (q ContestQuery) LikePattern() string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q.normalized)
	return "%" + escaped + "%"
}
