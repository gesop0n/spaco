package domain

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

const (
	maxRegistrationNoteLength = 500
	maxReviewNoteLength       = 1000
	// maxClockSkewは、端末とサーバーの時計のずれとして、未来の実施日時を許容する幅である。
	maxClockSkew = 5 * time.Minute
)

// ReviewItemは、ユーザーが復習対象として登録した問題で、Ankiのカードに相当する集約である。
// 再挑戦の記録、次回予定日の更新、一時停止をこの集約の中で整合させる。
type ReviewItem struct {
	id               ReviewItemID
	userID           identifier.UserID
	problemID        catalog.ProblemID
	registrationNote string
	registeredAt     time.Time
	pausedAt         *time.Time
	state            scheduler.State
	dueOn            scheduler.Day
	intervalDays     int
	memory           *scheduler.MemoryState
	reps             int
	lapses           int
	lastReviewedAt   *time.Time
}

// RegisterReviewItemは、問題を復習対象として登録する。
// 最初の復習日は、ユーザーのタイムゾーンで登録日の翌日とする。登録は再挑戦の回数に含めない。
func RegisterReviewItem(
	id ReviewItemID,
	userID identifier.UserID,
	problemID catalog.ProblemID,
	note string,
	registeredAt time.Time,
	location *time.Location,
) (ReviewItem, error) {
	if id.IsZero() || userID.IsZero() || problemID.IsZero() || location == nil {
		return ReviewItem{}, fmt.Errorf("%w: id, user, problem, and location are required", ErrInvalidReviewItem)
	}
	normalizedNote, err := normalizeNote(note, maxRegistrationNoteLength)
	if err != nil {
		return ReviewItem{}, err
	}
	registeredAt = toStoredTime(registeredAt)
	card := scheduler.NewCard(scheduler.DayOf(registeredAt, location).AddDays(1))
	return ReviewItem{
		id:               id,
		userID:           userID,
		problemID:        problemID,
		registrationNote: normalizedNote,
		registeredAt:     registeredAt,
		state:            card.State,
		dueOn:            card.Due,
	}, nil
}

// StoredReviewItemは、repositoryから復元する値である。
type StoredReviewItem struct {
	ID               ReviewItemID
	UserID           identifier.UserID
	ProblemID        catalog.ProblemID
	RegistrationNote string
	RegisteredAt     time.Time
	PausedAt         *time.Time
	State            scheduler.State
	DueOn            scheduler.Day
	IntervalDays     int
	Memory           *scheduler.MemoryState
	Reps             int
	Lapses           int
	LastReviewedAt   *time.Time
}

func RehydrateReviewItem(stored StoredReviewItem) (ReviewItem, error) {
	if stored.ID.IsZero() || stored.UserID.IsZero() || stored.ProblemID.IsZero() {
		return ReviewItem{}, fmt.Errorf("%w: id, user, and problem are required", ErrInvalidReviewItem)
	}
	if _, err := normalizeNote(stored.RegistrationNote, maxRegistrationNoteLength); err != nil {
		return ReviewItem{}, err
	}
	card := scheduler.Card{
		State:        stored.State,
		Due:          stored.DueOn,
		IntervalDays: stored.IntervalDays,
		Memory:       stored.Memory,
		Reps:         stored.Reps,
		Lapses:       stored.Lapses,
	}
	if err := card.Validate(); err != nil {
		return ReviewItem{}, fmt.Errorf("%w: %w", ErrInvalidReviewItem, err)
	}
	if (stored.State == scheduler.StateReview) != (stored.LastReviewedAt != nil) {
		return ReviewItem{}, fmt.Errorf("%w: last reviewed time must match the state", ErrInvalidReviewItem)
	}
	return ReviewItem{
		id:               stored.ID,
		userID:           stored.UserID,
		problemID:        stored.ProblemID,
		registrationNote: stored.RegistrationNote,
		registeredAt:     stored.RegisteredAt,
		pausedAt:         stored.PausedAt,
		state:            stored.State,
		dueOn:            stored.DueOn,
		intervalDays:     stored.IntervalDays,
		memory:           stored.Memory,
		reps:             stored.Reps,
		lapses:           stored.Lapses,
		lastReviewedAt:   stored.LastReviewedAt,
	}, nil
}

