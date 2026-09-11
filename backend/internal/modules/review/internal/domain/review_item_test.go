package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

var tokyo = mustLoadLocation("Asia/Tokyo")

func TestRegisterReviewItemSchedulesNextDayInUserTimeZone(t *testing.T) {
	t.Parallel()

	// 日本時間では9月11日23時30分なので、最初の復習は9月12日になる。
	registeredAt := time.Date(2026, 9, 11, 14, 30, 0, 123456789, time.UTC)
	item := mustRegister(t, "  解説の考え方を再現したい  ", registeredAt)
	if item.DueOn().String() != "2026-09-12" || item.State() != scheduler.StateNew || item.Reps() != 0 {
		t.Fatalf("registered item = %+v", item)
	}
	if item.RegistrationNote() != "解説の考え方を再現したい" {
		t.Fatalf("RegistrationNote() = %q", item.RegistrationNote())
	}
	if item.RegisteredAt().Nanosecond() != 123456000 {
		t.Fatalf("RegisteredAt() = %v, want microsecond precision", item.RegisteredAt())
	}

	if _, err := RegisterReviewItem(NewReviewItemID(), identifier.NewUserID(), mustProblemID(t), strings.Repeat("あ", 501), registeredAt, tokyo); !errors.Is(err, ErrInvalidNote) {
		t.Fatalf("long note error = %v, want ErrInvalidNote", err)
	}
}

func TestRecordUpdatesScheduleAndCreatesLog(t *testing.T) {
	t.Parallel()

	registeredAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	item := mustRegister(t, "", registeredAt)
	performedAt := time.Date(2026, 9, 11, 11, 0, 0, 0, time.UTC)
	outcome := mustOutcome(t, ResultIndependent, "")
	input := ReviewInput{RequestID: mustRequestID(t), Outcome: outcome, PerformedAt: performedAt, Note: " 自力で解けた "}

	recorded, log, err := item.Record(NewReviewLogID(), input, performedAt.Add(time.Hour), tokyo, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if recorded.State() != scheduler.StateReview || recorded.Reps() != 1 || recorded.Lapses() != 0 {
		t.Fatalf("recorded item = %+v", recorded)
	}
	if lastReviewedAt, ok := recorded.LastReviewedAt(); !ok || !lastReviewedAt.Equal(performedAt) {
		t.Fatalf("LastReviewedAt() = %v, %v", lastReviewedAt, ok)
	}
	performedOn := scheduler.DayOf(performedAt, tokyo)
	if recorded.DueOn() != performedOn.AddDays(recorded.IntervalDays()) || recorded.IntervalDays() < 1 {
		t.Fatalf("due = %s, interval = %d", recorded.DueOn(), recorded.IntervalDays())
	}
	if log.StateBefore() != scheduler.StateNew || log.ElapsedDays() != 0 || log.LastIntervalDays() != 0 ||
		log.IntervalDays() != recorded.IntervalDays() || log.NextDueOn() != recorded.DueOn() ||
		log.Note() != "自力で解けた" || log.Outcome() != outcome {
		t.Fatalf("log = %+v", log)
	}
	if memory, ok := recorded.Memory(); !ok || memory != log.Memory() {
		t.Fatalf("Memory() = %+v, %v; log memory = %+v", memory, ok, log.Memory())
	}

	// 元の集約は変更しない。
	if item.State() != scheduler.StateNew || item.Reps() != 0 {
		t.Fatalf("original item changed: %+v", item)
	}
}

func TestPreviewScheduleMatchesRecordedSchedule(t *testing.T) {
	t.Parallel()

	item := mustRegister(t, "", time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))
	first := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	item, _, err := item.Record(NewReviewLogID(), ReviewInput{
		RequestID:   mustRequestID(t),
		Outcome:     mustOutcome(t, ResultIndependent, DifficultyGood),
		PerformedAt: first,
	}, first, tokyo, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("first Record() error = %v", err)
	}

	performedAt := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	previews, err := item.PreviewSchedule(performedAt, performedAt, tokyo, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("PreviewSchedule() error = %v", err)
	}
	if len(previews) != 5 {
		t.Fatalf("len(previews) = %d, want 5", len(previews))
	}
	for _, preview := range previews {
		recorded, _, err := item.Record(NewReviewLogID(), ReviewInput{
			RequestID:   mustRequestID(t),
			Outcome:     preview.Outcome,
			PerformedAt: performedAt,
		}, performedAt, tokyo, scheduler.NewScheduler())
		if err != nil {
			t.Fatalf("Record() error = %v", err)
		}
		if recorded.DueOn() != preview.NextDueOn ||
			preview.DaysAfterPerformed != preview.NextDueOn.DaysSince(scheduler.DayOf(performedAt, tokyo)) {
			t.Fatalf("preview %+v, recorded due %s", preview, recorded.DueOn())
		}
	}
	// 解説を見たACは、ACできなかった場合と同じ予定になる。
	if previews[0].NextDueOn != previews[1].NextDueOn {
		t.Fatalf("assisted preview = %s, unsolved preview = %s", previews[1].NextDueOn, previews[0].NextDueOn)
	}
	if previews[2].NextDueOn >= previews[3].NextDueOn || previews[3].NextDueOn >= previews[4].NextDueOn {
		t.Fatalf("independent previews must be ordered hard < good < easy: %+v", previews[2:])
	}
}

func TestRecordRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	registeredAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	item := mustRegister(t, "", registeredAt)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	validInput := ReviewInput{
		RequestID:   mustRequestID(t),
		Outcome:     mustOutcome(t, ResultUnsolved, ""),
		PerformedAt: now,
	}

	reviewed, _, err := item.Record(NewReviewLogID(), validInput, now, tokyo, scheduler.NewScheduler())
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	for _, test := range []struct {
		name  string
		item  ReviewItem
		input ReviewInput
		want  error
	}{
		{name: "paused", item: item.Pause(now), input: validInput, want: ErrReviewItemPaused},
		{name: "future", item: item, input: withPerformedAt(validInput, now.Add(6*time.Minute)), want: ErrInvalidPerformedAt},
		{name: "before registration day", item: item, input: withPerformedAt(validInput, registeredAt.Add(-24*time.Hour)), want: ErrInvalidPerformedAt},
		{name: "before last review day", item: reviewed, input: withPerformedAt(validInput, now.Add(-24*time.Hour)), want: ErrInvalidPerformedAt},
		{name: "long note", item: item, input: withNote(validInput, strings.Repeat("a", 1001)), want: ErrInvalidNote},
		{name: "missing request id", item: item, input: ReviewInput{Outcome: validInput.Outcome, PerformedAt: now}, want: ErrInvalidReviewLog},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := test.item.Record(NewReviewLogID(), test.input, now, tokyo, scheduler.NewScheduler()); !errors.Is(err, test.want) {
				t.Fatalf("Record() error = %v, want %v", err, test.want)
			}
		})
	}

	// 端末の時計が数分進んでいても、記録できる。
	if _, _, err := item.Record(NewReviewLogID(), withPerformedAt(validInput, now.Add(4*time.Minute)), now, tokyo, scheduler.NewScheduler()); err != nil {
		t.Fatalf("Record() within clock skew error = %v", err)
	}
}

func TestPauseAndResumeKeepSchedule(t *testing.T) {
	t.Parallel()

	item := mustRegister(t, "", time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC))
	firstPause := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	paused := item.Pause(firstPause).Pause(firstPause.Add(time.Hour))
	if pausedAt, ok := paused.PausedAt(); !ok || !pausedAt.Equal(firstPause) {
		t.Fatalf("PausedAt() = %v, %v; want first pause time", pausedAt, ok)
	}
	resumed := paused.Resume()
	if resumed.Paused() || resumed.DueOn() != item.DueOn() {
		t.Fatalf("resumed item = %+v", resumed)
	}
}

