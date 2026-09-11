// Package atcoderproblemsは、AtCoder ProblemsのInformation APIからコンテスト・問題情報を取得する。
package atcoderproblems

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

const (
	DefaultBaseURL = "https://kenkoooo.com/atcoder/resources/"

	// AtCoder Problemsは、APIへのアクセスの間を1秒以上空けるよう求めている。
	requestInterval  = 1500 * time.Millisecond
	maxResponseBytes = 64 << 20
	userAgent        = "spaco-catalog-sync"
)

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	sleep      func(context.Context, time.Duration) error
}

var _ usecase.IAtCoderProblemsClient = (*Client)(nil)

func NewClient(httpClient *http.Client, baseURL string) (*Client, error) {
	if httpClient == nil {
		return nil, errors.New("create atcoder problems client: http client is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return nil, fmt.Errorf("create atcoder problems client: invalid base url %q", baseURL)
	}
	return &Client{httpClient: httpClient, baseURL: parsed, sleep: sleepContext}, nil
}

// FetchSnapshotは、コンテスト・問題・対応表の3つのファイルを、間隔を空けて順に取得する。
func (c *Client) FetchSnapshot(ctx context.Context) (domain.Snapshot, error) {
	var contests []contestJSON
	if err := c.getJSON(ctx, "contests.json", &contests); err != nil {
		return domain.Snapshot{}, err
	}
	if err := c.sleep(ctx, requestInterval); err != nil {
		return domain.Snapshot{}, err
	}
	var problems []problemJSON
	if err := c.getJSON(ctx, "problems.json", &problems); err != nil {
		return domain.Snapshot{}, err
	}
	if err := c.sleep(ctx, requestInterval); err != nil {
		return domain.Snapshot{}, err
	}
	var pairs []contestProblemJSON
	if err := c.getJSON(ctx, "contest-problem.json", &pairs); err != nil {
		return domain.Snapshot{}, err
	}
	return convertSnapshot(contests, problems, pairs), nil
}

type contestJSON struct {
	ID               string `json:"id"`
	StartEpochSecond int64  `json:"start_epoch_second"`
	DurationSecond   int64  `json:"duration_second"`
	Title            string `json:"title"`
	RateChange       string `json:"rate_change"`
}

type problemJSON struct {
	ID           string `json:"id"`
	ContestID    string `json:"contest_id"`
	ProblemIndex string `json:"problem_index"`
	Name         string `json:"name"`
}

type contestProblemJSON struct {
	ContestID    string `json:"contest_id"`
	ProblemID    string `json:"problem_id"`
	ProblemIndex string `json:"problem_index"`
}

// convertSnapshotは、形式が不正な行を取り込まずに数える。一部の不正な行で同期全体を止めない。
func convertSnapshot(contests []contestJSON, problems []problemJSON, pairs []contestProblemJSON) domain.Snapshot {
	var snapshot domain.Snapshot
	for _, raw := range contests {
		id, err := catalog.ParseContestID(raw.ID)
		if err != nil || raw.DurationSecond < 0 || raw.DurationSecond > math.MaxInt64/int64(time.Second) {
			snapshot.Skipped++
			continue
		}
		contest, err := domain.NewContest(
			id,
			raw.Title,
			time.Unix(raw.StartEpochSecond, 0),
			time.Duration(raw.DurationSecond)*time.Second,
			raw.RateChange,
		)
		if err != nil {
			snapshot.Skipped++
			continue
		}
		snapshot.Contests = append(snapshot.Contests, contest)
	}

	for _, raw := range problems {
		id, idErr := catalog.ParseProblemID(raw.ID)
		contestID, contestErr := catalog.ParseContestID(raw.ContestID)
		if idErr != nil || contestErr != nil {
			snapshot.Skipped++
			continue
		}
		snapshot.Problems = append(snapshot.Problems, catalog.NewProblem(id, contestID, raw.ProblemIndex, raw.Name))
	}

	for _, raw := range pairs {
		contestID, contestErr := catalog.ParseContestID(raw.ContestID)
		problemID, problemErr := catalog.ParseProblemID(raw.ProblemID)
		if contestErr != nil || problemErr != nil {
			snapshot.Skipped++
			continue
		}
		pair, err := domain.NewContestProblem(contestID, problemID, raw.ProblemIndex)
		if err != nil {
			snapshot.Skipped++
			continue
		}
		snapshot.ContestProblems = append(snapshot.ContestProblems, pair)
	}
	return snapshot
}

func (c *Client) getJSON(ctx context.Context, name string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.JoinPath(name).String(), nil)
	if err != nil {
		return fmt.Errorf("create %s request: %w", name, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", userAgent)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("get %s: %w", name, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("get %s: unexpected status %d", name, response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	return nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
