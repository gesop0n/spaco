package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	reviewv1 "github.com/gesop0n/spaco/backend/generated/spaco/review/v1"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type recordReviewStub func(context.Context, identifier.UserID, usecase.RecordReviewInput) (usecase.RecordReviewResult, error)

func (f recordReviewStub) Execute(ctx context.Context, userID identifier.UserID, input usecase.RecordReviewInput) (usecase.RecordReviewResult, error) {
	return f(ctx, userID, input)
}

type getTodayReviewsStub func(context.Context, identifier.UserID) (usecase.TodayReviews, error)

func (f getTodayReviewsStub) Execute(ctx context.Context, userID identifier.UserID) (usecase.TodayReviews, error) {
	return f(ctx, userID)
}

func TestRecordReviewConvertsRequestAndResponse(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	view, log := recordedReview(t, userID)
	performedAt := time.Date(2026, 9, 11, 11, 0, 0, 0, time.UTC)
	requestID := uuid.NewString()
	handler := newTestHandler(t, recordReviewStub(func(_ context.Context, gotUserID identifier.UserID, input usecase.RecordReviewInput) (usecase.RecordReviewResult, error) {
		if gotUserID != userID || input.RequestID != requestID || input.Result != domain.ResultIndependent ||
			input.Difficulty != domain.DifficultyHard || !input.PerformedAt.Equal(performedAt) || input.Note != "苦戦した" {
			t.Fatalf("RecordReviewInput = %+v", input)
		}
		return usecase.RecordReviewResult{Item: view, Log: log}, nil
	}), nil)

	response, err := handler.RecordReview(
		authentication.WithUserID(context.Background(), userID),
		connect.NewRequest(&reviewv1.RecordReviewRequest{
			ReviewItemId: view.Item.ID().String(),
			RequestId:    requestID,
			Result:       reviewv1.ReviewResult_REVIEW_RESULT_INDEPENDENT,
			Difficulty:   reviewv1.Difficulty_DIFFICULTY_HARD,
			PerformedAt:  timestamppb.New(performedAt),
			Note:         "苦戦した",
		}),
	)
	if err != nil {
		t.Fatalf("RecordReview() error = %v", err)
	}
	item := response.Msg.GetItem()
	if item.GetProblem().GetUrl() != "https://atcoder.jp/contests/abc350/tasks/abc350_d" ||
		item.GetDueOn() != view.Item.DueOn().String() || item.GetReviewCount() != 1 ||
		item.GetRegisteredOn() != view.RegisteredOn.String() {
		t.Fatalf("item = %+v", item)
	}
	if response.Msg.GetLog().GetDifficulty() != reviewv1.Difficulty_DIFFICULTY_HARD ||
		response.Msg.GetLog().GetNextDueOn() != log.NextDueOn().String() {
		t.Fatalf("log = %+v", response.Msg.GetLog())
	}
}

func TestRecordReviewRequiresAuthenticationAndPerformedAt(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t, recordReviewStub(func(context.Context, identifier.UserID, usecase.RecordReviewInput) (usecase.RecordReviewResult, error) {
		t.Fatal("use case must not be called")
		return usecase.RecordReviewResult{}, nil
	}), nil)
	request := connect.NewRequest(&reviewv1.RecordReviewRequest{})

	if _, err := handler.RecordReview(context.Background(), request); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("unauthenticated code = %v", connect.CodeOf(err))
	}
	ctx := authentication.WithUserID(context.Background(), identifier.NewUserID())
	if _, err := handler.RecordReview(ctx, request); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing performed_at code = %v", connect.CodeOf(err))
	}
}

func TestHandlerMapsUseCaseErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		cause error
		want  connect.Code
	}{
		{cause: domain.ErrReviewItemPaused, want: connect.CodeFailedPrecondition},
		{cause: usecase.ErrReviewItemNotFound, want: connect.CodeNotFound},
		{cause: usecase.ErrProblemNotFound, want: connect.CodeNotFound},
		{cause: usecase.ErrRequestIDConflict, want: connect.CodeAlreadyExists},
		{cause: domain.ErrInvalidPerformedAt, want: connect.CodeInvalidArgument},
		{cause: catalog.ErrInvalidProblemURL, want: connect.CodeInvalidArgument},
		{cause: usecase.ErrTooManyProblems, want: connect.CodeInvalidArgument},
		{cause: usecase.ErrProblemInfoMissing, want: connect.CodeInternal},
		{cause: context.DeadlineExceeded, want: connect.CodeDeadlineExceeded},
	} {
		handler := newTestHandler(t, nil, getTodayReviewsStub(func(context.Context, identifier.UserID) (usecase.TodayReviews, error) {
			return usecase.TodayReviews{}, test.cause
		}))
		_, err := handler.GetTodayReviews(
			authentication.WithUserID(context.Background(), identifier.NewUserID()),
			connect.NewRequest(&reviewv1.GetTodayReviewsRequest{}),
		)
		if code := connect.CodeOf(err); code != test.want {
			t.Errorf("GetTodayReviews(%v) code = %v, want %v", test.cause, code, test.want)
		}
	}
}

