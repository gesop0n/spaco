package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
)

// SearchContestsは、コンテストIDまたはコンテスト名でコンテストを探すユースケースである。
type SearchContests struct {
	repository ISearchContestsRepository
}

func NewSearchContests(repository ISearchContestsRepository) (*SearchContests, error) {
	if repository == nil {
		return nil, errors.New("create search contests use case: repository is required")
	}
	return &SearchContests{repository: repository}, nil
}

func (u *SearchContests) Execute(ctx context.Context, rawQuery string) ([]domain.Contest, error) {
	query, err := domain.NewContestQuery(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("search contests: %w", err)
	}
	contests, err := u.repository.SearchContests(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search contests: %w", err)
	}
	return contests, nil
}
