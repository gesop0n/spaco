package usecase

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// 日本時間では2026-09-12 00:30、ロサンゼルスでは2026-09-11 08:30。
var fixedNow = time.Date(2026, 9, 11, 15, 30, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

type repositoryStub struct {
	createReviewItems               func(context.Context, []domain.ReviewItem) ([]domain.ReviewItem, error)
	findReviewItemsByProblemIDs     func(context.Context, identifier.UserID, []catalog.ProblemID) ([]domain.ReviewItem, error)
	findLatestReviewLogs            func(context.Context, identifier.UserID, []domain.ReviewItemID) ([]domain.ReviewLog, error)
	listDueReviewItems              func(context.Context, identifier.UserID, scheduler.Day) ([]domain.ReviewItem, error)
	listUpcomingReviewItems         func(context.Context, identifier.UserID, scheduler.Day, int) ([]domain.ReviewItem, error)
	countReviewLogsPerformedBetween func(context.Context, identifier.UserID, time.Time, time.Time) (int, error)
	findReviewItemByID              func(context.Context, identifier.UserID, domain.ReviewItemID) (domain.ReviewItem, error)
	recordReview                    func(context.Context, identifier.UserID, domain.ReviewItemID, domain.RequestID, RecordReviewFunc) (RecordedReview, error)
	updateReviewItemPausedAt        func(context.Context, domain.ReviewItem) (domain.ReviewItem, error)
}

func (s repositoryStub) CreateReviewItems(ctx context.Context, items []domain.ReviewItem) ([]domain.ReviewItem, error) {
	return s.createReviewItems(ctx, items)
}

func (s repositoryStub) FindReviewItemsByProblemIDs(ctx context.Context, userID identifier.UserID, ids []catalog.ProblemID) ([]domain.ReviewItem, error) {
	return s.findReviewItemsByProblemIDs(ctx, userID, ids)
}

func (s repositoryStub) FindLatestReviewLogs(ctx context.Context, userID identifier.UserID, ids []domain.ReviewItemID) ([]domain.ReviewLog, error) {
	if s.findLatestReviewLogs == nil {
		return nil, nil
	}
	return s.findLatestReviewLogs(ctx, userID, ids)
}

func (s repositoryStub) ListDueReviewItems(ctx context.Context, userID identifier.UserID, today scheduler.Day) ([]domain.ReviewItem, error) {
	return s.listDueReviewItems(ctx, userID, today)
}

func (s repositoryStub) ListUpcomingReviewItems(ctx context.Context, userID identifier.UserID, today scheduler.Day, limit int) ([]domain.ReviewItem, error) {
	return s.listUpcomingReviewItems(ctx, userID, today, limit)
}

func (s repositoryStub) CountReviewLogsPerformedBetween(ctx context.Context, userID identifier.UserID, from, until time.Time) (int, error) {
	return s.countReviewLogsPerformedBetween(ctx, userID, from, until)
}

func (s repositoryStub) FindReviewItemByID(ctx context.Context, userID identifier.UserID, id domain.ReviewItemID) (domain.ReviewItem, error) {
	return s.findReviewItemByID(ctx, userID, id)
}

func (s repositoryStub) RecordReview(ctx context.Context, userID identifier.UserID, id domain.ReviewItemID, requestID domain.RequestID, record RecordReviewFunc) (RecordedReview, error) {
	return s.recordReview(ctx, userID, id, requestID, record)
}

func (s repositoryStub) UpdateReviewItemPausedAt(ctx context.Context, item domain.ReviewItem) (domain.ReviewItem, error) {
	return s.updateReviewItemPausedAt(ctx, item)
}

type catalogStub struct {
	problems      []catalog.Problem
	ensureCalled  *bool
	ensureProblem func(catalog.ProblemRef) catalog.Problem
}

func (s catalogStub) FindProblems(_ context.Context, ids []catalog.ProblemID) ([]catalog.Problem, error) {
	var found []catalog.Problem
	for _, problem := range s.problems {
		if slices.Contains(ids, problem.ID()) {
			found = append(found, problem)
		}
	}
	return found, nil
}

func (s catalogStub) EnsureProblem(_ context.Context, ref catalog.ProblemRef) (catalog.Problem, error) {
	if s.ensureCalled != nil {
		*s.ensureCalled = true
	}
	return s.ensureProblem(ref), nil
}

type timeZoneStub string

func (s timeZoneStub) TimeZone(context.Context, identifier.UserID) (string, error) {
	return string(s), nil
}

func TestRegisterProblemsSplitsRegisteredAndExistingItems(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	problemA, problemB := problem(t, "abc350_a"), problem(t, "abc350_b")
	existing := registerItem(t, userID, problemA.ID(), fixedNow.Add(-72*time.Hour))
	var createdItems []domain.ReviewItem
	registerProblems, err := NewRegisterProblems(repositoryStub{
		createReviewItems: func(_ context.Context, items []domain.ReviewItem) ([]domain.ReviewItem, error) {
			createdItems = items
			// abc350_aは登録済みなので、abc350_bだけを保存した結果を返す。
			return []domain.ReviewItem{items[0]}, nil
		},
		findReviewItemsByProblemIDs: func(_ context.Context, _ identifier.UserID, ids []catalog.ProblemID) ([]domain.ReviewItem, error) {
			if len(ids) != 1 || ids[0] != problemA.ID() {
				t.Fatalf("existing problem ids = %v", ids)
			}
			return []domain.ReviewItem{existing}, nil
		},
	}, catalogStub{problems: []catalog.Problem{problemA, problemB}}, timeZoneStub("Asia/Tokyo"), fixedClock)
	if err != nil {
		t.Fatalf("NewRegisterProblems() error = %v", err)
	}

	result, err := registerProblems.Execute(context.Background(), userID, []string{"abc350_b", " abc350_a ", "abc350_b"}, "共通メモ")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(createdItems) != 2 {
		t.Fatalf("created items = %d, want duplicated ids removed", len(createdItems))
	}
	if len(result.Registered) != 1 || result.Registered[0].Problem.ID() != problemB.ID() {
		t.Fatalf("Registered = %+v", result.Registered)
	}
	registered := result.Registered[0]
	// 日本時間で9月12日に登録したので、最初の復習は9月13日になる。
	if registered.RegisteredOn.String() != "2026-09-12" || registered.Item.DueOn().String() != "2026-09-13" ||
		registered.Item.RegistrationNote() != "共通メモ" {
		t.Fatalf("registered view = %+v", registered)
	}
	if len(result.AlreadyRegistered) != 1 || result.AlreadyRegistered[0].Item.ID() != existing.ID() {
		t.Fatalf("AlreadyRegistered = %+v", result.AlreadyRegistered)
	}
}

func TestRegisterProblemsValidatesBeforeSaving(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	repository := repositoryStub{
		createReviewItems: func(context.Context, []domain.ReviewItem) ([]domain.ReviewItem, error) {
			t.Fatal("invalid registration must not be saved")
			return nil, nil
		},
	}
	registerProblems, err := NewRegisterProblems(
		repository,
		catalogStub{problems: []catalog.Problem{problem(t, "abc350_a")}},
		timeZoneStub("Asia/Tokyo"),
		fixedClock,
	)
	if err != nil {
		t.Fatalf("NewRegisterProblems() error = %v", err)
	}

	tooMany := make([]string, 101)
	for index := range tooMany {
		tooMany[index] = "abc350_a"
	}
	for _, test := range []struct {
		name  string
		ids   []string
		note  string
		wants error
	}{
		{name: "empty", ids: nil, wants: ErrNoProblems},
		{name: "too many", ids: tooMany, wants: ErrTooManyProblems},
		{name: "invalid id", ids: []string{"abc 350"}, wants: catalog.ErrInvalidProblemID},
		{name: "unknown problem", ids: []string{"abc350_z"}, wants: ErrProblemNotFound},
		{name: "long note", ids: []string{"abc350_a"}, note: strings.Repeat("a", 501), wants: domain.ErrInvalidNote},
	} {
		if _, err := registerProblems.Execute(context.Background(), userID, test.ids, test.note); !errors.Is(err, test.wants) {
			t.Errorf("%s: error = %v, want %v", test.name, err, test.wants)
		}
	}
}

func TestRegisterProblemByURLReturnsExistingItemWithoutChangingIt(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	stub := problem(t, "abc350_d")
	existing := registerItem(t, userID, stub.ID(), fixedNow.Add(-72*time.Hour))
	registerProblemByURL, err := NewRegisterProblemByURL(repositoryStub{
		createReviewItems: func(context.Context, []domain.ReviewItem) ([]domain.ReviewItem, error) { return nil, nil },
		findReviewItemsByProblemIDs: func(context.Context, identifier.UserID, []catalog.ProblemID) ([]domain.ReviewItem, error) {
			return []domain.ReviewItem{existing}, nil
		},
	}, catalogStub{ensureProblem: func(catalog.ProblemRef) catalog.Problem { return stub }}, timeZoneStub("Asia/Tokyo"), fixedClock)
	if err != nil {
		t.Fatalf("NewRegisterProblemByURL() error = %v", err)
	}

	result, err := registerProblemByURL.Execute(context.Background(), userID, "https://atcoder.jp/contests/abc350/tasks/abc350_d", "")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.AlreadyRegistered || result.Item.Item.ID() != existing.ID() {
		t.Fatalf("Execute() = %+v", result)
	}
}

func TestRegisterProblemByURLRejectsInvalidURLBeforeCatalog(t *testing.T) {
	t.Parallel()

	ensureCalled := false
	registerProblemByURL, err := NewRegisterProblemByURL(
		repositoryStub{},
		catalogStub{ensureCalled: &ensureCalled},
		timeZoneStub("Asia/Tokyo"),
		fixedClock,
	)
	if err != nil {
		t.Fatalf("NewRegisterProblemByURL() error = %v", err)
	}
	for _, test := range []struct {
		url  string
		note string
		want error
	}{
		{url: "https://example.com/contests/abc350/tasks/abc350_d", want: catalog.ErrInvalidProblemURL},
		{url: "https://atcoder.jp/contests/abc350/tasks/abc350_d", note: strings.Repeat("a", 501), want: domain.ErrInvalidNote},
	} {
		if _, err := registerProblemByURL.Execute(context.Background(), identifier.NewUserID(), test.url, test.note); !errors.Is(err, test.want) {
			t.Errorf("Execute(%q) error = %v, want %v", test.url, err, test.want)
		}
	}
	if ensureCalled {
		t.Fatal("catalog must not be changed for invalid input")
	}
}

func TestGetTodayReviewsUsesUserTimeZone(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		timeZone  string
		wantToday string
		wantFrom  time.Time
	}{
		{timeZone: "Asia/Tokyo", wantToday: "2026-09-12", wantFrom: time.Date(2026, 9, 11, 15, 0, 0, 0, time.UTC)},
		{timeZone: "America/Los_Angeles", wantToday: "2026-09-11", wantFrom: time.Date(2026, 9, 11, 7, 0, 0, 0, time.UTC)},
	} {
		t.Run(test.timeZone, func(t *testing.T) {
			t.Parallel()

			userID := identifier.NewUserID()
			dueProblem := problem(t, "abc350_a")
			due := registerItem(t, userID, dueProblem.ID(), fixedNow.Add(-72*time.Hour))
			getTodayReviews, err := NewGetTodayReviews(repositoryStub{
				listDueReviewItems: func(_ context.Context, _ identifier.UserID, today scheduler.Day) ([]domain.ReviewItem, error) {
					if today.String() != test.wantToday {
						t.Fatalf("today = %s, want %s", today, test.wantToday)
					}
					return []domain.ReviewItem{due}, nil
				},
				listUpcomingReviewItems: func(_ context.Context, _ identifier.UserID, _ scheduler.Day, limit int) ([]domain.ReviewItem, error) {
					if limit != 3 {
						t.Fatalf("limit = %d, want 3", limit)
					}
					return nil, nil
				},
				countReviewLogsPerformedBetween: func(_ context.Context, _ identifier.UserID, from, until time.Time) (int, error) {
					if !from.Equal(test.wantFrom) || !until.Equal(test.wantFrom.Add(24*time.Hour)) {
						t.Fatalf("range = [%v, %v), want local midnight", from, until)
					}
					return 2, nil
				},
			}, catalogStub{problems: []catalog.Problem{dueProblem}}, timeZoneStub(test.timeZone), fixedClock)
			if err != nil {
				t.Fatalf("NewGetTodayReviews() error = %v", err)
			}

			result, err := getTodayReviews.Execute(context.Background(), userID)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Today.String() != test.wantToday || result.TimeZone != test.timeZone ||
				len(result.DueItems) != 1 || result.CompletedCount != 2 || len(result.UpcomingItems) != 0 {
				t.Fatalf("Execute() = %+v", result)
			}
		})
	}
}