func (item ReviewItem) ID() ReviewItemID                  { return item.id }
func (item ReviewItem) UserID() identifier.UserID         { return item.userID }
func (item ReviewItem) ProblemID() catalog.ProblemID      { return item.problemID }
func (item ReviewItem) RegistrationNote() string          { return item.registrationNote }
func (item ReviewItem) RegisteredAt() time.Time           { return item.registeredAt }
func (item ReviewItem) Paused() bool                      { return item.pausedAt != nil }
func (item ReviewItem) State() scheduler.State            { return item.state }
func (item ReviewItem) DueOn() scheduler.Day              { return item.dueOn }
func (item ReviewItem) IntervalDays() int                 { return item.intervalDays }
func (item ReviewItem) Lapses() int                       { return item.lapses }
func (item ReviewItem) PausedAt() (time.Time, bool)       { return valueOf(item.pausedAt) }
func (item ReviewItem) LastReviewedAt() (time.Time, bool) { return valueOf(item.lastReviewedAt) }

// Repsは、記録した再挑戦の回数を返す。登録は含めない。
func (item ReviewItem) Reps() int { return item.reps }

func (item ReviewItem) Memory() (scheduler.MemoryState, bool) {
	if item.memory == nil {
		return scheduler.MemoryState{}, false
	}
	return *item.memory, true
}

// ReviewInputは、ユーザーが入力した再挑戦の結果である。
type ReviewInput struct {
	RequestID   RequestID
	Outcome     Outcome
	PerformedAt time.Time
	Note        string
}

// Recordは、再挑戦の結果を記録し、FSRSで次回予定日を決める。
// 実施日は、実際に取り組んだ日時からユーザーのタイムゾーンで数える。
func (item ReviewItem) Record(
	logID ReviewLogID,
	input ReviewInput,
	now time.Time,
	location *time.Location,
	reviewScheduler scheduler.Scheduler,
) (ReviewItem, ReviewLog, error) {
	if logID.IsZero() || input.RequestID.IsZero() {
		return ReviewItem{}, ReviewLog{}, fmt.Errorf("%w: log and request ids are required", ErrInvalidReviewLog)
	}
	note, err := normalizeNote(input.Note, maxReviewNoteLength)
	if err != nil {
		return ReviewItem{}, ReviewLog{}, err
	}
	performedAt := toStoredTime(input.PerformedAt)
	before, after, err := item.schedule(input.Outcome, performedAt, now, location, reviewScheduler)
	if err != nil {
		return ReviewItem{}, ReviewLog{}, err
	}

	next := item
	next.state = after.State
	next.dueOn = after.Due
	next.intervalDays = after.IntervalDays
	next.memory = after.Memory
	next.reps = after.Reps
	next.lapses = after.Lapses
	if item.lastReviewedAt == nil || performedAt.After(*item.lastReviewedAt) {
		next.lastReviewedAt = &performedAt
	}

	log := ReviewLog{
		id:               logID,
		reviewItemID:     item.id,
		userID:           item.userID,
		requestID:        input.RequestID,
		outcome:          input.Outcome,
		performedAt:      performedAt,
		note:             note,
		stateBefore:      before.State,
		elapsedDays:      scheduler.ElapsedDays(before, scheduler.DayOf(performedAt, location)),
		lastIntervalDays: before.IntervalDays,
		intervalDays:     after.IntervalDays,
		memory:           *after.Memory,
		nextDueOn:        after.Due,
	}
	return next, log, nil
}

// SchedulePreviewは、ある結果で記録した場合の次回予定日である。
type SchedulePreview struct {
	Outcome            Outcome
	NextDueOn          scheduler.Day
	DaysAfterPerformed int
}

