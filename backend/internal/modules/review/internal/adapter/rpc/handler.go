// Package rpcは、reviewのユースケースをConnectRPCへ接続するadapterを提供する。
package rpc

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	reviewv1 "github.com/gesop0n/spaco/backend/generated/spaco/review/v1"
	"github.com/gesop0n/spaco/backend/generated/spaco/review/v1/reviewv1connect"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type IRegisterProblemsUseCase interface {
	Execute(context.Context, identifier.UserID, []string, string) (usecase.RegisterProblemsResult, error)
}

type IRegisterProblemByURLUseCase interface {
	Execute(context.Context, identifier.UserID, string, string) (usecase.RegisterProblemByURLResult, error)
}

type IListRegisteredProblemIDsUseCase interface {
	Execute(context.Context, identifier.UserID, []string) ([]catalog.ProblemID, error)
}

type IGetTodayReviewsUseCase interface {
	Execute(context.Context, identifier.UserID) (usecase.TodayReviews, error)
}

type IListReviewItemsUseCase interface {
	Execute(context.Context, identifier.UserID) (usecase.ReviewItemList, error)
}

type IGetReviewItemUseCase interface {
	Execute(context.Context, identifier.UserID, string) (usecase.ReviewItemView, error)
}

type IPreviewReviewScheduleUseCase interface {
	Execute(context.Context, identifier.UserID, string, time.Time) ([]domain.SchedulePreview, error)
}

type IRecordReviewUseCase interface {
	Execute(context.Context, identifier.UserID, usecase.RecordReviewInput) (usecase.RecordReviewResult, error)
}

type IChangeReviewItemPauseUseCase interface {
	Execute(context.Context, identifier.UserID, string) (usecase.ReviewItemView, error)
}

// UseCasesは、Handlerが呼び出すユースケースである。
type UseCases struct {
	RegisterProblems         IRegisterProblemsUseCase
	RegisterProblemByURL     IRegisterProblemByURLUseCase
	ListRegisteredProblemIDs IListRegisteredProblemIDsUseCase
	GetTodayReviews          IGetTodayReviewsUseCase
	ListReviewItems          IListReviewItemsUseCase
	GetReviewItem            IGetReviewItemUseCase
	PreviewReviewSchedule    IPreviewReviewScheduleUseCase
	RecordReview             IRecordReviewUseCase
	PauseReviewItem          IChangeReviewItemPauseUseCase
	ResumeReviewItem         IChangeReviewItemPauseUseCase
}

type Handler struct {
	useCases UseCases
}

var _ reviewv1connect.ReviewServiceHandler = (*Handler)(nil)

func NewHandler(useCases UseCases) (*Handler, error) {
	if useCases.RegisterProblems == nil || useCases.RegisterProblemByURL == nil ||
		useCases.ListRegisteredProblemIDs == nil || useCases.GetTodayReviews == nil ||
		useCases.ListReviewItems == nil || useCases.GetReviewItem == nil ||
		useCases.PreviewReviewSchedule == nil || useCases.RecordReview == nil ||
		useCases.PauseReviewItem == nil || useCases.ResumeReviewItem == nil {
		return nil, errors.New("create review handler: all use cases are required")
	}
	return &Handler{useCases: useCases}, nil
}

func (h *Handler) RegisterProblems(
	ctx context.Context,
	request *connect.Request[reviewv1.RegisterProblemsRequest],
) (*connect.Response[reviewv1.RegisterProblemsResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	result, err := h.useCases.RegisterProblems.Execute(ctx, userID, message.GetProblemIds(), message.GetNote())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.RegisterProblemsResponse{
		RegisteredItems:        reviewItemMessages(result.Registered),
		AlreadyRegisteredItems: reviewItemMessages(result.AlreadyRegistered),
	}), nil
}

