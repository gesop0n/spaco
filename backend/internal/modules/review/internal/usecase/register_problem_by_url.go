package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// RegisterProblemByURLResultは、URLから登録した復習対象である。
type RegisterProblemByURLResult struct {
	Item ReviewItemView
	// AlreadyRegisteredがtrueなら、既存の復習対象を変更せずに返している。
	AlreadyRegistered bool
}

// RegisterProblemByURLは、AtCoderの問題URLから復習対象を登録するユースケースである。
// AtCoder Problemsにまだ無い問題も、URLから読み取ったIDで登録する。
type RegisterProblemByURL struct {
	repository IRegisterProblemsRepository
	problems   IProblemEnsurer
	timeZones  ITimeZoneReader
	clock      Clock
}

func NewRegisterProblemByURL(
	repository IRegisterProblemsRepository,
	problems IProblemEnsurer,
	timeZones ITimeZoneReader,
	clock Clock,
) (*RegisterProblemByURL, error) {
	if err := requireDependencies(
		"register problem by url",
		dependency{name: "repository", value: repository},
		dependency{name: "problem ensurer", value: problems},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &RegisterProblemByURL{repository: repository, problems: problems, timeZones: timeZones, clock: clock}, nil
}

func (u *RegisterProblemByURL) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawURL string,
	note string,
) (RegisterProblemByURLResult, error) {
	if userID.IsZero() {
		return RegisterProblemByURLResult{}, errors.New("register problem by url: user id is required")
	}
	ref, err := catalog.ParseProblemURL(rawURL)
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
	}
	location, _, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
	}
	// 入力を検証してから問題情報を登録し、不正な入力でcatalogに行を作らない。
	item, err := domain.RegisterReviewItem(domain.NewReviewItemID(), userID, ref.ProblemID(), note, u.clock(), location)
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
	}
	problem, err := u.problems.EnsureProblem(ctx, ref)
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: ensure problem: %w", err)
	}
	problems := problemsByID([]catalog.Problem{problem})

	created, err := u.repository.CreateReviewItems(ctx, []domain.ReviewItem{item})
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
	}
	alreadyRegistered := len(created) == 0
	items := created
	latestLogs := map[domain.ReviewItemID]domain.ReviewLog{}
	if alreadyRegistered {
		items, err = u.repository.FindReviewItemsByProblemIDs(ctx, userID, []catalog.ProblemID{ref.ProblemID()})
		if err != nil {
			return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
		}
		if len(items) != 1 {
			return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", ErrReviewItemNotFound)
		}
		latestLogs, err = findLatestLogs(ctx, u.repository, userID, items)
		if err != nil {
			return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
		}
	}

	views, err := newReviewItemViews(items, problems, latestLogs, location)
	if err != nil {
		return RegisterProblemByURLResult{}, fmt.Errorf("register problem by url: %w", err)
	}
	return RegisterProblemByURLResult{Item: views[0], AlreadyRegistered: alreadyRegistered}, nil
}
