// Package rpcは、catalogのユースケースをConnectRPCへ接続するadapterを提供する。
package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalogv1 "github.com/gesop0n/spaco/backend/generated/spaco/catalog/v1"
	"github.com/gesop0n/spaco/backend/generated/spaco/catalog/v1/catalogv1connect"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/usecase"
)

// ISearchContestsUseCaseは、RPC adapterが必要とするコンテスト検索を表すinterfaceである。
type ISearchContestsUseCase interface {
	Execute(context.Context, string) ([]domain.Contest, error)
}

// IListContestProblemsUseCaseは、RPC adapterが必要とする問題一覧の取得を表すinterfaceである。
type IListContestProblemsUseCase interface {
	Execute(context.Context, string) (usecase.ContestProblems, error)
}

type Handler struct {
	searchContests      ISearchContestsUseCase
	listContestProblems IListContestProblemsUseCase
}

var _ catalogv1connect.CatalogServiceHandler = (*Handler)(nil)

func NewHandler(
	searchContests ISearchContestsUseCase,
	listContestProblems IListContestProblemsUseCase,
) (*Handler, error) {
	if searchContests == nil {
		return nil, errors.New("create catalog handler: search contests use case is required")
	}
	if listContestProblems == nil {
		return nil, errors.New("create catalog handler: list contest problems use case is required")
	}
	return &Handler{searchContests: searchContests, listContestProblems: listContestProblems}, nil
}

func (h *Handler) SearchContests(
	ctx context.Context,
	request *connect.Request[catalogv1.SearchContestsRequest],
) (*connect.Response[catalogv1.SearchContestsResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request is required"))
	}
	contests, err := h.searchContests.Execute(ctx, request.Msg.GetQuery())
	if err != nil {
		return nil, connectError(err)
	}
	messages := make([]*catalogv1.Contest, 0, len(contests))
	for _, contest := range contests {
		messages = append(messages, contestMessage(contest))
	}
	return connect.NewResponse(&catalogv1.SearchContestsResponse{Contests: messages}), nil
}

func (h *Handler) ListContestProblems(
	ctx context.Context,
	request *connect.Request[catalogv1.ListContestProblemsRequest],
) (*connect.Response[catalogv1.ListContestProblemsResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request is required"))
	}
	result, err := h.listContestProblems.Execute(ctx, request.Msg.GetContestId())
	if err != nil {
		return nil, connectError(err)
	}
	problems := make([]*catalogv1.Problem, 0, len(result.Problems))
	for _, problem := range result.Problems {
		problems = append(problems, problemMessage(problem))
	}
	return connect.NewResponse(&catalogv1.ListContestProblemsResponse{
		Contest:  contestMessage(result.Contest),
		Problems: problems,
	}), nil
}

func contestMessage(contest domain.Contest) *catalogv1.Contest {
	return &catalogv1.Contest{
		Id:      contest.ID().String(),
		Title:   contest.Title(),
		StartAt: timestamppb.New(contest.StartAt()),
	}
}

func problemMessage(problem catalog.Problem) *catalogv1.Problem {
	message := &catalogv1.Problem{
		Id:        problem.ID().String(),
		ContestId: problem.ContestID().String(),
		Url:       problem.URL(),
	}
	if index, ok := problem.Index(); ok {
		message.ProblemIndex = &index
	}
	if name, ok := problem.Name(); ok {
		message.Name = &name
	}
	return message
}

func connectError(err error) *connect.Error {
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, context.Canceled)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, context.DeadlineExceeded)
	case errors.Is(err, domain.ErrInvalidContestQuery), errors.Is(err, catalog.ErrInvalidContestID):
		return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid catalog request"))
	case errors.Is(err, usecase.ErrContestNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("contest not found"))
	default:
		// DBや内部構造の詳細はclientへ公開しない。
		return connect.NewError(connect.CodeInternal, errors.New("catalog operation failed"))
	}
}
