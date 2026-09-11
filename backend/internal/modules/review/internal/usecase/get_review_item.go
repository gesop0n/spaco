package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// GetReviewItemは、問題IDから復習対象を返すユースケースである。
type GetReviewItem struct {
	repository IGetReviewItemRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
}

func NewGetReviewItem(
	repository IGetReviewItemRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
) (*GetReviewItem, error) {
	if err := requireDependencies(
		"get review item",
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
	); err != nil {
		return nil, err
	}
	return &GetReviewItem{repository: repository, problems: problems, timeZones: timeZones}, nil
}

func (u *GetReviewItem) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawProblemID string,
) (ReviewItemView, error) {
	if userID.IsZero() {
		return ReviewItemView{}, errors.New("get review item: user id is required")
	}
	problemID, err := catalog.ParseProblemID(strings.TrimSpace(rawProblemID))
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	items, err := u.repository.FindReviewItemsByProblemIDs(ctx, userID, []catalog.ProblemID{problemID})
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	if len(items) == 0 {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", ErrReviewItemNotFound)
	}
	location, _, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	problems, err := findProblemsForItems(ctx, u.problems, items)
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	latestLogs, err := findLatestLogs(ctx, u.repository, userID, items)
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	views, err := newReviewItemViews(items, problems, latestLogs, location)
	if err != nil {
		return ReviewItemView{}, fmt.Errorf("get review item: %w", err)
	}
	return views[0], nil
}