func (h *Handler) RegisterProblemByUrl(
	ctx context.Context,
	request *connect.Request[reviewv1.RegisterProblemByUrlRequest],
) (*connect.Response[reviewv1.RegisterProblemByUrlResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	result, err := h.useCases.RegisterProblemByURL.Execute(ctx, userID, message.GetUrl(), message.GetNote())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.RegisterProblemByUrlResponse{
		Item:              reviewItemMessage(result.Item),
		AlreadyRegistered: result.AlreadyRegistered,
	}), nil
}

func (h *Handler) ListRegisteredProblemIds(
	ctx context.Context,
	request *connect.Request[reviewv1.ListRegisteredProblemIdsRequest],
) (*connect.Response[reviewv1.ListRegisteredProblemIdsResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	problemIDs, err := h.useCases.ListRegisteredProblemIDs.Execute(ctx, userID, message.GetProblemIds())
	if err != nil {
		return nil, connectError(err)
	}
	values := make([]string, len(problemIDs))
	for index, problemID := range problemIDs {
		values[index] = problemID.String()
	}
	return connect.NewResponse(&reviewv1.ListRegisteredProblemIdsResponse{ProblemIds: values}), nil
}

func (h *Handler) GetTodayReviews(
	ctx context.Context,
	request *connect.Request[reviewv1.GetTodayReviewsRequest],
) (*connect.Response[reviewv1.GetTodayReviewsResponse], error) {
	userID, _, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	result, err := h.useCases.GetTodayReviews.Execute(ctx, userID)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.GetTodayReviewsResponse{
		Today:               result.Today.String(),
		TimeZone:            result.TimeZone,
		DueItems:            reviewItemMessages(result.DueItems),
		CompletedTodayCount: int32(result.CompletedCount),
		UpcomingItems:       reviewItemMessages(result.UpcomingItems),
	}), nil
}

func (h *Handler) ListReviewItems(
	ctx context.Context,
	request *connect.Request[reviewv1.ListReviewItemsRequest],
) (*connect.Response[reviewv1.ListReviewItemsResponse], error) {
	userID, _, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	result, err := h.useCases.ListReviewItems.Execute(ctx, userID)
	if err != nil {
		return nil, connectError(err)
	}
	logs := make([]*reviewv1.ReviewLog, 0, len(result.Logs))
	for _, log := range result.Logs {
		logs = append(logs, reviewLogMessage(log))
	}
	return connect.NewResponse(&reviewv1.ListReviewItemsResponse{
		Today:    result.Today.String(),
		TimeZone: result.TimeZone,
		Items:    reviewItemMessages(result.Items),
		Logs:     logs,
	}), nil
}

func (h *Handler) GetReviewItem(
	ctx context.Context,
	request *connect.Request[reviewv1.GetReviewItemRequest],
) (*connect.Response[reviewv1.GetReviewItemResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	view, err := h.useCases.GetReviewItem.Execute(ctx, userID, message.GetProblemId())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.GetReviewItemResponse{Item: reviewItemMessage(view)}), nil
}

