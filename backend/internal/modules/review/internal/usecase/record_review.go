package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// RecordReviewInputは、再挑戦の結果の入力である。
type RecordReviewInput struct {
	ReviewItemID string
	RequestID    string
	Result       domain.Result
	Difficulty   domain.Difficulty
	PerformedAt  time.Time
	Note         string
}

// RecordReviewResultは、保存した記録と、更新後の復習対象である。
type RecordReviewResult struct {
	Item ReviewItemView
	Log  domain.ReviewLog
}

// RecordReviewは、再挑戦の結果を記録し、次回予定日を更新するユースケースである。
// 同じrequest_idで再送された場合は、新しく記録せず保存済みの結果を返す。
type RecordReview struct {
	repository IRecordReviewRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
	clock      Clock
	scheduler  scheduler.Scheduler
}

func NewRecordReview(
	repository IRecordReviewRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
	reviewScheduler scheduler.Scheduler,
) (*RecordReview, error) {
	if err := requireDependencies(
		"record review",
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &RecordReview{
		repository: repository,
		problems:   problems,
		timeZones:  timeZones,
		clock:      clock,
		scheduler:  reviewScheduler,
	}, nil
}

func (u *RecordReview) Execute(
	ctx context.Context,
	userID identifier.UserID,
	input RecordReviewInput,
) (RecordReviewResult, error) {
	if userID.IsZero() {
		return RecordReviewResult{}, errors.New("record review: user id is required")
	}
	reviewItemID, err := domain.ParseReviewItemID(input.ReviewItemID)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	requestID, err := domain.ParseRequestID(input.RequestID)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	outcome, err := domain.NewOutcome(input.Result, input.Difficulty)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	location, _, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}

	now := u.clock()
	recorded, err := u.repository.RecordReview(
		ctx,
		userID,
		reviewItemID,
		requestID,
		func(item domain.ReviewItem) (domain.ReviewItem, domain.ReviewLog, error) {
			return item.Record(domain.NewReviewLogID(), domain.ReviewInput{
				RequestID:   requestID,
				Outcome:     outcome,
				PerformedAt: input.PerformedAt,
				Note:        input.Note,
			}, now, location, u.scheduler)
		},
	)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}

	items := []domain.ReviewItem{recorded.Item}
	problems, err := findProblemsForItems(ctx, u.problems, items)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	latestLogs, err := findLatestLogs(ctx, u.repository, userID, items)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	views, err := newReviewItemViews(items, problems, latestLogs, location)
	if err != nil {
		return RecordReviewResult{}, fmt.Errorf("record review: %w", err)
	}
	return RecordReviewResult{Item: views[0], Log: recorded.Log}, nil
}