func TestRecordReviewRecordsInsideRepositoryCallback(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	recordedProblem := problem(t, "abc350_a")
	item := registerItem(t, userID, recordedProblem.ID(), fixedNow.Add(-72*time.Hour))
	requestID := uuid.NewString()
	var savedLog domain.ReviewLog
	recordReview, err := NewRecordReview(repositoryStub{
		recordReview: func(_ context.Context, _ identifier.UserID, id domain.ReviewItemID, gotRequestID domain.RequestID, record RecordReviewFunc) (RecordedReview, error) {
			if id != item.ID() || gotRequestID.String() != requestID {
				t.Fatalf("RecordReview(%s, %s)", id, gotRequestID)
			}
			next, log, err := record(item)
			if err != nil {
				return RecordedReview{}, err
			}
			savedLog = log
			return RecordedReview{Item: next, Log: log}, nil
		},
		findLatestReviewLogs: func(context.Context, identifier.UserID, []domain.ReviewItemID) ([]domain.ReviewLog, error) {
			return []domain.ReviewLog{savedLog}, nil
		},
	}, catalogStub{problems: []catalog.Problem{recordedProblem}}, timeZoneStub("Asia/Tokyo"), fixedClock, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("NewRecordReview() error = %v", err)
	}

	result, err := recordReview.Execute(context.Background(), userID, RecordReviewInput{
		ReviewItemID: item.ID().String(),
		RequestID:    requestID,
		Result:       domain.ResultIndependent,
		Difficulty:   domain.DifficultyEasy,
		PerformedAt:  fixedNow.Add(-time.Hour),
		Note:         "余裕を持って解けた",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Log.Outcome().Rating() != scheduler.Easy || result.Log.Note() != "余裕を持って解けた" {
		t.Fatalf("Log = %+v", result.Log)
	}
	if result.Item.Item.Reps() != 1 || result.Item.LastLog == nil || result.Item.LastLog.ID() != result.Log.ID() {
		t.Fatalf("Item = %+v", result.Item)
	}
}

func TestRecordReviewValidatesInputBeforeRepository(t *testing.T) {
	t.Parallel()

	recordReview, err := NewRecordReview(repositoryStub{
		recordReview: func(context.Context, identifier.UserID, domain.ReviewItemID, domain.RequestID, RecordReviewFunc) (RecordedReview, error) {
			t.Fatal("invalid input must not reach repository")
			return RecordedReview{}, nil
		},
	}, catalogStub{}, timeZoneStub("Asia/Tokyo"), fixedClock, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("NewRecordReview() error = %v", err)
	}
	valid := RecordReviewInput{
		ReviewItemID: domain.NewReviewItemID().String(),
		RequestID:    uuid.NewString(),
		Result:       domain.ResultAssisted,
		PerformedAt:  fixedNow,
	}
	for _, test := range []struct {
		name  string
		input RecordReviewInput
		want  error
	}{
		{name: "invalid item id", input: func() RecordReviewInput { input := valid; input.ReviewItemID = "x"; return input }(), want: domain.ErrInvalidID},
		{name: "nil request id", input: func() RecordReviewInput { input := valid; input.RequestID = uuid.Nil.String(); return input }(), want: domain.ErrInvalidID},
		{name: "difficulty for assisted", input: func() RecordReviewInput { input := valid; input.Difficulty = domain.DifficultyGood; return input }(), want: domain.ErrInvalidOutcome},
	} {
		if _, err := recordReview.Execute(context.Background(), identifier.NewUserID(), test.input); !errors.Is(err, test.want) {
			t.Errorf("%s: error = %v, want %v", test.name, err, test.want)
		}
	}
}

func TestPauseAndResumeReviewItem(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	pausedProblem := problem(t, "abc350_a")
	item := registerItem(t, userID, pausedProblem.ID(), fixedNow.Add(-72*time.Hour))
	stored := item
	repository := repositoryStub{
		findReviewItemByID: func(context.Context, identifier.UserID, domain.ReviewItemID) (domain.ReviewItem, error) {
			return stored, nil
		},
		updateReviewItemPausedAt: func(_ context.Context, updated domain.ReviewItem) (domain.ReviewItem, error) {
			stored = updated
			return updated, nil
		},
	}
	catalogProblems := catalogStub{problems: []catalog.Problem{pausedProblem}}
	pauseReviewItem, err := NewPauseReviewItem(repository, catalogProblems, timeZoneStub("Asia/Tokyo"), fixedClock)
	if err != nil {
		t.Fatalf("NewPauseReviewItem() error = %v", err)
	}
	resumeReviewItem, err := NewResumeReviewItem(repository, catalogProblems, timeZoneStub("Asia/Tokyo"), fixedClock)
	if err != nil {
		t.Fatalf("NewResumeReviewItem() error = %v", err)
	}

	paused, err := pauseReviewItem.Execute(context.Background(), userID, item.ID().String())
	if err != nil || !paused.Item.Paused() {
		t.Fatalf("Pause Execute() = %+v, %v", paused, err)
	}
	resumed, err := resumeReviewItem.Execute(context.Background(), userID, item.ID().String())
	if err != nil || resumed.Item.Paused() || resumed.Item.DueOn() != item.DueOn() {
		t.Fatalf("Resume Execute() = %+v, %v", resumed, err)
	}
}

func TestNewUseCaseRejectsMissingDependency(t *testing.T) {
	t.Parallel()

	if _, err := NewRegisterProblems(repositoryStub{}, catalogStub{}, timeZoneStub("UTC"), nil); err == nil {
		t.Fatal("NewRegisterProblems() accepted a nil clock")
	}
}

func problem(t *testing.T, id string) catalog.Problem {
	t.Helper()
	problemID, err := catalog.ParseProblemID(id)
	if err != nil {
		t.Fatalf("ParseProblemID() error = %v", err)
	}
	contestID, err := catalog.ParseContestID(strings.Split(id, "_")[0])
	if err != nil {
		t.Fatalf("ParseContestID() error = %v", err)
	}
	return catalog.NewProblem(problemID, contestID, "A", "Past ABCs")
}

func registerItem(t *testing.T, userID identifier.UserID, problemID catalog.ProblemID, registeredAt time.Time) domain.ReviewItem {
	t.Helper()
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	item, err := domain.RegisterReviewItem(domain.NewReviewItemID(), userID, problemID, "", registeredAt, tokyo)
	if err != nil {
		t.Fatalf("RegisterReviewItem() error = %v", err)
	}
	return item
}