func (h *Handler) PreviewReviewSchedule(
	ctx context.Context,
	request *connect.Request[reviewv1.PreviewReviewScheduleRequest],
) (*connect.Response[reviewv1.PreviewReviewScheduleResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	performedAt, err := timestampValue(message.GetPerformedAt())
	if err != nil {
		return nil, err
	}
	previews, err := h.useCases.PreviewReviewSchedule.Execute(ctx, userID, message.GetReviewItemId(), performedAt)
	if err != nil {
		return nil, connectError(err)
	}
	messages := make([]*reviewv1.SchedulePreview, 0, len(previews))
	for _, preview := range previews {
		messages = append(messages, &reviewv1.SchedulePreview{
			Result:             resultMessage(preview.Outcome.Result()),
			Difficulty:         difficultyMessage(preview.Outcome),
			NextDueOn:          preview.NextDueOn.String(),
			DaysAfterPerformed: int32(preview.DaysAfterPerformed),
		})
	}
	return connect.NewResponse(&reviewv1.PreviewReviewScheduleResponse{Previews: messages}), nil
}

func (h *Handler) RecordReview(
	ctx context.Context,
	request *connect.Request[reviewv1.RecordReviewRequest],
) (*connect.Response[reviewv1.RecordReviewResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	performedAt, err := timestampValue(message.GetPerformedAt())
	if err != nil {
		return nil, err
	}
	result, err := h.useCases.RecordReview.Execute(ctx, userID, usecase.RecordReviewInput{
		ReviewItemID: message.GetReviewItemId(),
		RequestID:    message.GetRequestId(),
		Result:       resultValue(message.GetResult()),
		Difficulty:   difficultyValue(message.GetDifficulty()),
		PerformedAt:  performedAt,
		Note:         message.GetNote(),
	})
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.RecordReviewResponse{
		Item: reviewItemMessage(result.Item),
		Log:  reviewLogMessage(result.Log),
	}), nil
}

func (h *Handler) PauseReviewItem(
	ctx context.Context,
	request *connect.Request[reviewv1.PauseReviewItemRequest],
) (*connect.Response[reviewv1.PauseReviewItemResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	view, err := h.useCases.PauseReviewItem.Execute(ctx, userID, message.GetReviewItemId())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.PauseReviewItemResponse{Item: reviewItemMessage(view)}), nil
}

func (h *Handler) ResumeReviewItem(
	ctx context.Context,
	request *connect.Request[reviewv1.ResumeReviewItemRequest],
) (*connect.Response[reviewv1.ResumeReviewItemResponse], error) {
	userID, message, err := authenticatedRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	view, err := h.useCases.ResumeReviewItem.Execute(ctx, userID, message.GetReviewItemId())
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&reviewv1.ResumeReviewItemResponse{Item: reviewItemMessage(view)}), nil
}

func authenticatedRequest[T any](
	ctx context.Context,
	request *connect.Request[T],
) (identifier.UserID, *T, error) {
	userID, ok := authentication.UserIDFromContext(ctx)
	if !ok {
		return identifier.UserID{}, nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required"),
		)
	}
	if request == nil || request.Msg == nil {
		return identifier.UserID{}, nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request is required"))
	}
	return userID, request.Msg, nil
}

func timestampValue(timestamp *timestamppb.Timestamp) (time.Time, error) {
	if timestamp == nil {
		return time.Time{}, connect.NewError(connect.CodeInvalidArgument, errors.New("performed_at is required"))
	}
	if err := timestamp.CheckValid(); err != nil {
		return time.Time{}, connect.NewError(connect.CodeInvalidArgument, errors.New("performed_at is invalid"))
	}
	return timestamp.AsTime(), nil
}

func resultValue(result reviewv1.ReviewResult) domain.Result {
	switch result {
	case reviewv1.ReviewResult_REVIEW_RESULT_INDEPENDENT:
		return domain.ResultIndependent
	case reviewv1.ReviewResult_REVIEW_RESULT_ASSISTED:
		return domain.ResultAssisted
	case reviewv1.ReviewResult_REVIEW_RESULT_UNSOLVED:
		return domain.ResultUnsolved
	default:
		// 未指定や未知の値は、ユースケースで入力エラーにする。
		return ""
	}
}

func resultMessage(result domain.Result) reviewv1.ReviewResult {
	switch result {
	case domain.ResultIndependent:
		return reviewv1.ReviewResult_REVIEW_RESULT_INDEPENDENT
	case domain.ResultAssisted:
		return reviewv1.ReviewResult_REVIEW_RESULT_ASSISTED
	case domain.ResultUnsolved:
		return reviewv1.ReviewResult_REVIEW_RESULT_UNSOLVED
	default:
		return reviewv1.ReviewResult_REVIEW_RESULT_UNSPECIFIED
	}
}

