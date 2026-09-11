package atcoderproblems

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSnapshotReadsResourcesWithIntervals(t *testing.T) {
	t.Parallel()

	resources := map[string]string{
		"/resources/contests.json": `[
			{"id":"APG4b","start_epoch_second":0,"duration_second":3153600000,"title":"C++入門 AtCoder Programming Guide for beginners (APG4b)","rate_change":"-"},
			{"id":"bad id","start_epoch_second":0,"duration_second":1,"title":"invalid","rate_change":"-"}
		]`,
		"/resources/problems.json": `[
			{"id":"APG4b_a","contest_id":"APG4b","problem_index":"A","name":"1.00.はじめに","title":"A. 1.00.はじめに"}
		]`,
		"/resources/contest-problem.json": `[
			{"contest_id":"APG4b","problem_id":"APG4b_a","problem_index":"A"},
			{"contest_id":"APG4b","problem_id":"APG4b_b","problem_index":""}
		]`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("User-Agent") != userAgent {
			t.Errorf("User-Agent = %q", request.Header.Get("User-Agent"))
		}
		body, ok := resources[request.URL.Path]
		if !ok {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(body))
	}))
	defer server.Close()

	client, err := NewClient(server.Client(), server.URL+"/resources/")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	var waits []time.Duration
	client.sleep = func(_ context.Context, duration time.Duration) error {
		waits = append(waits, duration)
		return nil
	}

	snapshot, err := client.FetchSnapshot(context.Background())
	if err != nil {
		t.Fatalf("FetchSnapshot() error = %v", err)
	}
	if len(waits) != 2 || waits[0] < time.Second || waits[1] < time.Second {
		t.Fatalf("waits = %v, want two waits of at least 1s", waits)
	}
	if len(snapshot.Contests) != 1 || len(snapshot.Problems) != 1 || len(snapshot.ContestProblems) != 1 || snapshot.Skipped != 2 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	// 常設コンテストの約100年という期間も、オーバーフローせずに読み取る。
	if got := snapshot.Contests[0].Duration(); got != 3153600000*time.Second {
		t.Fatalf("Duration() = %v", got)
	}
	if name, _ := snapshot.Problems[0].Name(); name != "1.00.はじめに" {
		t.Fatalf("problem name = %q", name)
	}
}

func TestFetchSnapshotReturnsErrorForUnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewClient(server.Client(), server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.FetchSnapshot(context.Background()); err == nil {
		t.Fatal("FetchSnapshot() error = nil, want status error")
	}
}
