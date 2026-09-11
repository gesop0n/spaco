package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

const maxProblemsPerRegistration = 100

// RegisterProblemsResultは、今回登録した復習対象と、すでに登録済みだった復習対象である。
type RegisterProblemsResult struct {
	Registered        []ReviewItemView
	AlreadyRegistered []ReviewItemView
}

// RegisterProblemsは、選択した問題をまとめて復習対象に登録するユースケースである。
// 同じ問題はユーザーごとに1件だけ登録し、登録済みの問題は変更しない。
type RegisterProblems struct {
	repository IRegisterProblemsRepository
	problems   IProblemFinder
	timeZones  ITimeZoneReader
	clock      Clock
}

func NewRegisterProblems(
	repository IRegisterProblemsRepository,
	problems IProblemFinder,
	timeZones ITimeZoneReader,
	clock Clock,
) (*RegisterProblems, error) {
	if err := requireDependencies(
		"register problems",
		dependency{name: "repository", value: repository},
		dependency{name: "problem finder", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &RegisterProblems{repository: repository, problems: problems, timeZones: timeZones, clock: clock}, nil
}

func (u *RegisterProblems) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawProblemIDs []string,
	note string,
) (RegisterProblemsResult, error) {
	if userID.IsZero() {
		return RegisterProblemsResult{}, errors.New("register problems: user id is required")
	}
	problemIDs, err := parseProblemIDs(rawProblemIDs, maxProblemsPerRegistration)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}
	location, _, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}

	now := u.clock()
	items := make([]domain.ReviewItem, 0, len(problemIDs))
	for _, problemID := range problemIDs {
		item, err := domain.RegisterReviewItem(domain.NewReviewItemID(), userID, problemID, note, now, location)
		if err != nil {
			return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
		}
		items = append(items, item)
	}

	found, err := u.problems.FindProblems(ctx, problemIDs)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: find problems: %w", err)
	}
	problems := problemsByID(found)
	for _, problemID := range problemIDs {
		if _, ok := problems[problemID]; !ok {
			return RegisterProblemsResult{}, fmt.Errorf("register problems: %w: %s", ErrProblemNotFound, problemID)
		}
	}

	created, err := u.repository.CreateReviewItems(ctx, items)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}
	createdProblemIDs := make(map[catalog.ProblemID]struct{}, len(created))
	for _, item := range created {
		createdProblemIDs[item.ProblemID()] = struct{}{}
	}
	var existingProblemIDs []catalog.ProblemID
	for _, problemID := range problemIDs {
		if _, ok := createdProblemIDs[problemID]; !ok {
			existingProblemIDs = append(existingProblemIDs, problemID)
		}
	}
	var existing []domain.ReviewItem
	if len(existingProblemIDs) > 0 {
		existing, err = u.repository.FindReviewItemsByProblemIDs(ctx, userID, existingProblemIDs)
		if err != nil {
			return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
		}
	}
	latestLogs, err := findLatestLogs(ctx, u.repository, userID, existing)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}

	registered, err := newReviewItemViews(orderByProblemIDs(created, problemIDs), problems, nil, location)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}
	alreadyRegistered, err := newReviewItemViews(orderByProblemIDs(existing, problemIDs), problems, latestLogs, location)
	if err != nil {
		return RegisterProblemsResult{}, fmt.Errorf("register problems: %w", err)
	}
	return RegisterProblemsResult{Registered: registered, AlreadyRegistered: alreadyRegistered}, nil
}
