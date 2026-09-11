package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ReviewItemListは、復習リストの画面に返す内容である。
type ReviewItemList struct {
	Today    scheduler.Day
	TimeZone string
	// Itemsは、登録日時が新しい順の復習対象である。
	Items []ReviewItemView
	// Logsは、実施日時が古い順の再挑戦の記録である。
	Logs []domain.ReviewLog
}

// ListReviewItemsは、登録済みの復習対象と再挑戦の履歴を返すユースケースである。
type ListReviewItems struct {
	repository IListReviewItemsRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
	clock      Clock
}

func NewListReviewItems(
	repository IListReviewItemsRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (*ListReviewItems, error) {
	if err := requireDependencies(
		"list review items",
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &ListReviewItems{repository: repository, problems: problems, timeZones: timeZones, clock: clock}, nil
}

func (u *ListReviewItems) Execute(ctx context.Context, userID identifier.UserID) (ReviewItemList, error) {
	if userID.IsZero() {
		return ReviewItemList{}, errors.New("list review items: user id is required")
	}
	location, timeZone, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return ReviewItemList{}, fmt.Errorf("list review items: %w", err)
	}
	items, err := u.repository.ListReviewItems(ctx, userID)
	if err != nil {
		return ReviewItemList{}, fmt.Errorf("list review items: %w", err)
	}
	logs, err := u.repository.ListReviewLogs(ctx, userID)
	if err != nil {
		return ReviewItemList{}, fmt.Errorf("list review items: %w", err)
	}
	problems, err := findProblemsForItems(ctx, u.problems, items)
	if err != nil {
		return ReviewItemList{}, fmt.Errorf("list review items: %w", err)
	}
	views, err := newReviewItemViews(items, problems, latestLogsByItem(logs), location)
	if err != nil {
		return ReviewItemList{}, fmt.Errorf("list review items: %w", err)
	}
	return ReviewItemList{
		Today:    scheduler.DayOf(u.clock(), location),
		TimeZone: timeZone,
		Items:    views,
		Logs:     logs,
	}, nil
}
