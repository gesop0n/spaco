package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// PauseReviewItemは、復習対象を一時停止するユースケースである。履歴と予定日は変えない。
type PauseReviewItem struct {
	changer pauseChanger
}

func NewPauseReviewItem(
	repository IChangeReviewItemPauseRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (*PauseReviewItem, error) {
	changer, err := newPauseChanger("pause review item", repository, problems, timeZones, clock)
	if err != nil {
		return nil, err
	}
	return &PauseReviewItem{changer: changer}, nil
}

func (u *PauseReviewItem) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawReviewItemID string,
) (ReviewItemView, error) {
	view, err := u.changer.change(ctx, userID, rawReviewItemID, func(item domain.ReviewItem, now time.Time) domain.ReviewItem {
		return item.Pause(now)
	})
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("pause review item: %w", err)
	}
	return view, nil
}

// ResumeReviewItemは、一時停止した復習対象を再開するユースケースである。
// 予定日を過ぎていれば、そのまま今日の復習に戻る。
type ResumeReviewItem struct {
	changer pauseChanger
}

func NewResumeReviewItem(
	repository IChangeReviewItemPauseRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (*ResumeReviewItem, error) {
	changer, err := newPauseChanger("resume review item", repository, problems, timeZones, clock)
	if err != nil {
		return nil, err
	}
	return &ResumeReviewItem{changer: changer}, nil
}

func (u *ResumeReviewItem) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawReviewItemID string,
) (ReviewItemView, error) {
	view, err := u.changer.change(ctx, userID, rawReviewItemID, func(item domain.ReviewItem, _ time.Time) domain.ReviewItem {
		return item.Resume()
	})
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("resume review item: %w", err)
	}
	return view, nil
}

type pauseChanger struct {
	repository IChangeReviewItemPauseRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
	clock      Clock
}

func newPauseChanger(
	useCase string,
	repository IChangeReviewItemPauseRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (pauseChanger, error) {
	if err := requireDependencies(
		useCase,
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return pauseChanger{}, err
	}
	return pauseChanger{repository: repository, problems: problems, timeZones: timeZones, clock: clock}, nil
}

func (c pauseChanger) change(
	ctx context.Context,
	userID identifier.UserID,
	rawReviewItemID string,
	change func(domain.ReviewItem, time.Time) domain.ReviewItem,
) (ReviewItemView, error) {
	if userID.IsZero() {
		return ReviewItemView{}, errors.New("user id is required")
	}
	reviewItemID, err := domain.ParseReviewItemID(rawReviewItemID)
	if err != nil {
		return ReviewItemView{}, err
	}
	item, err := c.repository.FindReviewItemByID(ctx, userID, reviewItemID)
	if err != nil {
		return ReviewItemView{}, err
	}
	updated, err := c.repository.UpdateReviewItemPausedAt(ctx, change(item, c.clock()))
	if err != nil {
		return ReviewItemView{}, err
	}

	location, _, err := loadLocation(ctx, c.timeZones, userID)
	if err != nil {
		return ReviewItemView{}, err
	}
	items := []domain.ReviewItem{updated}
	problems, err := findProblemsForItems(ctx, c.problems, items)
	if err != nil {
		return ReviewItemView{}, err
	}
	latestLogs, err := findLatestLogs(ctx, c.repository, userID, items)
	if err != nil {
		return ReviewItemView{}, err
	}
	views, err := newReviewItemViews(items, problems, latestLogs, location)
	if err != nil {
		return ReviewItemView{}, err
	}
	return views[0], nil
}