// PreviewScheduleは、performedAtに取り組んだと仮定し、結果ごとの次回予定日を返す。記録はしない。
// fuzzのseedは記録時と同じなので、同じ条件で記録すればプレビューと同じ日になる。
func (item ReviewItem) PreviewSchedule(
	performedAt time.Time,
	now time.Time,
	location *time.Location,
	reviewScheduler scheduler.Scheduler,
) ([]SchedulePreview, error) {
	performedAt = toStoredTime(performedAt)
	outcomes := Outcomes()
	previews := make([]SchedulePreview, 0, len(outcomes))
	for _, outcome := range outcomes {
		_, after, err := item.schedule(outcome, performedAt, now, location, reviewScheduler)
		if err != nil {
			return nil, err
		}
		previews = append(previews, SchedulePreview{
			Outcome:            outcome,
			NextDueOn:          after.Due,
			DaysAfterPerformed: after.Due.DaysSince(scheduler.DayOf(performedAt, location)),
		})
	}
	return previews, nil
}

// Pauseは、今日の復習に出さないよう一時停止する。すでに停止中なら停止した日時を変えない。
func (item ReviewItem) Pause(now time.Time) ReviewItem {
	if item.pausedAt != nil {
		return item
	}
	pausedAt := toStoredTime(now)
	item.pausedAt = &pausedAt
	return item
}

// Resumeは、一時停止を解除する。予定日は変えないため、期限を過ぎていれば今日の復習に戻る。
func (item ReviewItem) Resume() ReviewItem {
	item.pausedAt = nil
	return item
}

func (item ReviewItem) schedule(
	outcome Outcome,
	performedAt time.Time,
	now time.Time,
	location *time.Location,
	reviewScheduler scheduler.Scheduler,
) (scheduler.Card, scheduler.Card, error) {
	if item.pausedAt != nil {
		return scheduler.Card{}, scheduler.Card{}, ErrReviewItemPaused
	}
	if location == nil {
		return scheduler.Card{}, scheduler.Card{}, fmt.Errorf("%w: location is required", ErrInvalidReviewItem)
	}
	if performedAt.IsZero() || performedAt.After(now.Add(maxClockSkew)) {
		return scheduler.Card{}, scheduler.Card{}, fmt.Errorf("%w: must not be in the future", ErrInvalidPerformedAt)
	}
	performedOn := scheduler.DayOf(performedAt, location)
	if performedOn < scheduler.DayOf(item.registeredAt, location) {
		return scheduler.Card{}, scheduler.Card{}, fmt.Errorf("%w: must not be before registration", ErrInvalidPerformedAt)
	}
	before := item.card(location)
	if item.lastReviewedAt != nil && performedOn < before.LastReviewedOn {
		return scheduler.Card{}, scheduler.Card{}, fmt.Errorf("%w: must not be before the last review", ErrInvalidPerformedAt)
	}

	after, err := reviewScheduler.Next(before, outcome.Rating(), performedOn, scheduler.FuzzFromSeed(item.fuzzSeed()))
	if err != nil {
		return scheduler.Card{}, scheduler.Card{}, fmt.Errorf("%w: %w", ErrInvalidReviewItem, err)
	}
	return before, after, nil
}

func (item ReviewItem) card(location *time.Location) scheduler.Card {
	card := scheduler.Card{
		State:        item.state,
		Due:          item.dueOn,
		IntervalDays: item.intervalDays,
		Memory:       item.memory,
		Reps:         item.reps,
		Lapses:       item.lapses,
	}
	if item.lastReviewedAt != nil {
		card.LastReviewedOn = scheduler.DayOf(*item.lastReviewedAt, location)
	}
	return card
}

// fuzzSeedは、Ankiがカードの識別子と復習回数からseedを作るのに合わせた値である。
func (item ReviewItem) fuzzSeed() uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write(item.id.value[:])
	return hash.Sum64() + uint64(item.reps)
}

func normalizeNote(value string, maxLength int) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxLength {
		return "", fmt.Errorf("%w: must be valid text up to %d characters", ErrInvalidNote, maxLength)
	}
	return value, nil
}

// toStoredTimeは、PostgreSQLのtimestamptzと同じマイクロ秒の精度に揃える。
func toStoredTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Microsecond)
}

func valueOf(t *time.Time) (time.Time, bool) {
	if t == nil {
		return time.Time{}, false
	}
	return *t, true
}