func TestDifficultyValueRejectsUnknownEnum(t *testing.T) {
	t.Parallel()

	if _, err := domain.NewOutcome(domain.ResultIndependent, difficultyValue(reviewv1.Difficulty(99))); !errors.Is(err, domain.ErrInvalidOutcome) {
		t.Fatalf("unknown difficulty error = %v, want ErrInvalidOutcome", err)
	}
	if resultValue(reviewv1.ReviewResult(99)) != "" {
		t.Fatal("unknown result must not map to a valid result")
	}
}

func newTestHandler(t *testing.T, recordReview IRecordReviewUseCase, getTodayReviews IGetTodayReviewsUseCase) *Handler {
	t.Helper()
	if recordReview == nil {
		recordReview = recordReviewStub(func(context.Context, identifier.UserID, usecase.RecordReviewInput) (usecase.RecordReviewResult, error) {
			return usecase.RecordReviewResult{}, errors.New("unexpected call")
		})
	}
	if getTodayReviews == nil {
		getTodayReviews = getTodayReviewsStub(func(context.Context, identifier.UserID) (usecase.TodayReviews, error) {
			return usecase.TodayReviews{}, errors.New("unexpected call")
		})
	}
	handler, err := NewHandler(UseCases{
		RegisterProblems:         registerProblemsStub{},
		RegisterProblemByURL:     registerProblemByURLStub{},
		ListRegisteredProblemIDs: listRegisteredProblemIDsStub{},
		GetTodayReviews:          getTodayReviews,
		ListReviewItems:          listReviewItemsStub{},
		GetReviewItem:            changeReviewItemStub{},
		PreviewReviewSchedule:    previewReviewScheduleStub{},
		RecordReview:             recordReview,
		PauseReviewItem:          changeReviewItemStub{},
		ResumeReviewItem:         changeReviewItemStub{},
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

type registerProblemsStub struct{}

func (registerProblemsStub) Execute(context.Context, identifier.UserID, []string, string) (usecase.RegisterProblemsResult, error) {
	return usecase.RegisterProblemsResult{}, errors.New("unexpected call")
}

type registerProblemByURLStub struct{}

func (registerProblemByURLStub) Execute(context.Context, identifier.UserID, string, string) (usecase.RegisterProblemByURLResult, error) {
	return usecase.RegisterProblemByURLResult{}, errors.New("unexpected call")
}

type listRegisteredProblemIDsStub struct{}

func (listRegisteredProblemIDsStub) Execute(context.Context, identifier.UserID, []string) ([]catalog.ProblemID, error) {
	return nil, errors.New("unexpected call")
}

type listReviewItemsStub struct{}

func (listReviewItemsStub) Execute(context.Context, identifier.UserID) (usecase.ReviewItemList, error) {
	return usecase.ReviewItemList{}, errors.New("unexpected call")
}

type changeReviewItemStub struct{}

func (changeReviewItemStub) Execute(context.Context, identifier.UserID, string) (usecase.ReviewItemView, error) {
	return usecase.ReviewItemView{}, errors.New("unexpected call")
}

type previewReviewScheduleStub struct{}

func (previewReviewScheduleStub) Execute(context.Context, identifier.UserID, string, time.Time) ([]domain.SchedulePreview, error) {
	return nil, errors.New("unexpected call")
}

func recordedReview(t *testing.T, userID identifier.UserID) (usecase.ReviewItemView, domain.ReviewLog) {
	t.Helper()
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	problemID, _ := catalog.ParseProblemID("abc350_d")
	contestID, _ := catalog.ParseContestID("abc350")
	registeredAt := time.Date(2026, 9, 10, 11, 0, 0, 0, time.UTC)
	item, err := domain.RegisterReviewItem(domain.NewReviewItemID(), userID, problemID, "", registeredAt, tokyo)
	if err != nil {
		t.Fatalf("RegisterReviewItem() error = %v", err)
	}
	requestID, _ := domain.ParseRequestID(uuid.NewString())
	outcome, _ := domain.NewOutcome(domain.ResultIndependent, domain.DifficultyHard)
	performedAt := time.Date(2026, 9, 11, 11, 0, 0, 0, time.UTC)
	recorded, log, err := item.Record(domain.NewReviewLogID(), domain.ReviewInput{
		RequestID:   requestID,
		Outcome:     outcome,
		PerformedAt: performedAt,
	}, performedAt, tokyo, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	return usecase.ReviewItemView{
		Item:         recorded,
		Problem:      catalog.NewProblem(problemID, contestID, "D", "New Friends"),
		RegisteredOn: scheduler.DayOf(registeredAt, tokyo),
		LastLog:      &log,
	}, log
}