func TestOutcomeRating(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		result     Result
		difficulty Difficulty
		want       scheduler.Rating
	}{
		{result: ResultUnsolved, want: scheduler.Again},
		// 解説・ヒントを見たACは、自力で思い出せた成功として扱わない。
		{result: ResultAssisted, want: scheduler.Again},
		{result: ResultIndependent, difficulty: DifficultyHard, want: scheduler.Hard},
		{result: ResultIndependent, want: scheduler.Good},
		{result: ResultIndependent, difficulty: DifficultyEasy, want: scheduler.Easy},
	} {
		outcome := mustOutcome(t, test.result, test.difficulty)
		if outcome.Rating() != test.want {
			t.Errorf("%s/%s Rating() = %d, want %d", test.result, test.difficulty, outcome.Rating(), test.want)
		}
		restored, err := OutcomeFromRating(outcome.Result(), outcome.Rating())
		if err != nil || restored != outcome {
			t.Errorf("OutcomeFromRating() = %+v, %v; want %+v", restored, err, outcome)
		}
	}
	for _, test := range []struct {
		result     Result
		difficulty Difficulty
	}{
		{result: ResultAssisted, difficulty: DifficultyEasy},
		{result: ResultIndependent, difficulty: "normal"},
		{result: "solved"},
	} {
		if _, err := NewOutcome(test.result, test.difficulty); !errors.Is(err, ErrInvalidOutcome) {
			t.Errorf("NewOutcome(%q, %q) error = %v, want ErrInvalidOutcome", test.result, test.difficulty, err)
		}
	}
	if _, err := OutcomeFromRating(ResultAssisted, scheduler.Good); !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("OutcomeFromRating() error = %v, want ErrInvalidOutcome", err)
	}
}

func TestRehydrateReviewItemRejectsInconsistentState(t *testing.T) {
	t.Parallel()

	item := mustRegister(t, "", time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC))
	stored := StoredReviewItem{
		ID:           item.ID(),
		UserID:       item.UserID(),
		ProblemID:    item.ProblemID(),
		RegisteredAt: item.RegisteredAt(),
		State:        scheduler.StateReview,
		DueOn:        item.DueOn(),
		Reps:         1,
	}
	if _, err := RehydrateReviewItem(stored); !errors.Is(err, ErrInvalidReviewItem) {
		t.Fatalf("RehydrateReviewItem() error = %v, want ErrInvalidReviewItem", err)
	}
}

func withPerformedAt(input ReviewInput, performedAt time.Time) ReviewInput {
	input.PerformedAt = performedAt
	return input
}

func withNote(input ReviewInput, note string) ReviewInput {
	input.Note = note
	return input
}

func mustRegister(t *testing.T, note string, registeredAt time.Time) ReviewItem {
	t.Helper()
	item, err := RegisterReviewItem(NewReviewItemID(), identifier.NewUserID(), mustProblemID(t), note, registeredAt, tokyo)
	if err != nil {
		t.Fatalf("RegisterReviewItem() error = %v", err)
	}
	return item
}

func mustOutcome(t *testing.T, result Result, difficulty Difficulty) Outcome {
	t.Helper()
	outcome, err := NewOutcome(result, difficulty)
	if err != nil {
		t.Fatalf("NewOutcome() error = %v", err)
	}
	return outcome
}

func mustProblemID(t *testing.T) catalog.ProblemID {
	t.Helper()
	id, err := catalog.ParseProblemID("abc350_d")
	if err != nil {
		t.Fatalf("ParseProblemID() error = %v", err)
	}
	return id
}

func mustRequestID(t *testing.T) RequestID {
	t.Helper()
	id, err := ParseRequestID(uuid.NewString())
	if err != nil {
		t.Fatalf("ParseRequestID() error = %v", err)
	}
	return id
}

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return location
}
