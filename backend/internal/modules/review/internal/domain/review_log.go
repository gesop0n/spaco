package domain

import (
	"fmt"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ReviewLogは、1回の再挑戦の記録で、Ankiのrevlogに相当する。
// 結果に加え、FSRSの計算に使った経過日数と計算結果を残す。
type ReviewLog struct {
	id               ReviewLogID
	reviewItemID     ReviewItemID
	userID           identifier.UserID
	requestID        RequestID
	outcome          Outcome
	performedAt      time.Time
	note             string
	stateBefore      scheduler.State
	elapsedDays      int
	lastIntervalDays int
	intervalDays     int
	memory           scheduler.MemoryState
	nextDueOn        scheduler.Day
}

// StoredReviewLogは、repositoryから復元する値である。
type StoredReviewLog struct {
	ID               ReviewLogID
	ReviewItemID     ReviewItemID
	UserID           identifier.UserID
	RequestID        RequestID
	Result           Result
	Rating           scheduler.Rating
	PerformedAt      time.Time
	Note             string
	StateBefore      scheduler.State
	ElapsedDays      int
	LastIntervalDays int
	IntervalDays     int
	Memory           scheduler.MemoryState
	NextDueOn        scheduler.Day
}

func RehydrateReviewLog(stored StoredReviewLog) (ReviewLog, error) {
	if stored.ID.IsZero() || stored.ReviewItemID.IsZero() || stored.UserID.IsZero() || stored.RequestID.IsZero() {
		return ReviewLog{}, fmt.Errorf("%w: ids are required", ErrInvalidReviewLog)
	}
	outcome, err := OutcomeFromRating(stored.Result, stored.Rating)
	if err != nil {
		return ReviewLog{}, fmt.Errorf("%w: %w", ErrInvalidReviewLog, err)
	}
	if stored.StateBefore != scheduler.StateNew && stored.StateBefore != scheduler.StateReview {
		return ReviewLog{}, fmt.Errorf("%w: unknown state %q", ErrInvalidReviewLog, stored.StateBefore)
	}
	if stored.ElapsedDays < 0 || stored.LastIntervalDays < 0 || stored.IntervalDays < 1 {
		return ReviewLog{}, fmt.Errorf("%w: days are out of range", ErrInvalidReviewLog)
	}
	return ReviewLog{
		id:               stored.ID,
		reviewItemID:     stored.ReviewItemID,
		userID:           stored.UserID,
		requestID:        stored.RequestID,
		outcome:          outcome,
		performedAt:      stored.PerformedAt,
		note:             stored.Note,
		stateBefore:      stored.StateBefore,
		elapsedDays:      stored.ElapsedDays,
		lastIntervalDays: stored.LastIntervalDays,
		intervalDays:     stored.IntervalDays,
		memory:           stored.Memory,
		nextDueOn:        stored.NextDueOn,
	}, nil
}

func (log ReviewLog) ID() ReviewLogID               { return log.id }
func (log ReviewLog) ReviewItemID() ReviewItemID    { return log.reviewItemID }
func (log ReviewLog) UserID() identifier.UserID     { return log.userID }
func (log ReviewLog) RequestID() RequestID          { return log.requestID }
func (log ReviewLog) Outcome() Outcome              { return log.outcome }
func (log ReviewLog) PerformedAt() time.Time        { return log.performedAt }
func (log ReviewLog) Note() string                  { return log.note }
func (log ReviewLog) StateBefore() scheduler.State  { return log.stateBefore }
func (log ReviewLog) ElapsedDays() int              { return log.elapsedDays }
func (log ReviewLog) LastIntervalDays() int         { return log.lastIntervalDays }
func (log ReviewLog) IntervalDays() int             { return log.intervalDays }
func (log ReviewLog) Memory() scheduler.MemoryState { return log.memory }
func (log ReviewLog) NextDueOn() scheduler.Day      { return log.nextDueOn }
