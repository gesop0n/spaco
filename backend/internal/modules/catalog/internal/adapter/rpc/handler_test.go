package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	catalogv1 "github.com/gesop0n/spaco/backend/generated/spaco/catalog/v1"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

type searchContestsStub func(context.Context, string) ([]domain.Contest, error)

func (f searchContestsStub) Execute(ctx context.Context, query string) ([]domain.Contest, error) {
	return f(ctx, query)
}

type listContestProblemsStub func(context.Context, string) (usecase.ContestProblems, error)

func (f listContestProblemsStub) Execute(ctx context.Context, contestID string) (usecase.ContestProblems, error) {
	return f(ctx, contestID)
}

func TestListContestProblemsReturnsProblemsWithOptionalFields(t *testing.T) {
	t.Parallel()

	contestID, _ := catalog.ParseContestID("abc042")
	shared, _ := catalog.ParseProblemID("arc058_a")
	unfilled, _ := catalog.ParseProblemID("abc042_x")
	contest, err := domain.NewContest(contestID, "AtCoder Beginner Contest 042", time.Unix(1469275200, 0), 100*time.Minute, "All")
	if err != nil {
		t.Fatalf("NewContest() error = %v", err)
	}
	handler, err := NewHandler(
		searchContestsStub(func(context.Context, string) ([]domain.Contest, error) { return nil, nil }),
		listContestProblemsStub(func(_ context.Context, gotContestID string) (usecase.ContestProblems, error) {
			if gotContestID != "abc042" {
				t.Fatalf("contestID = %q", gotContestID)
			}
			return usecase.ContestProblems{
				Contest: contest,
				Problems: []catalog.Problem{
					catalog.NewProblem(shared, contestID, "C", "Iroha's Obsession"),
					catalog.NewProblem(unfilled, contestID, "", ""),
				},
			}, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	response, err := handler.ListContestProblems(
		context.Background(),
		connect.NewRequest(&catalogv1.ListContestProblemsRequest{ContestId: "abc042"}),
	)
	if err != nil {
		t.Fatalf("ListContestProblems() error = %v", err)
	}
	problems := response.Msg.GetProblems()
	if len(problems) != 2 || problems[0].GetProblemIndex() != "C" ||
		problems[0].GetUrl() != "https://atcoder.jp/contests/abc042/tasks/arc058_a" {
		t.Fatalf("problems = %+v", problems)
	}
	if problems[1].ProblemIndex != nil || problems[1].Name != nil {
		t.Fatalf("unfilled problem must not set optional fields: %+v", problems[1])
	}
	if response.Msg.GetContest().GetStartAt().AsTime() != contest.StartAt() {
		t.Fatalf("contest = %+v", response.Msg.GetContest())
	}
}

func TestHandlerMapsErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		cause error
		want  connect.Code
	}{
		{cause: domain.ErrInvalidContestQuery, want: connect.CodeInvalidArgument},
		{cause: catalog.ErrInvalidContestID, want: connect.CodeInvalidArgument},
		{cause: usecase.ErrContestNotFound, want: connect.CodeNotFound},
		{cause: errors.New("database is down"), want: connect.CodeInternal},
	} {
		handler, err := NewHandler(
			searchContestsStub(func(context.Context, string) ([]domain.Contest, error) { return nil, test.cause }),
			listContestProblemsStub(func(context.Context, string) (usecase.ContestProblems, error) {
				return usecase.ContestProblems{}, test.cause
			}),
		)
		if err != nil {
			t.Fatalf("NewHandler() error = %v", err)
		}
		_, err = handler.SearchContests(context.Background(), connect.NewRequest(&catalogv1.SearchContestsRequest{}))
		if code := connect.CodeOf(err); code != test.want {
			t.Errorf("SearchContests(%v) code = %v, want %v", test.cause, code, test.want)
		}
		_, err = handler.ListContestProblems(context.Background(), connect.NewRequest(&catalogv1.ListContestProblemsRequest{}))
		if code := connect.CodeOf(err); code != test.want {
			t.Errorf("ListContestProblems(%v) code = %v, want %v", test.cause, code, test.want)
		}
	}
}
