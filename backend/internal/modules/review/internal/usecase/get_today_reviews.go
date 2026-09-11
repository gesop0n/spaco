package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

const maxUpcomingReviewItems = 3

// TodayReviewsは、今日の復習の画面に返す内容である。
type TodayReviews struct {
	Today    scheduler.Day
	TimeZone string
	// DueItemsは、予定日が今日以前で、一時停止していない復習対象である。予定日が古い順。
	DueItems []ReviewItemView
	// CompletedCountは、ユーザーのタイムゾーンで今日に実施した再挑戦の記録数である。
	CompletedCount int
	// UpcomingItemsは、予定日が明日以降の復習対象である。予定日が近い順。
	UpcomingItems []ReviewItemView
}

// GetTodayReviewsは、今日の復習と、この先の予定を返すユースケースである。
// 予定日を過ぎても取り組んでいない復習対象は、今日の復習に残し続ける。
type GetTodayReviews struct {
	repository IGetTodayReviewsRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
	clock      Clock
}

func NewGetTodayReviews(
	repository IGetTodayReviewsRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (*GetTodayReviews, error) {
	if err := requireDependencies(
		"get today reviews",
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &GetTodayReviews{repository: repository, problems: problems, timeZones: timeZones, clock: clock}, nil
}

func (u *GetTodayReviews) Execute(ctx context.Context, userID identifier.UserID) (TodayReviews, error) {
	if userID.IsZero() {
		return TodayReviews{}, errors.New("get today reviews: user id is required")
	}
	location, timeZone, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	today := scheduler.DayOf(u.clock(), location)

	due, err := u.repository.ListDueReviewItems(ctx, userID, today)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	upcoming, err := u.repository.ListUpcomingReviewItems(ctx, userID, today, maxUpcomingReviewItems)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	completedCount, err := u.repository.CountReviewLogsPerformedBetween(
		ctx,
		userID,
		today.StartIn(location),
		today.AddDays(1).StartIn(location),
	)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}

	items := slices.Concat(due, upcoming)
	problems, err := findProblemsForItems(ctx, u.problems, items)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	latestLogs, err := findLatestLogs(ctx, u.repository, userID, items)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	dueViews, err := newReviewItemViews(due, problems, latestLogs, location)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	upcomingViews, err := newReviewItemViews(upcoming, problems, latestLogs, location)
	if err != nil {
		return TodayReviews{}, fmt.Errorf("get today reviews: %w", err)
	}
	return TodayReviews{
		Today:          today,
		TimeZone:       timeZone,
		DueItems:       dueViews,
		CompletedCount: completedCount,
		UpcomingItems:  upcomingViews,
	}, nil
}