func difficultyValue(difficulty reviewv1.Difficulty) domain.Difficulty {
	switch difficulty {
	case reviewv1.Difficulty_DIFFICULTY_HARD:
		return domain.DifficultyHard
	case reviewv1.Difficulty_DIFFICULTY_GOOD:
		return domain.DifficultyGood
	case reviewv1.Difficulty_DIFFICULTY_EASY:
		return domain.DifficultyEasy
	case reviewv1.Difficulty_DIFFICULTY_UNSPECIFIED:
		return ""
	default:
		// 未知の値を未指定として扱わず、ユースケースで入力エラーにする。
		return domain.Difficulty(difficulty.String())
	}
}

func difficultyMessage(outcome domain.Outcome) reviewv1.Difficulty {
	difficulty, ok := outcome.Difficulty()
	if !ok {
		return reviewv1.Difficulty_DIFFICULTY_UNSPECIFIED
	}
	switch difficulty {
	case domain.DifficultyHard:
		return reviewv1.Difficulty_DIFFICULTY_HARD
	case domain.DifficultyEasy:
		return reviewv1.Difficulty_DIFFICULTY_EASY
	default:
		return reviewv1.Difficulty_DIFFICULTY_GOOD
	}
}

func reviewItemMessages(views []usecase.ReviewItemView) []*reviewv1.ReviewItem {
	messages := make([]*reviewv1.ReviewItem, 0, len(views))
	for _, view := range views {
		messages = append(messages, reviewItemMessage(view))
	}
	return messages
}

func reviewItemMessage(view usecase.ReviewItemView) *reviewv1.ReviewItem {
	message := &reviewv1.ReviewItem{
		Id:               view.Item.ID().String(),
		Problem:          problemMessage(view.Problem),
		RegistrationNote: view.Item.RegistrationNote(),
		RegisteredAt:     timestamppb.New(view.Item.RegisteredAt()),
		RegisteredOn:     view.RegisteredOn.String(),
		DueOn:            view.Item.DueOn().String(),
		Paused:           view.Item.Paused(),
		ReviewCount:      int32(view.Item.Reps()),
	}
	if view.LastLog != nil {
		message.LastLog = reviewLogMessage(*view.LastLog)
	}
	return message
}

func reviewLogMessage(log domain.ReviewLog) *reviewv1.ReviewLog {
	return &reviewv1.ReviewLog{
		Id:           log.ID().String(),
		ReviewItemId: log.ReviewItemID().String(),
		Result:       resultMessage(log.Outcome().Result()),
		Difficulty:   difficultyMessage(log.Outcome()),
		PerformedAt:  timestamppb.New(log.PerformedAt()),
		Note:         log.Note(),
		NextDueOn:    log.NextDueOn().String(),
	}
}

func problemMessage(problem catalog.Problem) *reviewv1.Problem {
	message := &reviewv1.Problem{
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

var invalidArgumentErrors = []error{
	domain.ErrInvalidID,
	domain.ErrInvalidNote,
	domain.ErrInvalidOutcome,
	domain.ErrInvalidPerformedAt,
	catalog.ErrInvalidProblemID,
	catalog.ErrInvalidProblemURL,
	usecase.ErrNoProblems,
	usecase.ErrTooManyProblems,
}

func connectError(err error) *connect.Error {
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, context.Canceled)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, context.DeadlineExceeded)
	case errors.Is(err, domain.ErrReviewItemPaused):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("review item is paused"))
	case errors.Is(err, usecase.ErrReviewItemNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("review item not found"))
	case errors.Is(err, usecase.ErrProblemNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("problem not found"))
	case errors.Is(err, usecase.ErrRequestIDConflict):
		return connect.NewError(connect.CodeAlreadyExists, errors.New("request id is already used"))
	}
	for _, target := range invalidArgumentErrors {
		if errors.Is(err, target) {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid review request"))
		}
	}
	// DBや内部構造の詳細はclientへ公開しない。
	return connect.NewError(connect.CodeInternal, errors.New("review operation failed"))
}
